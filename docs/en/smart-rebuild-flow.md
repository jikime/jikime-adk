# Smart Rebuild Complete Flow

> **"Rebuild, not Migrate"** — Don't convert code, build it new.

## Overview

Smart Rebuild is an AI-powered workflow that **rebuilds** legacy sites (web builders, PHP, etc.) using modern technology stacks (Next.js, Spring Boot, etc.).

---

## Complete Workflow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           SMART REBUILD COMPLETE WORKFLOW                        │
│                                  Version 2.2.0                                   │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│  Phase 1: CAPTURE (Link Collection)                                             │
│  /jikime:smart-rebuild capture https://example.com                              │
├─────────────────────────────────────────────────────────────────────────────────┤
│  • Crawl site with Playwright                                                   │
│  • 🔴 Collect links only (Lazy Capture - no HTML/screenshots yet)               │
│  • Generate sitemap.json (captured: false)                                      │
│  • (Optional) --prefetch for full capture                                       │
│  • (Optional) --login for authenticated capture                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        ↓
┌─────────────────────────────────────────────────────────────────────────────────┐
│  Phase 2: ANALYZE (Analysis & Mapping)                                          │
│  /jikime:smart-rebuild analyze --source=./legacy-php --capture=./capture        │
├─────────────────────────────────────────────────────────────────────────────────┤
│  • Analyze legacy source code                                                   │
│  • URL ↔ Source file matching                                                   │
│  • Auto-classify static/dynamic pages                                           │
│  • Extract SQL queries → Identify API endpoints                                 │
│  • Generate mapping.json (source ↔ capture mapping)                             │
│  • 🔴 Generate api-mapping.json (commonApis + pageApis)                         │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        ↓
┌─────────────────────────────────────────────────────────────────────────────────┐
│  Phase 3: GENERATE (Code Generation) - Per-page iteration                       │
│  /jikime:smart-rebuild generate frontend --page 1                               │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        ↓
    ┌───────────────────────────────────────────────────────────────────────────┐
    │  Phase A: Frontend Project Initialization (🔴 Runs once on first page)    │
    ├───────────────────────────────────────────────────────────────────────────┤
    │  • Create Next.js + shadcn/ui project                                     │
    │  • Install dependencies                                                   │
    │  • Create styles/legacy/ folder                                           │
    └───────────────────────────────────────────────────────────────────────────┘
                                        ↓
    ┌───────────────────────────────────────────────────────────────────────────┐
    │  Phase B: Page Base Code Generation                                        │
    ├───────────────────────────────────────────────────────────────────────────┤
    │  Step 0: 🔴 Lazy Capture check (capture now if captured=false!)           │
    │  Step 1: Read sitemap.json                                                │
    │  Step 2: Read screenshot (visual analysis)                                │
    │  Step 2.5: 🔴 Section detection & sitemap.json update (for HITL matching!)│
    │  Step 3: Read HTML (extract text/images)                                  │
    │  Step 3.5: 🔴 Fetch original CSS (first page only)                        │
    │  Step 4: Generate section components (with data-section-id)               │
    │  Step 5: Generate page.tsx (compose section components)                   │
    └───────────────────────────────────────────────────────────────────────────┘
                                        ↓
    ┌───────────────────────────────────────────────────────────────────────────┐
    │  Phase C: Development Server Launch                                        │
    ├───────────────────────────────────────────────────────────────────────────┤
    │  • Run npm run dev                                                        │
    │  • 🔴 Accessible at localhost:3893 (default port)                         │
    └───────────────────────────────────────────────────────────────────────────┘
                                        ↓
    ┌───────────────────────────────────────────────────────────────────────────┐
    │  Phase D: AskUserQuestion - Choose Next Step                               │
    ├───────────────────────────────────────────────────────────────────────────┤
    │  "Page {N} base code complete. What's next?"                               │
    │                                                                            │
    │  options:                                                                  │
    │    ├── "HITL Fine-tuning"    → Phase E                                    │
    │    ├── "🔴 Backend Connect"  → Phase G (shown for dynamic pages only)     │
    │    ├── "Next Page"           → Phase B (next pending page)                │
    │    └── "Custom Input"        → Follow user instructions                   │
    └───────────────────────────────────────────────────────────────────────────┘
                    │
        ┌───────────┼───────────┬───────────┐
        ↓           ↓           ↓           ↓
    [HITL Adjust] [BE Connect] [Next Page] [Custom Input]
        │           │           │
        ↓           │           │
    ┌───────────────────────────────────────────────────────────────────────────┐
    │  Phase E: HITL Loop (Section-by-section comparison & modification)         │
    ├───────────────────────────────────────────────────────────────────────────┤
    │  🚨 HITL HARD RULES (Never violate!)                                       │
    │    • 🔴 Claude must NOT decide alone! Always ask the user!                 │
    │    • 🔴 No auto skip/approve even with high match rate!                    │
    │                                                                            │
    │  📍 Section Comparison Selector Rules:                                     │
    │    • Original page: Semantic selectors (header, .hero, #nav)               │
    │    • Local page: data-section-id ([data-section-id="01-header"])          │
    │                                                                            │
    │  E-1. Run hitl-refine.ts (capture & compare original vs local)            │
    │  E-2. Parse JSON result (overallMatch%, issues[], suggestions[])          │
    │  E-3. 🔴 AskUserQuestion: "Match rate {N}%. How to proceed?" (Required!)   │
    │       ├── Approve → E-5                                                   │
    │       ├── Needs fix → E-4 → 🔄 E-1 (re-capture & re-compare)              │
    │       └── Skip → E-5                                                      │
    │  E-4. Code modification (based on suggestions)                            │
    │  E-5. Check next section                                                  │
    │       ├── Remaining sections exist → Return to E-1                        │
    │       └── All sections complete → Phase F                                 │
    └───────────────────────────────────────────────────────────────────────────┘
        │                       │
        ↓                       ↓
    ┌───────────────────────────────────────────────────────────────────────────┐
    │  Phase F: Page Complete                                                    │
    ├───────────────────────────────────────────────────────────────────────────┤
    │  • Update sitemap.json (status = "completed")                             │
    │  • AskUserQuestion: "Proceed to next page?"                               │
    │       ├── Yes → Phase B (next pending page)                               │
    │       └── No → Exit                                                       │
    └───────────────────────────────────────────────────────────────────────────┘


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🔴 Phase G: Backend Integration (When dynamic page is selected)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

    ┌───────────────────────────────────────────────────────────────────────────┐
    │  🔴 Phase G-0: Backend Project Initialization (Runs once on first dynamic │
    │  page)                                                                     │
    │  /jikime:smart-rebuild backend-init --framework spring-boot                │
    ├───────────────────────────────────────────────────────────────────────────┤
    │                                                                            │
    │  G-0.1: Check backend project existence                                   │
    │         IF {output}/backend/ doesn't exist → Proceed to G-0.2             │
    │         ELSE → Skip to G-1                                                │
    │                                                                            │
    │  G-0.2: AskUserQuestion (Framework selection)                             │
    │         "Select backend framework"                                        │
    │         options:                                                           │
    │           ├── "Spring Boot (Java)" → spring-boot                           │
    │           ├── "FastAPI (Python)"   → fastapi                               │
    │           ├── "Go Fiber"           → go-fiber                              │
    │           └── "NestJS (Node.js)"   → nestjs                                │
    │                                                                            │
    │  G-0.3: Project Scaffolding                                               │
    │         • Create project with selected framework template                 │
    │         • Generate base directory structure                               │
    │                                                                            │
    │  G-0.4: Install Dependencies                                              │
    │         • Spring Boot: build.gradle dependencies                          │
    │         • FastAPI: requirements.txt / pyproject.toml                       │
    │         • Go Fiber: go.mod                                                 │
    │         • NestJS: package.json                                             │
    │                                                                            │
    │  G-0.5: DB Connection Setup                                               │
    │         • AskUserQuestion: "Enter DB type and connection info"            │
    │         • Configure application.yml / .env                                │
    │                                                                            │
    │  G-0.6: CORS + Common Settings                                            │
    │         • Allow localhost:3000                                            │
    │         • Common exception handling                                       │
    │         • Logging configuration                                           │
    │                                                                            │
    └───────────────────────────────────────────────────────────────────────────┘
                                        ↓
    ┌───────────────────────────────────────────────────────────────────────────┐
    │  Phase G-1: Common API Check                                               │
    │  /jikime:smart-rebuild generate backend --common-only                      │
    ├───────────────────────────────────────────────────────────────────────────┤
    │  • Check commonApis in api-mapping.json                                   │
    │  • IF ungenerated common APIs exist (generated: false):                   │
    │    → Generate auth APIs (login, logout, me)                               │
    │    → Generate common utilities                                            │
    │  • Update api-mapping.json (generated: true)                              │
    └───────────────────────────────────────────────────────────────────────────┘
                                        ↓
    ┌───────────────────────────────────────────────────────────────────────────┐
    │  Phase G-2: Page-specific API Generation                                   │
    │  /jikime:smart-rebuild generate backend --page 3 --skip-common             │
    ├───────────────────────────────────────────────────────────────────────────┤
    │  • Extract pageApis[{pageId}] from api-mapping.json                       │
    │  • For each API:                                                          │
    │    ├── Generate Controller                                                │
    │    ├── Generate Service                                                   │
    │    ├── Generate Repository                                                │
    │    └── Generate Entity (reference entities[])                             │
    │  • Update api-mapping.json (generated: true)                              │
    └───────────────────────────────────────────────────────────────────────────┘
                                        ↓
    ┌───────────────────────────────────────────────────────────────────────────┐
    │  Phase G-3: Frontend Connect                                               │
    │  /jikime:smart-rebuild generate connect --page 3                           │
    ├───────────────────────────────────────────────────────────────────────────┤
    │  • Set NEXT_PUBLIC_API_URL in .env.local                                  │
    │  • Create lib/api-client.ts (if not exists)                               │
    │  • Replace mock data → fetch API calls                                    │
    │  • Update api-mapping.json (connected: true)                              │
    └───────────────────────────────────────────────────────────────────────────┘
                                        ↓
    ┌───────────────────────────────────────────────────────────────────────────┐
    │  Phase G-4: Integration Test                                               │
    ├───────────────────────────────────────────────────────────────────────────┤
    │  • Start BE server:                                                       │
    │    ├── Spring Boot: ./gradlew bootRun                                      │
    │    ├── FastAPI: uvicorn main:app --reload                                  │
    │    ├── Go Fiber: go run main.go                                            │
    │    └── NestJS: npm run start:dev                                           │
    │  • Start FE server: npm run dev                                           │
    │  • AskUserQuestion: "Is the API integration working correctly?"           │
    │    ├── Working → G-5                                                      │
    │    └── Error → Debug and retry                                            │
    └───────────────────────────────────────────────────────────────────────────┘
                                        ↓
    ┌───────────────────────────────────────────────────────────────────────────┐
    │  Phase G-5: Integration Complete & Next Step                               │
    ├───────────────────────────────────────────────────────────────────────────┤
    │  • AskUserQuestion: "Integration complete! What's next?"                  │
    │    ├── "HITL Re-adjustment" → Phase E                                     │
    │    ├── "Next Page" → Phase B (next pending page)                          │
    │    └── "Custom Input"   → Follow user instructions                        │
    └───────────────────────────────────────────────────────────────────────────┘
```

---

## Phase G-0: Framework-specific Environment Configuration Details

### Supported Frameworks

| Framework | Language | CLI Command |
|-----------|----------|-------------|
| Spring Boot 3.x | Java 21 | `/jikime:smart-rebuild backend-init --framework spring-boot` |
| FastAPI | Python 3.12+ | `/jikime:smart-rebuild backend-init --framework fastapi` |
| Go Fiber | Go 1.22+ | `/jikime:smart-rebuild backend-init --framework go-fiber` |
| NestJS | Node.js 20+ | `/jikime:smart-rebuild backend-init --framework nestjs` |

### Framework Configuration Matrix

| Item | Spring Boot | FastAPI | Go Fiber | NestJS |
|------|-------------|---------|----------|--------|
| **Project Init** | Spring Initializr | `uv init` | `go mod init` | `nest new` |
| **Dependency File** | build.gradle | pyproject.toml | go.mod | package.json |
| **Config File** | application.yml | .env | config.yaml | .env |
| **DB ORM** | JPA/Hibernate | SQLAlchemy | GORM | TypeORM/Prisma |
| **Server Run** | `./gradlew bootRun` | `uvicorn main:app` | `go run main.go` | `npm run start:dev` |
| **Default Port** | 8080 | 8000 | 3001 | 3001 |

### Spring Boot Initialization Details

```bash
# G-0.3: Project creation
cd {output} && mkdir -p backend
cd {output}/backend && spring init \
  --dependencies=web,data-jpa,mysql,lombok,validation \
  --java-version=21 \
  --type=gradle-project \
  --name=api-server \
  .
```

**Directory Structure:**
```
backend/
├── build.gradle
├── settings.gradle
└── src/main/
    ├── java/com/example/api/
    │   ├── ApiApplication.java
    │   ├── config/
    │   │   ├── CorsConfig.java
    │   │   └── SecurityConfig.java
    │   ├── controller/
    │   ├── service/
    │   ├── repository/
    │   ├── entity/
    │   └── dto/
    └── resources/
        └── application.yml
```

**CorsConfig.java:**
```java
@Configuration
public class CorsConfig implements WebMvcConfigurer {
    @Override
    public void addCorsMappings(CorsRegistry registry) {
        registry.addMapping("/api/**")
            .allowedOrigins("http://localhost:3893")
            .allowedMethods("GET", "POST", "PUT", "DELETE", "OPTIONS")
            .allowedHeaders("*")
            .allowCredentials(true);
    }
}
```

### FastAPI Initialization Details

```bash
# G-0.3: Project creation
cd {output} && mkdir -p backend
cd {output}/backend && uv init
cd {output}/backend && uv add fastapi uvicorn sqlalchemy pymysql python-dotenv
```

**Directory Structure:**
```
backend/
├── pyproject.toml
├── .env
├── main.py
├── config.py
├── routers/
│   ├── __init__.py
│   └── auth.py
├── services/
├── models/
└── schemas/
```

**main.py:**
```python
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3893"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/health")
def health_check():
    return {"status": "ok"}
