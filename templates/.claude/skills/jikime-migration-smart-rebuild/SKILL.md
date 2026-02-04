# Smart Rebuild

> AI-powered legacy site rebuilding workflow

```yaml
version: 1.0.0
triggers:
  - smart-rebuild
  - site-rebuild
  - legacy-rebuild
skills: jikime-mobile-react-native, jikime-framework-nextjs@16, jikime-lang-java
```

## Overview

Smart Rebuild는 레거시 사이트를 **새로 구축**하는 AI 기반 워크플로우입니다.

**철학**: "Rebuild, not Migrate" — 코드를 변환하지 않고, 새로 만든다.

| 계층 | 전략 |
|------|------|
| UI | 스크린샷 기반 새로 생성 |
| API | 소스 참고하여 클린 코드로 |
| DB | 유지 + 점진적 개선 |

## Quick Reference

| Phase | 설명 | 도구 |
|-------|------|------|
| **Capture** | 사이트 크롤링 & 스크린샷 | Playwright |
| **Analyze** | 소스 분석 & 매핑 | AST Parser |
| **Generate** | 코드 생성 | Claude Code |

## Usage

```bash
# 전체 프로세스
/jikime:smart-rebuild https://example.com --source=./legacy-php

# 단계별 실행
/jikime:smart-rebuild capture https://example.com
/jikime:smart-rebuild analyze --source=./legacy-php
/jikime:smart-rebuild generate --mapping=./mapping.json
```

## 2-Track Strategy

### Track 1: Static Content
```
라이브 사이트 → Playwright 스크래핑 → Next.js 정적 페이지
```
- 소개, About, FAQ, 이용약관 등

### Track 2: Dynamic Content
```
소스 분석 → SQL 추출 → Backend API → Next.js 동적 페이지
```
- 회원 목록, 결제 내역, 게시판 등

## Files

| File | Purpose |
|------|---------|
| `rules/overview.md` | 전체 워크플로우 가이드 |
| `rules/phase-1-capture.md` | 캡처 단계 상세 |
| `rules/phase-2-analyze.md` | 분석 & 매핑 단계 |
| `rules/phase-3-generate.md` | 코드 생성 단계 |
| `rules/troubleshooting.md` | 문제 해결 가이드 |
| `scripts/` | CLI 도구 |

## Related

- [Smart Rebuild 상세 문서](../../../docs/smart-rebuild.md)
- F.R.I.D.A.Y. 마이그레이션 오케스트레이터
