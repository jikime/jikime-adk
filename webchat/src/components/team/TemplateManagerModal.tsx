'use client'

import { useState, useEffect, useCallback } from 'react'
import { useServer } from '@/contexts/ServerContext'
import { useLocale } from '@/contexts/LocaleContext'
import MonacoReactEditor from '@monaco-editor/react'
import { Button }     from '@/components/ui/button'
import { Badge }      from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Dialog, DialogContent, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import {
  Plus, Trash2, Save, ChevronRight, LayoutTemplate,
  Users, Code2, Search, FlaskConical, FileText, Sparkles, Loader2,
} from 'lucide-react'
import { cn } from '@/lib/utils'

// ── Default YAML template ────────────────────────────────────────────

const DEFAULT_YAML = `name: my-template
version: "1.0.0"
description: 'Describe your team template here'
default_budget: 200000

agents:
  - id: leader
    role: leader
    auto_spawn: true
    description: 'Coordinates the team and delegates tasks to workers'
    task: |
      Goal: {{goal}}

      You are the team leader for {{team_name}}.

      Your responsibilities:
      1. Analyze the goal and break it into concrete subtasks
      2. Create tasks and assign to workers:
         jikime team tasks create {{team_name}} "Task: <description>" \\
           --desc "Detailed requirements" \\
           --dod "Done when: <acceptance criteria>"
      3. Monitor progress:
         jikime team inbox receive {{team_name}}
      4. Synthesize results and report overall completion

  - id: worker-1
    role: worker
    auto_spawn: false
    description: 'Executes tasks assigned by the leader'
    task: |
      Goal: {{goal}}
      Team: {{team_name}}
      Agent: {{agent_id}}

      Check your inbox for task assignments:
        jikime team inbox receive {{team_name}}

      For each assigned task:
      1. Execute the task thoroughly
      2. Report completion to the leader:
         jikime team inbox send {{team_name}} leader "Completed: <summary>"
      3. Update task status:
         jikime team tasks update {{team_name}} <task-id> --status done
`

// ── Preset templates (YAML strings) ──────────────────────────────────

