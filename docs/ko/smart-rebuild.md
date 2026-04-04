# Smart Rebuild

> "AI로 스마트하게 레거시를 새로 구축"
>
> **"Rebuild, not Migrate"** — 코드를 변환하지 않고, 새로 만든다.

## 1. 개요

### 1.1 컨셉

Smart Rebuild는 기존 레거시 사이트(웹빌더, PHP 등)를 현대적인 기술 스택(Next.js, Java Spring Boot 등)으로 **새로 구축**하는 AI 기반 워크플로우입니다.

```
기존 마이그레이션: 소스 코드 분석 → 코드 변환 (레거시 패턴 유지)
Smart Rebuild:    스크린샷 + 소스 → AI가 새로 생성 (클린 코드)
```

### 1.2 핵심 철학

| 계층 | 전략 | 이유 |
|------|------|------|
| **UI** | 새로 만듦 | 레거시 프론트 코드 분석 가치 낮음 |
| **API** | 새로 만듦 | 소스 참고하여 클린 아키텍처로 |
| **DB** | 유지 + 점진적 개선 | 데이터 손실 위험 Zero |

### 1.3 적용 대상

- 웹빌더로 만든 사이트 (Wix, Squarespace, WordPress 등)
- 레거시 PHP 사이트
- jQuery 기반 사이트
- 기타 레거시 웹 애플리케이션

---

## 2. 2-Track 전략

페이지를 **정적/동적**으로 자동 분류하여 각각 다른 방식으로 처리합니다.

### 2.1 Track 1: 정적 콘텐츠

```
라이브 사이트 → Playwright 스크래핑 → Next.js 정적 페이지

적합한 페이지: 소개, About, FAQ, 이용약관, 공지사항
특징: DB 필요 없음, 콘텐츠만 옮기면 됨
```

### 2.2 Track 2: 동적 콘텐츠

```
소스 분석 → SQL 추출 → Backend API → Next.js 페이지

적합한 페이지: 회원 목록, 결제 내역, 게시판, 관리자
특징: DB 연동 필수, 비즈니스 로직 있음
```

### 2.3 자동 분류 기준

**동적 페이지 판단 기준:**
- SQL 쿼리 존재 (SELECT, INSERT, UPDATE, DELETE)
- DB 연결 함수 (mysqli_*, PDO, $wpdb)
- 세션 체크 ($_SESSION, session_start)
- POST 처리 ($_POST, $_REQUEST)
- 동적 파라미터 ($_GET['id'])

**정적 페이지 판단 기준:**
- 위 항목 모두 없음
- 순수 HTML + 약간의 PHP (include, require만)

---

## 3. 전체 워크플로우

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Phase 1: Capture (링크 수집) - Lazy Capture 방식                            │
├─────────────────────────────────────────────────────────────────────────────┤
│  Playwright로 사이트 크롤링                                                  │
│  ├── 🔴 링크만 수집 (HTML/스크린샷 캡처 안 함!)                              │
│  ├── sitemap.json 생성 (captured: false)                                    │
│  └── --prefetch 옵션 시에만 전체 캡처                                        │
└─────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│  Phase 2: Analyze (분석 & 매핑)                                              │
├─────────────────────────────────────────────────────────────────────────────┤
│  레거시 소스 분석                                                            │
│  ├── URL ↔ 소스 파일 매칭                                                   │
│  ├── 정적/동적 자동 분류                                                     │
│  ├── SQL 쿼리 추출 (동적인 경우)                                             │
│  ├── 🔴 API 의존성 추출 → api-mapping.json 생성                              │
│  └── mapping.json 생성                                                       │
└─────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│  Phase 3: Generate Frontend (페이지별 처리)                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│  Phase A: 프로젝트 초기화 (첫 페이지만)                                       │
│  Phase B: 페이지 기본 코드 생성                                               │
│    ├── Step 0: 🔴 Lazy Capture 체크 (captured=false면 캡처)                  │
│    ├── Step 1: sitemap.json 읽기                                             │
│    ├── Step 2: 스크린샷 읽기 (시각 분석)                                     │
│    ├── Step 3: HTML 읽기 (텍스트/이미지 추출)                                │
│    ├── Step 3.5: 🔴 원본 CSS Fetch (첫 페이지만)                             │
│    ├── Step 4: 🔴 섹션별 컴포넌트 생성 (data-section-id 포함!)               │
│    └── Step 5: page.tsx 생성 (섹션 컴포넌트 조합)                            │
│  Phase C: 개발 서버 실행                                                      │
│  Phase D: AskUserQuestion (다음 작업 선택)                                    │
│    ├── HITL 세부 조정 → Phase E                                              │
│    ├── 🔴 백엔드 연동 → Phase G (동적 페이지만)                              │
│    ├── 다음 페이지 → Phase B                                                 │
│    └── 직접 입력                                                              │
│  Phase E: HITL 루프 (섹션별 비교 & 수정)                                      │
│  Phase F: 페이지 완료                                                         │
│  🔴 Phase G: 백엔드 연동 (페이지별 점진적 연동)                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Phase 1: Capture (링크 수집)

