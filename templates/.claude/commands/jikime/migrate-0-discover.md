---
description: "[Step 0/4] 소스 프로젝트 탐색. 기술 스택, 아키텍처, 마이그레이션 복잡도 식별."
---

# Migration Step 0: Discover

**시작 단계**: 소스 코드를 탐색하고 기본 분석합니다.

## What This Command Does

1. **Technology Detection** - 언어, 프레임워크, 라이브러리 식별
2. **Architecture Analysis** - 구조, 패턴, 의존성 파악
3. **Complexity Assessment** - 마이그레이션 난이도 평가
4. **Migration Report** - 상세 분석 보고서 생성

## Usage

```bash
# Discover source codebase
/jikime:migrate-0-discover @legacy-app/

# Discover with target in mind
/jikime:migrate-0-discover @legacy-app/ --target nextjs

# Quick discovery (overview only)
/jikime:migrate-0-discover @legacy-app/ --quick
```

## Options

| Option | Description |
|--------|-------------|
| `@path` | Source code path to analyze |
| `--target` | Intended target stack (helps focus analysis) |
| `--quick` | Quick overview without deep analysis |
| `--output` | Save report to file |

## Output

```markdown
# Discovery Report

## Source Overview
- **Language**: PHP 7.4
- **Framework**: Laravel 8
- **Database**: MySQL 5.7
- **Frontend**: jQuery + Blade

## Complexity Score: 7/10 (Medium-High)

## Recommended Target
Based on analysis: **Next.js 15 + Prisma**

## Next Step
Run `/jikime:migrate-1-analyze` for deep analysis
```

## Agent Delegation

| Phase | Agent | Purpose |
|-------|-------|---------|
| Analysis | `source-analyzer` | Legacy code analysis |
| Architecture | `architect` | Pattern identification |

## Workflow

```
/jikime:migrate-0-discover  ← 현재
        ↓
/jikime:migrate-1-analyze
        ↓
/jikime:migrate-2-plan
        ↓
/jikime:migrate-3-execute
        ↓
/jikime:migrate-4-verify
```

## Next Step

탐색 후 다음 단계로:
```bash
/jikime:migrate-1-analyze
```

---

Version: 2.1.0
