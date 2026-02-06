# Smart Rebuild Execution Guide

상세 실행 절차, 코드 예시, 워크플로우 가이드.

---

## 🚨🚨🚨 UI 생성 핵심 원칙 (CRITICAL!) 🚨🚨🚨

**Claude는 반드시 HTML + 스크린샷을 보고 원본과 동일한 UI를 재현해야 합니다!**

### 🔴 HARD RULES

| # | 규칙 | 설명 |
|---|------|------|
| 1 | **스크린샷 필수 분석** | 코드 작성 전 반드시 스크린샷을 Read하고 시각적으로 분석 |
| 2 | **HTML 구조 복사** | `<header>`, `<nav>`, `<main>`, `<footer>` 구조 그대로 유지 |
| 3 | **원본 텍스트 유지** | HTML에서 추출한 텍스트를 번역 없이 원본 그대로 사용 |
| 4 | **원본 이미지 URL** | HTML의 `<img src="...">` URL을 그대로 사용 |
| 5 | **원본 CSS Fetch** | 원본 사이트의 CSS를 WebFetch로 가져와 `src/styles/`에 저장 |
| 6 | **섹션 컴포넌트 분리** | 섹션별로 `components/{route}/*-section.tsx` 파일 생성 |
| 7 | **섹션 식별자 필수** | 모든 주요 섹션에 `data-section-id` 속성 추가 (HITL 비교용) |
| 8 | **스크린샷 기반 스타일** | 색상, 폰트 크기, 간격은 스크린샷에서 추출 |
| 9 | **kebab-case 네이밍** | 폴더/파일명은 반드시 kebab-case (`about-us/`, `hero-section.tsx`) |

### ❌ 절대 하지 말 것

- ❌ 스크린샷 안 보고 기본 템플릿으로 대충 만들기
- ❌ HTML 내용 번역하기 (영어→한글, 한글→영어)
- ❌ 텍스트나 이미지 내용 상상해서 창작하기
- ❌ 원본과 다른 레이아웃이나 색상 사용하기
- ❌ PascalCase 폴더명 사용 (`AboutUs/` ❌ → `about-us/` ✅)
- ❌ 섹션에 `data-section-id` 빼먹기 (HITL 비교 불가!)

### ✅ 반드시 해야 할 것

```
1. Read: {capture}/sitemap.json         # 페이지 정보 확인
2. Read: {capture}/{screenshot_file}    # 🔴 스크린샷 시각 분석 (레이아웃, 색상, 간격)
3. Read: {capture}/{html_file}          # 🔴 HTML에서 텍스트, 이미지 URL 추출
4. Write: 코드 작성                      # 🔴 원본과 동일하게!
```

---

## Core Philosophy

| 계층 | 전략 | 이유 |
|------|------|------|
| **UI** | 스크린샷 → 새로 생성 | 레거시 프론트 코드 분석 가치 낮음 |
| **API** | 소스 참고 → 클린 코드 | 비즈니스 로직만 추출 |
| **DB** | 유지 + 점진적 개선 | 데이터 손실 Zero |

## 2-Track Strategy

### Track 1: Static Content (정적 콘텐츠)
```
라이브 사이트 → Playwright 스크래핑 → Next.js 정적 페이지
```
- 소개, About, FAQ, 이용약관 등
- DB 연동 없음, 콘텐츠만 이동

### Track 2: Dynamic Content (동적 콘텐츠)
```
소스 분석 → SQL 추출 → Backend API → Next.js 동적 페이지
```
- 회원 목록, 결제 내역, 게시판 등
- DB 연동 필수, 비즈니스 로직 있음

---

## Phase 1: Capture (링크 수집)

**목표:** Playwright로 라이브 사이트의 모든 링크를 수집하여 sitemap.json 생성

> **🔴 Lazy Capture 방식**: capture 단계에서는 **링크만 수집**합니다.
> 실제 HTML + 스크린샷 캡처는 `generate --page N` 단계에서 해당 페이지 처리 시 수행됩니다.

### 캡처 옵션

| 옵션 | 설명 | 기본값 |
|------|------|--------|
| `--merge` | 기존 sitemap.json에 새 route만 추가 | ✅ (기본) |
| `--force` | sitemap 새로 생성 (기존 덮어쓰기) | - |
| `--prefetch` | 모든 페이지 HTML + 스크린샷 미리 캡처 | - |
| `--clean` | 더 이상 존재하지 않는 route 제거 | - |

### 실행 절차 (기본: 링크만 수집)