### 4.1 Lazy Capture 방식

**기본 동작:** 링크만 수집하고, HTML + 스크린샷은 `generate --page N` 단계에서 캡처합니다.

| 옵션 | 동작 |
|------|------|
| (기본) | 링크만 수집 → `captured: false` |
| `--prefetch` | 모든 페이지 HTML + 스크린샷 캡처 → `captured: true` |

**장점:**
- 불필요한 캡처 시간 절약
- 페이지별 점진적 처리 가능
- 실제 필요한 페이지만 캡처

### 4.2 캡처 옵션

| 옵션 | 설명 | 기본값 |
|------|------|--------|
| `--merge` | 기존 sitemap.json 보존하면서 병합 (completed 페이지 유지) | - |
| `--include <patterns>` | 특정 URL 패턴만 캡처 (`--merge`와 함께 사용) | 전체 |
| `--prefetch` | 모든 페이지 HTML + 스크린샷 미리 캡처 | - |
| `--max-pages` | 최대 캡처 페이지 수 | `100` |
| `--login` | 로그인 필요 시 (브라우저 열림) | - |

### 4.3 단일 페이지 캡처 (`capture-page`)

특정 페이지 1개만 캡처하고 **sitemap.json에 자동 반영**합니다.

| 옵션 | 설명 |
|------|------|
| `<url>` | 캡처할 페이지 URL (직접 지정) |
| `--page <id>` | mapping.json의 페이지 ID로 URL 자동 조회 (예: `page_009`) |
| `--mapping <file>` | mapping.json 경로 (state에서 자동 탐색) |
| `--output <dir>` | 출력 디렉토리 (state에서 자동 탐색) |

```bash
# URL 직접 지정
smart-rebuild capture-page https://example.com/qna/list.php

# mapping.json의 page ID로 (경로 자동 탐색)
smart-rebuild capture-page --page page_009
```

**sitemap 반영 규칙:**
- 기존 URL과 일치하면 → screenshot/html/capturedAt **업데이트**
- 새 URL이면 → 새 항목 **추가** (ID 자동 부여)
- summary 카운트 **자동 재계산**

### 4.4 경로 자동 탐색 (`.smart-rebuild-state.json`)

capture + analyze 단계에서 자동 생성되는 state 파일이 경로 정보를 추적합니다.
**이후 단계에서 `--output`, `--mapping`, `--capture`, `--source` 옵션을 생략할 수 있습니다.**

```json
{
  "captureDir": "/path/to/capture",
  "sourceDir": "/path/to/source",
  "mappingFile": "/path/to/mapping.json",
  "baseUrl": "https://example.com"
}
```

**우선순위:** 사용자 입력 > state 파일 값 > 기본값

### 4.3 sitemap.json 구조

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
      "type": "static",
      "hasApi": false,
      "capturedAt": "2026-02-06T10:00:00Z"
    },
    {
      "id": 2,
      "url": "https://example.com/products",
      "title": "상품 목록",
      "captured": false,
      "screenshot": null,
      "html": null,
      "status": "pending",
      "type": "dynamic",
      "hasApi": true,
      "apis": ["/api/products"],
      "capturedAt": null
    }
  ]
}
```

---

## 5. Phase 2: Analyze (분석 & 매핑)

### 5.1 API 의존성 추출

레거시 소스에서 페이지별 필요한 API 엔드포인트를 자동으로 식별합니다.

```javascript
// PHP 파일에서 SQL 쿼리 추출
const sqlPatterns = [
  { pattern: /SELECT\s+.+\s+FROM\s+(\w+)/gi, method: 'GET' },
  { pattern: /INSERT\s+INTO\s+(\w+)/gi, method: 'POST' },
  { pattern: /UPDATE\s+(\w+)\s+SET/gi, method: 'PUT' },
  { pattern: /DELETE\s+FROM\s+(\w+)/gi, method: 'DELETE' },
];

