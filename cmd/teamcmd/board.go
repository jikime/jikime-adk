package teamcmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"jikime-adk/internal/team"
)

// boardSanitize replaces chars that are invalid in tmux session/window names.
func boardSanitize(s string) string {
	return strings.NewReplacer(" ", "-", "/", "-", ":", "-").Replace(s)
}

func newBoardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "board",
		Short: "View team activity board",
	}
	cmd.AddCommand(newBoardShowCmd())
	cmd.AddCommand(newBoardLiveCmd())
	cmd.AddCommand(newBoardOverviewCmd())
	cmd.AddCommand(newBoardAttachCmd())
	cmd.AddCommand(newBoardServeCmd())
	return cmd
}

func newBoardShowCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "show <team-name>",
		Short: "Show current board snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			td := teamDir(name)

			store, err := team.NewStore(filepath.Join(td, "tasks"))
			if err != nil {
				return err
			}
			reg, err := team.NewRegistry(filepath.Join(td, "registry"))
			if err != nil {
				return err
			}

			tasks, _ := store.List("", "")
			agents, _ := reg.List()

			if jsonOut {
				out := map[string]interface{}{
					"team":   name,
					"agents": agents,
					"tasks":  tasks,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Printf("╔══════════════════════════════════════════════════════╗\n")
			fmt.Printf("║  Team Board: %-39s║\n", name)
			fmt.Printf("╚══════════════════════════════════════════════════════╝\n\n")

			// Agents section
			fmt.Printf("Agents (%d):\n", len(agents))
			for _, a := range agents {
				alive := "❌"
				if ok, _ := reg.IsAlive(a.ID); ok {
					alive = "✅"
				}
				task := "-"
				if a.CurrentTaskID != "" {
					task = a.CurrentTaskID[:8]
				}
				fmt.Printf("  %s %-14s [%-8s]  role:%-12s task:%s\n",
					alive, a.ID, a.Status, a.Role, task)
			}

			// Tasks section
			counts := map[team.TaskStatus]int{}
			for _, t := range tasks {
				counts[t.Status]++
			}
			fmt.Printf("\nTasks (%d total):\n", len(tasks))
			fmt.Printf("  pending:%-4d  in_progress:%-4d  done:%-4d  failed:%-4d  blocked:%-4d\n",
				counts[team.TaskStatusPending],
				counts[team.TaskStatusInProgress],
				counts[team.TaskStatusDone],
				counts[team.TaskStatusFailed],
				counts[team.TaskStatusBlocked],
			)

			fmt.Printf("\nRecent tasks:\n")
			shown := 0
			for _, t := range tasks {
				if shown >= 10 {
					break
				}
				id := t.ID
				if len(id) > 8 {
					id = id[:8]
				}
				agent := t.AgentID
				if agent == "" {
					agent = "-"
				}
				fmt.Printf("  %s  [%-11s]  %-30s  agent:%s\n", id, t.Status, t.Title, agent)
				shown++
			}
			if len(tasks) > 10 {
				fmt.Printf("  ... and %d more\n", len(tasks)-10)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newBoardLiveCmd() *cobra.Command {
	var interval int
	cmd := &cobra.Command{
		Use:   "live <team-name>",
		Short: "Live-refresh board every N seconds (Ctrl+C to stop)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			td := teamDir(name)

			store, err := team.NewStore(filepath.Join(td, "tasks"))
			if err != nil {
				return err
			}
			reg, err := team.NewRegistry(filepath.Join(td, "registry"))
			if err != nil {
				return err
			}

			tick := time.NewTicker(time.Duration(interval) * time.Second)
			defer tick.Stop()

			printBoard := func() {
				// Clear screen (ANSI)
				fmt.Print("\033[2J\033[H")
				fmt.Printf("Team Board: %s  [%s]  (Ctrl+C to stop)\n\n",
					name, time.Now().Format("15:04:05"))

				tasks, _ := store.List("", "")
				agents, _ := reg.List()

				counts := map[team.TaskStatus]int{}
				for _, t := range tasks {
					counts[t.Status]++
				}
				fmt.Printf("Agents: %d  |  Tasks: pending:%d  wip:%d  done:%d  failed:%d\n\n",
					len(agents),
					counts[team.TaskStatusPending],
					counts[team.TaskStatusInProgress],
					counts[team.TaskStatusDone],
					counts[team.TaskStatusFailed],
				)

				for _, a := range agents {
					alive := "❌"
					if ok, _ := reg.IsAlive(a.ID); ok {
						alive = "✅"
					}
					fmt.Printf("  %s %-14s [%-8s]  role:%s\n", alive, a.ID, a.Status, a.Role)
				}
			}

			printBoard()
			for range tick.C {
				printBoard()
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&interval, "interval", "i", 3, "Refresh interval in seconds")
	return cmd
}

// newBoardAttachCmd creates a dashboard tmux session that links all agent windows
// for a team, giving a unified view without disrupting running agents.
func newBoardAttachCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "attach <team-name>",
		Short: "Open a tmux dashboard linking all agent windows for the team",
		Long: `Creates a board tmux session and links each agent's window into it.
Use Ctrl-b n/p to navigate between agents. The board session is read-only;
agents continue running in their own sessions unaffected.

Example:
  jikime team board attach my-team`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			prefix := "jikime-" + boardSanitize(name) + "-"
			boardSession := "jikime-" + boardSanitize(name) + "-board"

			// List all tmux sessions to find agents for this team.
			out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").Output()
			if err != nil {
				return fmt.Errorf("tmux list-sessions: %w (is tmux running?)", err)
			}

			var agentSessions []string
			for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || line == boardSession {
					continue
				}
				if strings.HasPrefix(line, prefix) {
					agentSessions = append(agentSessions, line)
				}
			}

			if len(agentSessions) == 0 {
				return fmt.Errorf("no active tmux sessions found for team %q (prefix: %s)", name, prefix)
			}

			// Kill stale board session if it exists.
			_ = exec.Command("tmux", "kill-session", "-t", boardSession).Run()

			// Create a fresh board session (starts with a temporary shell window).
			if out, err := exec.Command("tmux", "new-session", "-d", "-s", boardSession, "-n", "_board_").CombinedOutput(); err != nil {
				return fmt.Errorf("create board session: %w\n%s", err, out)
			}

			// Link each agent's window into the board session.
			// -t must point to an existing session only (no window name);
			// tmux appends the linked window automatically at the next index.
			linked := 0
			for _, sess := range agentSessions {
				agentName := strings.TrimPrefix(sess, prefix)
				srcWin := sess + ":" + boardSanitize(agentName)

				if err := exec.Command("tmux", "link-window", "-s", srcWin, "-t", boardSession).Run(); err != nil {
					// Fallback: try window index 0 if named window not found.
					srcWin0 := sess + ":0"
					if err2 := exec.Command("tmux", "link-window", "-s", srcWin0, "-t", boardSession).Run(); err2 != nil {
						fmt.Printf("  ⚠️  could not link %s: %v\n", sess, err2)
						continue
					}
				}
				linked++
			}

			// Remove the initial placeholder window.
			_ = exec.Command("tmux", "kill-window", "-t", boardSession+":_board_").Run()

			if linked == 0 {
				_ = exec.Command("tmux", "kill-session", "-t", boardSession).Run()
				return fmt.Errorf("failed to link any agent windows into board session")
			}

			// Move to first window.
			_ = exec.Command("tmux", "select-window", "-t", boardSession+":0").Run()

			fmt.Printf("📺 Board session: %s\n", boardSession)
			fmt.Printf("   %d agent windows linked (Ctrl-b n/p to switch, Ctrl-b d to detach)\n\n", linked)

			// Replace current process with tmux so it gets full TTY control.
			tmuxBin, err := exec.LookPath("tmux")
			if err != nil {
				return fmt.Errorf("tmux not found: %w", err)
			}
			if os.Getenv("TMUX") != "" {
				// Already inside tmux — switch the current client to the board session.
				return syscall.Exec(tmuxBin, []string{"tmux", "switch-client", "-t", boardSession}, os.Environ())
			}
			return syscall.Exec(tmuxBin, []string{"tmux", "attach-session", "-t", boardSession}, os.Environ())
		},
	}
}

