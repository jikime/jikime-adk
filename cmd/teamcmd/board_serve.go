package teamcmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

//go:embed board_spa.html
var boardHTML string

// newBoardServeCmd starts a local HTTP dashboard server.
// Endpoints:
//
//	GET /              -> HTML dashboard (React SPA)
//	GET /api/overview  -> JSON list of all teams
//	GET /api/team/:name -> JSON snapshot of a single team
//	GET /api/events/:name -> SSE stream (pushes team snapshot on interval)
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

			// --- GET / -> HTML SPA ---
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/" && r.URL.Path != "/index.html" {
					http.NotFound(w, r)
					return
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write([]byte(boardHTML))
			})

			// --- GET /api/overview -> [{name, description}, ...] ---
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
					data, err := os.ReadFile(filepath.Join(teamsDir, e.Name(), "config.json"))
					if err == nil {
						if jsonErr := json.Unmarshal(data, &cfg); jsonErr != nil {
							fmt.Fprintf(os.Stderr, "warning: failed to parse config for team %s: %v\n", e.Name(), jsonErr)
						}
					}
					list = append(list, teamMeta{Name: e.Name(), Description: cfg.Description})
				}
				if list == nil {
					list = []teamMeta{}
				}
				writeJSON2(w, list)
			})

			// --- GET /api/team/:name -> team snapshot ---
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

			// --- GET /api/events/:name -> SSE stream ---
			mux.HandleFunc("/api/events/", func(w http.ResponseWriter, r *http.Request) {
				name := strings.TrimPrefix(r.URL.Path, "/api/events/")
				if name == "" {
					http.Error(w, "team name required", http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")
				w.Header().Set("Access-Control-Allow-Origin", corsOrigin(r))

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
			fmt.Printf("jikime dashboard: http://%s\n", addr)
			if defaultTeam != "" {
				fmt.Printf("   Default team: %s\n", defaultTeam)
			}
			fmt.Println("   Press Ctrl+C to stop.")
			srv := &http.Server{
				Addr:         addr,
				Handler:      mux,
				ReadTimeout:  15 * time.Second,
				WriteTimeout: 60 * time.Second,
				IdleTimeout:  120 * time.Second,
			}
			return srv.ListenAndServe()
		},
	}
	cmd.Flags().IntVarP(&port, "port", "p", 8080, "HTTP server port")
	cmd.Flags().StringVar(&host, "host", "127.0.0.1", "Bind address")
	cmd.Flags().Float64VarP(&interval, "interval", "i", 2.0, "SSE push interval in seconds")
	return cmd
}

// corsOrigin returns the Origin header if it is a localhost address,
// otherwise falls back to http://localhost:4000.
func corsOrigin(r *http.Request) string {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return "http://localhost:4000"
	}
	if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
		return origin
	}
	return "http://localhost:4000"
}

// writeJSON2 serialises v as indented JSON with CORS headers.
func writeJSON2(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4000")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
