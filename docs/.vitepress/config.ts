import { defineConfig } from 'vitepress'

const koSidebar = [
  {
    text: '시작하기',
    items: [
      { text: '소개', link: '/ko/commands' },
      { text: '규칙 & 원칙', link: '/ko/rules' },
      { text: '배포 가이드', link: '/ko/deployment' },
      { text: '릴리스 가이드', link: '/ko/release' },
    ],
  },
  {
    text: '스킬 시스템',
    items: [
      { text: '스킬 목록', link: '/ko/skills-catalog' },
      { text: '스킬 개요', link: '/ko/skills' },
      { text: '스킬 만들기', link: '/ko/skill-create' },
      { text: '마케팅 스킬', link: '/ko/marketing' },
    ],
  },
  {
    text: 'AI 에이전트',
    items: [
      { text: 'Agents 레퍼런스', link: '/ko/agents' },
      { text: 'Agent Teams', link: '/ko/agents-team' },
      { text: 'J.A.R.V.I.S.', link: '/ko/jarvis' },
      { text: 'F.R.I.D.A.Y.', link: '/ko/friday' },
    ],
  },
  {
    text: '워크플로우',
    items: [
      { text: 'Harness Engineering', link: '/ko/harness-engineering' },
      { text: 'Harness Workflow', link: '/ko/harness-workflow' },
      { text: 'Harness Test Flow', link: '/ko/harness-test-flow' },
      { text: 'Ralph Loop', link: '/ko/ralph-loop' },
      { text: 'POC-First 개발', link: '/ko/poc-first' },
      { text: 'TDD & DDD', link: '/ko/tdd-ddd' },
      { text: 'PR 라이프사이클', link: '/ko/pr-lifecycle' },
      { text: 'Worktree 관리', link: '/ko/worktree' },
      { text: 'Task 포맷', link: '/ko/task-format' },
      { text: 'Sync 워크플로우', link: '/ko/sync' },
      { text: 'Site-Flow', link: '/ko/site-flow' },
    ],
  },
  {
    text: '마이그레이션',
    items: [
      { text: '마이그레이션 시스템', link: '/ko/migration' },
      { text: '마이그레이션 스킬', link: '/ko/migration-skill' },
      { text: 'Smart Rebuild', link: '/ko/smart-rebuild' },
      { text: 'Smart Rebuild 플로우', link: '/ko/smart-rebuild-flow' },
      { text: 'Playwright 검증', link: '/ko/migrate-playwright' },
    ],
  },
  {
    text: '시스템 레퍼런스',
    items: [
      { text: 'Hooks 시스템', link: '/ko/hooks' },
      { text: 'Auto-Memory', link: '/ko/auto-memory' },
      { text: 'Context 레퍼런스', link: '/ko/context' },
      { text: 'Codemap', link: '/ko/codemap' },
      { text: 'Statusline', link: '/ko/statusline' },
      { text: 'Provider Router', link: '/ko/provider-router' },
      { text: 'WSL 리눅스 폰트', link: '/wsl-linux-font' },
    ],
  },
  {
    text: 'Webchat',
    items: [
      { text: '설치 가이드', link: '/ko/webchat/installation' },
      { text: '사용법', link: '/ko/webchat/usage' },
      { text: '원격 서버 연결', link: '/ko/webchat/remote-server' },
      { text: '아키텍처', link: '/ko/webchat/architecture' },
      { text: '트러블슈팅', link: '/ko/webchat/troubleshooting' },
      { text: '업데이트 가이드', link: '/ko/webchat/update' },
    ],
  },
]

