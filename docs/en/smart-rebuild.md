# Smart Rebuild

> "Smartly rebuild legacy with AI"
>
> **"Rebuild, not Migrate"** — Don't convert code, create new.

## 1. Overview

### 1.1 Concept

Smart Rebuild is an AI-based workflow that **rebuilds** existing legacy sites (web builders, PHP, etc.) using modern technology stacks (Next.js, Java Spring Boot, etc.).

```
Traditional Migration: Source code analysis → Code conversion (legacy patterns retained)
Smart Rebuild:         Screenshot + Source → AI generates new (clean code)
```

### 1.2 Core Philosophy

| Layer | Strategy | Reason |
|-------|----------|--------|
| **UI** | Create new | Low value in analyzing legacy frontend code |
| **API** | Create new | Reference source for clean architecture |
| **DB** | Maintain + gradual improvement | Zero risk of data loss |

### 1.3 Target Applications

- Sites built with web builders (Wix, Squarespace, WordPress, etc.)
- Legacy PHP sites
- jQuery-based sites
- Other legacy web applications

---

## 2. 2-Track Strategy

Pages are automatically classified as **static/dynamic** and processed differently.

### 2.1 Track 1: Static Content

```
Live site → Playwright scraping → Next.js static pages

Suitable pages: Introduction, About, FAQ, Terms of Service, Announcements
Characteristics: No DB needed, just migrate content
```

### 2.2 Track 2: Dynamic Content

```
Source analysis → SQL extraction → Backend API → Next.js pages

Suitable pages: Member list, Payment history, Bulletin board, Admin
Characteristics: DB integration required, has business logic
```

### 2.3 Automatic Classification Criteria

**Dynamic page criteria:**
- SQL queries exist (SELECT, INSERT, UPDATE, DELETE)
- DB connection functions (mysqli_*, PDO, $wpdb)
- Session checks ($_SESSION, session_start)
- POST processing ($_POST, $_REQUEST)
- Dynamic parameters ($_GET['id'])

**Static page criteria:**
- None of the above items
- Pure HTML + minimal PHP (include, require only)

---

## 3. Overall Workflow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Phase 1: Capture (Link Collection) - Lazy Capture Method                   │
├─────────────────────────────────────────────────────────────────────────────┤
│  Crawling site with Playwright                                              │
│  ├── 🔴 Collect links only (No HTML/screenshot capture!)                    │
│  ├── Generate sitemap.json (captured: false)                                │
│  └── Full capture only with --prefetch option                               │
└─────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│  Phase 2: Analyze (Analysis & Mapping)                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│  Legacy source analysis                                                     │
│  ├── URL ↔ Source file matching                                             │
│  ├── Automatic static/dynamic classification                                │
│  ├── SQL query extraction (if dynamic)                                      │
│  ├── 🔴 API dependency extraction → Generate api-mapping.json               │
│  └── Generate mapping.json                                                  │
└─────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│  Phase 3: Generate Frontend (Page-by-page processing)                       │
├─────────────────────────────────────────────────────────────────────────────┤
│  Phase A: Project initialization (first page only)                          │
│  Phase B: Generate page base code                                           │
│    ├── Step 0: 🔴 Lazy Capture check (capture if captured=false)            │
│    ├── Step 1: Read sitemap.json                                            │
│    ├── Step 2: Read screenshot (visual analysis)                            │
│    ├── Step 3: Read HTML (text/image extraction)                            │
│    ├── Step 3.5: 🔴 Original CSS Fetch (first page only)                    │
│    ├── Step 4: 🔴 Generate section components (include data-section-id!)    │
│    └── Step 5: Generate page.tsx (combine section components)               │
│  Phase C: Run development server                                            │
│  Phase D: AskUserQuestion (select next action)                              │
│    ├── HITL fine-tuning → Phase E                                           │
│    ├── 🔴 Backend integration → Phase G (dynamic pages only)                │
│    ├── Next page → Phase B                                                  │
│    └── Custom input                                                         │
│  Phase E: HITL Loop (section-by-section comparison & modification)          │
│  Phase F: Page completion                                                   │
│  🔴 Phase G: Backend integration (page-by-page progressive integration)     │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Phase 1: Capture (Link Collection)

### 4.1 Lazy Capture Method

**Default behavior:** Collect links only, HTML + screenshots are captured at the `generate --page N` stage.

| Option | Behavior |
|--------|----------|
| (default) | Collect links only → `captured: false` |
| `--prefetch` | Capture all pages HTML + screenshots → `captured: true` |

