---
allowed-tools: [Read, Write, Edit, Bash, Glob, Grep, Task, WebFetch]
description: "AI-powered legacy site rebuilding - capture screenshots, analyze source, generate modern code"
argument-hint: "<url> [--source=<path>] [--phase capture|analyze|generate|all]"
---

# /jikime:smart-rebuild - Legacy Site Rebuilding

> **"Rebuild, not Migrate"** — 코드를 변환하지 않고, 새로 만든다.

## Purpose

레거시 사이트(웹빌더, PHP 등)를 스크린샷 + 소스 분석 기반으로 현대적 기술 스택(Next.js, Java Spring Boot)으로 **새로 구축**합니다.

## Usage

```bash
# 전체 워크플로우
/jikime:smart-rebuild https://example.com --source=./legacy-php

# 단계별 실행
/jikime:smart-rebuild https://example.com --phase=capture
/jikime:smart-rebuild --source=./legacy-php --phase=analyze
/jikime:smart-rebuild --mapping=./mapping.json --phase=generate
```

## Arguments

| Argument | Description | Default |
|----------|-------------|---------|
| `<url>` | 캡처할 사이트 URL | (required for capture) |
| `--source` | 레거시 소스 코드 경로 | `./source` |
| `--phase` | 실행 단계: `capture`, `analyze`, `generate`, `all` | `all` |
| `--output` | 출력 디렉토리 | `./smart-rebuild-output` |
| `--max-pages` | 최대 캡처 페이지 수 | `50` |
| `--auth` | 인증 세션 파일 (Playwright storageState) | - |
| `--backend` | 백엔드 타겟: `java`, `go`, `python` | `java` |
| `--frontend` | 프론트엔드 타겟: `nextjs`, `nuxt` | `nextjs` |

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

## Execution Workflow

### Phase 1: Capture (캡처)

**목표:** Playwright로 라이브 사이트의 모든 페이지 캡처

**실행 절차:**
1. Playwright 프로젝트 초기화 (없으면 생성)
2. 시작 URL에서 재귀적으로 내부 링크 수집
3. 각 페이지마다:
   - 전체 페이지 스크린샷 (fullPage: true)
   - 렌더링된 HTML 저장
   - 페이지 제목, H1 추출
4. `sitemap.json` 생성

**Playwright 크롤링 코드:**
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

**인증 페이지 처리:**
- `--auth` 옵션으로 Playwright storageState 파일 지정
- 수동 로그인 후 세션 저장: `context.storageState({ path: 'auth.json' })`

**출력:** `{output}/capture/sitemap.json`

---

### Phase 2: Analyze (분석 & 매핑)

**목표:** 소스 코드 분석하여 캡처와 매핑, 정적/동적 분류

**실행 절차:**
1. `sitemap.json` 로드
2. 소스 디렉토리의 모든 PHP/JSP/ASP 파일 스캔
3. URL ↔ 소스 파일 매칭:
   - 직접 경로 매칭: `/about` → `about.php`
   - index 매칭: `/products/` → `products/index.php`
   - 라우터 분석: `.htaccess`, `routes.php` 등
4. 각 소스 파일 분류:
   - **동적 판단 기준:**
     - SQL 쿼리 존재 (SELECT, INSERT, UPDATE, DELETE)
     - DB 함수 (mysqli_*, PDO, $wpdb)
     - 세션 사용 ($_SESSION, session_start)
     - POST 처리 ($_POST, $_REQUEST)
   - **정적 판단:** 위 항목 모두 없음
5. SQL 쿼리 추출 (동적 페이지)
6. `mapping.json` 생성

**분류 패턴:**
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

**출력:** `{output}/mapping.json`

---

### Phase 3: Generate (코드 생성)

**목표:** mapping.json 기반으로 현대적 코드 생성

**정적 페이지 생성:**
- 입력: 스크린샷 (UI 참고) + HTML (텍스트/이미지 추출)
- 출력: Next.js 정적 페이지 (app/[path]/page.tsx)
- Tailwind CSS 스타일링