// newBoardServeCmd starts a local HTTP dashboard server.
// Endpoints:
//
//	GET /              → HTML dashboard (React SPA)
//	GET /api/overview  → JSON list of all teams
//	GET /api/team/:name → JSON snapshot of a single team
//	GET /api/events/:name → SSE stream (pushes team snapshot on interval)
func newBoardServeCmd() *cobra.Command {
	var (
		port     int
		host     string
		interval float64
	)
	cmd := &cobra.Command{
		Use:   "serve [team-name]",
		Short: "Start a web dashboard server (http://localhost:8080)",
		Long: `Start a local HTTP server with a real-time web UI for monitoring team activity.
Open http://localhost:8080 in your browser after starting.

Example:
  jikime team board serve               # show all teams
  jikime team board serve my-team       # open to my-team
  jikime team board serve --port 9090   # custom port`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			defaultTeam := ""
			if len(args) > 0 {
				defaultTeam = args[0]
			}

			mux := http.NewServeMux()

			// --- GET / → HTML SPA ---
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/" && r.URL.Path != "/index.html" {
					http.NotFound(w, r)
					return
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write([]byte(boardHTML))
			})

			// --- GET /api/overview → [{name, description}, ...] ---
			mux.HandleFunc("/api/overview", func(w http.ResponseWriter, r *http.Request) {
				teamsDir := filepath.Join(dataDir(), "teams")
				entries, _ := os.ReadDir(teamsDir)
				type teamMeta struct {
					Name        string `json:"name"`
					Description string `json:"description,omitempty"`
				}
				var list []teamMeta
				for _, e := range entries {
					if !e.IsDir() {
						continue
					}
					cfg := struct {
						Description string `json:"description"`
					}{}
					data, _ := os.ReadFile(filepath.Join(teamsDir, e.Name(), "config.json"))
					_ = json.Unmarshal(data, &cfg)
					list = append(list, teamMeta{Name: e.Name(), Description: cfg.Description})
				}
				if list == nil {
					list = []teamMeta{}
				}
				writeJSON2(w, list)
			})

			// --- GET /api/team/:name → team snapshot ---
			mux.HandleFunc("/api/team/", func(w http.ResponseWriter, r *http.Request) {
				name := strings.TrimPrefix(r.URL.Path, "/api/team/")
				if name == "" {
					http.Error(w, `{"error":"team name required"}`, http.StatusBadRequest)
					return
				}
				data, err := collectBoardData(name)
				if err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)
					_, _ = fmt.Fprintf(w, `{"error":%q}`, err.Error())
					return
				}
				writeJSON2(w, data)
			})

			// --- GET /api/events/:name → SSE stream ---
			mux.HandleFunc("/api/events/", func(w http.ResponseWriter, r *http.Request) {
				name := strings.TrimPrefix(r.URL.Path, "/api/events/")
				if name == "" {
					http.Error(w, "team name required", http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")
				w.Header().Set("Access-Control-Allow-Origin", "*")

				flusher, ok := w.(http.Flusher)
				if !ok {
					http.Error(w, "streaming unsupported", http.StatusInternalServerError)
					return
				}

				tick := time.NewTicker(time.Duration(interval*1000) * time.Millisecond)
				defer tick.Stop()
				ctx := r.Context()

				push := func() {
					data, err := collectBoardData(name)
					var payload []byte
					if err != nil {
						payload, _ = json.Marshal(map[string]string{"error": err.Error()})
					} else {
						payload, _ = json.Marshal(data)
					}
					_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
					flusher.Flush()
				}

				push() // immediate first push
				for {
					select {
					case <-ctx.Done():
						return
					case <-tick.C:
						push()
					}
				}
			})

			addr := fmt.Sprintf("%s:%d", host, port)
			fmt.Printf("🌐 jikime dashboard: http://%s\n", addr)
			if defaultTeam != "" {
				fmt.Printf("   Default team: %s\n", defaultTeam)
			}
			fmt.Println("   Press Ctrl+C to stop.")
			return http.ListenAndServe(addr, mux)
		},
	}
	cmd.Flags().IntVarP(&port, "port", "p", 8080, "HTTP server port")
	cmd.Flags().StringVar(&host, "host", "127.0.0.1", "Bind address")
	cmd.Flags().Float64VarP(&interval, "interval", "i", 2.0, "SSE push interval in seconds")
	return cmd
}

