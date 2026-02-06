# Smart Rebuild Reference

Usage, Options, Supported Frameworks 참조 문서.

---

## Purpose

레거시 사이트(웹빌더, PHP 등)를 스크린샷 + 소스 분석 기반으로 현대적 기술 스택(Next.js, Java Spring Boot)으로 **새로 구축**합니다.

## Usage

```bash
# 전체 워크플로우 (권장)
/jikime:smart-rebuild https://example.com --source=./legacy-php --output=./rebuild-output

# Phase 1: 캡처
/jikime:smart-rebuild capture https://example.com --output=./rebuild-output/capture
/jikime:smart-rebuild capture https://example.com --login --output=./rebuild-output/capture

# Phase 2: 분석
/jikime:smart-rebuild analyze --source=./legacy-php --capture=./rebuild-output/capture

# Phase 3: 코드 생성 (페이지별)
/jikime:smart-rebuild generate frontend --page 1
/jikime:smart-rebuild generate frontend --next
/jikime:smart-rebuild generate frontend --status

# Phase 3: 백엔드 생성
/jikime:smart-rebuild generate backend --mapping=./rebuild-output/mapping.json

# Phase 3: 연동
/jikime:smart-rebuild generate connect --frontend-dir=./rebuild-output/frontend
```

---

## Subcommands

| Subcommand | Description |
|------------|-------------|
| (none) | 전체 워크플로우 실행 |
| `capture` | 사이트 크롤링 및 스크린샷 캡처 |
| `analyze` | 소스 분석 및 매핑 생성 |
| `generate frontend` | 프론트엔드 생성 (Mock 데이터 포함) |
| `generate backend` | 백엔드 API 생성 |
| `generate connect` | 프론트엔드와 백엔드 연동 |
| `generate hitl` | HITL 수동 실행 (generate frontend에 통합됨) |

---

## Options

### 전역 옵션

| Option | Description | Default |
|--------|-------------|---------|
| `--output` | 출력 디렉토리 | `./smart-rebuild-output` |
| `--source` | 레거시 소스 경로 | (required) |
| `--target` | 타겟 프론트엔드 프레임워크 | `nextjs16` |
| `--target-backend` | 타겟 백엔드 프레임워크 | `java` |
| `--ui-library` | UI 컴포넌트 라이브러리 | `shadcn` |

### 페이지별 처리 옵션

| Option | Description | Example |
|--------|-------------|---------|
| `--page [n]` | 특정 페이지 ID | `--page 1` |
| `--page [n-m]` | 페이지 범위 | `--page 1-5` |
| `--next` | 다음 pending 페이지 | `--next` |
| `--status` | 상태 조회 | `--status` |

### capture 옵션

> **🔴 Lazy Capture 방식**: 기본적으로 **링크만 수집**하고, HTML + 스크린샷은 `generate --page N` 단계에서 캡처합니다.

| Option | Description | Default |
|--------|-------------|---------|
| `<url>` | 캡처할 사이트 URL | (required) |
| `--merge` | 기존 sitemap.json에 새 route만 추가 | ✅ (기본) |
| `--force` | sitemap 새로 생성 (기존 덮어쓰기) | - |
| `--prefetch` | 🔴 모든 페이지 HTML + 스크린샷 미리 캡처 | - |
| `--clean` | 더 이상 존재하지 않는 route 제거 | - |
| `--max-pages` | 최대 캡처 페이지 수 | `100` |
| `--concurrency` | 동시 처리 수 | `5` |
| `--login` | 로그인 필요 시 | - |
| `--auth` | 기존 세션 파일 재사용 | - |
| `--exclude` | 제외 URL 패턴 | `/admin/*,/api/*` |
| `--no-dedupe` | 템플릿 중복 제거 비활성화 | `false` |

**Lazy Capture 동작:**
- 기본: 링크만 수집 → `captured: false`
- `--prefetch` 사용 시: 모든 페이지 HTML + 스크린샷 캡처 → `captured: true`

### analyze 옵션

| Option | Description | Default |
|--------|-------------|---------|
| `--source` | 레거시 소스 경로 | (required) |
| `--capture` | 캡처 디렉토리 | `./capture` |
| `--output` | 매핑 파일 출력 | `./mapping.json` |
| `--framework` | 소스 프레임워크 오버라이드 | 자동 감지 |
| `--db-schema` | DB 스키마 파일 | - |
| `--db-from-env` | .env에서 스키마 추출 | - |

### generate frontend 옵션

| Option | Description | Default |
|--------|-------------|---------|
| `--mapping` | 매핑 파일 | `./mapping.json` |
| `--output` | 출력 디렉토리 | `./output/frontend` |
| `--capture` | 캡처 디렉토리 | `./capture` |
| `--target` | 타겟 프레임워크 | `nextjs16` |
| `--ui-library` | UI 라이브러리 | `shadcn` |

### generate hitl 옵션