// 테이블명 → API 엔드포인트 변환
// members → /api/members
// product_list → /api/products
```

### 5.2 api-mapping.json 구조

```json
{
  "version": "1.0",
  "createdAt": "2026-02-06T10:00:00Z",
  "sourceFramework": "php-pure",
  "targetBackend": "java",

  "commonApis": [
    {
      "path": "/api/auth/login",
      "method": "POST",
      "required": true,
      "sourceFile": "login.php",
      "generated": false,
      "connected": false
    },
    {
      "path": "/api/users/me",
      "method": "GET",
      "required": true,
      "sourceFile": "session.php",
      "generated": false,
      "connected": false
    }
  ],

  "pageApis": {
    "1": [],
    "3": [
      {
        "path": "/api/products",
        "method": "GET",
        "sourceFile": "product_list.php",
        "table": "products",
        "params": ["category", "page", "limit"],
        "generated": false,
        "connected": false
      }
    ]
  },

  "entities": [
    {
      "name": "Product",
      "table": "products",
      "fields": [
        { "name": "id", "type": "BIGINT", "javaType": "Long" },
        { "name": "name", "type": "VARCHAR(255)", "javaType": "String" },
        { "name": "price", "type": "DECIMAL(10,2)", "javaType": "BigDecimal" }
      ]
    }
  ]
}
```

**필드 설명:**

| 필드 | 설명 |
|------|------|
| `commonApis` | 모든 페이지에서 공통으로 필요한 API (인증 등) |
| `commonApis[].required` | true면 첫 동적 페이지 연동 시 반드시 생성 |
| `pageApis` | 페이지 ID별 필요한 API 목록 |
| `*.generated` | API 생성 완료 여부 |
| `*.connected` | 프론트엔드 연동 완료 여부 |

---

## 6. Phase 3: Generate Frontend

### 6.1 HARD RULES (절대 위반 금지!)

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
| 10 | **섹션 감지 → sitemap 저장** | 원본 HTML 분석 시 섹션 정보를 sitemap.json에 저장 (HITL 매칭용) |

### 6.2 개발 서버 포트

| 서버 | 포트 | 설명 |
|------|------|------|
| **Frontend (Next.js)** | `3893` | 기본 포트 (package.json에 설정됨) |
| **Backend (Spring Boot)** | `8080` | 기본 포트 |
| **Backend (FastAPI)** | `8000` | 기본 포트 |
| **Backend (Go Fiber/NestJS)** | `3001` | 기본 포트 |

### 6.3 파일/폴더 네이밍 규칙

| 대상 | 규칙 | ✅ 올바른 예시 | ❌ 잘못된 예시 |
|------|------|---------------|---------------|
| **라우트 폴더** | kebab-case | `about-us/`, `contact-form/` | `aboutUs/`, `ContactForm/` |
| **페이지 파일** | page.tsx (고정) | `about-us/page.tsx` | `AboutUs.tsx` |
| **컴포넌트 파일** | kebab-case | `header-nav.tsx`, `hero-section.tsx` | `HeaderNav.tsx` |

### 6.4 섹션 감지 & sitemap.json 업데이트

**Phase B Step 2.5에서 HTML 분석 시 섹션을 감지하고 sitemap.json에 저장:**

| 우선순위 | 원본 HTML 셀렉터 | 섹션 ID | 섹션 이름 |
|---------|-----------------|---------|----------|
| 1 | `header`, `#header`, `.header`, `[role="banner"]` | `01` | `header` |
| 2 | `nav`, `#nav`, `.gnb`, `[role="navigation"]` | `02` | `nav` |
| 3 | `.hero`, `.visual`, `.banner`, `.main-visual` | `03` | `hero` |
| 4 | `main`, `#main`, `.content`, `[role="main"]` | `04` | `main` |
| 5 | `section`, `.section` | `05+` | `section-N` |
| 6 | `aside`, `.sidebar`, `[role="complementary"]` | `..` | `sidebar` |
| 7 | `footer`, `#footer`, `[role="contentinfo"]` | `..` | `footer` |

**sitemap.json에 sections 배열 추가:**
```json
{
  "pages": [{
    "id": 1,
    "url": "https://example.com/",
    "sections": [
      { "id": "01", "name": "header", "label": "헤더", "selector": "header" },
      { "id": "02", "name": "nav", "label": "내비게이션", "selector": "#gnb" },
      { "id": "03", "name": "hero", "label": "메인 비주얼", "selector": ".hero" },
      { "id": "04", "name": "main", "label": "메인 콘텐츠", "selector": "main" },
      { "id": "05", "name": "footer", "label": "푸터", "selector": "footer" }
    ]
  }]
}
```