**Advantages:**
- Save unnecessary capture time
- Enable page-by-page progressive processing
- Capture only pages actually needed

### 4.2 Capture Options

| Option | Description | Default |
|--------|-------------|---------|
| `--merge` | Preserve existing sitemap.json and merge new results (completed pages kept) | - |
| `--include <patterns>` | Capture only specific URL patterns (use with `--merge`) | all |
| `--prefetch` | Pre-capture all pages HTML + screenshots | - |
| `--max-pages` | Maximum number of pages to capture | `100` |
| `--login` | When login is required (browser opens) | - |

### 4.3 Single Page Capture (`capture-page`)

Capture a single page and **auto-update sitemap.json**.

| Option | Description |
|--------|-------------|
| `<url>` | Page URL to capture (direct) |
| `--page <id>` | Auto-resolve URL from mapping.json page ID (e.g. `page_009`) |
| `--mapping <file>` | mapping.json path (auto-detected from state) |
| `--output <dir>` | Output directory (auto-detected from state) |

```bash
# Direct URL
smart-rebuild capture-page https://example.com/qna/list.php

# By mapping.json page ID (paths auto-detected)
smart-rebuild capture-page --page page_009
```

**Sitemap update rules:**
- Matching URL exists → **update** screenshot/html/capturedAt
- New URL → **append** new entry (auto ID assignment)
- Summary counts **auto-recalculated**

### 4.4 Auto Path Resolution (`.smart-rebuild-state.json`)

A state file auto-generated during capture + analyze tracks all path information.
**Subsequent commands can omit `--output`, `--mapping`, `--capture`, `--source` options.**

```json
{
  "captureDir": "/path/to/capture",
  "sourceDir": "/path/to/source",
  "mappingFile": "/path/to/mapping.json",
  "baseUrl": "https://example.com"
}
```

**Priority:** User input > state file value > default

### 4.3 sitemap.json Structure

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
      "title": "Homepage",
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
      "title": "Product List",
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

## 5. Phase 2: Analyze (Analysis & Mapping)

### 5.1 API Dependency Extraction

Automatically identify required API endpoints per page from legacy source.

```javascript
// Extract SQL queries from PHP files
const sqlPatterns = [
  { pattern: /SELECT\s+.+\s+FROM\s+(\w+)/gi, method: 'GET' },
  { pattern: /INSERT\s+INTO\s+(\w+)/gi, method: 'POST' },
  { pattern: /UPDATE\s+(\w+)\s+SET/gi, method: 'PUT' },
  { pattern: /DELETE\s+FROM\s+(\w+)/gi, method: 'DELETE' },
];

// Table name → API endpoint conversion
// members → /api/members
// product_list → /api/products
```

### 5.2 api-mapping.json Structure

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

**Field descriptions:**

| Field | Description |
|-------|-------------|
| `commonApis` | APIs required commonly across all pages (authentication, etc.) |
| `commonApis[].required` | If true, must be generated when integrating first dynamic page |
| `pageApis` | List of APIs required per page ID |
| `*.generated` | Whether API generation is complete |
| `*.connected` | Whether frontend integration is complete |

---

## 6. Phase 3: Generate Frontend

### 6.1 HARD RULES (Absolutely no violations!)

| # | Rule | Description |
|---|------|-------------|
| 1 | **Mandatory screenshot analysis** | Must Read and visually analyze screenshot before writing code |
| 2 | **Copy HTML structure** | Maintain `<header>`, `<nav>`, `<main>`, `<footer>` structure as-is |
| 3 | **Preserve original text** | Use text extracted from HTML without translation |
| 4 | **Original image URLs** | Use `<img src="...">` URLs from HTML as-is |
| 5 | **Original CSS Fetch** | Fetch CSS from original site via WebFetch and save to `src/styles/` |
| 6 | **Separate section components** | Create `components/{route}/*-section.tsx` files per section |
| 7 | **Mandatory section identifier** | Add `data-section-id` attribute to all major sections (for HITL comparison) |
| 8 | **Screenshot-based styling** | Extract colors, font sizes, spacing from screenshot |
| 9 | **kebab-case naming** | Folder/file names must be kebab-case (`about-us/`, `hero-section.tsx`) |
| 10 | **Section detection → save to sitemap** | Save section info to sitemap.json when analyzing original HTML (for HITL matching) |

### 6.2 Development Server Ports