**1단계: sitemap.json 확인**
```
IF sitemap.json 존재 AND --force 아님:
  → 기존 sitemap 로드
  → 증분 모드 (새 링크만 추가)
ELSE:
  → 새로운 sitemap 생성
```

**2단계: 링크 크롤링 (HTML/스크린샷 캡처 안 함!)**
1. Playwright 브라우저 초기화
2. 시작 URL 방문
3. 페이지 내 `<a href>` 태그에서 내부 링크 수집
4. 수집된 링크 재귀적으로 방문 & 링크 수집
5. 각 URL 정규화 (trailing slash, query params 제거)
6. 중복 제거

**3단계: sitemap.json 생성/업데이트**
```
- 발견된 모든 URL을 pages 배열에 추가
- captured: false (아직 캡처 안 됨)
- status: pending
```

### URL 정규화 규칙

```javascript
function normalizeUrl(url) {
  const parsed = new URL(url);
  parsed.search = '';  // query params 제거
  parsed.hash = '';    // hash 제거
  let path = parsed.pathname;
  if (path !== '/' && path.endsWith('/')) {
    path = path.slice(0, -1);  // trailing slash 제거
  }
  parsed.pathname = path;
  return parsed.toString();
}
```

### 링크 수집 코드

```javascript
const { chromium } = require('playwright');

async function collectLinks(startUrl, baseUrl, maxPages = 100) {
  const browser = await chromium.launch();
  const visited = new Set();
  const toVisit = [normalizeUrl(startUrl)];
  const pages = [];

  while (toVisit.length > 0 && pages.length < maxPages) {
    const url = toVisit.shift();
    if (visited.has(url)) continue;
    visited.add(url);

    const page = await browser.newPage();
    try {
      await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 15000 });

      // 페이지 제목 추출
      const title = await page.title();

      // 내부 링크 수집 (HTML/스크린샷 캡처 안 함!)
      const links = await page.$$eval('a[href]', (anchors, base) =>
        anchors.map(a => a.href).filter(h => h.startsWith(base) && !h.includes('#')),
        baseUrl
      );

      // 새 링크들 큐에 추가
      for (const link of links) {
        const normalized = normalizeUrl(link);
        if (!visited.has(normalized) && !toVisit.includes(normalized)) {
          toVisit.push(normalized);
        }
      }

      pages.push({
        id: pages.length + 1,
        url: url,
        title: title,
        captured: false,      // 🔴 아직 캡처 안 됨
        screenshot: null,
        html: null,
        status: 'pending',
        links: [...new Set(links.map(normalizeUrl))]
      });

    } catch (e) {
      console.error(`Failed to visit: ${url}`, e.message);
    } finally {
      await page.close();
    }
  }

  await browser.close();
  return pages;
}
```

### --prefetch 옵션 (전체 미리 캡처)

일괄 생성이나 오프라인 작업이 필요한 경우:

```bash
/jikime:smart-rebuild capture https://example.com --prefetch
```

이 옵션 사용 시:
- 모든 페이지 HTML + 스크린샷 미리 캡처
- `captured: true`로 설정
- 기존 방식과 동일하게 동작

### 상태별 처리 (증분 모드)

| 기존 상태 | 새 크롤링에서 발견 | 처리 |
|----------|------------------|------|
| 있음 | O | 유지 (건너뛰기) |
| (없음) | O | **추가** (새 route) |
| 있음 | X | 유지 (삭제 안 함) |

### Playwright 크롤링 코드 (레거시 - --prefetch용)

```javascript
const { chromium } = require('playwright');

async function capturePage(browser, url, baseUrl, outputDir) {
  const page = await browser.newPage();

  await page.goto(url, { waitUntil: 'networkidle', timeout: 30000 });

  // Lazy loading 해결: 자동 스크롤
  await page.evaluate(async () => {
    await new Promise((resolve) => {
      let total = 0;
      const timer = setInterval(() => {
        window.scrollBy(0, 500);
        total += 500;
        if (total >= document.body.scrollHeight || total >= 30000) {
          clearInterval(timer);
          window.scrollTo(0, 0);
          resolve();
        }
      }, 100);
    });
  });

  // 스크린샷 + HTML 저장
  const filename = url.replace(/https?:\/\//, '').replace(/[^a-zA-Z0-9]/g, '_').slice(0, 80);
  await page.screenshot({ path: `${outputDir}/${filename}.png`, fullPage: true });
  const html = await page.content();
  require('fs').writeFileSync(`${outputDir}/${filename}.html`, html);

  // 내부 링크 수집
  const links = await page.$$eval('a[href]', (anchors, base) =>
    anchors.map(a => a.href).filter(h => h.startsWith(base) && !h.includes('#')),
    baseUrl
  );

  return { url, filename, links: [...new Set(links)] };
}
```