**동적 페이지 생성:**

1. **Backend (Java Spring Boot):**
   - Entity: SQL 테이블 → JPA Entity
   - Repository: JpaRepository 인터페이스
   - Controller: REST API 엔드포인트

2. **Frontend (Next.js):**
   - Server Component로 데이터 페칭
   - API 엔드포인트 연동
   - 반응형 UI 컴포넌트

**SQL → Java 타입 매핑:**
| SQL | Java |
|-----|------|
| INT | Long |
| VARCHAR | String |
| TEXT | @Lob String |
| DATETIME | LocalDateTime |
| ENUM | enum + @Enumerated |
| DECIMAL | BigDecimal |

**출력:**
- `{output}/backend/` - Java Spring Boot 프로젝트
- `{output}/frontend/` - Next.js 프로젝트

---

## Output Structure

```
smart-rebuild-output/
├── capture/
│   ├── sitemap.json          # 캡처 결과 인덱스
│   ├── *.png                  # 페이지 스크린샷
│   └── *.html                 # 페이지 HTML
│
├── mapping.json               # 소스 ↔ 캡처 매핑
│
├── backend/
│   └── src/main/java/com/example/
│       ├── entity/            # JPA Entity
│       ├── repository/        # Repository
│       └── controller/        # REST Controller
│
└── frontend/
    ├── app/                   # Next.js App Router
    │   ├── page.tsx           # 홈
    │   ├── about/page.tsx     # 정적
    │   └── members/page.tsx   # 동적
    └── components/            # 공통 컴포넌트
```

## EXECUTION DIRECTIVE

CRITICAL: Execute these steps in sequence.

**Step 1: Parse Arguments**
- Extract URL, source path, phase, and options from $ARGUMENTS
- Default phase to "all" if not specified
- Validate required arguments for each phase

**Step 2: Setup Output Directory**
```bash
mkdir -p {output}/capture
```

**Step 3: Execute Phase(s)**

IF phase is "capture" or "all":
1. Create Playwright crawler script in scratchpad
2. Run with Bash: `cd {scratchpad} && npm install playwright && node crawl.js`
3. Copy results to {output}/capture/
4. Report: "캡처 완료: {n}개 페이지"

IF phase is "analyze" or "all":
1. Read sitemap.json from capture
2. Use Glob to find source files: `**/*.php`, `**/*.jsp`, etc.
3. For each captured URL:
   - Match to source file
   - Read source and classify (static/dynamic)
   - Extract SQL queries if dynamic
4. Generate mapping.json
5. Report: "분석 완료: 정적 {n}, 동적 {m}"

IF phase is "generate" or "all":
1. Read mapping.json
2. For each static page:
   - Read captured HTML
   - Generate Next.js page component
3. For each dynamic page:
   - Extract SQL queries
   - Generate Java Entity/Repository/Controller
   - Generate Next.js page with API calls
4. Report: "생성 완료: Backend {n} entities, Frontend {m} pages"

**Step 4: Summary Report**
- Total pages processed
- Static vs Dynamic breakdown
- Generated files list
- Next steps recommendation

## Related Skills

- `jikime-migration-smart-rebuild` - 상세 문서 및 참조 코드
- `jikime-framework-nextjs@16` - Next.js 코드 생성 패턴
- `jikime-lang-java` - Java Spring Boot 패턴

## Troubleshooting

| 문제 | 해결 |
|------|------|
| 페이지 로드 타임아웃 | `timeout` 증가, `waitUntil: 'domcontentloaded'` |
| Lazy loading 이미지 누락 | 스크롤 거리/속도 조절 |
| 인증 필요 페이지 | `--auth` 옵션으로 세션 파일 지정 |
| URL ↔ 소스 매칭 실패 | 라우터 파일 분석, 수동 매핑 추가 |
