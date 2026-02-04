---
allowed-tools: [Read, Write, Edit, Bash, Glob, Grep, Task, WebFetch]
description: "AI-powered legacy site rebuilding - capture screenshots, analyze source, generate modern code"
argument-hint: "[capture|analyze|generate] <url> [options]"
---

# /jikime:smart-rebuild - Legacy Site Rebuilding

> **"Rebuild, not Migrate"** — 코드를 변환하지 않고, 새로 만든다.

## Purpose

레거시 사이트(웹빌더, PHP 등)를 스크린샷 + 소스 분석 기반으로 현대적 기술 스택(Next.js, Java Spring Boot)으로 **새로 구축**합니다.

## Usage

```bash
# 전체 워크플로우
/jikime:smart-rebuild https://example.com --source=./legacy-php

# Phase 1: 캡처
/jikime:smart-rebuild capture https://example.com --output=./capture

# Phase 1-1: 인증 세션 저장 (필요시)
/jikime:smart-rebuild capture --login --save-auth=auth.json

# Phase 1-2: 인증 페이지 캡처
/jikime:smart-rebuild capture https://example.com --auth=auth.json

# Phase 2: 분석 & 매핑
/jikime:smart-rebuild analyze --source=./legacy-php --capture=./capture

# Phase 3: 코드 생성
/jikime:smart-rebuild generate --mapping=./mapping.json --backend=java --frontend=nextjs
```

## Subcommands

| Subcommand | Description |
|------------|-------------|
| (none) | 전체 워크플로우 실행 (capture → analyze → generate) |
| `capture` | 사이트 크롤링 및 스크린샷 캡처 |
| `analyze` | 소스 분석 및 매핑 생성 |
| `generate` | 코드 생성 |

## Options

### capture 옵션
| Option | Description | Default |
|--------|-------------|---------|
| `<url>` | 캡처할 사이트 URL | (required) |
| `--output` | 출력 디렉토리 | `./capture` |
| `--max-pages` | 최대 캡처 페이지 수 | `100` |
| `--concurrency` | 동시 처리 수 | `5` |
| `--auth` | 인증 세션 파일 | - |
| `--login` | 수동 로그인 모드 | - |
| `--save-auth` | 세션 저장 파일 | - |
| `--exclude` | 제외 URL 패턴 | `/admin/*,/api/*` |

### analyze 옵션
| Option | Description | Default |
|--------|-------------|---------|
| `--source` | 레거시 소스 경로 | (required) |
| `--capture` | 캡처 디렉토리 | `./capture` |
| `--output` | 매핑 파일 출력 | `./mapping.json` |

### generate 옵션
| Option | Description | Default |
|--------|-------------|---------|
| `--mapping` | 매핑 파일 | (required) |
| `--backend` | 백엔드: `java`, `go`, `python` | `java` |
| `--frontend` | 프론트엔드: `nextjs`, `nuxt` | `nextjs` |
| `--output-backend` | 백엔드 출력 | `./backend` |
| `--output-frontend` | 프론트엔드 출력 | `./frontend` |

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

CRITICAL: Execute pre-built scripts from the skill folder.

**Scripts Location:**
```
.claude/skills/jikime-migration-smart-rebuild/scripts/
├── package.json
├── bin/smart-rebuild.ts      # CLI 엔트리포인트
├── capture/crawl.ts          # Playwright 크롤러
├── analyze/classify.ts       # 정적/동적 분류
└── generate/frontend.ts      # 코드 생성
```

**Step 1: Parse Arguments**
- Parse $ARGUMENTS to detect subcommand: `capture`, `analyze`, `generate`, or none (full workflow)
- Extract URL and options based on subcommand

**Step 2: Locate and Setup Scripts**
```bash
SCRIPTS_DIR=".claude/skills/jikime-migration-smart-rebuild/scripts"

# Install dependencies if needed
if [ ! -d "$SCRIPTS_DIR/node_modules" ]; then
  cd "$SCRIPTS_DIR" && npm install
fi
```

**Step 3: Execute Based on Subcommand**

**Case: No subcommand (전체 워크플로우)**
```bash
# /jikime:smart-rebuild https://example.com --source=./legacy-php
cd "$SCRIPTS_DIR" && npx ts-node bin/smart-rebuild.ts run {url} \
  --source={source} \
  --output={output}
```

**Case: capture**
```bash
# /jikime:smart-rebuild capture https://example.com --output=./capture
cd "$SCRIPTS_DIR" && npx ts-node bin/smart-rebuild.ts capture {url} \
  --output={output} \
  --max-pages={maxPages} \
  --concurrency={concurrency} \
  [--auth={auth}] \
  [--exclude={exclude}]

# /jikime:smart-rebuild capture --login --save-auth=auth.json
cd "$SCRIPTS_DIR" && npx ts-node bin/smart-rebuild.ts capture {url} \
  --login --save-auth={saveAuth}
```

**Case: analyze**
```bash
# /jikime:smart-rebuild analyze --source=./legacy-php --capture=./capture
cd "$SCRIPTS_DIR" && npx ts-node bin/smart-rebuild.ts analyze \
  --source={source} \
  --capture={capture} \
  --output={output}
```

**Case: generate**
```bash
# /jikime:smart-rebuild generate --mapping=./mapping.json --backend=java
cd "$SCRIPTS_DIR" && npx ts-node bin/smart-rebuild.ts generate \
  --mapping={mapping} \
  --backend={backend} \
  --frontend={frontend} \
  --output-backend={outputBackend} \
  --output-frontend={outputFrontend}
```

**Step 4: Report Results**
- Parse CLI output and report to user in conversation language
- Include: 캡처 페이지 수, 정적/동적 분류 결과, 생성된 파일 목록

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