const PRESETS: Array<{ label: string; icon: React.ReactNode; yaml: string }> = [
  {
    label: 'Dev Team',
    icon:  <Code2 className="w-3.5 h-3.5" />,
    yaml: `name: dev-team
version: "1.0.0"
description: 'Leader + Backend + Frontend + Tester. Full-stack development team.'
default_budget: 300000

agents:
  - id: leader
    role: leader
    auto_spawn: true
    description: 'Tech lead — coordinates development and reviews deliverables'
    task: |
      Goal: {{goal}}

      You are the tech lead for {{team_name}}.

      Steps:
      1. Break the goal into backend / frontend / test tasks
      2. Create and assign tasks:
         jikime team tasks create {{team_name}} "Backend: <feature>" \\
           --desc "API & service implementation" --dod "Unit tested, API documented"
         jikime team tasks create {{team_name}} "Frontend: <feature>" \\
           --desc "UI component & integration" --dod "Reviewed, responsive"
         jikime team tasks create {{team_name}} "Test: <scope>" \\
           --desc "Integration & E2E tests" --dod "Coverage >= 80%"
      3. Review completed work:
         jikime team inbox receive {{team_name}}
      4. Report overall completion

  - id: backend
    role: worker
    auto_spawn: true
    description: 'Implements API endpoints and business logic'
    task: |
      Goal: {{goal}}
      Team: {{team_name}} | Agent: {{agent_id}}

      Check inbox for backend tasks:
        jikime team inbox receive {{team_name}}

      For each task:
      1. Implement the API / service / repository layer
      2. Write unit tests
      3. Report completion:
         jikime team inbox send {{team_name}} leader "Backend done: <summary>"

  - id: frontend
    role: worker
    auto_spawn: true
    description: 'Builds UI components and integrates with backend APIs'
    task: |
      Goal: {{goal}}
      Team: {{team_name}} | Agent: {{agent_id}}

      Check inbox for frontend tasks:
        jikime team inbox receive {{team_name}}

      For each task:
      1. Build the React component / page
      2. Connect to backend APIs
      3. Ensure responsive design and accessibility
      4. Report completion:
         jikime team inbox send {{team_name}} leader "Frontend done: <summary>"

  - id: tester
    role: reviewer
    auto_spawn: false
    description: 'Writes integration tests and validates the full system'
    task: |
      Goal: {{goal}}
      Team: {{team_name}} | Agent: {{agent_id}}

      Check inbox for test tasks:
        jikime team inbox receive {{team_name}}

      For each task:
      1. Write integration and E2E tests
      2. Run the test suite and report coverage
      3. Flag any issues to the leader:
         jikime team inbox send {{team_name}} leader "Test results: <summary>"
`,
  },
  {
    label: 'Research Team',
    icon:  <Search className="w-3.5 h-3.5" />,
    yaml: `name: research-team
version: "1.0.0"
description: 'Leader + 2 Researchers + Analyst. Research, gather info, and synthesize findings.'
default_budget: 250000

agents:
  - id: leader
    role: leader
    auto_spawn: true
    description: 'Defines research questions and coordinates the research team'
    task: |
      Goal: {{goal}}

      You are the research lead for {{team_name}}.

      Steps:
      1. Break the goal into concrete research questions
      2. Assign research tasks:
         jikime team tasks create {{team_name}} "Research: <topic>" \\
           --desc "What to investigate" --dod "Summary with 3+ sources"
      3. Monitor progress:
         jikime team inbox receive {{team_name}}
      4. Request synthesis from analyst once research is complete

  - id: researcher-1
    role: worker
    auto_spawn: true
    description: 'Conducts primary research and web searches'
    task: |
      Goal: {{goal}}
      Team: {{team_name}} | Agent: {{agent_id}}

      Check inbox for research assignments:
        jikime team inbox receive {{team_name}}

      For each topic:
      1. Search for relevant information (use WebSearch)
      2. Collect key facts, data points, and sources
      3. Report findings:
         jikime team inbox send {{team_name}} leader "Research done: <summary>"

  - id: researcher-2
    role: worker
    auto_spawn: false
    description: 'Conducts secondary research and cross-validation'
    task: |
      Goal: {{goal}}
      Team: {{team_name}} | Agent: {{agent_id}}

      Check inbox for research assignments:
        jikime team inbox receive {{team_name}}

      Focus on cross-validating findings from researcher-1 and exploring additional angles.
      Report your findings to the leader.

  - id: analyst
    role: reviewer
    auto_spawn: false
    description: 'Synthesizes research findings into a coherent report'
    task: |
      Goal: {{goal}}
      Team: {{team_name}} | Agent: {{agent_id}}

      After receiving all research summaries from the leader:
      1. Synthesize findings into a structured report
      2. Highlight key insights and recommendations
      3. Identify gaps or conflicting information
      4. Deliver the final report:
         jikime team inbox send {{team_name}} leader "Analysis complete: <report>"
`,
  },
  {
    label: 'Code Review Team',
    icon:  <FlaskConical className="w-3.5 h-3.5" />,
    yaml: `name: code-review-team
version: "1.0.0"
description: 'Developer + Security Reviewer + Quality Reviewer. Thorough code review workflow.'
default_budget: 200000

agents:
  - id: leader
    role: leader
    auto_spawn: true
    description: 'Coordinates code review workflow'
    task: |
      Goal: {{goal}}

      You are the review coordinator for {{team_name}}.

      Steps:
      1. Identify the code / PR scope to review
      2. Assign review tasks:
         jikime team tasks create {{team_name}} "Security Review: <scope>" \\
           --desc "Check for vulnerabilities, auth issues, injection risks" \\
           --dod "OWASP checklist complete"
         jikime team tasks create {{team_name}} "Quality Review: <scope>" \\
           --desc "Check code quality, patterns, maintainability" \\
           --dod "Review comments documented"
      3. Collect review results and compile final report

  - id: developer
    role: worker
    auto_spawn: true
    description: 'Implements requested changes based on review feedback'
    task: |
      Goal: {{goal}}
      Team: {{team_name}} | Agent: {{agent_id}}

      Check inbox for change requests:
        jikime team inbox receive {{team_name}}

      For each feedback item:
      1. Understand the issue and implement the fix
      2. Verify the fix doesn't introduce regressions
      3. Report back:
         jikime team inbox send {{team_name}} leader "Fixed: <issue> — <how>"

  - id: security-reviewer
    role: reviewer
    auto_spawn: true
    description: 'Reviews code for security vulnerabilities (OWASP Top 10)'
    task: |
      Goal: {{goal}}
      Team: {{team_name}} | Agent: {{agent_id}}

      Perform a thorough security review:
      1. Check for injection vulnerabilities (SQL, XSS, command injection)
      2. Review authentication and authorization logic
      3. Identify exposed secrets or sensitive data leaks
      4. Check for insecure dependencies
      5. Report findings:
         jikime team inbox send {{team_name}} leader "Security review: <findings>"

  - id: quality-reviewer
    role: reviewer
    auto_spawn: false
    description: 'Reviews code quality, patterns, and maintainability'
    task: |
      Goal: {{goal}}
      Team: {{team_name}} | Agent: {{agent_id}}

      Perform a code quality review:
      1. Check SOLID principles and design patterns
      2. Identify code duplication and refactoring opportunities
      3. Verify test coverage and test quality
      4. Check documentation completeness
      5. Report findings:
         jikime team inbox send {{team_name}} leader "Quality review: <findings>"
`,
  },
  {
    label: 'Content Team',
    icon:  <FileText className="w-3.5 h-3.5" />,
    yaml: `name: content-team
version: "1.0.0"
description: 'Leader + Writer + Editor. Content creation and editorial workflow.'
default_budget: 150000

agents:
  - id: leader
    role: leader
    auto_spawn: true
    description: 'Defines content strategy and coordinates the team'
    task: |
      Goal: {{goal}}

      You are the content lead for {{team_name}}.

      Steps:
      1. Define the content scope and target audience
      2. Assign writing tasks:
         jikime team tasks create {{team_name}} "Write: <topic>" \\
           --desc "Audience, tone, key points to cover" \\
           --dod "Draft complete, 500-1000 words"
      3. Assign editing tasks once drafts are ready:
         jikime team tasks create {{team_name}} "Edit: <topic>" \\
           --desc "Review for clarity, grammar, SEO" \\
           --dod "Final version approved"
      4. Compile and publish final content

  - id: writer
    role: worker
    auto_spawn: true
    description: 'Creates content drafts based on briefs'
    task: |
      Goal: {{goal}}
      Team: {{team_name}} | Agent: {{agent_id}}

      Check inbox for writing assignments:
        jikime team inbox receive {{team_name}}

      For each assignment:
      1. Research the topic thoroughly
      2. Write a structured, engaging draft
      3. Include relevant examples and data
      4. Submit draft:
         jikime team inbox send {{team_name}} leader "Draft ready: <topic>"

  - id: editor
    role: reviewer
    auto_spawn: false
    description: 'Edits drafts for clarity, grammar, and SEO'
    task: |
      Goal: {{goal}}
      Team: {{team_name}} | Agent: {{agent_id}}

      Check inbox for editing assignments:
        jikime team inbox receive {{team_name}}

      For each draft:
      1. Review for clarity, flow, and grammar
      2. Optimize for SEO (keywords, headings, meta description)
      3. Ensure consistent tone and style
      4. Return edited version:
         jikime team inbox send {{team_name}} leader "Edited: <topic> — ready to publish"
`,
  },
]