const enSidebar = [
  {
    text: 'Getting Started',
    items: [
      { text: 'Commands', link: '/en/commands' },
      { text: 'Rules & Principles', link: '/en/rules' },
      { text: 'Deployment', link: '/en/deployment' },
      { text: 'Release Guide', link: '/en/release' },
    ],
  },
  {
    text: 'Skills System',
    items: [
      { text: 'Skills Catalog', link: '/en/skills-catalog' },
      { text: 'Skills Overview', link: '/en/skills' },
      { text: 'Create a Skill', link: '/en/skill-create' },
      { text: 'Marketing Skills', link: '/en/marketing' },
    ],
  },
  {
    text: 'AI Agents',
    items: [
      { text: 'Agents Reference', link: '/en/agents' },
      { text: 'Agent Teams', link: '/en/agents-team' },
      { text: 'J.A.R.V.I.S.', link: '/en/jarvis' },
      { text: 'F.R.I.D.A.Y.', link: '/en/friday' },
    ],
  },
  {
    text: 'Workflows',
    items: [
      { text: 'Harness Engineering', link: '/en/harness-engineering' },
      { text: 'Harness Workflow', link: '/en/harness-workflow' },
      { text: 'Harness Test Flow', link: '/en/harness-test-flow' },
      { text: 'Ralph Loop', link: '/en/ralph-loop' },
      { text: 'POC-First Dev', link: '/en/poc-first' },
      { text: 'TDD & DDD', link: '/en/tdd-ddd' },
      { text: 'PR Lifecycle', link: '/en/pr-lifecycle' },
      { text: 'Worktree', link: '/en/worktree' },
      { text: 'Task Format', link: '/en/task-format' },
      { text: 'Sync Workflow', link: '/en/sync' },
      { text: 'Site-Flow', link: '/en/site-flow' },
    ],
  },
  {
    text: 'Migration',
    items: [
      { text: 'Migration System', link: '/en/migration' },
      { text: 'Migration Skill', link: '/en/migration-skill' },
      { text: 'Smart Rebuild', link: '/en/smart-rebuild' },
      { text: 'Smart Rebuild Flow', link: '/en/smart-rebuild-flow' },
      { text: 'Playwright Verify', link: '/en/migrate-playwright' },
    ],
  },
  {
    text: 'Reference',
    items: [
      { text: 'Hooks System', link: '/en/hooks' },
      { text: 'Auto-Memory', link: '/en/auto-memory' },
      { text: 'Context Reference', link: '/en/context' },
      { text: 'Codemap', link: '/en/codemap' },
      { text: 'Statusline', link: '/en/statusline' },
      { text: 'Provider Router', link: '/en/provider-router' },
    ],
  },
]

export default defineConfig({
  title: 'JikiME-ADK',
  description: 'AI Development Kit — Skills, Agents, Workflows for Claude Code',
  base: '/jikime-adk/',
  ignoreDeadLinks: true,

  head: [
    ['link', { rel: 'icon', href: '/jikime-adk/favicon.ico' }],
  ],

  locales: {
    root: {
      label: 'Select Language',
      lang: 'en',
    },
    ko: {
      label: '한국어',
      lang: 'ko',
      link: '/ko/',
      themeConfig: {
        nav: [
          { text: '홈', link: '/ko/' },
          { text: '스킬', link: '/ko/skills-catalog' },
          { text: 'Webchat', link: '/ko/webchat/installation' },
          {
            text: '언어',
            items: [
              { text: '한국어', link: '/ko/' },
              { text: 'English', link: '/en/' },
            ],
          },
        ],
        sidebar: {
          '/ko/': koSidebar,
        },
      },
    },
    en: {
      label: 'English',
      lang: 'en',
      link: '/en/',
      themeConfig: {
        nav: [
          { text: 'Home', link: '/en/' },
          { text: 'Skills', link: '/en/skills-catalog' },
          {
            text: 'Language',
            items: [
              { text: '한국어', link: '/ko/' },
              { text: 'English', link: '/en/' },
            ],
          },
        ],
        sidebar: {
          '/en/': enSidebar,
        },
      },
    },
  },

  themeConfig: {
    logo: '/logo.svg',
    socialLinks: [
      { icon: 'github', link: 'https://github.com/jikime/jikime-adk' },
    ],
    search: {
      provider: 'local',
    },
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2025 JikiME',
    },
  },
})