### sitemap.json 구조 (Lazy Capture)

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
      "status": "pending",
      "type": "static",
      "capturedAt": "2026-02-06T10:00:00Z",
      "completedAt": null,
      "links": ["https://example.com/about", "..."]
    },
    {
      "id": 2,
      "url": "https://example.com/about",
      "title": "About Us",
      "captured": false,
      "screenshot": null,
      "html": null,
      "status": "pending",
      "type": null,
      "capturedAt": null,
      "completedAt": null,
      "links": []
    }
  ]
}
```

**필드 설명:**

| 필드 | 설명 |
|------|------|
| `createdAt` | sitemap 최초 생성 시간 (링크 수집 시점) |
| `updatedAt` | 마지막 업데이트 시간 |
| `summary.captured` | HTML + 스크린샷 캡처 완료된 페이지 수 |
| `page.captured` | 🔴 **해당 페이지 캡처 여부** (false면 generate 시 캡처) |
| `page.screenshot` | 캡처된 경우 파일명, 미캡처 시 null |
| `page.html` | 캡처된 경우 파일명, 미캡처 시 null |
| `page.capturedAt` | 해당 페이지 실제 캡처 시간 |

---

## Phase 2: Analyze (분석 & 매핑)

**목표:** 소스 코드 분석하여 캡처와 매핑, 정적/동적 분류

### 분류 패턴

```javascript
const dynamicPatterns = [
  /SELECT\s+.+\s+FROM/gi,
  /INSERT\s+INTO/gi,
  /UPDATE\s+.+\s+SET/gi,
  /DELETE\s+FROM/gi,
  /mysqli_query|\$pdo->query|\$wpdb->/g,
  /\$_SESSION|session_start/g,
  /\$_POST|\$_REQUEST/g,
];
```

---

## Phase 3: Generate Frontend - 전체 워크플로우

**CRITICAL:** Claude Code가 직접 수행합니다.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Phase A: 프로젝트 초기화 (첫 페이지만)                                        │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  Phase B: 페이지 기본 코드 생성 (🔴 HTML + 스크린샷 필수!)                     │
│  ──────────────────────────────────────────────────────────────────────────  │
│  1. Read: sitemap.json                                                       │
│  2. Read: {screenshot} → 🔴 레이아웃, 색상, 간격 시각 분석                    │
│  3. Read: {html} → 🔴 텍스트, 이미지 URL 추출 (번역 금지!)                    │
│  4. Write: 전체 페이지 코드 (🔴 원본과 동일하게!)                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  Phase C: 개발 서버 실행                                                      │
│  ──────────────────────────────────────────────────────────────────────────  │
│  Bash: cd {output}/frontend && npm run dev &                                  │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  Phase D: AskUserQuestion                                                     │
│  ──────────────────────────────────────────────────────────────────────────  │
│  "페이지 {N} 기본 코드 완료. 다음 작업은?"                                     │
│  options: [HITL 세부 조정, 다음 페이지, 직접 입력]                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    ▼                 ▼                 ▼
            [HITL 세부 조정]    [다음 페이지]     [직접 입력]
                    │
                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  Phase E: HITL 루프 (🔴 모든 섹션 완료까지 반복!)                              │
│  ──────────────────────────────────────────────────────────────────────────  │
│                                                                               │
│   ┌─────────────────────────────────────────────────────────────────────┐    │
│   │  E-1. hitl-refine.ts 실행 (Bash 필수!)                              │    │
│   │       cd "{SCRIPTS_DIR}" && npx ts-node --transpile-only \          │    │
│   │         generate/hitl-refine.ts --capture={capture} --page={pageId} │    │
│   └─────────────────────────────────────────────────────────────────────┘    │
│                                      │                                        │
│                                      ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐    │
│   │  E-2. JSON 결과 파싱                                                │    │
│   │       <!-- HITL_RESULT_JSON_START --> ... <!-- ...END --> 사이      │    │
│   │       { overallMatch, issues[], suggestions[] } 추출                │    │
│   └─────────────────────────────────────────────────────────────────────┘    │
│                                      │                                        │
│                                      ▼                                        │
│   ┌─────────────────────────────────────────────────────────────────────┐    │
│   │  E-3. AskUserQuestion                                               │    │
│   │       "{섹션} 일치율 {N}%. 어떻게 처리할까요?"                       │    │
│   │       options: [승인, 수정 필요, 스킵]                               │    │
│   └─────────────────────────────────────────────────────────────────────┘    │
│                                      │                                        │
│              ┌───────────────────────┼───────────────────────┐               │
│              ▼                       ▼                       ▼               │
│         [승인]                 [수정 필요]                [스킵]             │
│              │                       │                       │               │
│              │               ┌───────┴───────┐               │               │
│              │               ▼               │               │               │
│              │    ┌─────────────────────┐    │               │               │
│              │    │ E-4. 코드 수정      │    │               │               │
│              │    │ (suggestions 기반)  │    │               │               │
│              │    │ Edit: 해당 파일     │    │               │               │
│              │    └─────────────────────┘    │               │               │
│              │               │               │               │               │
│              │               ▼               │               │               │
│              │    ┌─────────────────────┐    │               │               │
│              │    │ 🔄 E-1로 돌아가기   │────┘               │               │
│              │    │ (재캡처 & 재비교)   │                    │               │
│              │    └─────────────────────┘                    │               │
│              │                                               │               │
│              └───────────────────┬───────────────────────────┘               │
│                                  ▼                                            │
│   ┌─────────────────────────────────────────────────────────────────────┐    │
│   │  E-5. 다음 섹션 체크                                                │    │
│   │       IF 남은 섹션 있음 → E-1로 돌아가기                            │    │
│   │       ELSE → Phase F (페이지 완료)                                  │    │
│   └─────────────────────────────────────────────────────────────────────┘    │
│                                                                               │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  Phase F: 페이지 완료 & 다음 페이지                                           │
│  ──────────────────────────────────────────────────────────────────────────  │
│  1. sitemap.json 업데이트 (status = "completed")                             │
│  2. AskUserQuestion: "페이지 {N} 완료! 다음 페이지로 진행할까요?"             │
│     - "예" → Phase B로 (다음 pending 페이지)                                  │
│     - "아니오" → 종료                                                         │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Phase A: 프로젝트 초기화 (첫 페이지만)

**조건**: `{output}/frontend/package.json`이 없으면 실행

```bash
# Next.js 프로젝트 생성
npx create-next-app@latest {output}/frontend \
  --typescript --tailwind --eslint --app --src-dir --import-alias "@/*" --use-npm

