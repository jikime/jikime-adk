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

## 3. 워크플로우

```
┌─────────────────────────────────────────────────────────────────┐
│  Phase 1: Capture (캡처)                                         │
├─────────────────────────────────────────────────────────────────┤
│  Playwright로 사이트 크롤링                                      │
│  ├── 모든 페이지 URL 수집 (재귀적)                               │
│  ├── 각 페이지 스크린샷 (fullPage)                               │
│  ├── HTML 저장                                                   │
│  └── sitemap.json 생성                                           │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│  Phase 2: Analyze (분석 & 매핑)                                  │
├─────────────────────────────────────────────────────────────────┤
│  레거시 소스 분석                                                │
│  ├── URL ↔ 소스 파일 매칭                                       │
│  ├── 정적/동적 자동 분류                                         │
│  ├── SQL 쿼리 추출 (동적인 경우)                                 │
│  ├── DB 스키마 분석                                              │
│  └── mapping.json 생성                                           │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│  Phase 3: Generate (코드 생성)                                   │
├─────────────────────────────────────────────────────────────────┤
│  정적 페이지:                                                    │
│  └── 스크린샷 + HTML → Next.js 정적 페이지                       │
│                                                                  │
│  동적 페이지:                                                    │
│  ├── SQL → Java Entity/Repository/Controller                    │
│  └── 스크린샷 + HTML → Next.js 페이지 (API 연동)                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. Phase 1: Capture (캡처)

### 4.1 Playwright 크롤링

```typescript
const { chromium } = require('playwright');

async function crawlAndCapture(startUrl: string) {
  const browser = await chromium.launch();
  const baseUrl = new URL(startUrl).origin;

  const visited = new Set<string>();
  const toVisit: string[] = [startUrl];
  const results = [];

  while (toVisit.length > 0) {
    const batch = toVisit.splice(0, 5); // 동시 5개 처리

    const promises = batch.map(async (url) => {
      if (visited.has(url)) return null;
      visited.add(url);
      return await capturePage(browser, url, baseUrl);
    });

    const batchResults = await Promise.all(promises);

    for (const result of batchResults) {
      if (!result) continue;
      results.push(result);

      // 새로운 링크 추가
      for (const link of result.links) {
        if (!visited.has(link) && !toVisit.includes(link)) {
          toVisit.push(link);
        }
      }
    }
  }

  await browser.close();
  return results;
}
```

### 4.2 페이지 캡처

```typescript
async function capturePage(browser, url, baseUrl) {
  const page = await browser.newPage();

  await page.goto(url, { waitUntil: 'networkidle' });

  // Lazy loading 해결: 전체 스크롤
  await autoScroll(page);

  // 전체 페이지 스크린샷
  await page.screenshot({
    path: `./output/${filename}.png`,
    fullPage: true
  });

  // HTML 저장
  const html = await page.content();

  // 내부 링크 수집
  const links = await page.$$eval('a[href]', (anchors, base) => {
    return anchors
      .map(a => a.href)
      .filter(href => href.startsWith(base));
  }, baseUrl);

  return { url, screenshot, html, links };
}
```

### 4.3 Lazy Loading 처리

```typescript
async function autoScroll(page) {
  await page.evaluate(async () => {
    await new Promise((resolve) => {
      let totalHeight = 0;
      const distance = 500;
      const maxHeight = 50000;

      const timer = setInterval(() => {
        window.scrollBy(0, distance);
        totalHeight += distance;

        if (totalHeight >= document.body.scrollHeight || totalHeight >= maxHeight) {
          clearInterval(timer);
          window.scrollTo(0, 0);
          resolve();
        }
      }, 100);
    });
  });
}
```

### 4.4 인증 페이지 처리

`--login` 옵션을 사용하면 로그인과 캡처가 한 번에 진행됩니다.

```bash
# 인증 필요 시: 로그인 → 캡처 한 번에 진행
/jikime:smart-rebuild capture https://example.com --login --output=./capture
```

**동작 방식:**
1. 브라우저가 열림 (headless: false)
2. 사용자가 직접 로그인 수행
3. 터미널에서 **Enter 입력** → 세션 자동 저장
4. headless 모드로 전환하여 캡처 진행

```typescript
// --login 옵션 처리 내부 로직
async function crawlAndCapture(url: string, options: CaptureOptions) {
  if (options.login) {
    // 1. 브라우저 열고 로그인 페이지 이동
    const browser = await chromium.launch({ headless: false });
    const page = await context.newPage();
    await page.goto(url);

    // 2. 사용자 로그인 대기
    await waitForUserInput('로그인 완료 후 Enter를 누르세요...');

    // 3. 세션 저장
    await context.storageState({ path: `${outputDir}/auth.json` });

    // 4. headless 모드로 재시작하여 캡처 진행
    await browser.close();
    browser = await chromium.launch({ headless: true });
    context = await browser.newContext({ storageState: sessionFile });
  }

  // 캡처 진행...
}
```

**세션 재사용 (반복 캡처 시):**
```bash
# 이전에 저장된 세션 파일 사용
/jikime:smart-rebuild capture https://example.com --auth=./capture/auth.json
```

### 4.5 출력: sitemap.json

```json
{
  "baseUrl": "https://example.com",
  "capturedAt": "2026-02-04T10:00:00Z",
  "totalPages": 47,
  "pages": [
    {
      "url": "https://example.com/about",
      "screenshot": "about.png",
      "html": "about.html",
      "title": "회사 소개",
      "links": ["/", "/contact", "/products"]
    }
  ]
}
```

---

## 5. Phase 2: Analyze (분석 & 매핑)

### 5.1 소스 분석 알고리즘

```typescript
interface PageAnalysis {
  path: string;
  type: 'static' | 'dynamic';
  reason: string[];
  dbQueries: string[];
}

