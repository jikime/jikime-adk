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
    notGitRepo: string; selectProject: string
  }
}

export const messages: Record<Locale, Messages> = { ko, en, zh, ja }