```

### Go Fiber Initialization Details

```bash
# G-0.3: Project creation
cd {output} && mkdir -p backend
cd {output}/backend && go mod init api-server
cd {output}/backend && go get github.com/gofiber/fiber/v2
cd {output}/backend && go get gorm.io/gorm gorm.io/driver/mysql
```

**Directory Structure:**
```
backend/
├── go.mod
├── go.sum
├── main.go
├── config/
│   └── config.go
├── handlers/
├── services/
├── models/
└── middleware/
    └── cors.go
```

**main.go:**
```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
    app := fiber.New()

    app.Use(cors.New(cors.Config{
        AllowOrigins: "http://localhost:3893",
        AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
    }))

    app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok"})
    })

    app.Listen(":3001")
}
```

### NestJS Initialization Details

```bash
# G-0.3: Project creation
cd {output} && npx @nestjs/cli new backend --package-manager npm
cd {output}/backend && npm install @nestjs/typeorm typeorm mysql2
cd {output}/backend && npm install @nestjs/config
```

**Directory Structure:**
```
backend/
├── package.json
├── .env
├── src/
│   ├── main.ts
│   ├── app.module.ts
│   ├── auth/
│   │   ├── auth.controller.ts
│   │   ├── auth.service.ts
│   │   └── auth.module.ts
│   └── common/
└── nest-cli.json
```

**main.ts:**
```typescript
import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);

  app.enableCors({
    origin: 'http://localhost:3893',
    credentials: true,
  });

  await app.listen(3001);
}
bootstrap();
```

---

## CLI Command Summary

### Phase 1: Capture

```bash
# Basic (collect links only)
/jikime:smart-rebuild capture https://example.com --output=./capture