| Option | Description | Default |
|--------|-------------|---------|
| `--capture` | 캡처 디렉토리 | `./capture` |
| `--page` | 처리할 페이지 ID | (다음 pending) |
| `--section` | 처리할 섹션 ID | (다음 pending) |
| `--responsive` | 반응형 테스트 | `false` |
| `--status` | 진행 상황 확인 | `false` |
| `--approve=ID` | 섹션 승인 | - |
| `--skip=ID` | 섹션 스킵 | - |
| `--reset` | 상태 초기화 | `false` |

---

## Supported Frameworks

### Source (레거시)

| 프레임워크 | 자동 감지 | 매칭 전략 |
|-----------|----------|----------|
| `php-pure` | ✅ index.php 기반 | 파일 기반 라우팅 |
| `wordpress` | ✅ wp-config.php | 테마/플러그인 기반 |
| `laravel` | ✅ artisan CLI | routes/web.php |
| `codeigniter` | ✅ application/controllers | Controllers/Views |
| `symfony` | ✅ symfony.lock | src/Controller |

### Target (생성)

| 구분 | 프레임워크 | 기본값 | 연동 Skill |
|------|-----------|--------|------------|
| Frontend | `nextjs16` | ✅ | `jikime-framework-nextjs@16` |
| Frontend | `nextjs15` | - | `jikime-framework-nextjs@15` |
| Frontend | `react` | - | `jikime-domain-frontend` |
| Backend | `java` | ✅ | `jikime-lang-java` |
| Backend | `go` | - | `jikime-lang-go` |
| Backend | `python` | - | `jikime-lang-python` |

### UI Library

| Value | 설명 | 연동 Skill |
|-------|------|------------|
| `shadcn` | shadcn/ui (Recommended) | `jikime-library-shadcn` |
| `mui` | Material UI | (향후 지원) |
| `legacy-css` | 레거시 CSS 복사 (비권장) | - |

---

## 파일 네이밍 규칙

| 파일 유형 | 규칙 | 예시 |
|----------|------|------|
| 페이지/라우트 | kebab-case | `about-us/page.tsx` |
| 컴포넌트 | kebab-case | `header-nav.tsx` |
| Java 클래스 | PascalCase | `MemberEntity.java` |
| Go 파일 | snake_case | `member_handler.go` |
| Python 파일 | snake_case | `member_router.py` |

---

## sitemap.json 구조 (Lazy Capture)

```json
{
  "baseUrl": "https://example.com",
  "createdAt": "2026-02-05T10:00:00Z",
  "updatedAt": "2026-02-06T14:30:00Z",
  "totalPages": 15,
  "summary": {
    "pending": 13,
    "in_progress": 1,
    "completed": 1,
    "captured": 2
  },
  "pages": [
    {
      "id": 1,
      "url": "https://example.com/",
      "title": "홈페이지",
      "captured": true,
      "screenshot": "page_1_home.png",
      "html": "page_1_home.html",
      "status": "completed",
      "capturedAt": "2026-02-06T10:00:00Z"
    },
    {
      "id": 2,
      "url": "https://example.com/about",
      "title": "About Us",
      "captured": false,
      "screenshot": null,
      "html": null,
      "status": "pending",
      "capturedAt": null
    }
  ]
}
```

**주요 필드:**
| 필드 | 설명 |
|------|------|
| `summary.captured` | HTML + 스크린샷 캡처 완료된 페이지 수 |
| `page.captured` | 🔴 해당 페이지 캡처 여부 (false면 generate 시 캡처) |
| `page.capturedAt` | 해당 페이지 실제 캡처 시간 |

---

## Output Structure

```
{output}/
├── capture/
│   ├── sitemap.json     # 캡처 인덱스 + captured 상태
│   ├── *.png            # 스크린샷 (캡처된 페이지만)
│   └── *.html           # HTML (캡처된 페이지만)
├── mapping.json         # 소스 ↔ 캡처 매핑
├── backend/
│   └── src/main/java/   # Spring Boot
└── frontend/
    └── src/
        ├── app/                    # Next.js App Router
        │   ├── page.tsx            # 홈 (섹션 컴포넌트 조합)
        │   └── about-us/page.tsx   # 섹션 컴포넌트 import
        ├── styles/                 # 🔴 원본 CSS 저장
        │   ├── legacy/             # fetch한 CSS 파일들
        │   └── legacy-imports.css
        └── components/             # 🔴 섹션 컴포넌트
            ├── common/             # 공통 (헤더, 푸터)
            ├── home/               # 홈 페이지 섹션들
            └── about-us/           # about-us 섹션들
                ├── hero-section.tsx
                └── team-section.tsx
```

---

## Troubleshooting

### 캡처 실패
- Playwright 브라우저 설치 확인: `npx playwright install chromium`
- 타임아웃 조정: `--timeout=60000`

### 로그인 필요 사이트
- `--login` 옵션 사용
- 브라우저에서 로그인 완료 후 Enter

### HITL 스크립트 실행 안 됨
- SCRIPTS_DIR 경로 확인
- npm install 실행 여부 확인

---

Version: 1.0.0
