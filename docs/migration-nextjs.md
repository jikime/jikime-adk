# Next.js Migration System

Legacy 프론트엔드 프로젝트를 Next.js 16 App Router로 마이그레이션하기 위한 완전 가이드입니다.

---

## Overview

JikiME-ADK의 Next.js Migration System은 Vue.js, React CRA, Angular, Svelte 등 다양한 레거시 프로젝트를 현대적인 Next.js 16 + App Router 아키텍처로 전환하는 워크플로우를 제공합니다.

### Target Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| Framework | Next.js 16 | App Router architecture |
| Language | TypeScript 5.x | Type safety |
| Styling | Tailwind CSS 4.x | Utility-first CSS |
| UI Components | shadcn/ui | Accessible components |
| Icons | lucide-react | Icon library |
| State | Zustand | State management |
| Form | react-hook-form + zod | Form validation |

### Supported Source Frameworks

| Framework | Versions | Detection Pattern |
|-----------|----------|-------------------|
| Vue.js | 2.x, 3.x | `vue` in package.json, `.vue` files |
| React CRA | 16+, 17+, 18+ | `react-scripts` in package.json |
| React Vite | All | `@vitejs/plugin-react` in vite.config |
| Angular | 12+ ~ 17+ | `@angular/core` in package.json |
| Svelte | 3.x, 4.x | `svelte` in package.json |

---

## System Complete Status

Next.js Migration System의 전체 구성 요소 현황입니다.

### Skills (jikime-migrate-to-nextjs/) - 24개 파일