> **CRITICAL:** 이 섹션 정보는 HITL 비교 시 원본↔로컬 매칭에 사용됩니다!

### 6.5 섹션 컴포넌트 분리

**모든 섹션은 별도 컴포넌트 파일로 분리하고, page.tsx에서 조합합니다.**

```
src/
├── app/
│   └── about-us/
│       └── page.tsx              # 섹션 컴포넌트 조합
│
└── components/
    └── about-us/                 # 페이지별 컴포넌트 폴더 (kebab-case!)
        ├── hero-section.tsx      # data-section-id="01-hero"
        ├── team-section.tsx      # data-section-id="02-team"
        └── contact-section.tsx   # data-section-id="03-contact"
```

**섹션 컴포넌트 예시:**
```tsx
// components/about-us/hero-section.tsx
export function HeroSection() {
  return (
    <section data-section-id="01-hero" className="...">
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
// app/about-us/page.tsx
import { HeroSection } from '@/components/about-us/hero-section';
import { TeamSection } from '@/components/about-us/team-section';
import { ContactSection } from '@/components/about-us/contact-section';

export default function AboutUsPage() {
  return (
    <div>
      <HeroSection />
      <TeamSection />
      <ContactSection />
    </div>
  );
}
```

### 6.6 원본 CSS Fetch

**첫 페이지 생성 시 원본 CSS를 가져와서 저장합니다.**

```
src/styles/
├── legacy/              # 원본 사이트에서 가져온 CSS
│   ├── main.css
│   └── style.css
└── legacy-imports.css   # 레거시 CSS 통합 import
```

**layout.tsx에서 import:**
```tsx
// src/app/layout.tsx
import '@/styles/legacy-imports.css';  // 🔴 레거시 CSS
import './globals.css';                 // Tailwind
```

---

## 7. Phase E: HITL 루프 (Human-In-The-Loop)

### 7.1 HITL HARD RULES (절대 위반 금지!)

| # | 규칙 | 설명 |
|---|------|------|
| 1 | **🔴 혼자 결정 금지** | Claude는 절대 혼자서 승인/스킵 결정하면 안 됨! |
| 2 | **🔴 AskUserQuestion 필수** | 모든 섹션 비교 후 반드시 사용자에게 물어봐야 함! |
| 3 | **🔴 사용자 응답 대기** | 사용자가 선택할 때까지 다음 단계 진행 금지! |
| 4 | **🔴 자동 skip 금지** | 일치율이 높아도 사용자 확인 없이 skip 금지! |
| 5 | **🔴 자동 approve 금지** | 일치율 100%여도 사용자 확인 필수! |

> **왜 중요한가?** HITL은 Human-in-the-Loop의 약자입니다. 사람(Human)이 루프 안에 있어야 합니다!
> Claude가 혼자 결정하면 HITL이 아니라 그냥 자동화입니다.

### 7.2 섹션 비교 셀렉터 규칙

| 대상 | 셀렉터 방식 | 예시 |
|------|------------|------|
| **원본 페이지** | 시맨틱 셀렉터 | `header`, `.hero`, `#nav` |
| **로컬 페이지** | data-section-id | `[data-section-id="01-header"]` |

> **이유:** 원본과 로컬의 HTML 구조가 다를 수 있으므로, 로컬은 생성 시 추가한 `data-section-id`로 매칭합니다.

### 7.3 워크플로우

```
E-1. hitl-refine.ts 실행 (Bash)
     → 원본 사이트 캡처 + 로컬 사이트 캡처 + DOM 비교
         ↓
E-2. JSON 결과 파싱
     → overallMatch%, issues[], suggestions[] 추출
         ↓
E-3. AskUserQuestion
     "{섹션} 일치율 {N}%. 어떻게 처리할까요?"
     options: [승인, 수정 필요, 스킵]
         ↓
E-4. 응답별 처리
     승인 → E-5
     수정 필요 → 코드 Edit → E-1로 돌아가기 (재캡처!)
     스킵 → E-5
         ↓
E-5. 다음 섹션 체크
     남은 섹션 있음 → E-1로 돌아가기
     모든 섹션 완료 → Phase F
```

### 7.4 data-section-id 규칙

**HITL 비교를 위해 모든 섹션에 `data-section-id` 속성 필수!**