function classifyPage(phpFile: string): PageAnalysis {
  const content = readFile(phpFile);
  const reasons = [];
  const dbQueries = [];

  // 1. SQL 쿼리 체크
  const sqlPatterns = [
    /SELECT\s+.+\s+FROM/gi,
    /INSERT\s+INTO/gi,
    /UPDATE\s+.+\s+SET/gi,
    /DELETE\s+FROM/gi,
  ];

  for (const pattern of sqlPatterns) {
    const matches = content.match(pattern);
    if (matches) {
      dbQueries.push(...matches);
      reasons.push('SQL 쿼리 발견');
    }
  }

  // 2. DB 연결 함수 체크
  if (/mysqli_query|\$pdo->query|\$wpdb->/g.test(content)) {
    reasons.push('DB 연결 함수');
  }

  // 3. 세션 체크
  if (/\$_SESSION|session_start/g.test(content)) {
    reasons.push('세션 사용');
  }

  // 4. POST 처리 체크
  if (/\$_POST|\$_REQUEST/g.test(content)) {
    reasons.push('POST 데이터 처리');
  }

  return {
    path: phpFile,
    type: reasons.length > 0 ? 'dynamic' : 'static',
    reason: reasons,
    dbQueries,
  };
}
```

### 5.2 출력: mapping.json

```json
{
  "project": {
    "name": "example-migration",
    "sourceUrl": "https://example.com",
    "sourcePath": "./legacy-php"
  },

  "summary": {
    "totalPages": 47,
    "static": 12,
    "dynamic": 35
  },

  "pages": [
    {
      "id": "page_001",

      "capture": {
        "url": "https://example.com/about",
        "screenshot": "captures/about.png",
        "html": "captures/about.html"
      },

      "source": {
        "file": "about.php",
        "type": "static",
        "reason": []
      },

      "output": {
        "frontend": {
          "path": "/app/about/page.tsx",
          "type": "static-page"
        }
      }
    },

    {
      "id": "page_002",

      "capture": {
        "url": "https://example.com/members",
        "screenshot": "captures/members.png",
        "html": "captures/members.html"
      },

      "source": {
        "file": "members/list.php",
        "type": "dynamic",
        "reason": ["SQL 쿼리 발견", "세션 사용"]
      },

      "database": {
        "queries": [
          {
            "raw": "SELECT * FROM members WHERE status = 'active'",
            "table": "members",
            "type": "SELECT"
          }
        ]
      },

      "output": {
        "backend": {
          "entity": "Member.java",
          "repository": "MemberRepository.java",
          "controller": "MemberController.java",
          "endpoint": "GET /api/members"
        },
        "frontend": {
          "path": "/app/members/page.tsx",
          "type": "dynamic-page",
          "apiCalls": ["GET /api/members"]
        }
      }
    }
  ],

  "database": {
    "tables": [
      {
        "name": "members",
        "columns": [
          {"name": "id", "type": "INT", "primary": true},
          {"name": "email", "type": "VARCHAR(255)"},
          {"name": "name", "type": "VARCHAR(100)"},
          {"name": "status", "type": "ENUM('active','inactive')"}
        ]
      }
    ]
  }
}
```

---

## 6. Phase 3: Generate (코드 생성)

### 6.1 정적 페이지 생성

**입력:**
- 스크린샷 (UI 디자인 참고)
- HTML (텍스트, 이미지 추출)

**출력:**
```tsx
// app/about/page.tsx
export default function AboutPage() {
  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">회사 소개</h1>
      <div className="prose max-w-none">
        {/* HTML에서 추출한 콘텐츠 */}
      </div>
    </div>
  );
}
```

### 6.2 동적 페이지 생성

**Backend (Java Spring Boot):**

```java
// Member.java
@Entity
@Table(name = "members")
public class Member {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    private String email;
    private String name;
    private String status;
}

