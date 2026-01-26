# skill-create 명령어

Claude Code 스킬을 생성하는 범용 스킬 생성기입니다.

## 개요

`skill-create`는 다양한 유형의 전문 스킬을 생성합니다. Claude Code Skills 공식 스펙을 따르며, Progressive Disclosure 패턴을 적용합니다.

**위치**: `/templates/.claude/commands/jikime/skill-create.md`

---

## 사용법

```bash
/jikime:skill-create --type <type> --name <name> [--enhance-only]
```

| 인자 | 필수 | 옵션 | 설명 |
|------|------|------|------|
| `--type` | Yes | `lang`, `platform`, `domain`, `workflow`, `library`, `framework` | 스킬 유형 |
| `--name` | Yes | 임의의 이름 | 스킬 이름 |
| `--enhance-only` | No | - | 기존 스킬 개선만 (새로 생성 안함) |

### 예시

```bash
# Rust 언어 전문가 스킬 생성
/jikime:skill-create --type lang --name rust

# Firebase 플랫폼 스킬 생성
/jikime:skill-create --type platform --name firebase

# 보안 도메인 스킬 생성
/jikime:skill-create --type domain --name security

# CI/CD 워크플로우 스킬 생성
/jikime:skill-create --type workflow --name ci-cd

# Prisma 라이브러리 스킬 생성
/jikime:skill-create --type library --name prisma

# Remix 프레임워크 스킬 생성
/jikime:skill-create --type framework --name remix

# 기존 Python 스킬 개선
/jikime:skill-create --type lang --name python --enhance-only
```

---

## 스킬 유형별 생성 구조

### `lang` (언어 전문가)

```
jikime-lang-{name}/
├── SKILL.md              # 메인 스킬 파일
├── examples.md           # 프로덕션 코드 예제
└── reference.md          # 완전한 API 레퍼런스
```

**용도**: 프로그래밍 언어별 문법, 패턴, 베스트 프랙티스

**예시**: `jikime-lang-typescript`, `jikime-lang-python`, `jikime-lang-rust`

### `platform` (플랫폼 통합)

```
jikime-platform-{name}/
├── SKILL.md              # 메인 스킬 파일
├── setup.md              # 설정 및 구성 가이드
└── reference.md          # API 및 통합 레퍼런스
```

**용도**: 클라우드 플랫폼, SaaS 서비스 통합

**예시**: `jikime-platform-vercel`, `jikime-platform-supabase`, `jikime-platform-firebase`

### `domain` (도메인 전문가)

```
jikime-domain-{name}/
├── SKILL.md              # 메인 스킬 파일
├── patterns.md           # 도메인 특화 패턴
└── examples.md           # 구현 예제
```

**용도**: 특정 기술 영역의 전문 지식

**예시**: `jikime-domain-frontend`, `jikime-domain-backend`, `jikime-domain-security`

### `workflow` (워크플로우)

```
jikime-workflow-{name}/
├── SKILL.md              # 메인 스킬 파일
├── steps.md              # 워크플로우 단계 및 페이즈
└── examples.md           # 워크플로우 예제
```

**용도**: 개발 프로세스, 자동화, CI/CD 패턴

**예시**: `jikime-workflow-tdd`, `jikime-workflow-ddd`, `jikime-workflow-ci-cd`

### `library` (라이브러리 전문가)

```
jikime-library-{name}/
├── SKILL.md              # 메인 스킬 파일
├── examples.md           # 사용 예제
└── reference.md          # API 레퍼런스
```

**용도**: 특정 라이브러리/패키지의 사용법

**예시**: `jikime-library-zod`, `jikime-library-prisma`, `jikime-library-shadcn`

### `framework` (프레임워크 전문가)

```
jikime-framework-{name}/
├── SKILL.md              # 메인 스킬 파일
├── patterns.md           # 프레임워크 패턴
└── upgrade.md            # 버전 업그레이드 가이드
```

**용도**: 프레임워크별 컨벤션, 라우팅, 컴포넌트

**예시**: `jikime-framework-nextjs`, `jikime-framework-remix`, `jikime-framework-nuxt`

---

## 실행 워크플로우

### Phase 1: Context7 Research

스킬 유형에 따라 Context7에서 관련 문서 조회:

1. 라이브러리 ID 조회: `resolve-library-id`
2. 문서 쿼리: `query-docs`
3. 수집: API 패턴, 베스트 프랙티스, 일반적인 함정, 버전 정보

### Phase 2: Skill Discovery

기존 스킬 검색:
- 패턴: `jikime-{type}-{name}` 또는 `jikime-{name}`
- 존재 + `--enhance-only` → 개선 계획 준비
- 존재 + 플래그 없음 → 사용자에게 질문 (개선 / 새로 생성 / 취소)
- 없음 → 새로 생성

