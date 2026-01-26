# migration-skill 명령어

프레임워크 마이그레이션 스킬을 생성하는 Claude Code 스킬 생성기입니다.

## 개요

`migration-skill`은 레거시 프로젝트를 현대 프레임워크로 마이그레이션하기 위한 전문 스킬을 생성합니다. Claude Code Skills 공식 스펙을 따르며, F.R.I.D.A.Y. 마이그레이션 오케스트레이터와 통합됩니다.

**위치**: `/templates/.claude/commands/jikime/migration-skill.md`

---

## 사용법

```bash
/jikime:migration-skill --from <source> --to <target> [--enhance-only]
```

| 인자 | 필수 | 옵션 | 설명 |
|------|------|------|------|
| `--from` | Yes | `cra`, `vue`, `angular`, `svelte`, `jquery`, `php` | 소스 프레임워크 |
| `--to` | Yes | `nextjs`, `nuxt`, `react`, `vue` | 타겟 프레임워크 |
| `--enhance-only` | No | - | 기존 스킬 개선만 (새로 생성 안함) |

### 예시

```bash
# CRA에서 Next.js로 마이그레이션 스킬 생성
/jikime:migration-skill --from cra --to nextjs

# Vue에서 Nuxt로 마이그레이션 스킬 생성
/jikime:migration-skill --from vue --to nuxt

# 기존 Angular→React 스킬 개선
/jikime:migration-skill --from angular --to react --enhance-only
```

---

## 생성되는 구조

```
jikime-migration-{from}-to-{to}/
├── SKILL.md                    # 메인 스킬 파일 (필수)
├── modules/
│   ├── {from}-patterns.md      # 상세 변환 패턴
│   ├── migration-scenarios.md  # 일반적인 마이그레이션 시나리오
│   └── troubleshooting.md      # 문제 해결 가이드
├── examples/
│   ├── before-after.md         # Before/After 코드 비교
│   └── sample-migration.md     # 전체 마이그레이션 예제
└── scripts/
    └── analyze.sh              # 분석 스크립트 (선택)
```

---

## 실행 워크플로우

### Phase 1: Context7 Research

- 타겟 프레임워크 라이브러리 ID 조회
- 마이그레이션 문서 쿼리
- Codemod CLI 도구 수집
- 점진적 마이그레이션 전략 수집
- 일반적인 함정 식별

### Phase 2: Skill Discovery

- 기존 스킬 검색: `jikime-migration-to-{to}` 또는 `jikime-migrate-{from}-to-{to}`
- 존재하면 → 개선 계획 준비
- 없으면 → 템플릿으로 새로 생성

### Phase 3: Skill Structure Generation

- 완전한 스킬 디렉토리 생성
- 패턴, 시나리오, 예제, 스크립트 파일 생성

### Phase 4: SKILL.md Template Generation

공식 frontmatter 형식으로 생성:

```yaml
---
name: migrate-{from}-to-{to}
description: "{From} to {To} migration specialist..."
argument-hint: [source-path]
disable-model-invocation: false
user-invocable: true
allowed-tools: Read, Grep, Glob, Edit, Write
context: fork
agent: Explore
---
```

### Phase 5: Pattern Module Generation

`modules/{from}-patterns.md` 생성:
- 공식 마이그레이션 도구 & codemods
- 점진적 마이그레이션 전략
- 패턴 매핑 테이블
- 일반적인 함정 & 해결책

---

## 지원 프레임워크

### 소스 프레임워크

| 프레임워크 | Alias | 감지 패턴 |
|-----------|-------|----------|
| Create React App | `cra` | `react-scripts` in package.json |
| Vue.js | `vue` | `vue` in package.json |
| Angular | `angular` | `@angular/core` in package.json |
| Svelte | `svelte` | `svelte` in package.json |
| jQuery | `jquery` | `jquery` in package.json 또는 `$()` 패턴 |
| PHP/Laravel | `php` | `composer.json` 존재 |

### 타겟 프레임워크

| 프레임워크 | Alias | 기본 버전 |
|-----------|-------|----------|
| Next.js | `nextjs` | 16 (App Router) |
| Nuxt | `nuxt` | 3 |
| React | `react` | 19 |
| Vue | `vue` | 3.5 |

---

## SKILL.md Frontmatter

### 공식 필드