// MemberRepository.java
public interface MemberRepository extends JpaRepository<Member, Long> {
    List<Member> findByStatus(String status);
}

// MemberController.java
@RestController
@RequestMapping("/api/members")
public class MemberController {

    @Autowired
    private MemberRepository memberRepository;

    @GetMapping
    public List<Member> getActiveMembers() {
        return memberRepository.findByStatus("active");
    }
}
```

**Frontend (Next.js):**

```tsx
// app/members/page.tsx
async function MembersPage() {
  const members = await fetch('http://api.example.com/api/members')
    .then(res => res.json());

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">회원 목록</h1>
      <div className="grid grid-cols-3 gap-4">
        {members.map(member => (
          <MemberCard key={member.id} member={member} />
        ))}
      </div>
    </div>
  );
}
```

---

## 7. 스킬 구조

```
skills/jikime-migration-smart-rebuild/
├── SKILL.md                    # 스킬 정의
├── rules/
│   ├── overview.md             # 전체 워크플로우 가이드
│   ├── phase-1-capture.md      # 캡처 단계 상세
│   ├── phase-2-analyze.md      # 분석 단계 상세
│   ├── phase-3-generate.md     # 생성 단계 상세
│   └── troubleshooting.md      # 문제 해결
│
└── scripts/                    # CLI 도구
    ├── package.json
    ├── bin/
    │   └── smart-rebuild.ts    # CLI 엔트리포인트
    ├── capture/
    │   ├── crawl.ts            # 사이트 크롤링
    │   ├── auth.ts             # 인증 처리
    │   └── screenshot.ts       # 스크린샷 캡처
    ├── analyze/
    │   ├── classify.ts         # 정적/동적 분류
    │   ├── match.ts            # 소스 ↔ 캡처 매칭
    │   └── extract-sql.ts      # SQL 쿼리 추출
    └── generate/
        ├── frontend.ts         # Next.js 코드 생성
        └── backend.ts          # Java API 코드 생성