```
{순번}-{섹션명}
예: 01-header, 02-nav, 03-hero, 04-features, 05-footer
```

| 원본 HTML | 로컬 React |
|-----------|------------|
| `<header id="main-header">` | `<header data-section-id="01-header">` |
| `<section class="hero">` | `<section data-section-id="02-hero">` |
| `<footer>` | `<footer data-section-id="05-footer">` |

---

## 8. Phase G: 백엔드 연동 (페이지별 점진적 연동)

### 8.1 개요

**기존 방식 문제점:**
```
모든 FE 완료 → BE 일괄 생성 → 일괄 연동
→ 피드백 루프가 너무 김, 문제 발견이 늦음
```

**새로운 방식:**
```
1페이지 FE 완료 → 해당 페이지 API 생성 → 연동 → 즉시 확인
→ 빠른 피드백, 조기 문제 발견
```

### 8.2 Phase G 워크플로우

```
┌─────────────────────────────────────────────────────────────────┐
│  Phase G: 백엔드 연동 (페이지별 점진적 연동)                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  G-1. 공통 API 체크                                              │
│       IF api-mapping.json의 commonApis 중 미생성 API 있음:       │
│         → 공통 API 먼저 생성 (인증, 사용자 정보 등)               │
│                                                                  │
│  G-2. 페이지 전용 API 생성                                       │
│       - api-mapping.json에서 pageApis[{pageId}] 추출            │
│       - Spring Boot: Controller + Service + Repository 생성     │
│       - Entity 클래스 생성 (entities[] 참조)                     │
│                                                                  │
│  G-3. Frontend Connect                                           │
│       - Mock 데이터 → fetch API 호출로 교체                      │
│       - .env.local에 NEXT_PUBLIC_API_URL 설정                   │
│                                                                  │
│  G-4. 통합 테스트                                                │
│       - BE 서버 실행: ./gradlew bootRun                          │
│       - FE 서버 실행: npm run dev                                │
│       - 실제 동작 확인                                           │
│                                                                  │
│  G-5. AskUserQuestion                                            │
│       "연동 완료! 다음 작업은?"                                  │
│       options: [HITL 재조정, 다음 페이지, 직접 입력]              │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 8.3 generate backend 옵션

| Option | Description | Default |
|--------|-------------|---------|
| `--api-mapping` | API 매핑 파일 | `./api-mapping.json` |
| `--page <id>` | 특정 페이지 API만 생성 | (전체) |
| `--common-only` | 공통 API만 생성 (인증 등) | - |
| `--skip-common` | 공통 API 스킵 (이미 생성된 경우) | - |

```bash
# 공통 API 먼저 생성
/jikime:smart-rebuild generate backend --common-only

# 특정 페이지 API만 생성
/jikime:smart-rebuild generate backend --page 3 --skip-common
```

### 8.4 generate connect 옵션

| Option | Description | Default |
|--------|-------------|---------|
| `--frontend-dir` | 프론트엔드 디렉토리 | `./output/frontend` |
| `--page <id>` | 특정 페이지만 연동 | (전체) |
| `--api-url` | 백엔드 API URL | `http://localhost:8080` |

---

## 9. 출력 구조

```
{output}/
├── capture/
│   ├── sitemap.json          # 캡처 인덱스 + captured 상태
│   ├── *.png                 # 스크린샷 (캡처된 페이지만)
│   ├── *.html                # HTML (캡처된 페이지만)
│   └── hitl/                 # HITL 비교 결과
│
├── mapping.json              # 소스 ↔ 캡처 매핑
├── api-mapping.json          # 🔴 API 의존성 매핑
│
├── backend/                  # Spring Boot 프로젝트
│   └── src/main/java/com/example/api/
│       ├── controller/
│       │   ├── AuthController.java
│       │   └── ProductController.java
│       ├── service/
│       ├── repository/
│       └── entity/
│
└── frontend/                 # Next.js 프로젝트
    ├── .env.local            # API_URL 설정
    └── src/
        ├── app/
        │   ├── page.tsx
        │   └── about-us/page.tsx
        ├── lib/
        │   └── api-client.ts
        ├── styles/
        │   ├── legacy/       # 원본 CSS
        │   └── legacy-imports.css
        └── components/
            ├── common/
            └── about-us/
                ├── hero-section.tsx
                └── team-section.tsx
```

---

## 10. CLI 명령어

### 10.1 전체 프로세스

```bash
/jikime:smart-rebuild https://example.com --source=./legacy-php
```