# shadcn/ui 초기화
cd {output}/frontend && npx shadcn@latest init --defaults
cd {output}/frontend && npx shadcn@latest add button card input dialog table tabs alert badge form label textarea select

# 원본 CSS 저장 폴더 생성
mkdir -p {output}/frontend/src/styles/legacy
```

**styles 폴더 구조:**
```
src/styles/
├── legacy/              # 원본 사이트에서 가져온 CSS
│   ├── main.css         # 메인 스타일시트
│   ├── components.css   # 컴포넌트 스타일
│   └── fonts.css        # 폰트 정의
└── index.css            # 레거시 CSS import 통합
```

---

## Phase B: 페이지 기본 코드 생성 (🔴 CRITICAL!)

### Step 0: 페이지 캡처 확인 (🔴 Lazy Capture!)

> **generate 시점에 해당 페이지를 캡처합니다.** capture 단계에서는 링크만 수집했기 때문입니다.

```
Read: {capture}/sitemap.json
      ↓
IF page.captured === false:
  → 해당 페이지 HTML + 스크린샷 캡처 실행
  → sitemap.json 업데이트 (captured: true)
ELSE:
  → Step 1로 진행
```

**페이지 캡처 스크립트 실행:**
```bash
cd "{SCRIPTS_DIR}" && npx ts-node --transpile-only \
  bin/capture-page.ts --url={page.url} --output={capture} --page-id={page.id}
```

**캡처 완료 후 sitemap.json 업데이트:**
```json
{
  "id": 1,
  "captured": true,
  "screenshot": "page_1_home.png",
  "html": "page_1_home.html",
  "capturedAt": "2026-02-06T10:00:00Z"
}
```

### Step 1: sitemap.json 읽기

```
Read: {capture}/sitemap.json
```

페이지 정보 확인:
- id, url, title
- captured (true 확인)
- screenshot 파일명
- html 파일명
- status (pending → in_progress로 변경)

### Step 2: 스크린샷 읽기 (🔴 필수!)

```
Read: {capture}/{page.screenshot}
```

**시각 분석 항목:**
- 전체 레이아웃 구조 (헤더 위치, 사이드바 유무, 푸터 스타일)
- 색상 팔레트 (primary, secondary, background, text)
- 간격과 여백 (px 단위 추정)
- 폰트 크기와 굵기
- 컴포넌트 배치와 정렬

### Step 3: HTML 파일 읽기 (🔴 필수!)

```
Read: {capture}/{page.html}
```

**추출 항목:**
- 시맨틱 구조: `<header>`, `<nav>`, `<main>`, `<aside>`, `<footer>`
- CSS 클래스: `flex`, `grid`, `container`, `wrapper`, `col-*`
- **실제 텍스트 콘텐츠 (번역 없이 원본 그대로!)**
- **이미지 URL: `<img src="...">` 그대로 복사**
- **CSS 링크: `<link rel="stylesheet" href="...">` URL 추출**

### Step 3.5: 원본 CSS Fetch & 저장 (🔴 첫 페이지에서 필수!)

**첫 페이지 생성 시 원본 CSS를 가져와서 저장합니다.**

```
1. HTML에서 CSS URL 추출:
   <link rel="stylesheet" href="https://example.com/css/main.css">
   <link rel="stylesheet" href="https://example.com/css/style.css">