// writeJSON2 serialises v as indented JSON with CORS headers.
func writeJSON2(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// boardTaskItem is the wire format for a task card in the web UI.
type boardTaskItem struct {
	ID        string   `json:"id"`
	Subject   string   `json:"subject"`
	Owner     string   `json:"owner,omitempty"`
	BlockedBy []string `json:"blockedBy,omitempty"`
}

// boardMember is the wire format for an agent member card.
type boardMember struct {
	Name       string `json:"name"`
	User       string `json:"user,omitempty"`
	AgentType  string `json:"agentType"`
	InboxCount int    `json:"inboxCount"`
}

// boardMessage is the wire format for a message in the event log panel.
type boardMessage struct {
	From      string `json:"from"`
	To        string `json:"to,omitempty"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Content   string `json:"content"`
}

// collectBoardData builds the full board payload for a team.
func collectBoardData(name string) (map[string]interface{}, error) {
	td := teamDir(name)
	if _, err := os.Stat(td); os.IsNotExist(err) {
		return nil, fmt.Errorf("team %q not found", name)
	}

	// Config
	cfg := struct {
		LeaderID    string `json:"leader_id"`
		Description string `json:"description"`
	}{}
	data, _ := os.ReadFile(filepath.Join(td, "config.json"))
	_ = json.Unmarshal(data, &cfg)

	// Agents
	reg, _ := team.NewRegistry(filepath.Join(td, "registry"))
	agents, _ := reg.List()

	leaderName := cfg.LeaderID
	members := make([]boardMember, 0, len(agents))
	for _, a := range agents {
		if leaderName == "" && a.Role == "leader" {
			leaderName = a.ID
		}
		ib, _ := team.NewInbox(team.InboxDir(td, a.ID))
		inboxCount := 0
		if ib != nil {
			inboxCount, _ = ib.Count()
		}
		members = append(members, boardMember{
			Name:       a.ID,
			AgentType:  string(a.Role),
			InboxCount: inboxCount,
		})
	}

	// Tasks
	store, _ := team.NewStore(filepath.Join(td, "tasks"))
	allTasks, _ := store.List("", "")

	taskGroups := map[string][]boardTaskItem{
		"pending":     {},
		"in_progress": {},
		"done":        {},
		"failed":      {},
		"blocked":     {},
	}
	summary := map[string]int{
		"pending": 0, "in_progress": 0, "done": 0, "failed": 0, "blocked": 0,
	}
	for _, t := range allTasks {
		status := string(t.Status)
		item := boardTaskItem{
			ID:      t.ID,
			Subject: t.Title,
			Owner:   t.AgentID,
		}
		if _, ok := taskGroups[status]; ok {
			taskGroups[status] = append(taskGroups[status], item)
			summary[status]++
		}
	}

	// Messages (event log, last 50)
	ti := team.NewTeamInbox(td)
	logMsgs, _ := ti.EventLog(50, "")
	messages := make([]boardMessage, 0, len(logMsgs))
	for _, m := range logMsgs {
		messages = append(messages, boardMessage{
			From:      m.From,
			To:        m.To,
			Type:      string(m.Kind),
			Timestamp: m.SentAt.Format(time.RFC3339),
			Content:   m.Body,
		})
	}

	return map[string]interface{}{
		"team": map[string]string{
			"name":        name,
			"leaderName":  leaderName,
			"description": cfg.Description,
		},
		"members":     members,
		"tasks":       taskGroups,
		"taskSummary": summary,
		"messages":    messages,
	}, nil
}

// boardHTML is the single-file React SPA served at GET /.
// Adapted from ClawTeam's dashboard, updated for jikime task statuses.
const boardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>jikime</title>
<style>
  @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
  :root {
    --bg:#000;--surface:#0a0a0a;--surface2:#111;--surface3:#1a1a1a;
    --border:rgba(255,255,255,.08);--border-hover:rgba(255,255,255,.15);
    --text:#f5f5f7;--text-secondary:#86868b;--text-tertiary:#6e6e73;
    --green:#30d158;--orange:#ff9f0a;--red:#ff453a;--blue:#0a84ff;--purple:#bf5af2;
    --radius:12px;--radius-sm:8px;
  }
  *{box-sizing:border-box;margin:0;padding:0}
  body{font-family:'Inter',-apple-system,BlinkMacSystemFont,sans-serif;background:var(--bg);color:var(--text);min-height:100vh;-webkit-font-smoothing:antialiased}
</style>
<script crossorigin src="https://unpkg.com/react@18/umd/react.production.min.js"></script>
<script crossorigin src="https://unpkg.com/react-dom@18/umd/react-dom.production.min.js"></script>
<script src="https://unpkg.com/@babel/standalone/babel.min.js"></script>
</head>
<body>
<div id="root"></div>
<script type="text/babel">
const {useState,useEffect,useRef,useMemo}=React;
const S={
  nav:{position:'sticky',top:0,zIndex:100,background:'rgba(0,0,0,.72)',backdropFilter:'saturate(180%) blur(20px)',WebkitBackdropFilter:'saturate(180%) blur(20px)',borderBottom:'1px solid var(--border)',padding:'0 32px',height:52,display:'flex',alignItems:'center',gap:16},
  navTitle:{fontSize:15,fontWeight:600,letterSpacing:'-.01em',color:'var(--text)'},
  select:{background:'var(--surface2)',color:'var(--text)',border:'1px solid var(--border)',borderRadius:8,padding:'6px 12px',fontSize:13,outline:'none',cursor:'pointer'},
  statusDot:c=>({width:8,height:8,borderRadius:'50%',marginLeft:'auto',background:c?'var(--green)':'var(--red)',boxShadow:c?'0 0 8px var(--green)':'none'}),
  statusText:{fontSize:12,color:'var(--text-tertiary)',marginLeft:6},
  container:{maxWidth:1200,margin:'0 auto',padding:'24px 32px'},
  summaryGrid:{display:'grid',gridTemplateColumns:'repeat(5,1fr)',gap:12,marginBottom:24},
  summaryCard:{background:'var(--surface)',border:'1px solid var(--border)',borderRadius:'var(--radius)',padding:'20px 16px',textAlign:'center'},
  summaryNum:c=>({fontSize:32,fontWeight:700,letterSpacing:'-.02em',color:c,lineHeight:1}),
  summaryLabel:{fontSize:11,fontWeight:500,color:'var(--text-tertiary)',textTransform:'uppercase',letterSpacing:'.06em',marginTop:6},
  section:{background:'var(--surface)',border:'1px solid var(--border)',borderRadius:'var(--radius)',marginBottom:20,overflow:'hidden'},
  sectionHeader:{padding:'14px 20px',borderBottom:'1px solid var(--border)',display:'flex',alignItems:'center',justifyContent:'space-between'},
  sectionTitle:{fontSize:13,fontWeight:600,color:'var(--text)',letterSpacing:'-.01em'},
  sectionBadge:{fontSize:11,fontWeight:500,color:'var(--text-tertiary)',background:'var(--surface3)',borderRadius:10,padding:'2px 8px'},
  membersGrid:{display:'grid',gridTemplateColumns:'repeat(auto-fill,minmax(200px,1fr))',gap:1,background:'var(--border)'},
  memberCard:{background:'var(--surface)',padding:'14px 20px'},
  memberName:{fontSize:14,fontWeight:600,color:'var(--text)'},
  memberType:{fontSize:12,color:'var(--text-tertiary)',marginTop:2},
  inboxBadge:n=>({display:'inline-block',marginTop:6,fontSize:11,fontWeight:500,padding:'2px 8px',borderRadius:10,background:n>0?'rgba(255,69,58,.15)':'var(--surface3)',color:n>0?'var(--red)':'var(--text-tertiary)'}),
  msgList:{maxHeight:360,overflowY:'auto'},
  msgItem:{padding:'12px 20px',borderBottom:'1px solid var(--border)'},
  msgMeta:{display:'flex',alignItems:'center',gap:8,fontSize:12,marginBottom:4},
  msgTag:{fontSize:10,fontWeight:600,textTransform:'uppercase',letterSpacing:'.04em',padding:'2px 6px',borderRadius:4,background:'var(--surface3)',color:'var(--text-tertiary)'},
  msgFrom:{color:'var(--text)',fontWeight:500},
  msgTo:{color:'var(--text-tertiary)'},
  msgTime:{color:'var(--text-tertiary)',marginLeft:'auto',fontSize:11},
  msgContent:{fontSize:13,color:'var(--text-secondary)',lineHeight:1.5,whiteSpace:'pre-wrap',wordBreak:'break-word'},
  msgEmpty:{padding:32,textAlign:'center',fontSize:13,color:'var(--text-tertiary)'},
  kanban:{display:'grid',gridTemplateColumns:'repeat(5,1fr)',gap:12},
  kanbanCol:{background:'var(--surface)',border:'1px solid var(--border)',borderRadius:'var(--radius)',minHeight:200,overflow:'hidden'},
  kanbanHeader:c=>({padding:'12px 16px',borderBottom:'2px solid '+c,fontSize:11,fontWeight:600,textTransform:'uppercase',letterSpacing:'.06em',color:c,display:'flex',justifyContent:'space-between'}),
  kanbanBody:{padding:8},
  taskCard:{background:'var(--surface2)',border:'1px solid var(--border)',borderRadius:'var(--radius-sm)',padding:'10px 12px',marginBottom:6,cursor:'default'},
  taskId:{fontSize:11,fontFamily:'SF Mono,Menlo,monospace',color:'var(--text-tertiary)'},
  taskSubject:{fontSize:13,fontWeight:500,color:'var(--text)',marginTop:2,lineHeight:1.4},
  taskOwner:{fontSize:11,color:'var(--text-tertiary)',marginTop:4},
  emptyCol:{padding:24,textAlign:'center',fontSize:12,color:'var(--text-tertiary)'},
  teamInfo:{display:'flex',alignItems:'baseline',gap:16,marginBottom:20},
  teamName:{fontSize:28,fontWeight:700,letterSpacing:'-.03em',color:'var(--text)'},
  teamMeta:{fontSize:13,color:'var(--text-tertiary)'},
};

const COLS=[
  {key:'pending',label:'Pending',color:'var(--orange)'},
  {key:'in_progress',label:'In Progress',color:'var(--blue)'},
  {key:'done',label:'Done',color:'var(--green)'},
  {key:'failed',label:'Failed',color:'var(--red)'},
  {key:'blocked',label:'Blocked',color:'var(--purple)'},
];

function SummaryCards({summary}){
  return(<div style={S.summaryGrid}>{COLS.map(c=>(
    <div key={c.key} style={S.summaryCard}>
      <div style={S.summaryNum(c.color)}>{summary[c.key]||0}</div>
      <div style={S.summaryLabel}>{c.label}</div>
    </div>
  ))}</div>);
}

function Members({members}){
  return(<div style={S.section}>
    <div style={S.sectionHeader}><span style={S.sectionTitle}>Members</span><span style={S.sectionBadge}>{members.length}</span></div>
    <div style={S.membersGrid}>{members.map(m=>(
      <div key={m.name} style={S.memberCard}>
        <div style={S.memberName}>{m.name}</div>
        <div style={S.memberType}>{m.agentType}</div>
        <span style={S.inboxBadge(m.inboxCount)}>{m.inboxCount>0?m.inboxCount+' msg':'inbox empty'}</span>
      </div>
    ))}</div>
  </div>);
}

function Messages({messages}){
  const sorted=useMemo(()=>[...messages].reverse(),[messages]);
  return(<div style={S.section}>
    <div style={S.sectionHeader}><span style={S.sectionTitle}>Messages</span><span style={S.sectionBadge}>{messages.length}</span></div>
    <div style={S.msgList}>{sorted.length===0?(<div style={S.msgEmpty}>No messages</div>):sorted.map((m,i)=>{
      const ts=(m.timestamp||'').slice(11,19);const dt=(m.timestamp||'').slice(5,10);
      return(<div key={i} style={S.msgItem}>
        <div style={S.msgMeta}><span style={S.msgTag}>{m.type||'msg'}</span><span style={S.msgFrom}>{m.from}</span><span style={S.msgTo}>→ {m.to||'all'}</span><span style={S.msgTime}>{dt} {ts}</span></div>
        <div style={S.msgContent}>{m.content||''}</div>
      </div>);
    })}</div>
  </div>);
}

function Kanban({tasks}){
  return(<div style={S.kanban}>{COLS.map(col=>{
    const items=tasks[col.key]||[];
    return(<div key={col.key} style={S.kanbanCol}>
      <div style={S.kanbanHeader(col.color)}><span>{col.label}</span><span>{items.length}</span></div>
      <div style={S.kanbanBody}>{items.length===0?(<div style={S.emptyCol}>&mdash;</div>):items.map(t=>(
        <div key={t.id} style={S.taskCard}>
          <div style={S.taskId}>#{(t.id||'').slice(0,8)}</div>
          <div style={S.taskSubject}>{t.subject||''}</div>
          <div style={S.taskOwner}>{t.owner||'-'}</div>
        </div>
      ))}</div>
    </div>);
  })}</div>);
}

function Dashboard({data}){
  const{team,members=[],tasks={},taskSummary={},messages=[]}=data;
  return(<div style={S.container}>
    <div style={S.teamInfo}>
      <span style={S.teamName}>{team.name}</span>
      <span style={S.teamMeta}>led by {team.leaderName||'?'} &middot; {members.length} member{members.length!==1?'s':''}{team.description?' — '+team.description:''}</span>
    </div>
    <SummaryCards summary={taskSummary}/>
    <Members members={members}/>
    <Messages messages={messages}/>
    <Kanban tasks={tasks}/>
  </div>);
}

function App(){
  const[teams,setTeams]=useState([]);
  const[current,setCurrent]=useState('');
  const[data,setData]=useState(null);
  const[connected,setConnected]=useState(false);
  const evtRef=useRef(null);

  useEffect(()=>{
    fetch('/api/overview').then(r=>r.json()).then(list=>{
      setTeams(list);
      if(list.length===1)setCurrent(list[0].name);
    }).catch(()=>{});
  },[]);

  useEffect(()=>{
    if(evtRef.current){evtRef.current.close();evtRef.current=null;}
    if(!current){setData(null);setConnected(false);return;}
    const src=new EventSource('/api/events/'+encodeURIComponent(current));
    src.onmessage=e=>{const d=JSON.parse(e.data);if(!d.error){setData(d);setConnected(true);}};
    src.onerror=()=>setConnected(false);
    evtRef.current=src;
    return()=>src.close();
  },[current]);

  return(<>
    <nav style={S.nav}>
      <span style={S.navTitle}>jikime</span>
      <select style={S.select} value={current} onChange={e=>setCurrent(e.target.value)}>
        <option value="">Select Team</option>
        {teams.map(t=>(<option key={t.name} value={t.name}>{t.name}{t.description?' — '+t.description:''}</option>))}
      </select>
      <div style={{display:'flex',alignItems:'center',marginLeft:'auto'}}>
        <div style={S.statusDot(connected)}/><span style={S.statusText}>{connected?'Live':'Disconnected'}</span>
      </div>
    </nav>
    {data?<Dashboard data={data}/>:(
      <div style={{...S.container,textAlign:'center',paddingTop:120}}>
        <div style={{fontSize:48,fontWeight:700,letterSpacing:'-.04em',color:'var(--text)'}}>jikime</div>
        <div style={{fontSize:15,color:'var(--text-tertiary)',marginTop:8}}>Select a team to get started</div>
      </div>
    )}
  </>);
}
ReactDOM.createRoot(document.getElementById('root')).render(<App/>);
</script>
</body>
</html>`

func newBoardOverviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "overview",
		Short: "Show overview of all teams",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := os.UserHomeDir()
			teamsDir := filepath.Join(dataDir(), "teams")
			entries, err := os.ReadDir(teamsDir)
			if os.IsNotExist(err) {
				fmt.Printf("No teams found in %s\n", teamsDir)
				_ = home
				return nil
			}
			if err != nil {
				return err
			}

			fmt.Printf("%-20s  %-8s  %-6s  %s\n", "TEAM", "AGENTS", "TASKS", "TEMPLATE")
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				teamName := e.Name()
				td := filepath.Join(teamsDir, teamName)

				// Load config
				cfg := struct {
					Template string `json:"template"`
				}{}
				data, _ := os.ReadFile(filepath.Join(td, "config.json"))
				_ = json.Unmarshal(data, &cfg)

				// Count agents
				regDir := filepath.Join(td, "registry")
				agentFiles, _ := os.ReadDir(regDir)
				agentCount := len(agentFiles)

				// Count tasks
				taskFiles, _ := os.ReadDir(filepath.Join(td, "tasks"))
				taskCount := len(taskFiles)

				fmt.Printf("%-20s  %-8d  %-6d  %s\n",
					teamName, agentCount, taskCount, cfg.Template)
			}
			return nil
		},
	}
}
