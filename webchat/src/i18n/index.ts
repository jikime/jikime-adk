import ko from './locales/ko'
import en from './locales/en'
import zh from './locales/zh'
import ja from './locales/ja'

export type Locale = 'ko' | 'en' | 'zh' | 'ja'

export const LOCALES: { id: Locale; label: string; flag: string }[] = [
  { id: 'ko', label: '한국어', flag: '🇰🇷' },
  { id: 'en', label: 'English', flag: '🇺🇸' },
  { id: 'zh', label: '中文',   flag: '🇨🇳' },
  { id: 'ja', label: '日本語', flag: '🇯🇵' },
]

// Use a structural type so all locale objects satisfy it
export type Messages = {
  layout: {
    tabs: { chat: string; terminal: string; files: string; git: string }
    theme: { toDark: string; toLight: string }
  }
  sidebar: {
    noServer: string; connected: string; connecting: string; addServer: string
    projects: string; noProjects: string; noProjectsHint: string; newChat: string
    settings: string; deleteProject: string; deleteSession: string
    save: string; cancel: string; delete: string
    serverName: string; serverHost: string; secureConnection: string
    settingsTitle: string; defaultModel: string; permissionMode: string
    autoAllow: string; confirmEach: string; autoAllowDesc: string; confirmEachDesc: string
    gitPatTitle: string; gitPatPlaceholder: string; gitPatDesc: string
    activeModel: string
    deleteProjectTitle: string; deleteProjectDesc: string
    deleteSessionCount: (n: number) => string
    deleteProjectWarning: string
    deleteSessionTitle: string; deleteSessionDesc: string; deleteSessionWarning: string
    models: {
      opusDesc: string; sonnetDesc: string; opus45Desc: string
      sonnet45Desc: string; haikuDesc: string
    }
  }
  chat: {
    permissionRequest: string; loadingHistory: string; selectProject: string
    projectLabel: (path: string) => string
    placeholderStreaming: string; placeholder: string; attachFile: string
    extendedThinking: string; moreModels: string
    extendedThinkingTitle: string; extendedThinkingDesc: string
    voiceStop: string; voiceStart: string; thinking: string
    allow: string; alwaysAllow: string; deny: string
    emptyState: string; stopResponding: string; abortedBanner: string; retry: string
    selectSession: string; selectSessionHint: string
  }
  shell: { connecting: string; ready: string; dead: string; error: string }
  files: {
    selectProject: string; search: string; noResults: string; noFiles: string
    selectFile: string; selectFileHint: string; unsavedChanges: string
    saved: string; saving: string; save: string; readError: string; readErrorGeneric: string
  }
  git: {
    noChanges: string; changes: string; log: string; branch: string; selectAll: string
    loading: string; selectFileDiff: string; commitPlaceholder: string
    selectFilePlaceholder: string; noCommits: string; noBranches: string; commit: string
    notGitRepo: string; selectProject: string; issues: string
    // issues panel
    issuesSelectProject: string; issuesNoPat: string; issuesRepoError: string
    issuesNoIssues: string; issuesPatHint: string
    issuesPollingTarget: string; issuesOther: string
    issuesProcessing: string; issuesPollingLabel: string
    issuesAutoPolling: string; issuesIntervalSuffix: string
    issuesActiveCount: (n: number) => string
    issuesLastCheck: string; issuesLastCheckAgo: (s: number) => string; issuesChecking: string
    issuesStopPolling: string; issuesStartPolling: string; issuesStarting: string; issuesStop: string
    issuesPolling: string
    issuesNewIssue: string; issuesTitlePlaceholder: string; issuesBodyPlaceholder: string
    issuesCreate: string; issuesCreating: string
    issuesAutoProcessing: string
    issuesManualProcess: string; issuesStopManual: string
    issuesLogRunning: string; issuesLogDone: string; issuesLogError: string
    // harness setup
    harnessSetup: string; harnessSetupTitle: string; harnessSetupNoWorkflow: string
    harnessSlug: string; harnessLabel: string; harnessWorkspaceRoot: string
    harnessPort: string; harnessMaxAgents: string; harnessMode: string
    harnessModeBasic: string; harnessModeJikiMe: string; harnessAutoDetected: string
    harnessGenerate: string; harnessGenerating: string
    harnessStart: string; harnessRunning: string
  }
  team: {
    // Toolbar & tabs
    tabBoard: string; tabAgents: string
    selectTeam: string; noTeams: string
    addTask: string; refresh: string; runTeam: string; newTeam: string
    live: string; offline: string
    members: (n: number) => string
    ledBy: (name: string) => string
    // Empty states
    selectProject: string; selectProjectHint: string
    selectTeamHint: (project: string) => string
    selectTeamDesc: string
    noAgents: string; noAgentsHint: string
    // Board sections
    sectionMembers: string; sectionMessages: string; sectionTasks: string
    noMembers: string; noMessages: string; inboxEmpty: string; inboxMsg: (n: number) => string
    // Kanban status columns
    statusPending: string; statusInProgress: string; statusBlocked: string
    statusDone: string; statusFailed: string
    // Team create modal
    createTitle: string; createNameLabel: string
    createTemplateLabel: string; createWorkersLabel: string
    createBudgetLabel: string; createBudgetHint: string
    createTemplateCustom: string; createTemplateManage: string
    creating: string; createFailed: string; createBtn: string
    // Task add modal
    addTaskTitle: string; taskTitleLabel: string; taskDescLabel: string
    taskTitlePlaceholder: string; taskDescPlaceholder: string
    adding: string; addBtn: string
    // Team serve/run modal
    runTitle: string; goalLabel: string; goalPlaceholder: string
    worktreeLabel: string; running: string; runFailed: string; runBtn: string
    // Template manager
    templateTitle: string; templateSidebarHeader: string
    templateNewBlank: string; templatePresetsTitle: string; templatePresetsHint: string
    templateBuiltinNotice: string; templateBack: string
    templateDelete: string; templateSave: string; templateSaving: string
    templateSaveFailed: string; templateNoNameError: string
    templateDeleteConfirm: (name: string) => string
    templateNone: string; builtinBadge: string; templateSelectPrompt: string
    newTemplateLabel: string
    // AI generation
    aiGenerate: string; aiGenerating: string; aiGenerateBtn: string; aiPromptPlaceholder: string
    // Common
    cancel: string
  }
}

export const messages: Record<Locale, Messages> = { ko, en, zh, ja }