2. WebFetch로 각 CSS 파일 가져오기:
   WebFetch: https://example.com/css/main.css
   WebFetch: https://example.com/css/style.css

3. src/styles/legacy/ 폴더에 저장:
   Write: {output}/frontend/src/styles/legacy/main.css
   Write: {output}/frontend/src/styles/legacy/style.css

4. 통합 import 파일 생성:
   Write: {output}/frontend/src/styles/legacy-imports.css
```

**legacy-imports.css 내용:**
```css
/* 원본 사이트에서 가져온 레거시 CSS */
@import './legacy/main.css';
@import './legacy/style.css';

/* 필요시 오버라이드 */
/* .legacy-override { ... } */
```

**layout.tsx에서 import:**
```tsx
// src/app/layout.tsx
import '@/styles/legacy-imports.css';  // 🔴 레거시 CSS
import './globals.css';                 // Tailwind
```

**CSS 클래스 사용 방식:**
| 상황 | 사용 방법 |
|------|----------|
| 원본과 동일한 스타일 필요 | 레거시 CSS 클래스 그대로 사용 |
| Tailwind로 충분한 경우 | Tailwind 클래스 사용 |
| 커스텀 스타일 필요 | `globals.css`에 추가 |

**⚠️ 주의사항:**
- 첫 페이지에서만 CSS를 fetch (이후 페이지는 재사용)
- 상대 경로 CSS URL은 절대 경로로 변환 후 fetch
- CSS 내부의 `url()` 경로도 원본 도메인 기준으로 유지
- 폰트 파일 URL은 원본 그대로 사용 (또는 로컬 복사)

### Step 4: React 코드 작성 (🔴 원본과 동일하게!)

```
Write: {output}/frontend/src/app/{route}/page.tsx
```

**🔴 파일/폴더 네이밍 규칙 (HARD RULE!):**

| 대상 | 규칙 | ✅ 올바른 예시 | ❌ 잘못된 예시 |
|------|------|---------------|---------------|
| **라우트 폴더** | kebab-case | `about-us/`, `contact-form/` | `aboutUs/`, `ContactForm/` |
| **페이지 파일** | page.tsx (고정) | `about-us/page.tsx` | `AboutUs.tsx` |
| **컴포넌트 파일** | kebab-case | `header-nav.tsx`, `user-card.tsx` | `HeaderNav.tsx`, `UserCard.tsx` |
| **유틸리티 파일** | kebab-case | `date-utils.ts`, `api-client.ts` | `dateUtils.ts`, `ApiClient.ts` |

**URL → 폴더 변환 규칙:**
- `/about-us` → `app/about-us/page.tsx`
- `/products/category` → `app/products/category/page.tsx`
- `/contact_us` → `app/contact-us/page.tsx` (underscore → hyphen)
- `/AboutPage` → `app/about-page/page.tsx` (PascalCase → kebab-case)

**작성 원칙:**
- HTML 구조 → React 컴포넌트 구조로 변환
- HTML 텍스트 → 원본 그대로 JSX에 삽입 (번역 금지!)
- HTML 이미지 URL → `<img src="원본URL">` 또는 Next.js Image
- 스크린샷 색상 → Tailwind 클래스 또는 CSS 변수
- 스크린샷 레이아웃 → Tailwind flex/grid 클래스

### 🔴 섹션 식별자 규칙 (HITL 비교를 위해 필수!)

**문제:** 로컬 React 코드에 섹션 식별자가 없으면 HITL 스크립트가 섹션별 캡처를 할 수 없음!

**해결:** 모든 주요 섹션에 `data-section-id` 속성 추가

| 원본 HTML | 로컬 React | data-section-id |
|-----------|------------|-----------------|
| `<header id="main-header">` | `<header data-section-id="01-header">` | `01-header` |
| `<nav class="main-nav">` | `<nav data-section-id="02-nav">` | `02-nav` |
| `<section class="hero">` | `<section data-section-id="03-hero">` | `03-hero` |
| `<main>` | `<main data-section-id="04-main">` | `04-main` |
| `<aside>` | `<aside data-section-id="05-sidebar">` | `05-sidebar` |
| `<footer>` | `<footer data-section-id="06-footer">` | `06-footer` |

**섹션 ID 네이밍 규칙:**
```
{순번}-{섹션명}
예: 01-header, 02-nav, 03-hero, 04-features, 05-testimonials, 06-footer
```

**원본 HTML에서 섹션 추출 방법:**
1. 시맨틱 태그: `<header>`, `<nav>`, `<main>`, `<section>`, `<aside>`, `<footer>`
2. ID 속성: `id="hero"`, `id="features"` 등
3. 클래스명: `class="section-*"`, `class="block-*"` 등
4. 명확한 구분선 (큰 여백, 배경색 변화)

### 🔴 섹션 컴포넌트 분리 규칙 (필수!)

**모든 섹션은 별도 컴포넌트 파일로 분리하고, page.tsx에서 조합합니다.**

**폴더 구조:**
```
src/
├── app/
│   └── about-us/
│       └── page.tsx              # 섹션 컴포넌트 조합
│
└── components/
    └── about-us/                 # 🔴 페이지별 컴포넌트 폴더 (kebab-case!)
        ├── header-section.tsx    # 01-header
        ├── nav-section.tsx       # 02-nav
        ├── hero-section.tsx      # 03-hero
        ├── features-section.tsx  # 04-features
        └── footer-section.tsx    # 05-footer