| 필드 | 설명 |
|------|------|
| `name` | 스킬 표시 이름 (`/slash-command`가 됨) |
| `description` | 자동 로드 트리거 키워드 |
| `argument-hint` | 자동완성 힌트 |
| `disable-model-invocation` | Claude 자동 호출 제어 |
| `user-invocable` | 사용자 메뉴 표시 여부 |
| `allowed-tools` | 권한 프롬프트 없이 사용 가능한 도구 |
| `model` | 모델 선택 (opus, sonnet, haiku) |
| `context` | 실행 컨텍스트 (fork, inline) |
| `agent` | 서브에이전트 타입 (Explore, Plan 등) |
| `hooks` | 스킬 범위 훅 |

### 예시

```yaml
---
name: migrate-cra-to-nextjs
description: "CRA to Next.js 16 migration specialist. Handles react-scripts removal, App Router migration, SSR/SSG patterns."
argument-hint: [source-path]
user-invocable: true
allowed-tools: Read, Grep, Glob, Edit, Write
context: fork
agent: Explore
---
```

---

## 동적 컨텍스트 주입

스킬 내에서 프로젝트 정보를 동적으로 수집:

```markdown
### Current Dependencies
!`cat package.json 2>/dev/null | grep -A 20 '"dependencies"' || echo "No package.json"`

### Framework Detection
!`ls -la src/ 2>/dev/null | head -20 || echo "No src directory"`
```

### 문자열 치환

- `$ARGUMENTS`: 호출 시 전달된 모든 인자
- `${CLAUDE_SESSION_ID}`: 현재 세션 ID

---

## Progressive Disclosure

스킬은 3단계 로딩 시스템 사용:

| 레벨 | 토큰 | 내용 |
|------|------|------|
| **Level 1** | ~100 | 메타데이터/frontmatter만 |
| **Level 2** | ~5K | 전체 마크다운 본문 |
| **Level 3+** | 가변 | 번들된 참조 파일 (온디맨드 로드) |

---

## 스킬 저장 위치

| 위치 | 경로 | 적용 범위 |
|------|------|----------|
| Enterprise | Managed Settings | 모든 조직 사용자 |
| Personal | `~/.claude/skills/<name>/SKILL.md` | 모든 프로젝트 |
| Project | `.claude/skills/<name>/SKILL.md` | 현재 프로젝트만 |
| Plugin | `<plugin>/skills/<name>/SKILL.md` | 플러그인 활성화 시 |

---

## 품질 체크리스트

스킬 생성 완료 전 확인:

- [ ] Context7 최신 문서 조회 완료
- [ ] Frontmatter가 공식 스펙 준수
- [ ] Description에 사용 시점 명시
- [ ] 주요 마이그레이션 패턴 문서화
- [ ] 동적 컨텍스트 주입으로 프로젝트 분석
- [ ] 코드 예제 문법 정확
- [ ] 점진적 마이그레이션 전략 포함
- [ ] 공식 도구/codemods 문서화
- [ ] 문제 해결 섹션 포함
- [ ] SKILL.md 500줄 이하
- [ ] 버전 및 changelog 업데이트
- [ ] 지원 파일 적절히 링크

---

## 관련 명령어

### 마이그레이션 워크플로우

| 명령어 | 설명 |
|--------|------|
| `/jikime:migrate-0-discover` | 소스 프로젝트 발견 |
| `/jikime:migrate-1-analyze` | 상세 분석 |
| `/jikime:migrate-2-plan` | 마이그레이션 계획 |
| `/jikime:migrate-3-execute` | 마이그레이션 실행 |
| `/jikime:migrate-4-verify` | 검증 |
| `/jikime:friday` | 전체 마이그레이션 오케스트레이션 |

### 관련 스킬

| 스킬 | 설명 |
|------|------|
| `jikime-migration-to-nextjs` | Legacy → Next.js 16 |
| `jikime-migration-angular-to-nextjs` | Angular → Next.js |
| `jikime-migration-jquery-to-react` | jQuery → React |
| `jikime-migration-patterns-auth` | 인증 마이그레이션 패턴 |
| `jikime-migration-ast-grep` | AST 기반 코드 변환 |

---

## Context7 쿼리 템플릿

```
Migration Guide: "{from} to {to} migration guide official"
Pattern Query: "{from} {pattern} equivalent in {to}"
Best Practices: "{to} performance best practices migration"
```

---

Version: 1.0.0
Last Updated: 2026-01-26