# Full pre-capture
/jikime:smart-rebuild capture https://example.com --prefetch --output=./capture

# Login required
/jikime:smart-rebuild capture https://example.com --login --output=./capture
```

### Phase 2: Analyze

```bash
/jikime:smart-rebuild analyze --source=./legacy-php --capture=./capture
```

### Phase 3: Generate

```bash
# Frontend (per page)
/jikime:smart-rebuild generate frontend --page 1
/jikime:smart-rebuild generate frontend --next
/jikime:smart-rebuild generate frontend --status

# 🔴 Backend initialization (NEW!)
/jikime:smart-rebuild backend-init --framework spring-boot
/jikime:smart-rebuild backend-init --framework fastapi
/jikime:smart-rebuild backend-init --framework go-fiber
/jikime:smart-rebuild backend-init --framework nestjs

# Backend API generation (per page)
/jikime:smart-rebuild generate backend --common-only
/jikime:smart-rebuild generate backend --page 3 --skip-common

# Frontend-Backend connection (per page)
/jikime:smart-rebuild generate connect --page 3
```

---

## Complete Flow Summary

```
1. CAPTURE     → Link collection (Lazy Capture)
2. ANALYZE     → Analysis & mapping, generate api-mapping.json
3. GENERATE    → Per-page iteration
   ├── Phase A → FE project init (first page only)
   ├── Phase B → Page code generation
   ├── Phase C → Development server launch
   ├── Phase D → Choose next step
   │   ├── HITL adjustment → Phase E
   │   ├── BE connection  → Phase G
   │   └── Next page → Phase B
   ├── Phase E → HITL loop (section-by-section comparison)
   ├── Phase F → Page complete
   └── Phase G → Backend integration
       ├── G-0 → 🔴 BE project init (first dynamic page only)
       ├── G-1 → Common API generation
       ├── G-2 → Page-specific API generation
       ├── G-3 → FE-BE connection
       ├── G-4 → Integration test
       └── G-5 → Choose next step
```

---

## Related Documents

| Document | Path | Description |
|----------|------|-------------|
| Command Definition | `templates/.claude/commands/jikime/smart-rebuild.md` | Slash command definition |
| Execution Procedure | `templates/.claude/rules/jikime/smart-rebuild-execution.md` | Detailed execution procedure, code examples |
| Option Reference | `templates/.claude/rules/jikime/smart-rebuild-reference.md` | Options, frameworks, output structure |
| Overview Document | `docs/smart-rebuild.md` | Concept and usage summary |

---

**Created:** 2026-02-09
**Version:** 2.2.0
**Change History:**
- v2.2.0: Added HITL HARD RULES, added section ID matching system, added Step 2.5 section detection, standardized dev server port to 3893
- v2.1.0: Added Phase G-0 (backend initialization), added `backend-init` subcommand
- v2.0.0: Added Phase G (per-page progressive backend integration)
- v1.0.0: Initial version