```

**섹션 컴포넌트 예시 (`components/about-us/hero-section.tsx`):**
```tsx
// Section: 03-hero
// Generated from: https://example.com/about-us

export function HeroSection() {
  return (
    <section data-section-id="03-hero" className="...">
      {/* 🔴 원본 HTML 텍스트 그대로! */}
      <h1>About Our Company</h1>
      <p>We are a leading provider of...</p>
      <img src="https://example.com/images/hero.jpg" alt="Hero" />
    </section>
  );
}
```

**page.tsx 템플릿:**
```tsx
// Generated from: {url}
// Original title: {title}

import { HeaderSection } from '@/components/{route}/header-section';
import { NavSection } from '@/components/{route}/nav-section';
import { HeroSection } from '@/components/{route}/hero-section';
import { FeaturesSection } from '@/components/{route}/features-section';
import { FooterSection } from '@/components/{route}/footer-section';

export default function AboutUsPage() {
  return (
    <div className="...">
      {/* 🔴 섹션 컴포넌트 조합 - data-section-id는 각 컴포넌트 내부에! */}
      <HeaderSection />
      <NavSection />
      <main data-section-id="00-main" className="...">
        <HeroSection />
        <FeaturesSection />
      </main>
      <FooterSection />
    </div>
  );
}
```

**컴포넌트 파일 네이밍 규칙:**

| 섹션 ID | 컴포넌트 파일명 | export 이름 |
|---------|----------------|-------------|
| `01-header` | `header-section.tsx` | `HeaderSection` |
| `02-nav` | `nav-section.tsx` | `NavSection` |
| `03-hero` | `hero-section.tsx` | `HeroSection` |
| `04-features` | `features-section.tsx` | `FeaturesSection` |
| `05-testimonials` | `testimonials-section.tsx` | `TestimonialsSection` |
| `06-footer` | `footer-section.tsx` | `FooterSection` |

**⚠️ 주의사항:**
- 컴포넌트 폴더명은 라우트와 동일하게 kebab-case (`about-us/`, `contact-form/`)
- 컴포넌트 파일명은 kebab-case (`hero-section.tsx`)
- export 이름은 PascalCase (`HeroSection`)
- `data-section-id`는 각 섹션 컴포넌트 내부의 루트 요소에 추가!
- 공통 컴포넌트(헤더, 푸터)는 `components/common/`에 별도 관리 가능

**⚠️ HITL 주의:** `data-section-id`가 없는 섹션은 HITL 비교에서 제외됩니다!

---

## Phase C: 개발 서버 실행

```bash
cd {output}/frontend && npm run dev &
sleep 3  # 서버 시작 대기
```

---

## Phase D: AskUserQuestion

```
AskUserQuestion:
  question: "페이지 {N} 기본 코드 생성 완료. 다음 작업은?"
  header: "페이지 완료"
  options:
    - label: "HITL 세부 조정"
      description: "원본과 로컬을 섹션별로 비교하고 수정"
    - label: "다음 페이지"
      description: "현재 페이지 완료 처리, 다음 페이지로"
    - label: "직접 입력"
      description: "수정할 내용을 직접 입력"