| Server | Port | Description |
|--------|------|-------------|
| **Frontend (Next.js)** | `3893` | Default port (set in package.json) |
| **Backend (Spring Boot)** | `8080` | Default port |
| **Backend (FastAPI)** | `8000` | Default port |
| **Backend (Go Fiber/NestJS)** | `3001` | Default port |

### 6.3 File/Folder Naming Rules

| Target | Rule | ✅ Correct Example | ❌ Wrong Example |
|--------|------|-------------------|-----------------|
| **Route folders** | kebab-case | `about-us/`, `contact-form/` | `aboutUs/`, `ContactForm/` |
| **Page files** | page.tsx (fixed) | `about-us/page.tsx` | `AboutUs.tsx` |
| **Component files** | kebab-case | `header-nav.tsx`, `hero-section.tsx` | `HeaderNav.tsx` |

### 6.4 Section Detection & sitemap.json Update

**Detect sections when analyzing HTML in Phase B Step 2.5 and save to sitemap.json:**

| Priority | Original HTML Selector | Section ID | Section Name |
|----------|----------------------|------------|--------------|
| 1 | `header`, `#header`, `.header`, `[role="banner"]` | `01` | `header` |
| 2 | `nav`, `#nav`, `.gnb`, `[role="navigation"]` | `02` | `nav` |
| 3 | `.hero`, `.visual`, `.banner`, `.main-visual` | `03` | `hero` |
| 4 | `main`, `#main`, `.content`, `[role="main"]` | `04` | `main` |
| 5 | `section`, `.section` | `05+` | `section-N` |
| 6 | `aside`, `.sidebar`, `[role="complementary"]` | `..` | `sidebar` |
| 7 | `footer`, `#footer`, `[role="contentinfo"]` | `..` | `footer` |

**Add sections array to sitemap.json:**
```json
{
  "pages": [{
    "id": 1,
    "url": "https://example.com/",
    "sections": [
      { "id": "01", "name": "header", "label": "Header", "selector": "header" },
      { "id": "02", "name": "nav", "label": "Navigation", "selector": "#gnb" },
      { "id": "03", "name": "hero", "label": "Main Visual", "selector": ".hero" },
      { "id": "04", "name": "main", "label": "Main Content", "selector": "main" },
      { "id": "05", "name": "footer", "label": "Footer", "selector": "footer" }
    ]
  }]
}
```

> **CRITICAL:** This section information is used for original↔local matching during HITL comparison!

### 6.5 Section Component Separation

**All sections are separated into individual component files and combined in page.tsx.**

```
src/
├── app/
│   └── about-us/
│       └── page.tsx              # Combine section components
│
└── components/
    └── about-us/                 # Page-specific component folder (kebab-case!)
        ├── hero-section.tsx      # data-section-id="01-hero"
        ├── team-section.tsx      # data-section-id="02-team"
        └── contact-section.tsx   # data-section-id="03-contact"
```

**Section component example:**
```tsx
// components/about-us/hero-section.tsx
export function HeroSection() {
  return (
    <section data-section-id="01-hero" className="...">
      {/* 🔴 Original HTML text as-is! */}
      <h1>About Our Company</h1>
      <p>We are a leading provider of...</p>
      <img src="https://example.com/images/hero.jpg" alt="Hero" />
    </section>
  );
}
```

**page.tsx template:**
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

### 6.6 Original CSS Fetch

**Fetch and save original CSS when generating the first page.**

```
src/styles/
├── legacy/              # CSS fetched from original site
│   ├── main.css
│   └── style.css
└── legacy-imports.css   # Unified legacy CSS import
```

**Import in layout.tsx:**
```tsx
// src/app/layout.tsx
import '@/styles/legacy-imports.css';  // 🔴 Legacy CSS
import './globals.css';                 // Tailwind
```

---

## 7. Phase E: HITL Loop (Human-In-The-Loop)

### 7.1 HITL HARD RULES (Absolutely no violations!)

| # | Rule | Description |
|---|------|-------------|
| 1 | **🔴 No solo decisions** | Claude must never approve/skip alone! |
| 2 | **🔴 AskUserQuestion required** | Must ask user after every section comparison! |
| 3 | **🔴 Wait for user response** | No proceeding to next step until user selects! |
| 4 | **🔴 No auto skip** | No skipping without user confirmation even with high match rate! |
| 5 | **🔴 No auto approve** | User confirmation required even at 100% match rate! |

> **Why is this important?** HITL stands for Human-in-the-Loop. A human must be in the loop!
> If Claude decides alone, it's not HITL, it's just automation.