| 카테고리 | 파일 수 | 내용 |
|---------|--------|------|
| **SKILL.md** | 1 | 메인 스킬 정의 (v1.7.0) |
| **modules/** | 7 | project-initialization, migration-flow-tutorial, cheatsheet, migration-scenario, nextjs16-patterns, react-patterns, vue-patterns |
| **templates/pre-migration/** | 7 | 00_cover ~ 06_baseline_report (Whitepaper 생성용) |
| **templates/post-migration/** | 8 | 00_cover ~ 07_maintenance_guide (결과 리포트용) |
| **templates/project-skill-template** | 1 | 프로젝트별 커스텀 스킬 템플릿 |

### Commands (migrate 시리즈) - 6개 파일

| 파일 | 용도 |
|-----|------|
| `migrate.md` | 메인 오케스트레이터 |
| `migrate-0-discover.md` | 소스 프레임워크 자동 탐지 |
| `migrate-1-analyze.md` | 마이그레이션 영향도 분석 |
| `migrate-2-plan.md` | 단계별 마이그레이션 계획 |
| `migrate-3-execute.md` | 실제 코드 변환 실행 |
| `migrate-4-verify.md` | 마이그레이션 결과 검증 |

### 위치 정보

```
templates/.claude/
├── skills/
│   └── jikime-migrate-to-nextjs/
│       ├── SKILL.md                      # 메인 스킬
│       ├── modules/                      # 7개 모듈
│       └── templates/                    # 16개 템플릿
│           ├── project-skill-template.md
│           ├── pre-migration/            # 7개 백서 템플릿
│           └── post-migration/           # 8개 리포트 템플릿
│
└── commands/jikime/
    ├── migrate.md                        # 메인 명령어
    ├── migrate-0-discover.md
    ├── migrate-1-analyze.md
    ├── migrate-2-plan.md
    ├── migrate-3-execute.md
    └── migrate-4-verify.md
```

---

## Skills 구조

### Core Skill: `jikime-migrate-to-nextjs`

**위치**: `templates/.claude/skills/jikime-migrate-to-nextjs/`

```
jikime-migrate-to-nextjs/
├── SKILL.md                    # 메인 스킬 정의 (v1.7.0)
├── modules/
│   ├── project-initialization.md   # 프로젝트 초기화 가이드 ⭐
│   ├── migration-flow-tutorial.md  # CLI 명령어 튜토리얼 ⭐
│   ├── cheatsheet.md               # 빠른 참조 가이드 ⭐
│   ├── migration-scenario.md       # 전체 마이그레이션 시나리오
│   ├── nextjs16-patterns.md        # Next.js 16 패턴
│   ├── react-patterns.md           # React → Next.js 전환
│   └── vue-patterns.md             # Vue → Next.js 전환
└── templates/
    ├── project-skill-template.md   # 프로젝트별 SKILL.md 템플릿
    ├── pre-migration/              # 사전 분석 백서 템플릿
    └── post-migration/             # 완료 보고서 템플릿
```

### Version-Specific Skills

| Skill | Location | Purpose |
|-------|----------|---------|
| `jikime-nextjs@14` | `templates/.claude/skills/jikime-nextjs@14/` | Next.js 14 baseline |
| `jikime-nextjs@15` | `templates/.claude/skills/jikime-nextjs@15/` | 14→15 breaking changes |
| `jikime-nextjs@16` | `templates/.claude/skills/jikime-nextjs@16/` | 15→16 new features |

### Domain Pattern Skills

| Skill | Location | Purpose |
|-------|----------|---------|
| `jikime-migration-patterns-auth` | `templates/.claude/skills/jikime-migration-patterns-auth/` | 인증 마이그레이션 패턴 |
| `jikime-library-vercel-ai-sdk` | `templates/.claude/skills/jikime-library-vercel-ai-sdk/` | Vercel AI SDK 통합 |

---

## Migration Workflow

### 전체 흐름

```
/jikime:migrate-analyze "./my-vue-app"
    │
    └─▶ ./migrations/my-vue-app/as_is_spec.md
                │
/jikime:migrate-to-nextjs plan my-vue-app
                │
                └─▶ ./migrations/my-vue-app/migration_plan.md
                            │
/jikime:migrate-to-nextjs skill my-vue-app
                            │
                            └─▶ .claude/skills/my-vue-app/SKILL.md
                                        │
/jikime:migrate-to-nextjs run my-vue-app [--output ./out]
                                        │
                                        └─▶ ./migrations/my-vue-app/out/
                                            (migrated project)
```

### Phase 0: Analyze (현황 분석)

```bash
# 레거시 프로젝트 분석
/jikime:migrate-analyze "./my-vue-app"

# 백서 생성 포함
/jikime:migrate-analyze "./my-vue-app" --whitepaper --client "ABC Corp"
```

**Output**: `./migrations/my-vue-app/as_is_spec.md`

### Phase 1: Plan (마이그레이션 계획)

```bash
/jikime:migrate-to-nextjs plan my-vue-app
```

**Output**: `./migrations/my-vue-app/migration_plan.md`

### Phase 2: Skill (프로젝트별 스킬 생성)

```bash
/jikime:migrate-to-nextjs skill my-vue-app
```

**Output**: `.claude/skills/my-vue-app/SKILL.md`

생성된 SKILL.md는 자동으로 다음 스킬들을 참조합니다:
- `jikime-nextjs@16` (필수)
- `jikime-migration-patterns-auth` (인증 코드 감지 시)
- `jikime-platform-*` (플랫폼 감지 시)

### Phase 3: Run (마이그레이션 실행)

```bash
# 기본 실행
/jikime:migrate-to-nextjs run my-vue-app

# 출력 경로 지정
/jikime:migrate-to-nextjs run my-vue-app --output ./new-project

# 완료 보고서 생성
/jikime:migrate-to-nextjs run my-vue-app --whitepaper-report --client "ABC Corp"
```

---

## Project Initialization (Quick Start)

새 Next.js 16 프로젝트를 처음부터 시작하거나, 마이그레이션 대상 프로젝트를 생성할 때 사용합니다.

### One-Liner Setup

```bash
# 1. 프로젝트 생성
npx create-next-app@latest my-project --typescript --tailwind --eslint --app --src-dir

# 2. 폴더 이동
cd my-project

# 3. shadcn/ui 초기화
npx shadcn@latest init

# 4. 기본 컴포넌트 설치
npx shadcn@latest add button card input label form dialog toast

# 5. 필수 패키지 설치
npm install zustand react-hook-form @hookform/resolvers zod lucide-react

# 6. 개발 서버 실행
npm run dev
```

### shadcn/ui 초기화 옵션

```
✔ Which style would you like to use? › New York
✔ Which color would you like to use as the base color? › Neutral
✔ Would you like to use CSS variables for theming? › Yes
```

| 옵션 | 권장값 | 설명 |
|------|--------|------|
| Style | **New York** | 모던하고 깔끔한 스타일 |
| Base color | **Neutral** | 범용적인 그레이 톤 |
| CSS variables | **Yes** | 다크모드, 테마 지원 |

### 컴포넌트 카테고리별 설치

```bash
# 기본 UI
npx shadcn@latest add button card input label textarea

# 폼 관련
npx shadcn@latest add form checkbox select switch radio-group

# 레이아웃/네비게이션
npx shadcn@latest add dialog dropdown-menu sheet tabs navigation-menu

# 피드백/상태
npx shadcn@latest add alert badge toast skeleton progress

# 데이터 표시
npx shadcn@latest add table avatar separator
```

### 마이그레이션용 추천 세트 (한 번에 설치)

```bash
npx shadcn@latest add \
  button card input label textarea \
  form checkbox select switch \
  dialog dropdown-menu sheet tabs \
  alert badge toast skeleton \
  table avatar separator
```

---

## Component Mapping Reference

### Vue → Next.js

| Vue Pattern | Next.js Equivalent |
|-------------|-------------------|
| `<template>` | JSX/TSX |
| `<script setup>` | Function component |
| `<style scoped>` | CSS Modules / Tailwind |
| `ref()`, `reactive()` | `useState` |
| `computed()` | `useMemo` |
| `watch()` | `useEffect` |
| `onMounted()` | `useEffect(() => {}, [])` |
| `v-if` / `v-else` | `{condition && ...}` |
| `v-for` | `.map()` |
| `v-model` | Controlled component |
| Vue Router | App Router |
| Vuex / Pinia | Zustand |

### React CRA → Next.js

| CRA Pattern | Next.js Equivalent |
|-------------|-------------------|
| `src/index.js` entry | `app/layout.tsx` |
| `react-router-dom` | App Router (file-based) |
| `BrowserRouter` | Remove (built-in) |
| `useNavigate()` | `useRouter()` from `next/navigation` |
| `<Link>` | `<Link>` from `next/link` |
| `process.env.REACT_APP_*` | `process.env.NEXT_PUBLIC_*` |

### State Migration Decision Tree

```
Is the state...
│
├─ Global (app-wide)?
│   ├─ Complex with actions? → Zustand
│   ├─ Simple shared state? → React Context
│   └─ Server state? → React Query / SWR
│
├─ Component-local?
│   ├─ Primitive value? → useState
│   ├─ Object/Array? → useState with immutable updates
│   └─ Derived value? → useMemo
│
└─ Form state?
    ├─ Simple form? → useState + controlled
    └─ Complex form? → react-hook-form + zod
```

---

## Skill Reference System

`migrate-to-nextjs skill` 명령어로 생성되는 프로젝트별 SKILL.md는 자동으로 관련 스킬들을 참조합니다.

### Auto-Linked Skills

| Skill | Purpose | Auto-Load Condition |
|-------|---------|---------------------|
| `jikime-nextjs@16` | Next.js 16 App Router 패턴 | 항상 |
| `jikime-nextjs@15` | Next.js 15 Breaking Changes | async params 감지 시 |
| `jikime-migration-patterns-auth` | 인증 마이그레이션 | 인증 코드 감지 시 |

### Conditional Reference Skills

| Skill | Trigger Condition |
|-------|-------------------|
| `jikime-platform-clerk` | Clerk import 감지 |
| `jikime-platform-supabase` | Supabase import 감지 |
| `jikime-library-vercel-ai-sdk` | AI SDK import 감지 |
| `jikime-library-shadcn` | shadcn/ui 사용 시 (기본) |

### How It Works

```
migrate-to-nextjs skill my-app
    │
    ├─▶ 1. as_is_spec.md 분석
    │
    ├─▶ 2. 사용 기술 감지 (Auth, State, Styling)
    │
    ├─▶ 3. 관련 스킬 자동 매핑
    │       - jikime-nextjs@16 (필수)
    │       - jikime-migration-patterns-auth (조건부)
    │       - jikime-platform-* (조건부)
    │
    └─▶ 4. SKILL.md 생성 (Reference Skills 섹션 포함)
```

---

## Whitepaper Generation

### Pre-Migration Whitepaper (사전 분석 보고서)

```bash
# 전체 백서 패키지 생성
/jikime:migrate-analyze "./my-vue-app" --whitepaper --client "ABC Corp"

# 다국어 지원
/jikime:migrate-analyze "./my-vue-app" --whitepaper --client "ABC Corp" --lang en
/jikime:migrate-analyze "./my-vue-app" --whitepaper --client "株式会社ABC" --lang ja
```

**Output Structure** (`./whitepaper/`):

| Document | Purpose | Audience |
|----------|---------|----------|
| `00_cover.md` | 표지, 목차 | 모든 독자 |
| `01_executive_summary.md` | 경영진 요약 | 의사결정자 |
| `02_feasibility_report.md` | 타당성 보고서 | PM, Tech Lead |
| `03_architecture_report.md` | AS-IS/TO-BE 비교 | 개발팀, 아키텍트 |
| `04_complexity_matrix.md` | 복잡도, 공수 산정 | PM, Tech Lead |
| `05_migration_roadmap.md` | 상세 일정 | 전체 팀 |
| `06_baseline_report.md` | 보안/성능 현황 | 보안팀, DevOps |

### Post-Migration Whitepaper (완료 보고서)

```bash
# 마이그레이션 완료 후 결과 보고서
/jikime:migrate-to-nextjs run my-vue-app --whitepaper-report --client "ABC Corp"
```

**Output Structure** (`./whitepaper-report/`):

| Document | Purpose | Audience |
|----------|---------|----------|
| `00_cover.md` | 핵심 성과 요약 | 모든 독자 |
| `01_executive_summary.md` | ROI 요약 | 경영진 |
| `02_performance_comparison.md` | Before/After 비교 | 기술팀 |
| `03_security_improvement.md` | 보안 개선 사항 | 보안팀 |
| `04_code_quality_report.md` | 코드 품질 비교 | 개발팀 |
| `05_architecture_evolution.md` | 아키텍처 진화 | 아키텍트 |
| `06_cost_benefit_analysis.md` | 비용 효과 분석 | 재무팀 |
| `07_maintenance_guide.md` | 유지보수 가이드 | DevOps |

---

## Output Structure

### Migration Artifacts

```
./migrations/{project-name}/
├── as_is_spec.md           # Phase 0: 현황 분석
├── migration_plan.md       # Phase 1: 마이그레이션 계획
├── component_mapping.yaml  # 컴포넌트 매핑 테이블
├── SKILL.md                # Phase 2: 프로젝트별 스킬
└── progress.yaml           # 진행 상황 추적
```

### Migrated Project Structure

```
{output-dir}/{project-name}/
├── src/
│   ├── app/
│   │   ├── layout.tsx
│   │   ├── page.tsx
│   │   └── {routes}/
│   ├── components/
│   │   ├── ui/           # shadcn 컴포넌트
│   │   └── {migrated}/   # 마이그레이션된 컴포넌트
│   ├── lib/
│   │   └── utils.ts
│   └── stores/           # Zustand 스토어
├── public/
├── package.json
├── tailwind.config.ts
├── tsconfig.json
└── next.config.ts
```

---

## Agent Delegation

### Phase별 에이전트 배정

| Phase | Command | Primary Agent | Supporting |
|-------|---------|---------------|------------|
| 0 | migrate-analyze | Explore | expert-frontend |
| 1 | migrate-to-nextjs plan | manager-spec | manager-strategy |
| 2 | migrate-to-nextjs skill | builder-skill | expert-frontend |
| 3 | migrate-to-nextjs run | manager-ddd | expert-frontend, expert-testing |

---

## Git Strategy

```yaml
branch_prefix: "migrate/"
commit_pattern: "migrate({scope}): {description}"

examples:
  - "migrate(analyze): complete AS_IS analysis for my-vue-app"
  - "migrate(Header): convert Header.vue to Header.tsx"
  - "migrate(routing): implement App Router navigation"
  - "migrate(complete): finish my-vue-app migration"
```

---

## Troubleshooting

### shadcn init 실패

```bash
# 캐시 클리어 후 재시도
npm cache clean --force
npx shadcn@latest init
```

### Tailwind 스타일 미적용

```typescript
// tailwind.config.ts 확인
const config = {
  content: [
    './src/**/*.{js,ts,jsx,tsx,mdx}',  // src 폴더 포함 확인
  ],
  // ...
}
```

### TypeScript 경로 오류

```json
// tsconfig.json 확인
{
  "compilerOptions": {
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

### 빌드 실패

```bash
# 타입 체크
npx tsc --noEmit

# 빌드 테스트
npm run build
```

---

## Quick Reference

### 필수 명령어 요약

```bash
# 분석
/jikime:migrate-analyze "./legacy-app"

# 계획
/jikime:migrate-to-nextjs plan legacy-app

# 스킬 생성
/jikime:migrate-to-nextjs skill legacy-app

# 실행
/jikime:migrate-to-nextjs run legacy-app
```

### 프로젝트 초기화 요약

```bash
npx create-next-app@latest my-project --typescript --tailwind --eslint --app --src-dir
cd my-project
npx shadcn@latest init
npx shadcn@latest add button card input label form dialog toast
npm install zustand react-hook-form @hookform/resolvers zod lucide-react
npm run dev
```

---

## Related Documentation

| Document | Description |
|----------|-------------|
| `@modules/project-initialization.md` | 프로젝트 초기화 상세 가이드 |
| `@modules/migration-flow-tutorial.md` | CLI 명령어 튜토리얼 |
| `@modules/cheatsheet.md` | 빠른 참조 카드 |
| `@modules/migration-scenario.md` | 전체 마이그레이션 시나리오 |
| `docs/migration.md` | 마이그레이션 시스템 개요 |

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.8.0 | 2026-01-22 | Added System Complete Status section with file inventory |
| 1.7.0 | 2026-01-22 | Added project-initialization.md (create-next-app, shadcn init) |
| 1.6.0 | 2026-01-22 | Added Skill Reference System - auto-references related skills |
| 1.5.0 | 2026-01-22 | Added migration-flow-tutorial.md, cheatsheet.md |
| 1.4.0 | 2026-01-21 | Added --whitepaper-output and --lang options |
| 1.3.0 | 2026-01-21 | Reorganized templates (pre/post-migration) |
| 1.2.0 | 2026-01-20 | Added Post-Migration Whitepaper |
| 1.1.0 | 2026-01-20 | Added Pre-Migration Whitepaper |
| 1.0.0 | 2026-01-19 | Initial release |

---

Version: 1.8.0
Last Updated: 2026-01-22
