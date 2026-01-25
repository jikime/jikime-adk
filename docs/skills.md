# JikiME-ADK 스킬 시스템

JikiME-ADK의 스킬 시스템 구조와 관리 도구에 대한 문서입니다.

## 개요

스킬은 Claude Code가 특정 도메인이나 작업에 대한 전문 지식을 제공하는 모듈입니다. Progressive Disclosure 패턴을 통해 토큰을 효율적으로 사용합니다.

## 스킬 구조

```
templates/.claude/skills/
├── _template/                    # 스킬 템플릿
│   ├── SKILL.md
│   └── tests/
│       ├── README.md
│       └── examples.yaml
├── jikime-lang-typescript/       # 언어 스킬
│   └── SKILL.md
├── jikime-domain-frontend/       # 도메인 스킬
│   └── SKILL.md
├── jikime-workflow-spec/         # 워크플로우 스킬
│   └── SKILL.md
└── ...
```

### 스킬 명명 규칙

```
jikime-{domain}-{name}
```

| Domain | 설명 | 예시 |
|--------|------|------|
| `lang` | 프로그래밍 언어 | `jikime-lang-typescript` |
| `domain` | 개발 도메인 | `jikime-domain-frontend` |
| `workflow` | 워크플로우 | `jikime-workflow-spec` |
| `platform` | 플랫폼/서비스 | `jikime-platform-vercel` |
| `framework` | 프레임워크 | `jikime-framework-nextjs@16` |
| `library` | 라이브러리 | `jikime-library-zod` |
| `foundation` | 핵심 기반 | `jikime-foundation-core` |
| `marketing` | 마케팅 | `jikime-marketing-seo` |
| `tool` | 도구 | `jikime-tool-ast-grep` |
| `migration` | 마이그레이션 | `jikime-migration-to-nextjs` |

### SKILL.md 구조

```yaml
---
name: jikime-example-skill
description: 스킬에 대한 간단한 설명
version: 1.0.0
tags:
  - example
  - tutorial

# Progressive Disclosure 설정
progressive_disclosure:
  enabled: true
  level1_tokens: ~100
  level2_tokens: ~5000

# 트리거 조건
triggers:
  keywords: ["예시", "example"]
  phases: ["plan", "run"]
  agents: ["manager-spec"]
  languages: ["typescript"]
---

# 스킬 본문

여기에 스킬의 상세 내용을 작성합니다.
```

## 스킬 관리 도구

### 1. 스킬 카탈로그 생성

모든 스킬의 메타데이터를 스캔하여 카탈로그를 생성합니다.

```bash
python3 scripts/generate_skill_catalog.py
```

**생성 파일:**
- `skills-catalog.yaml` - 기계 판독용 카탈로그
- `docs/skills-catalog.md` - 문서용 카탈로그

**언제 실행하나요?**

| 상황 | 재생성 필요 |
|------|------------|
| 새 스킬 추가 | ✅ 예 |
| 스킬 메타데이터(frontmatter) 수정 | ✅ 예 |
| 스킬 본문만 수정 | ❌ 아니오 |
| 스킬 삭제 | ✅ 예 |

### 2. 스킬 메타데이터 검증

모든 스킬의 frontmatter가 스키마를 준수하는지 검증합니다.

```bash
# 모든 스킬 검증
python3 scripts/validate_skills.py

# 특정 스킬만 검증
python3 scripts/validate_skills.py --skill jikime-marketing-seo

# 상세 출력
python3 scripts/validate_skills.py --verbose
```

**검증 항목:**
- 필수 필드: `name`, `description`, `version`
- 이름 패턴: `jikime-{domain}-{name}` 형식
- 버전 형식: semver (프레임워크 스킬은 예외)
- 트리거 설정: phases, keywords 유효성

### 3. 스킬 테스트

스킬의 테스트 예시와 트리거 설정을 검증합니다.

```bash
# 모든 스킬 테스트
python3 scripts/test_skills.py

# 특정 스킬만 테스트
python3 scripts/test_skills.py --skill jikime-marketing-seo

# 상세 출력
python3 scripts/test_skills.py --verbose
```

## 새 스킬 추가하기

### 1. 템플릿 복사

```bash
cp -r templates/.claude/skills/_template templates/.claude/skills/jikime-{domain}-{name}
```

### 2. SKILL.md 작성

`_template/SKILL.md`를 참고하여 frontmatter와 본문을 작성합니다.

### 3. 검증

```bash
python3 scripts/validate_skills.py --skill jikime-{domain}-{name}
```

### 4. 카탈로그 업데이트

```bash
python3 scripts/generate_skill_catalog.py
```

## 스킬 테스트 작성하기

### 테스트 파일 구조

```
skills/jikime-example/
├── SKILL.md
└── tests/
    └── examples.yaml
```

### examples.yaml 형식