### 7.2 Section Comparison Selector Rules

| Target | Selector Method | Example |
|--------|----------------|---------|
| **Original page** | Semantic selector | `header`, `.hero`, `#nav` |
| **Local page** | data-section-id | `[data-section-id="01-header"]` |

> **Reason:** Since HTML structure may differ between original and local, local uses `data-section-id` added during generation for matching.

### 7.3 Workflow

```
E-1. Run hitl-refine.ts (Bash)
     → Original site capture + Local site capture + DOM comparison
         ↓
E-2. Parse JSON results
     → Extract overallMatch%, issues[], suggestions[]
         ↓
E-3. AskUserQuestion
     "{Section} match rate {N}%. How should we handle this?"
     options: [Approve, Needs modification, Skip]
         ↓
E-4. Process by response
     Approve → E-5
     Needs modification → Edit code → Return to E-1 (recapture!)
     Skip → E-5
         ↓
E-5. Check next section
     Remaining sections exist → Return to E-1
     All sections complete → Phase F
```

### 7.4 data-section-id Rules

**`data-section-id` attribute required on all sections for HITL comparison!**

```
{sequence}-{section-name}
e.g.: 01-header, 02-nav, 03-hero, 04-features, 05-footer
```

| Original HTML | Local React |
|---------------|-------------|
| `<header id="main-header">` | `<header data-section-id="01-header">` |
| `<section class="hero">` | `<section data-section-id="02-hero">` |
| `<footer>` | `<footer data-section-id="05-footer">` |

---

## 8. Phase G: Backend Integration (Page-by-page Progressive Integration)

### 8.1 Overview

**Problem with existing approach:**
```
Complete all FE → Batch create BE → Batch integration
→ Feedback loop too long, late problem discovery
```

**New approach:**
```
Complete page 1 FE → Generate that page's API → Integrate → Verify immediately
→ Fast feedback, early problem discovery
```

### 8.2 Phase G Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│  Phase G: Backend Integration (Page-by-page Progressive)        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  G-1. Check common APIs                                          │
│       IF ungenerated APIs exist in api-mapping.json commonApis: │
│         → Generate common APIs first (auth, user info, etc.)    │
│                                                                  │
│  G-2. Generate page-specific APIs                                │
│       - Extract pageApis[{pageId}] from api-mapping.json        │
│       - Spring Boot: Generate Controller + Service + Repository │
│       - Generate Entity classes (reference entities[])          │
│                                                                  │
│  G-3. Frontend Connect                                           │
│       - Replace mock data → fetch API calls                      │
│       - Set NEXT_PUBLIC_API_URL in .env.local                    │
│                                                                  │
│  G-4. Integration test                                           │
│       - Run BE server: ./gradlew bootRun                         │
│       - Run FE server: npm run dev                               │
│       - Verify actual operation                                  │
│                                                                  │
│  G-5. AskUserQuestion                                            │
│       "Integration complete! What's next?"                       │
│       options: [HITL re-adjustment, Next page, Custom input]     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 8.3 generate backend Options

| Option | Description | Default |
|--------|-------------|---------|
| `--api-mapping` | API mapping file | `./api-mapping.json` |
| `--page <id>` | Generate only specific page APIs | (all) |
| `--common-only` | Generate common APIs only (auth, etc.) | - |
| `--skip-common` | Skip common APIs (if already generated) | - |

```bash
# Generate common APIs first
/jikime:smart-rebuild generate backend --common-only

# Generate specific page API only
/jikime:smart-rebuild generate backend --page 3 --skip-common
```

### 8.4 generate connect Options

| Option | Description | Default |
|--------|-------------|---------|
| `--frontend-dir` | Frontend directory | `./output/frontend` |
| `--page <id>` | Integrate specific page only | (all) |
| `--api-url` | Backend API URL | `http://localhost:8080` |

---

## 9. Output Structure

```
{output}/
├── capture/
│   ├── sitemap.json          # Capture index + captured status
│   ├── *.png                 # Screenshots (captured pages only)
│   ├── *.html                # HTML (captured pages only)
│   └── hitl/                 # HITL comparison results
│
├── mapping.json              # Source ↔ Capture mapping
├── api-mapping.json          # 🔴 API dependency mapping
│
├── backend/                  # Spring Boot project
│   └── src/main/java/com/example/api/
│       ├── controller/
│       │   ├── AuthController.java
│       │   └── ProductController.java
│       ├── service/
│       ├── repository/
│       └── entity/
│
└── frontend/                 # Next.js project
    ├── .env.local            # API_URL configuration
    └── src/
        ├── app/
        │   ├── page.tsx
        │   └── about-us/page.tsx
        ├── lib/
        │   └── api-client.ts
        ├── styles/
        │   ├── legacy/       # Original CSS
        │   └── legacy-imports.css
        └── components/
            ├── common/
            └── about-us/
                ├── hero-section.tsx
                └── team-section.tsx
```