### Phase 3: SKILL.md Template Generation

공식 frontmatter 형식으로 생성:

```yaml
---
name: jikime-{type}-{name}
description: "{Name} {type} specialist covering..."
version: 1.0.0
tags: ["{type}", "{name}"]
triggers:
  keywords: ["{name}"]
  phases: ["run"]
  agents: [relevant agents]
progressive_disclosure:
  enabled: true
  level1_tokens: ~100
  level2_tokens: ~5000
user-invocable: false
allowed-tools: [Read, Grep, Glob, Context7 MCP tools]
---
```

### Phase 4: Supporting Files Generation

유형별 지원 파일 생성:

| 유형 | 생성 파일 |
|------|----------|
| lang | examples.md + reference.md |
| platform | setup.md + reference.md |
| domain | patterns.md + examples.md |
| workflow | steps.md + examples.md |
| library | examples.md + reference.md |
| framework | patterns.md + upgrade.md |

### Phase 5: Progressive Disclosure Integration

SKILL.md에서 지원 파일 참조:

```markdown
## Advanced Patterns

For comprehensive documentation, see:

- examples.md for production-ready code examples
- reference.md for complete API reference
```

---

## SKILL.md Frontmatter

### 공식 필드

| 필드 | 설명 |
|------|------|
| `name` | 스킬 이름 (`jikime-{type}-{name}`) |
| `description` | 사용 시점 포함 설명 |
| `version` | 시맨틱 버전 |
| `tags` | 분류 태그 |
| `triggers` | Level 2 로딩 트리거 조건 |
| `progressive_disclosure` | Progressive Disclosure 설정 |
| `user-invocable` | 사용자 직접 호출 가능 여부 |
| `allowed-tools` | 허용된 도구 목록 |

---

## Progressive Disclosure

스킬은 3단계 로딩 시스템 사용:

| 레벨 | 토큰 | 내용 |
|------|------|------|
| **Level 1** | ~100 | 메타데이터/frontmatter만 |
| **Level 2** | ~5K | 전체 SKILL.md 본문 |
| **Level 3+** | 가변 | 번들된 참조 파일 (온디맨드 로드) |

---

## 스킬 저장 위치

| 위치 | 경로 | 적용 범위 |
|------|------|----------|
| Personal | `~/.claude/skills/<name>/SKILL.md` | 모든 프로젝트 |
| Project | `.claude/skills/<name>/SKILL.md` | 현재 프로젝트만 |
| Plugin | `<plugin>/skills/<name>/SKILL.md` | 플러그인 활성화 시 |

**기본값**: 프로젝트 레벨 `.claude/skills/jikime-{type}-{name}/`

---

## 품질 체크리스트

스킬 생성 완료 전 확인:

- [ ] Context7 최신 문서 조회 완료
- [ ] Frontmatter가 공식 스펙 준수
- [ ] Description에 사용 시점 명시 ("Use when...")
- [ ] Progressive Disclosure 설정 완료
- [ ] SKILL.md 500줄 이하
- [ ] 지원 파일 적절히 링크
- [ ] 코드 예제 문법 정확
- [ ] Context7 라이브러리 매핑 문서화
- [ ] 관련 스킬 식별
- [ ] 문제 해결 섹션 포함
- [ ] 버전 및 changelog 업데이트

---

## 관련 명령어

| 명령어 | 설명 |
|--------|------|
| `/jikime:migration-skill` | 마이그레이션 전용 스킬 생성 |
| `jikime-adk skill list` | 모든 스킬 목록 조회 |
| `jikime-adk skill info <name>` | 스킬 상세 정보 조회 |
| `jikime-adk skill search <keyword>` | 스킬 검색 |

---

## migration-skill과의 차이점

| 항목 | skill-create | migration-skill |
|------|--------------|-----------------|
| **용도** | 범용 스킬 생성 | 마이그레이션 전용 |
| **유형** | 6가지 (lang, platform, domain, workflow, library, framework) | 마이그레이션만 |
| **생성 구조** | 유형별 다름 | modules/, examples/, scripts/ 고정 |
| **Context7 쿼리** | 유형별 API/패턴 | 마이그레이션 가이드 |
| **워크플로우 통합** | 일반 개발 | F.R.I.D.A.Y. 마이그레이션 |

---

## Context7 쿼리 템플릿

```
# 언어 스킬
Query: "{name} language features best practices"

# 플랫폼 스킬
Query: "{name} SDK API integration guide"

# 도메인 스킬
Query: "{name} architecture patterns best practices"

# 워크플로우 스킬
Query: "{name} workflow automation CI/CD"

# 라이브러리 스킬
Query: "{name} library API usage examples"

# 프레임워크 스킬
Query: "{name} framework conventions routing"
```

---

Version: 1.0.0
Last Updated: 2026-01-26