### 10.2 단계별 실행

```bash
# Phase 1: 캡처 (링크만 수집 - 기본)
/jikime:smart-rebuild capture https://example.com --output=./capture

# Phase 1: 캡처 (전체 미리 캡처)
/jikime:smart-rebuild capture https://example.com --prefetch --output=./capture

# Phase 1: 캡처 (로그인 필요)
/jikime:smart-rebuild capture https://example.com --login --output=./capture

# Phase 1: 캡처 (기존 sitemap 유지하면서 재크롤링)
/jikime:smart-rebuild capture https://example.com --merge

# Phase 1: 캡처 (특정 URL 패턴만 선택적 캡처)
/jikime:smart-rebuild capture https://example.com --merge --include "/qna/*,/review/*"

# Phase 1: 단일 페이지 캡처 (sitemap 자동 반영)
/jikime:smart-rebuild capture-page https://example.com/qna/list.php
/jikime:smart-rebuild capture-page --page page_009    # mapping.json ID로 자동 조회

# Phase 2: 분석 & 매핑 (경로 자동 탐색)
/jikime:smart-rebuild analyze --source=./legacy-php --capture=./capture
/jikime:smart-rebuild analyze    # state에서 경로 자동 탐색

# Phase 3: 프론트엔드 생성 (페이지별, 경로 자동 탐색)
/jikime:smart-rebuild generate frontend --page 1
/jikime:smart-rebuild generate frontend --next
/jikime:smart-rebuild generate frontend --status

# Phase 3: 백엔드 생성 (페이지별)
/jikime:smart-rebuild generate backend --common-only
/jikime:smart-rebuild generate backend --page 3 --skip-common

# Phase 3: 연동 (페이지별)
/jikime:smart-rebuild generate connect --page 3
```

---

## 11. Troubleshooting

### 캡처 실패
- Playwright 브라우저 설치 확인: `npx playwright install chromium`
- 타임아웃 조정: `--timeout=60000`

### 로그인 필요 사이트
- `--login` 옵션 사용
- 브라우저에서 로그인 완료 후 Enter

### HITL 스크립트 실행 안 됨
- SCRIPTS_DIR 경로 확인
- npm install 실행 여부 확인

### CORS 오류
```
Access to fetch at 'http://localhost:8080/api/...' has been blocked by CORS policy
```
**해결:** Spring Boot의 `CorsConfig.java` 확인, `allowedOrigins`에 `http://localhost:3893` 추가

### API 연결 실패
```
Error: fetch failed / ECONNREFUSED
```
**해결:**
- 백엔드 서버 실행 여부 확인: `./gradlew bootRun`
- `.env.local`의 `NEXT_PUBLIC_API_URL` 확인

### DB 연결 오류
```
Cannot acquire connection from data source
```
**해결:** `application.yml`의 DB 설정 확인, DB 서버 실행 여부 확인

---

## 12. 기존 F.R.I.D.A.Y.와의 관계

| 항목 | F.R.I.D.A.Y. | Smart Rebuild |
|------|-------------|---------------|
| **접근 방식** | 코드 변환 | 새로 구축 |
| **UI 처리** | 코드 분석 → 변환 | 스크린샷 → 새로 생성 |
| **로직 처리** | 코드 변환 | 소스 참고 → 새로 생성 |
| **적합한 경우** | 구조화된 레거시 코드 | 빌더/스파게티 코드 |
| **결과물** | 변환된 코드 | 클린 코드 |

**두 방식은 상호 보완적이며, 상황에 따라 선택하여 사용**

---

## 13. 참고 문서

- `templates/.claude/commands/jikime/smart-rebuild.md` - 명령어 정의
- `templates/.claude/rules/jikime/smart-rebuild-execution.md` - 상세 실행 절차
- `templates/.claude/rules/jikime/smart-rebuild-reference.md` - 옵션 및 참조

---

**작성일:** 2026-02-09
**버전:** 2.3.0
**변경 이력:**
- v2.3.0: `capture-page` sitemap 자동 반영, `--page ID` mapping.json 연동, `capture --merge` 기존 sitemap 병합, `.smart-rebuild-state.json` 경로 자동 탐색
- v2.2.0: HITL HARD RULES 추가, 섹션 ID 매칭 시스템 추가, 개발 서버 포트 3893으로 표준화, sections 배열 구조 추가
- v2.0.0: Phase G (페이지별 점진적 백엔드 연동), Lazy Capture 방식 추가
- v1.0.0: 초기 버전