```

---

## Phase E: HITL 루프 (🔴 핵심 워크플로우!)

**"HITL 세부 조정" 선택 시 실행**

### E-1: hitl-refine.ts 실행 (🔴 Bash 필수!)

**Claude는 반드시 이 Bash 명령을 실행해야 합니다!**

```bash
cd "{SCRIPTS_DIR}" && npx ts-node --transpile-only \
  generate/hitl-refine.ts --capture={capture} --page={pageId}
```

**출력 예시:**
```
✅ 캡처 및 비교 완료!
📊 일치율: 85%
⚠️ 발견된 차이점:
   1. 배경색 차이: 원본(#fff) vs 로컬(#f5f5f5)
   2. 폰트 크기 차이: 원본(16px) vs 로컬(14px)
💡 수정 제안:
   1. 배경색을 #fff로 변경
   2. 폰트 크기를 16px로 변경
🎯 자동 추천: 검토 필요

<!-- HITL_RESULT_JSON_START -->
{
  "sectionId": "01",
  "sectionName": "header",
  "comparison": {
    "overallMatch": 85,
    "issues": [
      "배경색 차이: 원본(#fff) vs 로컬(#f5f5f5)",
      "폰트 크기 차이: 원본(16px) vs 로컬(14px)"
    ],
    "suggestions": [
      "배경색을 #fff로 변경",
      "폰트 크기를 16px로 변경"
    ]
  },
  "claudeInstructions": {
    "recommendation": "needs_review",
    "questionOptions": ["승인", "수정 필요", "스킵"]
  }
}
<!-- HITL_RESULT_JSON_END -->
```

### E-2: JSON 결과 파싱

`<!-- HITL_RESULT_JSON_START -->` ~ `<!-- HITL_RESULT_JSON_END -->` 사이 JSON 추출:

```typescript
interface HITLResult {
  sectionId: string;
  sectionName: string;
  comparison: {
    overallMatch: number;  // 0-100
    issues: string[];
    suggestions: string[];
  };
  claudeInstructions: {
    recommendation: 'approve' | 'needs_review' | 'needs_fix';
    questionOptions: string[];
  };
}
```

### E-3: AskUserQuestion

```
AskUserQuestion:
  question: "{sectionName} 섹션 비교 결과: 일치율 {overallMatch}%. {issues[0]}"
  header: "HITL"
  options:
    - "승인" (recommendation이 "approve"면 Recommended)
    - "수정 필요"
    - "스킵"
```

### E-4: 응답별 처리

| 응답 | 처리 |
|------|------|
| **승인** | → E-5 (다음 섹션으로) |
| **수정 필요** | → suggestions 기반으로 코드 Edit → 🔄 **E-1로 돌아가기** (재캡처!) |
| **스킵** | → E-5 (다음 섹션으로) |

**"수정 필요" 선택 시 루프:**
```
E-4. 코드 수정 (Edit)
     │
     ▼
🔄 E-1로 돌아가기 (재캡처 & 재비교)
     │
     ▼
E-2. JSON 파싱
     │
     ▼
E-3. AskUserQuestion
     │
     ▼
[승인/수정 필요/스킵]
     │
     ... (승인 또는 스킵될 때까지 반복)
```

### E-5: 섹션 완료 체크

```
IF 남은 pending 섹션 있음:
  → E-1로 돌아가기 (다음 섹션 처리)
ELSE:
  → Phase F (페이지 완료)
```

---

## Phase F: 페이지 완료 & 다음 페이지

### F-1: sitemap.json 업데이트

```json
{
  "pages": [
    {
      "id": 1,
      "status": "completed",
      "completedAt": "2026-02-05T10:30:00Z"
    }
  ],
  "summary": {
    "pending": 14,
    "completed": 1
  }
}
```

### F-2: 결과 보고

```markdown
## Page {N} 완료 ✅

| 항목 | 값 |
|------|-----|
| URL | {page.url} |
| 섹션 수 | 5개 (승인: 4, 스킵: 1) |
| 생성 파일 | app/{route}/page.tsx |

## 전체 진행률
- 완료: 1/15 (6.7%)
- 대기 중: 14
```

### F-3: 다음 페이지 질문

```
AskUserQuestion:
  question: "페이지 {N} 완료! 다음 페이지로 진행할까요?"
  header: "페이지 완료"
  options:
    - label: "예"
      description: "다음 pending 페이지로 진행"
    - label: "아니오"
      description: "여기서 종료"
```

---

## Phase 3b: Generate Backend

**목적:** Java Spring Boot API 생성

```bash
/jikime:smart-rebuild generate backend --mapping=./mapping.json
```

**SQL → Java 타입 매핑:**
| SQL | Java |
|-----|------|
| BIGINT | Long |
| INT | Integer |
| VARCHAR | String |
| TEXT | String |
| DATETIME | LocalDateTime |
| DECIMAL | BigDecimal |
| BOOLEAN | Boolean |

---

## Phase 3c: Generate Connect

**목적:** Mock 데이터를 실제 API 호출로 교체

```tsx
// Before: Mock
async function getMembers() {
  return Promise.resolve(mockMembers);
}

// After: Real API
async function getMembers() {
  const res = await fetch(`${process.env.API_URL}/api/members`);
  return res.json();
}
```

---

## CLI 명령어 참조

### Capture

```bash
cd "{SCRIPTS_DIR}" && npx ts-node --transpile-only bin/smart-rebuild.ts capture {url} \
  --output={output} \
  [--login] \
  [--max-pages=100]
```

### Analyze

```bash
cd "{SCRIPTS_DIR}" && npx ts-node --transpile-only bin/smart-rebuild.ts analyze \
  --source={source} \
  --capture={capture} \
  --output={output}
```

### HITL 수동 명령어

```bash
# 섹션 캡처 & 비교
cd "{SCRIPTS_DIR}" && npx ts-node --transpile-only generate/hitl-refine.ts \
  --capture={capture} --page={pageId}

# 섹션 승인
cd "{SCRIPTS_DIR}" && npx ts-node --transpile-only generate/hitl-refine.ts \
  --capture={capture} --page={pageId} --approve={sectionId}

# 섹션 스킵
cd "{SCRIPTS_DIR}" && npx ts-node --transpile-only generate/hitl-refine.ts \
  --capture={capture} --page={pageId} --skip={sectionId}

# 상태 확인
cd "{SCRIPTS_DIR}" && npx ts-node --transpile-only generate/hitl-refine.ts \
  --capture={capture} --status

# 반응형 테스트
cd "{SCRIPTS_DIR}" && npx ts-node --transpile-only generate/hitl-refine.ts \
  --capture={capture} --page={pageId} --responsive
```

---

## Output Structure

```
{output}/
├── capture/
│   ├── sitemap.json             # 캡처 인덱스 + 페이지 상태
│   ├── *.png                    # 페이지 스크린샷
│   ├── *.html                   # 페이지 HTML
│   └── hitl/                    # HITL 캡처 결과
│       └── page_{N}/
│           └── section_{id}_{name}/
│               ├── original.png
│               └── local.png
│
├── mapping.json                 # 소스 ↔ 캡처 매핑
│
├── backend/                     # (generate backend 시)
│   └── src/main/java/
│
└── frontend/                    # Next.js 프로젝트
    └── src/
        ├── app/
        │   ├── layout.tsx
        │   ├── globals.css
        │   ├── page.tsx             # 홈페이지
        │   ├── about-us/
        │   │   └── page.tsx         # 섹션 컴포넌트 조합
        │   └── {routes}/page.tsx
        │
        ├── styles/                  # 🔴 원본 CSS 저장 폴더
        │   ├── legacy/              # 원본 사이트에서 fetch한 CSS
        │   │   ├── main.css
        │   │   └── style.css
        │   └── legacy-imports.css   # 레거시 CSS 통합 import
        │
        └── components/              # 🔴 섹션 컴포넌트 폴더
            ├── common/              # 공통 컴포넌트 (헤더, 푸터 등)
            │   ├── header-section.tsx
            │   └── footer-section.tsx
            ├── home/                # 홈페이지 섹션
            │   ├── hero-section.tsx
            │   └── features-section.tsx
            └── about-us/            # about-us 페이지 섹션
                ├── hero-section.tsx
                ├── team-section.tsx
                └── contact-section.tsx
```

---

## 이미지 비교 결과표 템플릿

```markdown
| 항목 | 원본 | 로컬 | 상태 |
|------|------|------|------|
| 레이아웃 | 2열 그리드 | 2열 그리드 | ✅ 일치 |
| 헤더 색상 | #1a365d | #1e40af | ⚠️ 유사 |
| 폰트 크기 | 16px | 14px | ❌ 다름 |
| 이미지 | 표시됨 | 깨짐 | ❌ 수정필요 |
```

---

Version: 2.0.0