// ── Types ─────────────────────────────────────────────────────────────

interface TemplateItem {
  name:          string
  description:   string
  defaultBudget: number
}

interface Props {
  open:    boolean
  onClose: () => void
}

// ── Monaco editor options ─────────────────────────────────────────────

const EDITOR_OPTIONS = {
  minimap:               { enabled: false },
  fontSize:              13,
  lineNumbers:           'on' as const,
  wordWrap:              'on' as const,
  scrollBeyondLastLine:  false,
  tabSize:               2,
  renderLineHighlight:   'line' as const,
  fontFamily:            'Menlo, Monaco, "Courier New", monospace',
  padding:               { top: 12, bottom: 12 },
  scrollbar:             { verticalScrollbarSize: 4, horizontalScrollbarSize: 4 },
}

// ── Component ─────────────────────────────────────────────────────────

export default function TemplateManagerModal({ open, onClose }: Props) {
  const { getApiUrl } = useServer()
  const { t } = useLocale()

  const [templates,    setTemplates]    = useState<TemplateItem[]>([])
  const [selected,     setSelected]     = useState<string | null>(null)
  const [mode,         setMode]         = useState<'list' | 'edit'>('list')
  const [yamlContent,  setYamlContent]  = useState('')
  const [isNew,        setIsNew]        = useState(false)
  const [busy,         setBusy]         = useState(false)
  const [error,        setError]        = useState('')
  const [aiOpen,       setAiOpen]       = useState(false)
  const [aiPrompt,     setAiPrompt]     = useState('')
  const [aiLoading,    setAiLoading]    = useState(false)

  const BUILTIN = ['leader-worker', 'leader-worker-reviewer', 'parallel-workers']

  const loadTemplates = useCallback(async () => {
    try {
      const res  = await fetch(getApiUrl('/api/template/list'))
      const data = await res.json() as { templates: TemplateItem[] }
      setTemplates(data.templates || [])
    } catch { /* */ }
  }, [getApiUrl])

  useEffect(() => { if (open) loadTemplates() }, [open, loadTemplates])

  // ── Handlers ──────────────────────────────────────────────────────

  async function handleSelect(name: string) {
    try {
      const res  = await fetch(getApiUrl(`/api/template/${encodeURIComponent(name)}/yaml`))
      const data = await res.json() as { yaml?: string }
      setYamlContent(data.yaml ?? '')
    } catch {
      setYamlContent('')
    }
    setSelected(name)
    setIsNew(false)
    setMode('edit')
    setError('')
  }

  function handleNew() {
    setSelected(null)
    setYamlContent(DEFAULT_YAML)
    setIsNew(true)
    setMode('edit')
    setError('')
  }

  function handlePreset(yaml: string) {
    setSelected(null)
    setYamlContent(yaml)
    setIsNew(true)
    setMode('edit')
    setError('')
  }

  async function handleSave() {
    const nameMatch = yamlContent.match(/^name:\s*(.+)$/m)
    const name = nameMatch?.[1]?.trim().replace(/^['"]|['"]$/g, '') ?? ''
    if (!name) { setError(t.team.templateNoNameError); return }
    setBusy(true); setError('')
    try {
      const res = await fetch(getApiUrl('/api/template/save'), {
        method:  'POST',
        headers: { 'Content-Type': 'application/json' },
        body:    JSON.stringify({ name, yaml: yamlContent }),
      })
      if (!res.ok) {
        const e = await res.json() as { error?: string }
        setError(e.error || t.team.templateSaveFailed); return
      }
      await loadTemplates()
      setSelected(name)
      setIsNew(false)
    } catch (e) {
      setError(String(e))
    } finally {
      setBusy(false)
    }
  }

  async function handleDelete(name: string) {
    if (!confirm(t.team.templateDeleteConfirm(name))) return
    try {
      await fetch(getApiUrl(`/api/template/${encodeURIComponent(name)}`), { method: 'DELETE' })
      await loadTemplates()
      setMode('list')
      setSelected(null)
    } catch { /* */ }
  }

  function goList() { setMode('list'); setError(''); setAiOpen(false); setAiPrompt('') }

  async function handleAiGenerate() {
    if (!aiPrompt.trim()) return
    setAiLoading(true)
    setAiOpen(false)
    setError('')
    let accumulated = ''
    try {
      const res = await fetch(getApiUrl('/api/template/generate'), {
        method:  'POST',
        headers: { 'Content-Type': 'application/json' },
        body:    JSON.stringify({
          prompt:       aiPrompt.trim(),
          existingYaml: yamlContent || undefined,
        }),
      })
      const reader  = res.body!.getReader()
      const decoder = new TextDecoder()
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        const lines = decoder.decode(value).split('\n')
        for (const line of lines) {
          if (!line.startsWith('data: ')) continue
          try {
            const data = JSON.parse(line.slice(6)) as { chunk?: string; done?: boolean; error?: string }
            if (data.error) { setError(data.error); break }
            if (data.done)  break
            if (data.chunk) { accumulated += data.chunk; setYamlContent(accumulated) }
          } catch { /* skip malformed line */ }
        }
      }
      setAiPrompt('')
      setIsNew(true)
    } catch (e) {
      setError(String(e))
    } finally {
      setAiLoading(false)
    }
  }

  const currentName = yamlContent.match(/^name:\s*(.+)$/m)?.[1]?.trim().replace(/^['"]|['"]$/g, '') ?? ''
  const isBuiltin   = selected ? BUILTIN.includes(selected) : false

  // ── Render ────────────────────────────────────────────────────────

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) { onClose(); goList() } }}>
      <DialogContent className="w-[50vw] max-w-[50vw] sm:max-w-[50vw] h-[50vh] flex flex-col gap-0 p-0 overflow-hidden">

        {/* Header */}
        <DialogHeader className="px-5 py-3 border-b border-border shrink-0">
          <div className="flex items-center justify-between">
            <DialogTitle className="flex items-center gap-2 text-base">
              <LayoutTemplate className="w-4 h-4 text-muted-foreground" />
              {t.team.templateTitle}
              {mode === 'edit' && currentName && (
                <span className="text-sm font-normal text-muted-foreground ml-1">— {currentName}</span>
              )}
            </DialogTitle>
            {mode === 'edit' && (
              <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={goList}>
                {t.team.templateBack}
              </Button>
            )}
          </div>
        </DialogHeader>

        <div className="flex flex-1 min-h-0 overflow-hidden">

          {/* ── Left: Template List ── */}
          <div className="w-52 shrink-0 border-r border-border flex flex-col">
            <div className="px-3 py-2 flex items-center justify-between border-b border-border shrink-0">
              <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">{t.team.templateSidebarHeader}</span>
              <Button size="icon" variant="ghost" className="h-6 w-6" onClick={handleNew}>
                <Plus className="w-3.5 h-3.5" />
              </Button>
            </div>
            <ScrollArea className="flex-1">
              {templates.map((tmpl) => (
                <button
                  key={tmpl.name}
                  onClick={() => handleSelect(tmpl.name)}
                  className={cn(
                    'w-full text-left px-3 py-2.5 flex items-center justify-between gap-1 group transition-colors',
                    selected === tmpl.name && mode === 'edit'
                      ? 'bg-accent text-accent-foreground'
                      : 'hover:bg-muted/50',
                  )}
                >
                  <div className="min-w-0">
                    <div className="text-xs font-medium truncate">{tmpl.name}</div>
                    {BUILTIN.includes(tmpl.name) && (
                      <Badge variant="outline" className="text-[9px] h-3.5 px-1 mt-0.5">{t.team.builtinBadge}</Badge>
                    )}
                  </div>
                  <ChevronRight className="w-3 h-3 shrink-0 opacity-0 group-hover:opacity-50" />
                </button>
              ))}
              {templates.length === 0 && (
                <p className="px-3 py-4 text-xs text-muted-foreground/60 text-center">
                  {t.team.templateNone}
                </p>
              )}
            </ScrollArea>
          </div>

          {/* ── Right: Presets or Monaco Editor ── */}
          <div className="flex-1 min-w-0 flex flex-col overflow-hidden">

            {/* List mode: preset grid */}
            {mode === 'list' && (
              <div className="flex-1 overflow-auto p-6">
                <h3 className="text-sm font-semibold text-foreground mb-1">{t.team.templatePresetsTitle}</h3>
                <p className="text-xs text-muted-foreground mb-4">
                  {t.team.templatePresetsHint}
                </p>
                <div className="grid grid-cols-2 xl:grid-cols-4 gap-3 mb-6">
                  {PRESETS.map((p) => (
                    <button
                      key={p.label}
                      onClick={() => handlePreset(p.yaml)}
                      className="text-left rounded-lg border border-border hover:border-primary/40 hover:bg-accent/50 p-3 transition-all"
                    >
                      <div className="flex items-center gap-2 mb-2">
                        <div className="w-6 h-6 rounded bg-primary/10 flex items-center justify-center text-primary">
                          {p.icon}
                        </div>
                        <span className="text-xs font-semibold text-foreground">{p.label}</span>
                      </div>
                      <p className="text-[11px] text-muted-foreground leading-relaxed line-clamp-2">
                        {p.yaml.match(/^description:\s*['"]?([^'"\\n]+)/m)?.[1] ?? ''}
                      </p>
                    </button>
                  ))}
                </div>

                <div className="text-center py-6 border-t border-border">
                  <div className="w-10 h-10 rounded-full bg-muted flex items-center justify-center mx-auto mb-2">
                    <Users className="w-5 h-5 text-muted-foreground/50" />
                  </div>
                  <p className="text-xs text-muted-foreground mb-3">
                    {t.team.templateSelectPrompt}
                  </p>
                  <Button size="sm" variant="outline" className="text-xs h-7" onClick={handleNew}>
                    <Plus className="w-3 h-3 mr-1" /> {t.team.templateNewBlank}
                  </Button>
                </div>
              </div>
            )}

            {/* Edit mode: Monaco YAML editor */}
            {mode === 'edit' && (
              <>
                {/* read-only notice for built-in templates */}
                {isBuiltin && (
                  <div className="shrink-0 px-4 py-1.5 bg-muted/50 border-b border-border text-xs text-muted-foreground">
                    {t.team.templateBuiltinNotice}
                  </div>
                )}

                {/* Monaco */}
                <div className="flex-1 min-h-0">
                  <MonacoReactEditor
                    height="100%"
                    language="yaml"
                    value={yamlContent}
                    onChange={(val) => setYamlContent(val ?? '')}
                    theme="vs-dark"
                    options={{
                      ...EDITOR_OPTIONS,
                      readOnly: false,
                    }}
                  />
                </div>

                {/* Bottom action bar */}
                <div className="shrink-0 px-5 border-t border-border">
                  {/* AI 프롬프트 입력 영역 */}
                  {aiOpen && (
                    <div className="flex items-center gap-2 py-2 border-b border-border/50">
                      <Sparkles className="w-3.5 h-3.5 text-purple-400 shrink-0" />
                      <input
                        autoFocus
                        value={aiPrompt}
                        onChange={(e) => setAiPrompt(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') handleAiGenerate()
                          if (e.key === 'Escape') setAiOpen(false)
                        }}
                        placeholder={t.team.aiPromptPlaceholder}
                        className="flex-1 h-7 px-2 text-xs bg-muted border border-purple-500/30 rounded outline-none focus:border-purple-500/60 placeholder:text-muted-foreground/50"
                      />
                      <Button
                        size="sm" className="h-7 text-xs bg-purple-600 hover:bg-purple-500 text-white"
                        onClick={handleAiGenerate}
                        disabled={!aiPrompt.trim()}
                      >
                        {t.team.aiGenerateBtn}
                      </Button>
                      <Button size="sm" variant="ghost" className="h-7 text-xs" onClick={() => setAiOpen(false)}>
                        {t.team.cancel}
                      </Button>
                    </div>
                  )}

                  <div className="flex items-center justify-between gap-3 py-3">
                    <div className="flex items-center gap-3 min-w-0">
                      {error && <span className="text-xs text-red-500 truncate">{error}</span>}
                      {!error && (
                        <span className="text-xs text-muted-foreground/60">
                          {isNew ? t.team.newTemplateLabel : `~/.jikime/templates/${selected}.yaml`}
                        </span>
                      )}
                    </div>
                    <div className="flex gap-2 shrink-0">
                      {/* AI 생성 버튼 */}
                      <Button
                        size="sm" variant="outline"
                        className="h-7 text-xs gap-1.5 border-purple-500/30 text-purple-400 hover:bg-purple-500/10"
                        onClick={() => { setAiOpen((v) => !v); setError('') }}
                        disabled={busy || aiLoading}
                      >
                        {aiLoading
                          ? <Loader2 className="w-3 h-3 animate-spin" />
                          : <Sparkles className="w-3 h-3" />
                        }
                        {aiLoading ? t.team.aiGenerating : t.team.aiGenerate}
                      </Button>

                      {!isNew && !isBuiltin && (
                        <Button
                          size="sm" variant="outline"
                          className="h-7 text-xs text-red-500 border-red-500/30 hover:bg-red-500/10"
                          onClick={() => selected && handleDelete(selected)}
                          disabled={busy}
                        >
                          <Trash2 className="w-3 h-3" /> {t.team.templateDelete}
                        </Button>
                      )}
                      <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={goList} disabled={busy}>
                        {t.team.cancel}
                      </Button>
                      <Button size="sm" className="h-7 text-xs gap-1.5" onClick={handleSave} disabled={busy}>
                        <Save className="w-3 h-3" />
                        {busy ? t.team.templateSaving : t.team.templateSave}
                      </Button>
                    </div>
                  </div>
                </div>
              </>
            )}
          </div>
        </div>

      </DialogContent>
    </Dialog>
  )
}