```

---

## 8. CLI 명령어

### 8.1 전체 프로세스

```bash
/jikime:smart-rebuild https://example.com --source=./legacy-php
```

### 8.2 단계별 실행

```bash
# Phase 1: 캡처 (인증 불필요)
/jikime:smart-rebuild capture https://example.com --output=./capture

# Phase 1: 캡처 (인증 필요 - 로그인 후 캡처 진행)
/jikime:smart-rebuild capture https://example.com --login --output=./capture

# Phase 2: 분석 & 매핑
/jikime:smart-rebuild analyze --source=./legacy-php --capture=./capture

# Phase 3: 코드 생성
/jikime:smart-rebuild generate --mapping=./mapping.json --backend=java --frontend=nextjs
```

### 8.3 옵션

**capture 옵션:**
| 옵션 | 설명 | 기본값 |
|------|------|--------|
| `--output` | 출력 디렉토리 | `./capture` |
| `--max-pages` | 최대 캡처 페이지 수 | `100` |
| `--concurrency` | 동시 처리 수 | `5` |
| `--login` | 로그인 필요 시 (브라우저 열림 → 로그인 → 캡처) | - |
| `--auth` | 기존 세션 파일 재사용 | - |
| `--exclude` | 제외할 URL 패턴 | `/admin/*,/api/*` |

**analyze 옵션:**
| 옵션 | 설명 | 기본값 |
|------|------|--------|
| `--source` | 레거시 소스 경로 | `./source` |
| `--capture` | 캡처 디렉토리 | `./capture` |
| `--output` | 매핑 파일 출력 | `./mapping.json` |
| `--db-schema` | DB 스키마 파일 (prisma, sql, json) | - |
| `--db-from-env` | .env의 DATABASE_URL에서 스키마 추출 | - |

**generate 옵션:**
| 옵션 | 설명 | 기본값 |
|------|------|--------|
| `--mapping` | 매핑 파일 | `./mapping.json` |
| `--backend` | 백엔드 타겟 | `java` |
| `--frontend` | 프론트엔드 타겟 | `nextjs` |
| `--output-backend` | 백엔드 출력 디렉토리 | `./backend` |
| `--output-frontend` | 프론트엔드 출력 디렉토리 | `./frontend` |

---

## 9. 기존 F.R.I.D.A.Y.와의 관계

| 항목 | F.R.I.D.A.Y. | Smart Rebuild |
|------|-------------|---------------|
| **접근 방식** | 코드 변환 | 새로 구축 |
| **UI 처리** | 코드 분석 → 변환 | 스크린샷 → 새로 생성 |
| **로직 처리** | 코드 변환 | 소스 참고 → 새로 생성 |
| **적합한 경우** | 구조화된 레거시 코드 | 빌더/스파게티 코드 |
| **결과물** | 변환된 코드 | 클린 코드 |

**두 방식은 상호 보완적이며, 상황에 따라 선택하여 사용**

---

## 10. 향후 확장

### 10.1 지원 소스 확장
- PHP (완료)
- ASP.NET
- JSP
- Ruby on Rails

### 10.2 지원 타겟 확장
- Backend: Java, Node.js, Go, Python
- Frontend: Next.js, Nuxt.js, SvelteKit

### 10.3 AI 기능 강화
- 스크린샷에서 디자인 토큰 자동 추출
- 컴포넌트 자동 분류 및 생성
- 비즈니스 로직 자동 추론

---

## 11. 참고

### 11.1 테스트 결과

**테스트 사이트:** https://wvctesol.com

```
✅ 크롤링 완료: 22개 페이지 (약 1분)
✅ 전체 페이지 스크린샷 캡처 성공
✅ HTML 저장 성공
✅ sitemap.json 생성 성공
✅ 88개+ 내부 링크 자동 발견
```

### 11.2 관련 문서

- F.R.I.D.A.Y. 마이그레이션 오케스트레이터
- JikiME-ADK 스킬 개발 가이드
- Playwright 공식 문서

---

**작성일:** 2026-02-04
**버전:** 1.1.0