```yaml
# 스킬 테스트 예시
name: jikime-example-skill
version: 1.0.0

# 트리거 키워드 (SKILL.md와 동일하거나 테스트용으로 정의)
keywords:
  - 예시
  - example

# 테스트 케이스 (test_N_name, test_N_input, test_N_expected 형식)
test_1_name: 기본 테스트
test_1_input: 예시를 보여줘
test_1_expected: 예시 설명, 코드 샘플

test_2_name: 영어 입력
test_2_input: Show me an example
test_2_expected: Example explanation

# 트리거 검증
should_trigger:
  - 예시 코드 작성
  - example usage

should_not_trigger:
  - 관련 없는 주제
  - 다른 스킬 키워드
```

### 테스트 검증 항목

| 검증 유형 | 설명 |
|----------|------|
| 테스트 구조 | `test_N_input`과 `test_N_expected`가 모두 있는지 |
| 트리거 확인 | 입력이 키워드를 트리거하는지 |
| should_trigger | 트리거되어야 하는 입력이 실제로 트리거되는지 |
| should_not_trigger | 트리거되지 않아야 하는 입력이 트리거되지 않는지 |

## 관련 파일

| 파일 | 설명 |
|------|------|
| `scripts/generate_skill_catalog.py` | 카탈로그 생성 스크립트 |
| `scripts/validate_skills.py` | 메타데이터 검증 스크립트 |
| `scripts/test_skills.py` | 테스트 실행 스크립트 |
| `schemas/skill-frontmatter.schema.json` | frontmatter JSON 스키마 |
| `skills-catalog.yaml` | 생성된 카탈로그 (YAML) |
| `docs/skills-catalog.md` | 생성된 카탈로그 (Markdown) |

## 버전 관리 정책

### Semantic Versioning (SemVer)

스킬은 [Semantic Versioning](https://semver.org/)을 따릅니다.

```
MAJOR.MINOR.PATCH
```

| 버전 변경 | 언제 올리나요? | 예시 |
|----------|---------------|------|
| **MAJOR** | Breaking change (호환성 깨짐) | 트리거 키워드 대폭 변경, 필수 섹션 삭제 |
| **MINOR** | 새 기능 추가 (하위 호환) | 새 패턴 추가, 예시 확장, 섹션 추가 |
| **PATCH** | 버그 수정, 오타 수정 | 오타 수정, 링크 수정, 설명 명확화 |

### 예외: 프레임워크 버전 스킬

`jikime-framework-*` 스킬은 대상 프레임워크 버전을 version 필드에 사용합니다.

```yaml
# jikime-framework-nextjs@16/SKILL.md
name: jikime-framework-nextjs@16
version: "16"  # Next.js 16 버전을 의미
```

### 버전 업데이트 가이드라인

#### 1. PATCH 업데이트 (1.0.0 → 1.0.1)

```yaml
# 변경 전
version: 1.0.0

# 변경 후
version: 1.0.1
```

**해당 경우:**
- 오타 수정
- 설명 문구 개선
- 깨진 링크 수정
- 코드 예시 오류 수정

#### 2. MINOR 업데이트 (1.0.0 → 1.1.0)

```yaml
# 변경 전
version: 1.0.0

# 변경 후
version: 1.1.0
```

**해당 경우:**
- 새로운 패턴/예시 추가
- 새로운 섹션 추가
- 트리거 키워드 추가 (기존 유지)
- Works Well With 스킬 추가

#### 3. MAJOR 업데이트 (1.0.0 → 2.0.0)

```yaml
# 변경 전
version: 1.0.0

# 변경 후
version: 2.0.0
```

**해당 경우:**
- 트리거 키워드 대폭 변경/삭제
- 필수 섹션 구조 변경
- 스킬의 목적/범위 변경
- 다른 스킬과 통합/분리

### 스킬 간 의존성

#### Works Well With 섹션

스킬 간 연관 관계는 `## Works Well With` 섹션에 문서화합니다.

```markdown
## Works Well With

- **jikime-lang-typescript**: TypeScript 타입 정의 패턴
- **jikime-platform-vercel**: 배포 최적화
```

**주의사항:**
- 의존성은 **문서용 참조**이며, 자동 로딩되지 않음
- 필요한 스킬은 명시적으로 `Skill("스킬명")`으로 로딩
- 순환 의존성 주의 (A → B → A)

#### 의존성 관리 원칙

| 원칙 | 설명 |
|------|------|
| **명시적 로딩** | 필요한 스킬만 명시적으로 로딩 (토큰 효율) |
| **느슨한 결합** | 다른 스킬 없이도 독립적으로 작동 가능하게 설계 |
| **문서화** | 연관 스킬은 Works Well With에 명시 |

### Progressive Disclosure 토큰 가이드라인

| 토큰 범위 | 권장 용도 |
|----------|----------|
| ~2000 | 일반 스킬 (lang, library) |
| ~3000-5000 | 복잡한 스킬 (foundation, workflow) |
| ~5000+ | 대규모 스킬 (분할 고려) |

**권장사항:**
- 가능하면 ~2000-3000 토큰 유지
- 5000+ 토큰이면 스킬 분할 검토
- Quick Reference 섹션은 핵심 내용만

## 참고

- 모든 스크립트는 Python 표준 라이브러리만 사용 (외부 의존성 없음)
- 스킬 카탈로그는 현재 59개 스킬을 포함
- Progressive Disclosure로 토큰 효율성 67%+ 개선