---

## 10. CLI Commands

### 10.1 Full Process

```bash
/jikime:smart-rebuild https://example.com --source=./legacy-php
```

### 10.2 Step-by-step Execution

```bash
# Phase 1: Capture (links only - default)
/jikime:smart-rebuild capture https://example.com --output=./capture

# Phase 1: Capture (prefetch all)
/jikime:smart-rebuild capture https://example.com --prefetch --output=./capture

# Phase 1: Capture (login required)
/jikime:smart-rebuild capture https://example.com --login --output=./capture

# Phase 1: Capture (merge with existing sitemap)
/jikime:smart-rebuild capture https://example.com --merge

# Phase 1: Capture (selective URL patterns only)
/jikime:smart-rebuild capture https://example.com --merge --include "/qna/*,/review/*"

# Phase 1: Single page capture (auto-updates sitemap)
/jikime:smart-rebuild capture-page https://example.com/qna/list.php
/jikime:smart-rebuild capture-page --page page_009    # Auto-resolve from mapping.json

# Phase 2: Analyze & Mapping (paths auto-detected from state)
/jikime:smart-rebuild analyze --source=./legacy-php --capture=./capture
/jikime:smart-rebuild analyze    # Auto-detect paths from state

# Phase 3: Generate frontend (page by page, paths auto-detected)
/jikime:smart-rebuild generate frontend --page 1
/jikime:smart-rebuild generate frontend --next
/jikime:smart-rebuild generate frontend --status

# Phase 3: Generate backend (page by page)
/jikime:smart-rebuild generate backend --common-only
/jikime:smart-rebuild generate backend --page 3 --skip-common

# Phase 3: Connect (page by page)
/jikime:smart-rebuild generate connect --page 3
```

---

## 11. Troubleshooting

### Capture Failure
- Check Playwright browser installation: `npx playwright install chromium`
- Adjust timeout: `--timeout=60000`

### Sites Requiring Login
- Use `--login` option
- Complete login in browser, then press Enter

### HITL Script Not Running
- Check SCRIPTS_DIR path
- Check if npm install was run

### CORS Error
```
Access to fetch at 'http://localhost:8080/api/...' has been blocked by CORS policy
```
**Solution:** Check Spring Boot's `CorsConfig.java`, add `http://localhost:3893` to `allowedOrigins`

### API Connection Failure
```
Error: fetch failed / ECONNREFUSED
```
**Solution:**
- Check if backend server is running: `./gradlew bootRun`
- Check `NEXT_PUBLIC_API_URL` in `.env.local`

### DB Connection Error
```
Cannot acquire connection from data source
```
**Solution:** Check DB settings in `application.yml`, verify DB server is running

---

## 12. Relationship with Existing F.R.I.D.A.Y.

| Item | F.R.I.D.A.Y. | Smart Rebuild |
|------|-------------|---------------|
| **Approach** | Code conversion | Rebuild |
| **UI Processing** | Code analysis → Conversion | Screenshot → New generation |
| **Logic Processing** | Code conversion | Reference source → New generation |
| **Suitable for** | Structured legacy code | Builder/spaghetti code |
| **Output** | Converted code | Clean code |

**Both approaches are complementary and should be selected based on the situation**

---

## 13. Reference Documents

- `templates/.claude/commands/jikime/smart-rebuild.md` - Command definitions
- `templates/.claude/rules/jikime/smart-rebuild-execution.md` - Detailed execution procedures
- `templates/.claude/rules/jikime/smart-rebuild-reference.md` - Options and references

---

**Created:** 2026-02-09
**Version:** 2.3.0
**Change History:**
- v2.3.0: `capture-page` auto sitemap sync, `--page ID` mapping.json lookup, `capture --merge` existing sitemap preservation, `.smart-rebuild-state.json` auto path resolution
- v2.2.0: Added HITL HARD RULES, Added section ID matching system, Standardized dev server port to 3893, Added sections array structure
- v2.0.0: Phase G (page-by-page progressive backend integration), Added Lazy Capture method
- v1.0.0: Initial version
