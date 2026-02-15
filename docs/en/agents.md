# JikiME-ADK Agents Reference

A catalog of specialized agents in JikiME-ADK.

---

## Overview

JikiME-ADK provides 57 specialized agents:
- **Manager Agents (12)**: Workflow coordination and process management
- **Specialist Agents (37)**: Domain-specific specialized tasks
- **Designer Agents (1)**: UI/UX design and design systems
- **Orchestration Agents (3)**: Multi-agent coordination and task distribution
- **Builder Agents (4)**: Creation of new agents/commands/skills/plugins

### Agent Map

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      JikiME-ADK Agent Catalog (57)                       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─ Manager Agents (Workflow Coordination) ─────────────────────────────┐  │
│  │                                                                      │
│  │  manager-spec        SPEC document generation (EARS format)          │
│  │  manager-strategy    Implementation strategy planning                │
│  │  manager-ddd         DDD implementation (ANALYZE-PRESERVE-IMPROVE)   │
│  │  manager-project     Project initialization and configuration        │
│  │  manager-docs        Document synchronization                        │
│  │  manager-quality     Quality verification (TRUST 5)                  │
│  │  manager-git         Git workflow                                    │
│  │  manager-claude-code Claude Code configuration management            │
│  │  manager-database    DB schema design, query optimization            │
│  │  manager-dependency  Package updates, vulnerability management       │
│  │  manager-data        Data pipeline, ETL                              │
│  │  manager-context     Context/token management                        │
│  │                                                                      │
│  └──────────────────────────────────────────────────────────────────────┘
│                                                                          │
│  ┌─ Specialist Agents (Domain Experts) ──────────────────────────────────┐│
│  │                                                                      │
│  │  [Core]                                                              │
│  │  architect          System architecture design                       │
│  │  backend            API development, server logic                    │
│  │  frontend           React components, UI implementation              │
│  │  security-auditor   Security audit (OWASP)                           │
│  │  devops             CI/CD, infrastructure, deployment automation     │
│  │  optimizer          Performance optimization, bottleneck analysis    │
│  │  debugger           Debugging, error analysis                        │
│  │                                                                      │
│  │  [Testing]                                                           │
│  │  e2e-tester         E2E testing (Playwright)                         │
│  │  test-guide         Test strategy and guidance                       │
│  │                                                                      │
│  │  [Code Quality]                                                      │
│  │  refactorer         Refactoring/cleanup                              │
│  │  build-fixer        Build/type error fixing                          │
│  │  reviewer           Code review, PR review                           │
│  │  documenter         API/code documentation                           │
│  │  planner            Implementation planning                          │
│  │                                                                      │
│  │  [Language/Framework Specialists]                                    │
│  │  migrator           Legacy modernization, framework migration        │
│  │  specialist-angular Angular 15+, NgRx, RxJS, micro frontends         │
│  │  specialist-api     REST/GraphQL API design, OpenAPI                 │
│  │  specialist-java    Java 21+, Spring Boot, JPA                       │
│  │  specialist-javascript ES2023+, Node.js 20+, async patterns          │
│  │  specialist-spring  Spring Security, Data, Cloud                     │
│  │  specialist-nextjs  Next.js App Router, RSC, Server Actions          │
│  │  specialist-go      Go, Fiber/Gin, GORM                              │
│  │  specialist-php     PHP 8.3+, Laravel, Symfony                       │
│  │  specialist-postgres PostgreSQL, pgvector, RLS, JSONB                │
│  │  specialist-python  Python 3.11+, FastAPI, Django                    │
│  │  specialist-rust    Rust 2021, memory safety, system programming     │
│  │  specialist-sql     PostgreSQL, MySQL, SQL Server, Oracle            │
│  │  specialist-typescript TypeScript 5.0+, advanced types, e2e type safety │
│  │  specialist-vue     Vue 3, Composition API, Nuxt 3, Pinia            │
│  │  specialist-graphql GraphQL, Apollo Federation, subscriptions        │
│  │  specialist-microservices Microservices, Kubernetes, service mesh    │
│  │  specialist-mobile  React Native, Flutter, mobile apps               │
│  │  specialist-electron Electron desktop apps, cross-platform           │
│  │  specialist-websocket WebSocket, Socket.IO, real-time communication  │
│  │  fullstack          Full-stack development, DB → API → UI            │
│  │                                                                      │
│  │  [Research]                                                          │
│  │  analyst            Technical research, competitive analysis         │
│  │  explorer           Codebase exploration, search                     │
│  │                                                                      │
│  └──────────────────────────────────────────────────────────────────────┘
│                                                                          │
│  ┌─ Designer Agents (UI/UX Experts) ─────────────────────────────────────┐│
│  │                                                                      │
│  │  designer-ui        UI design systems, component libraries, a11y    │
│  │                                                                      │
│  └──────────────────────────────────────────────────────────────────────┘
│                                                                          │
│  ┌─ Orchestration Agents (Multi-Agent Coordination) ─────────────────────┐ │
│  │                                                                      │
│  │  orchestrator       Workflow orchestration, pipeline coordination    │
│  │  coordinator        Multi-agent coordination, result integration     │
│  │  dispatcher         Task distribution, load balancing, priority scheduling │
│  │                                                                      │
│  └──────────────────────────────────────────────────────────────────────┘
│                                                                          │
│  ┌─ Builder Agents (Creation Tools) ──────────────────────────────────┐   │
│  │                                                                      │
│  │  agent-builder      New agent definition creation                    │
│  │  command-builder    New slash command creation                       │
│  │  skill-builder      New skill definition creation                    │
│  │  plugin-builder     New plugin package creation                      │
│  │                                                                      │
│  └──────────────────────────────────────────────────────────────────────┘
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Manager Agents

### manager-spec

**Role**: SPEC Document Generation Expert

| Property | Value |
|----------|-------|
| Model | inherit |
| Tools | Read, Write, Edit, MultiEdit, Bash, Glob, Grep, TodoWrite, WebFetch, Context7 |
| Skills | jikime-foundation-claude, jikime-foundation-core, jikime-workflow-spec |

**Core Functions**:
- EARS format requirements document generation
- 3-file SPEC directory structure (`spec.md`, `plan.md`, `acceptance.md`)
- Given-When-Then acceptance criteria writing
- Domain-specific expert delegation recommendations

**SPEC ID Format**: `SPEC-{DOMAIN}-{NUMBER}` (e.g., SPEC-AUTH-001)

**When to Use**:
- When defining new feature requirements
- When executing `/jikime:1-plan` command

---

### manager-strategy

**Role**: Implementation Strategy Planning Expert

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Write, Edit, Grep, Glob, Bash, WebFetch, WebSearch, TodoWrite, Task, Skill, Context7 |
| Skills | jikime-foundation-claude, jikime-foundation-core, jikime-workflow-spec, jikime-workflow-project |

**Core Functions**:
- SPEC analysis and interpretation
- Library version selection (using Context7)
- Technical decisions and trade-off analysis
- Task decomposition

**Strategic Thinking Framework**:
1. **Phase 0**: Assumption audit (Hard vs Soft constraint classification)
2. **Phase 0.5**: First Principles decomposition (Five Whys)
3. **Phase 0.75**: Alternative generation (Conservative/Balanced/Aggressive)

**When to Use**:
- After SPEC analysis for implementation strategy planning
- When executing `/jikime:2-run` command

---

### manager-ddd

**Role**: DDD (Domain-Driven Development) Implementation Expert

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep, Glob, TodoWrite, Task, Skill, Context7 |
| Skills | jikime-foundation-claude, jikime-foundation-core, jikime-workflow-ddd, jikime-tool-ast-grep, jikime-workflow-testing |

**Core Functions**:
- ANALYZE-PRESERVE-IMPROVE DDD cycle execution
- Characterization test generation
- Behavior-preserving refactoring
- AST-grep based code analysis

**DDD Cycle**:

| Phase | Purpose | Key Activities |
|-------|---------|----------------|
| ANALYZE | Understand current state | Domain boundary identification, coupling/cohesion analysis |
| PRESERVE | Build safety net | Existing test verification, characterization test generation |
| IMPROVE | Incremental improvement | Atomic transformations, immediate test verification |

**When to Use**:
- When refactoring existing code
- When code improvement requires behavior preservation

---

### manager-project

**Role**: Project Initialization Expert

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Write, Edit, Grep, Glob, Bash, TodoWrite, Task, Skill, AskUserQuestion, Context7 |
| Skills | jikime-foundation-claude, jikime-foundation-core, jikime-workflow-project |

**Core Functions**:
- Project mode detection (New/Existing/Migration)
- User preference collection (AskUserQuestion)
- JikiME configuration structure creation
- Tech stack detection and documentation

**Generated Files**:
```
.jikime/
├── config/
│   ├── language.yaml      # Language settings
│   ├── user.yaml          # User settings
│   └── quality.yaml       # Quality settings
├── project/
│   ├── product.md         # Product information
│   ├── structure.md       # Project structure
│   └── tech.md            # Tech stack
└── specs/                 # SPEC documents
```

**When to Use**:
- When initializing a new project
- When executing `/jikime:0-project` command

---

### manager-docs

**Role**: Document Synchronization Expert

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep, Glob, TodoWrite |
| Skills | jikime-foundation-claude, jikime-foundation-core |

**Core Functions**:
- Code change analysis and document synchronization
- README, CODEMAP generation/update
- SPEC status synchronization
- API documentation

**Document Types**:

| Type | Location | Purpose |
|------|----------|---------|
| README.md | Project root | Overview, getting started guide |
| CODEMAPS/ | docs/ | Architecture overview, module structure |
| SPEC Status | .jikime/specs/ | Implementation status tracking |

**When to Use**:
- When updating documents after code changes
- When executing `/jikime:3-sync` command

---

### manager-quality

**Role**: Quality Verification Expert

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep, Glob, TodoWrite, Task, Skill, Context7 |
| Permission | bypassPermissions |
| Skills | jikime-foundation-claude, jikime-foundation-core, jikime-workflow-testing, jikime-tool-ast-grep |

**Core Functions**:
- TRUST 5 framework compliance verification
- Test/lint/type check execution
- Security scanning
- PostToolUse Hooks integration

**TRUST 5 Framework**:

| Principle | Verification Items |
|-----------|-------------------|
| **T**ested | Unit coverage >= 80%, all tests passing |
| **R**eadable | Functions < 50 lines, files < 400 lines, nesting < 4 levels |
| **U**nified | Consistent code style, DRY principle |
| **S**ecured | No hardcoded secrets, input validation |
| **T**rackable | Meaningful commits, SPEC traceability |

**When to Use**:
- After code changes for quality verification
- Automatically executed in `/jikime:2-run` Phase 2.5

---

### manager-git

**Role**: Git Workflow Expert

| Property | Value |
|----------|-------|
| Model | haiku |
| Tools | Bash, Read, Write, Edit, Grep, Glob, TodoWrite, Task, Skill |
| Skills | jikime-foundation-claude, jikime-foundation-core, jikime-workflow-project |

**Core Functions**:
- Git strategy by Personal/Team mode
- DDD phase-based commit messages
- Checkpoint system
- PR management (Team mode)

**Workflow Modes**:

| Mode | Branch Strategy | Commit Style |
|------|-----------------|--------------|
| Personal | Direct commit to main | Checkpoint tags |
| Team | feature/* → PR → main | PR-based |

**Checkpoint Format**: `jikime_cp/SPEC-XXX/phase_name`

**When to Use**:
- When committing/pushing code
- When creating PRs

---

### manager-database

**Role**: Database Management Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- DB schema design and normalization
- Query performance optimization
- Index strategy planning
- Backup and recovery planning
- Security and access control

**Checklist**:
- [ ] Query execution time < 100ms
- [ ] Index hit rate > 99%
- [ ] Connection pool optimization
- [ ] Backup verification complete
- [ ] Recovery procedure tested

**When to Use**:
- When designing/changing DB schema
- When query performance issues occur

---

### manager-dependency

**Role**: Dependency Management Expert

| Property | Value |
|----------|-------|
| Model | haiku |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- Dependency audit and analysis
- Security vulnerability remediation
- Version compatibility management
- Update strategy planning

**Update Strategies**:

| Strategy | Risk Level | Use Case |
|----------|------------|----------|
| Patch Only | Low | Production hotfixes |
| Minor Updates | Medium | Regular maintenance |
| Major Updates | High | Planned upgrades |
| Security Only | Low | Security vulnerability discovered |

**When to Use**:
- When dependency updates are needed
- When security vulnerabilities are discovered

---

### manager-data

**Role**: Data Engineering Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- Data pipeline design and implementation
- ETL/ELT process optimization
- Data quality and validation
- Data modeling (dimensional/document)

**Pipeline Patterns**:

| Pattern | Use Case | Tools |
|---------|----------|-------|
| Batch ETL | Daily/hourly loads | Airflow, dbt |
| Streaming | Real-time data | Kafka, Flink |
| CDC | Change capture | Debezium |
| ELT | Cloud warehouses | Snowflake, BigQuery |

**When to Use**:
- When designing data pipelines
- When data quality issues occur

---

### manager-context

**Role**: Context/Token Management Expert

| Property | Value |
|----------|-------|
| Model | haiku |
| Tools | Read, Write, Edit, Glob, Grep |

**Core Functions**:
- Context window optimization
- Session state management
- Token budget management
- Context retrieval and loading

**Token Management Strategy**:

| Zone | Usage | Action |
|------|-------|--------|
| Green | 0-60% | Normal operation |
| Yellow | 60-75% | Start compression |
| Orange | 75-85% | Archive non-essential items |
| Red | 85-95% | Active optimization |
| Critical | 95%+ | Emergency measures |

**When to Use**:
- When context optimization is needed
- When managing token budget

---

## Specialist Agents

### architect

**Role**: System Architecture Expert

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Grep, Glob |

**Core Functions**:
- System architecture design
- Technical trade-off evaluation
- ADR (Architecture Decision Record) writing
- Scalability/maintainability review

**Architecture Principles**:

| Principle | Description |
|-----------|-------------|
| Modularity | High cohesion, low coupling |
| Scalability | Horizontally scalable design |
| Maintainability | Easy to understand and test structure |
| Security | Defense in depth |

**When to Use**:
- When designing large-scale features
- When executing `/jikime:architect` command

---

### planner

**Role**: Implementation Planning Expert

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Grep, Glob |

**Core Functions**:
- Implementation planning for complex features
- Requirements analysis
- Step decomposition and prioritization
- Risk assessment

**Planning Process**:
1. Requirements analysis (understand feature request, define success criteria)
2. Architecture review (analyze existing codebase)
3. Step decomposition (file paths, dependencies, complexity)
4. Implementation order determination

**When to Use**:
- Before complex feature implementation
- When planning refactoring

---

### build-fixer

**Role**: Build/Type Error Resolution Expert

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep, Glob |

**Core Principle**: **Pass build with minimal changes** - No refactoring, only error fixes

**Common Error Patterns Fixed**:

| Error Type | Solution |
|------------|----------|
| Parameter has 'any' type | Add type annotation |
| Object is possibly 'undefined' | Use optional chaining (`?.`) |
| Cannot find module | Check path or use relative path |
| Hook called conditionally | Call hook at top level |

**Success Criteria**:
- `tsc --noEmit` passes
- `npm run build` succeeds
- Minimal line changes (less than 5% of affected files)

**When to Use**:
- When build errors occur
- When executing `/jikime:build-fix` command

---

### reviewer

**Role**: Code Review Expert

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Grep, Glob, Bash |

**Review Checklist**:

| Severity | Review Items |
|----------|--------------|
| CRITICAL | Hardcoded secrets, SQL Injection, XSS |
| HIGH | Large functions (50+ lines), deep nesting (4+ levels), missing error handling |
| MEDIUM | Inefficient algorithms, unnecessary re-renders |

**Approval Criteria**:

| Status | Condition |
|--------|-----------|
| Approve | No CRITICAL, HIGH issues |
| Warning | Only MEDIUM issues |
| Block | CRITICAL or HIGH issues present |

**When to Use**:
- After code changes for review
- During PR review

---

### refactorer

**Role**: Refactoring/Cleanup Expert

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep, Glob |

**Core Functions**:
- Unused code detection and removal
- Duplicate code consolidation
- Dependency cleanup
- DELETION_LOG.md documentation

**Analysis Tools**:
```bash
npx knip        # Unused exports/files/dependencies
npx depcheck    # Unused npm dependencies
npx ts-prune    # Unused TypeScript exports
```

**Safety Checklist**:
- Grep search all references
- Check dynamic imports
- Verify public API status
- Run all tests

**When to Use**:
- When cleaning up code
- When executing `/jikime:refactor` command

---

### security-auditor

**Role**: Security Audit Expert

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep, Glob |

**OWASP Top 10 Checklist**:

| Vulnerability | Inspection Items |
|---------------|------------------|
| Injection | Parameterized queries usage |
| Broken Authentication | Hash comparison usage |
| Sensitive Data Exposure | Environment variable usage |
| XSS | textContent vs innerHTML usage |
| SSRF | URL validation |
| Insufficient Authorization | Permission verification |

**Severity Classification**:

| Severity | Action |
|----------|--------|
| CRITICAL | Fix immediately |
| HIGH | Fix before deployment |
| MEDIUM | Fix if possible |
| LOW | Review and decide |

**When to Use**:
- When performing security audits
- When executing `/jikime:security` command

---

### test-guide

**Role**: Test Guidance Expert

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep |

**TDD Workflow (Red-Green-Refactor)**:
1. **RED**: Write tests first
2. **GREEN**: Minimal implementation to pass
3. **REFACTOR**: Improve

**Test Types**:

| Type | Target | Required |
|------|--------|----------|
| Unit | Individual functions/modules | Yes |
| Integration | API endpoints | Yes |
| E2E | User flows | Core only |

**Required Coverage**: 80%+

**When to Use**:
- When test writing guidance is needed
- When executing `/jikime:test` command

---

### e2e-tester

**Role**: E2E Testing Expert (Playwright)

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep, Glob |

**Core Functions**:
- Page Object Model pattern application
- Flaky test prevention
- Artifact (screenshots, videos) configuration
- Cross-browser testing

**Success Criteria**:
- All critical journey tests pass: 100%
- Overall pass rate > 95%
- Flaky rate < 5%
- Test time < 10 minutes

**When to Use**:
- When creating/running E2E tests
- When executing `/jikime:e2e` command

---

### documenter

**Role**: Documentation Expert

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Grep, Glob |

**Core Principle**: **Single Source of Truth** - Generate from code, minimize manual writing

**Document Structure**:
```
docs/
├── README.md           # Project overview
├── CODEMAPS/           # Code maps
│   ├── INDEX.md
│   ├── frontend.md
│   └── backend.md
└── GUIDES/             # Guides
    └── api.md
```

**When to Use**:
- When creating/updating documentation
- When executing `/jikime:docs` command

---

### migrator

**Role**: Legacy Modernization Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- Legacy code analysis and evaluation
- Tech stack modernization
- Framework migration
- Incremental transformation strategy

**Migration Patterns**:

| Pattern | Description | Use Case |
|---------|-------------|----------|
| Strangler Fig | Gradual component replacement | Large monoliths |
| Branch by Abstraction | Replace after abstraction | Core dependencies |
| Parallel Run | Run both versions simultaneously | Critical systems |
| Feature Toggle | Switch between implementations | Gradual rollout |

**When to Use**:
- When modernizing legacy systems
- When migrating frameworks

---

### specialist-java

**Role**: Java Development Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- Java 21+ development (leveraging latest features)
- Spring Boot 3.x applications
- JPA/Hibernate optimization
- Microservices design

**Java 21+ Key Features**:

| Feature | Use Case | Example |
|---------|----------|---------|
| Virtual Threads | High concurrency | `Thread.startVirtualThread()` |
| Pattern Matching | Type checking | `if (obj instanceof String s)` |
| Records | Data carriers | `record User(String name)` |
| Sealed Classes | Domain modeling | `sealed interface Shape` |

**When to Use**:
- When developing Java/Spring Boot
- When implementing enterprise applications

---

### specialist-spring

**Role**: Spring Ecosystem Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- Spring Boot 3.x configuration
- Spring Security setup
- Spring Data JPA optimization
- Spring Cloud microservices

**Key Configuration Points**:

| Area | Key Settings |
|------|--------------|
| Security | OAuth2/JWT, CORS, CSRF |
| Data | HikariCP, EntityGraph, Batch |
| Actuator | Health checks, metrics |
| Cloud | Config Server, Discovery |

**When to Use**:
- When setting up Spring-based projects
- When configuring Spring Security

---

### specialist-nextjs

**Role**: Next.js Development Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- Next.js 14/15/16 App Router
- React Server Components (RSC)
- Server Actions
- Performance optimization

**App Router Conventions**:

| File | Purpose |
|------|---------|
| `page.tsx` | Route UI component |
| `layout.tsx` | Shared layout |
| `loading.tsx` | Loading UI (Suspense) |
| `error.tsx` | Error boundary |
| `route.ts` | API endpoint |

**When to Use**:
- When developing Next.js
- When implementing RSC/Server Actions

---

### specialist-go

**Role**: Go Development Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- Go 1.22+ development
- Fiber/Gin web frameworks
- GORM database access
- Concurrency programming

**Go Patterns**:

| Pattern | Description | Example |
|---------|-------------|---------|
| Error Wrapping | Add context to errors | `fmt.Errorf("op: %w", err)` |
| Options Pattern | Flexible configuration | `WithTimeout(time.Second)` |
| Interface Segregation | Small interfaces | `type Reader interface` |
| Worker Pool | Limited concurrency | `sem := make(chan struct{}, n)` |

**When to Use**:
- When developing Go microservices
- When developing CLI tools

---

### specialist-postgres

**Role**: PostgreSQL Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- PostgreSQL 16+ advanced features
- pgvector embedding search
- Row Level Security (RLS)
- JSONB data type

**Advanced Feature Examples**:

| Feature | Purpose |
|---------|---------|
| pgvector | AI embedding similarity search |
| RLS | Row-level access control |
| JSONB | Semi-structured data storage |
| Partitioning | Large table performance |

**When to Use**:
- When using PostgreSQL advanced features
- When tuning database performance

---

### specialist-angular

**Role**: Angular Development Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- Angular 15+ development (standalone components, signals)
- NgRx state management (Store, Effects, Selectors)
- RxJS reactive programming
- Micro frontend architecture (Module Federation)

**Angular Latest Features**:

| Feature | Use Case | Example |
|---------|----------|---------|
| Standalone Components | Module-free components | `@Component({ standalone: true })` |
| Signals | Reactive state | `signal()`, `computed()`, `effect()` |
| Control Flow | Template control | `@if`, `@for`, `@switch` |
| Deferrable Views | Lazy loading | `@defer { }` |

**When to Use**:
- When developing Angular applications
- When implementing NgRx state management

---

### specialist-javascript

**Role**: JavaScript Development Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- ES2023+ latest feature utilization
- Node.js 20+ development
- Async patterns (Promise, async/await)
- Browser/server JavaScript integration

**ES2023+ Key Features**:

| Feature | Description | Example |
|---------|-------------|---------|
| Optional Chaining | Safe property access | `obj?.prop?.method?.()` |
| Nullish Coalescing | Default value handling | `value ?? 'default'` |
| Private Fields | Class encapsulation | `#privateField` |
| Top-level Await | Module async | `await import()` |

**When to Use**:
- When developing JavaScript applications
- When developing Node.js servers/CLIs

---

### specialist-php

**Role**: PHP Development Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- PHP 8.3+ latest features (Typed class constants, Override attribute)
- Laravel 11 / Symfony 7 frameworks
- Async PHP (Swoole, ReactPHP)
- Eloquent ORM / Doctrine ORM

**PHP 8.3+ Key Features**:

| Feature | Description | Example |
|---------|-------------|---------|
| Typed Class Constants | Constant type declaration | `public const string NAME` |
| #[Override] | Method override indication | `#[Override]` |
| json_validate() | JSON validation | `json_validate($json)` |
| Readonly Classes | Immutable classes | `readonly class User` |

**When to Use**:
- When developing PHP/Laravel/Symfony projects
- When modernizing legacy PHP

---

### specialist-python

**Role**: Python Development Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- Python 3.11+ latest feature utilization
- FastAPI / Django web frameworks
- Async programming (asyncio)
- Data science (pandas, numpy)

**Python 3.11+ Key Features**:

| Feature | Description | Example |
|---------|-------------|---------|
| Exception Groups | Multiple exception handling | `except* ExceptionGroup` |
| Task Groups | Structured concurrency | `async with TaskGroup()` |
| TOML Parser | Config file parsing | `tomllib.load()` |
| Pattern Matching | Structural patterns | `match value: case ...` |

**When to Use**:
- When developing Python/FastAPI/Django projects
- When implementing data pipelines

---

### specialist-rust

**Role**: Rust Development Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- Rust 2021 edition development
- Ownership and memory safety patterns
- Async programming (Tokio, async-std)
- System programming and FFI

**Rust Core Patterns**:

| Pattern | Description | Use Case |
|---------|-------------|----------|
| Ownership | Memory ownership transfer | Resource transfer |
| Borrowing | Access through references | Temporary access |
| Lifetimes | Reference validity scope | Complex reference relationships |
| RAII | Automatic resource release | File, lock management |

**When to Use**:
- When developing Rust applications
- When developing system programming/CLIs

---

### specialist-sql

**Role**: Multi-Database SQL Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- Cross-platform SQL optimization (PostgreSQL, MySQL, SQL Server, Oracle)
- Execution plan analysis and tuning
- Index strategy planning
- Data warehouse patterns

**Advanced Query Patterns**:

| Pattern | Purpose |
|---------|---------|
| CTEs | Complex query readability |
| Recursive Queries | Hierarchical data processing |
| Window Functions | Analysis and ranking |
| PIVOT/UNPIVOT | Data transformation |

**When to Use**:
- When optimizing query performance
- When designing data warehouses

---

### specialist-typescript

**Role**: TypeScript Development Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- TypeScript 5.0+ advanced features
- Type-level programming
- End-to-end type safety (tRPC, Prisma)
- Monorepo TypeScript configuration

**Advanced Type Patterns**:

| Pattern | Purpose |
|---------|---------|
| Conditional Types | Flexible type inference |
| Mapped Types | Type transformation |
| Template Literals | String manipulation |
| Branded Types | Domain modeling |

**When to Use**:
- When developing TypeScript projects
- When implementing advanced type patterns

---

### specialist-vue

**Role**: Vue Development Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- Vue 3 Composition API
- Nuxt 3 full-stack framework
- Pinia state management
- Reactivity system optimization

**Vue 3 Core Features**:

| Feature | Description | Example |
|---------|-------------|---------|
| Composition API | Logic reuse | `setup()`, `<script setup>` |
| Reactivity | Reactive data | `ref()`, `reactive()`, `computed()` |
| Teleport | DOM movement | `<Teleport to="body">` |
| Suspense | Async components | `<Suspense>` |

**When to Use**:
- When developing Vue/Nuxt applications
- When implementing Pinia state management

---

### specialist-api

**Role**: API Design Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Glob, Grep |

**Core Functions**:
- REST API design (resource modeling, versioning)
- GraphQL schema design
- OpenAPI 3.0/3.1 spec writing
- API gateway patterns

**API Design Principles**:

| Principle | Description |
|-----------|-------------|
| Resource-Centric | Noun-based URIs, CRUD mapping |
| Versioning | URL or header-based versioning |
| Error Handling | RFC 7807 Problem Details |
| Pagination | Cursor/Offset based |

**When to Use**:
- When designing and documenting APIs
- When writing OpenAPI specs

---

### specialist-graphql

**Role**: GraphQL Architecture Expert

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- GraphQL schema design and evolution
- Apollo Federation architecture
- N+1 query prevention (DataLoader)
- Subscription implementation

**Optimization Strategies**:

| Problem | Solution |
|---------|----------|
| N+1 Queries | DataLoader batching |
| Deep Queries | Depth limiting |
| Large Responses | Complexity analysis |
| Repeated Queries | Persisted queries |

**When to Use**:
- When designing GraphQL APIs
- When configuring Apollo Federation

---

### specialist-microservices

**Role**: Microservices Architecture Expert

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- Service boundary definition (DDD-based)
- Communication pattern design (sync/async)
- Service mesh configuration (Istio)
- Distributed data management

**Design Principles**:

| Principle | Description |
|-----------|-------------|
| Single Responsibility | One business function per service |
| Database per Service | No shared databases |
| API-First | Define contracts before implementation |
| Stateless | External state storage |

**When to Use**:
- When designing microservices
- When configuring Kubernetes deployments

---

### specialist-mobile

**Role**: Mobile Development Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- React Native 0.82+ / Flutter development
- Cross-platform code sharing (80%+)
- Offline-first architecture
- Native module integration

**Performance Goals**:

| Metric | Target |
|--------|--------|
| Cold Start | < 1.5 seconds |
| Memory Usage | < 120MB |
| Frame Rate | 60 FPS |
| App Size | < 40MB |

**When to Use**:
- When developing React Native / Flutter apps
- When implementing offline features

---

### specialist-electron

**Role**: Electron Desktop App Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- Electron Forge / Builder configuration
- Main-Renderer IPC communication
- Native module integration
- Auto-update (electron-updater)

**Security Checklist**:

| Item | Description |
|------|-------------|
| contextIsolation | true (required) |
| nodeIntegration | false (recommended) |
| sandbox | true (recommended) |
| CSP | Strict policy |

**When to Use**:
- When developing Electron apps
- When configuring desktop deployments

---

### specialist-websocket

**Role**: Real-time Communication Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- WebSocket server architecture
- Socket.IO + Redis clustering
- Connection management and scaling
- Presence and message history

**Performance Goals**:

| Metric | Target |
|--------|--------|
| Connections per Node | 50K concurrent |
| Message Latency | < 10ms p99 |
| Throughput | 100K msg/sec |
| Reconnection Time | < 2 seconds |

**When to Use**:
- When implementing real-time features
- When developing chat/notification systems

---

### fullstack

**Role**: Full-stack Development Expert

| Property | Value |
|----------|-------|
| Model | opus |
| Tools | Read, Write, Edit, Bash, Glob, Grep, TodoWrite, Task |

**Core Functions**:
- End-to-end feature implementation (DB → API → UI)
- Coordination of other specialist agents
- Full-stack architecture decisions
- Integration test strategy

**Work Scope**:

| Layer | Responsibility |
|-------|----------------|
| DB | Schema design, query optimization |
| Backend | API design, business logic |
| Frontend | UI components, state management |
| DevOps | Deployment, monitoring |

**When to Use**:
- When implementing end-to-end features
- When working across multiple layers

---

### analyst

**Role**: Research and Analysis Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Grep, Glob, WebFetch, WebSearch |

**Core Functions**:
- Technical research and evaluation
- Competitive analysis
- Decision support
- Knowledge synthesis

**Analysis Frameworks**:

| Type | Purpose | Output |
|------|---------|--------|
| Technical | Technology evaluation | Comparison matrix |
| Competitive | Market positioning | Competitor profiles |
| Feasibility | Project viability | Risk/opportunity report |
| Trend | Future outlook | Trend analysis |

**When to Use**:
- When researching before technical decisions
- When competitive analysis is needed

---

### explorer

**Role**: Codebase Exploration Expert

| Property | Value |
|----------|-------|
| Model | haiku |
| Tools | Read, Grep, Glob |

**Core Functions**:
- Code pattern search
- Implementation discovery
- Usage tracking
- Architecture exploration

**Search Strategies**:

| Strategy | Use Case | Tool |
|----------|----------|------|
| File Pattern | Find by filename | Glob |
| Content Search | Find code patterns | Grep |
| Definition | Find declarations | Grep + Read |
| Usage | Find references | Grep |

**When to Use**:
- When exploring code
- When executing `/jikime:learn` command

---

## Orchestration Agents

### orchestrator

**Role**: Workflow Orchestration Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- Workflow design and coordination
- Pipeline coordination
- Process automation
- State management

**Workflow Patterns**:

| Pattern | Use Case | Description |
|---------|----------|-------------|
| Sequential | Dependent steps | A → B → C |
| Parallel | Independent steps | A, B, C simultaneously |
| Fan-out/Fan-in | Distributed processing | Split, process, aggregate |
| Saga | Distributed transactions | Compensating actions |
| Pipeline | Data processing | Stage passage |

**When to Use**:
- When coordinating complex multi-step processes
- When automating workflows

---

### coordinator

**Role**: Multi-Agent Coordinator

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Bash, Glob, Grep |

**Core Functions**:
- Agent team composition
- Task distribution and scheduling
- Result integration and synthesis
- Conflict resolution

**Coordination Patterns**:

| Pattern | Description |
|---------|-------------|
| Parallel Execution | Multiple agents run simultaneously |
| Sequential Chain | Agent chain sequential execution |
| Fan-out/Fan-in | Distribute then integrate results |

**Agent Selection Matrix**:

| Task Type | Primary Agent | Supporting Agents |
|-----------|---------------|-------------------|
| Architecture | architect | analyst, specialist-* |
| Implementation | backend/frontend | debugger, test-guide |
| Migration | migrator | specialist-*, manager-database |
| Security | security-auditor | backend, analyst |

**When to Use**:
- When multiple agent collaboration is needed
- When decomposing complex tasks

---

### dispatcher

**Role**: Task Distribution Expert

| Property | Value |
|----------|-------|
| Model | haiku |
| Tools | Read, Write, Edit, Glob, Grep |

**Core Functions**:
- Task queue management
- Load balancing
- Priority scheduling
- Resource optimization

**Distribution Strategies**:

| Strategy | Use Case | Description |
|----------|----------|-------------|
| Round-Robin | Equal distribution | Agent rotation |
| Weighted | Capacity-based | Proportional to capability |
| Least-Loaded | Load balancing | Route to least busy |
| Affinity | Skill matching | Assign to experts |
| Priority | Deadline-based | Urgent tasks first |

**Performance Metrics**:
- Distribution latency < 50ms
- Load variance < 10%
- Task completion rate > 99%
- Deadline compliance > 95%

**When to Use**:
- When distributing bulk tasks
- When managing agent load

---

## Builder Agents

### agent-builder

**Role**: Agent Definition Creation

Creates new agent definition files in JikiME-ADK format.

**Generated File**: `.claude/agents/jikime/{agent-name}.md`

---

### command-builder

**Role**: Slash Command Creation

Creates new slash command definition files.

**Generated File**: `.claude/commands/jikime/{command-name}.md`

---

### skill-builder

**Role**: Skill Definition Creation

Creates skill files following the Progressive Disclosure pattern.

**Generated File**: `.claude/skills/jikime-{type}-{name}/SKILL.md`

---

### plugin-builder

**Role**: Plugin Package Creation

Creates JikiME-ADK plugin package structure.

**Generated Directory**: `packages/{plugin-name}/`

---

## Designer Agents

### designer-ui

**Role**: UI Design and Design System Expert

| Property | Value |
|----------|-------|
| Model | sonnet |
| Tools | Read, Write, Edit, Glob, Grep |

**Core Functions**:
- Design system creation and maintenance
- Component library architecture
- Design token management
- Accessibility compliance (WCAG 2.1 AA)

**Design System Architecture**:

| Layer | Components |
|-------|------------|
| Design Tokens | Colors, typography, spacing, shadows |
| Core Components | Button, Input, Card, Modal, Table |
| Patterns | Navigation, Forms, Data Display |
| Templates | Page layouts, app shells |

**Accessibility Standards**:

| Item | Requirement |
|------|-------------|
| Color Contrast | Normal text 4.5:1, large text 3:1 |
| Keyboard | All interactive elements focusable |
| Screen Reader | Semantic HTML, ARIA labels |
| Motion | prefers-reduced-motion support |

**When to Use**:
- When building design systems
- When developing component libraries
- When improving accessibility

---

## Agent Selection Guide

### Selection Decision Tree

```
1. Read-only codebase exploration?
   → Use explorer subagent

2. Need external documentation/API research?
   → Use WebSearch, WebFetch, Context7 MCP tools

3. Need domain expertise?
   → Use specialist subagent (backend, frontend, debugger, etc.)

4. Need language/framework expertise?
   → Use specialist-[lang] subagent (specialist-java, specialist-go, etc.)

5. Need workflow coordination?
   → Use manager-[workflow] subagent

6. Complex multi-step task?
   → Use manager-strategy subagent

7. Multi-agent coordination?
   → Use coordinator or orchestrator subagent

8. Legacy migration?
   → Use migrator subagent

9. Create new agent/command/skill?
   → Use [type]-builder subagent
```

### Command → Agent Mapping

| Command | Primary Agent |
|---------|---------------|
| `/jikime:0-project` | manager-project |
| `/jikime:1-plan` | manager-spec |
| `/jikime:2-run` | manager-strategy → manager-ddd |
| `/jikime:3-sync` | manager-docs → manager-git |
| `/jikime:jarvis` | (J.A.R.V.I.S. orchestration) |
| `/jikime:build-fix` | build-fixer |
| `/jikime:loop` | debugger → refactorer |
| `/jikime:architect` | architect |
| `/jikime:docs` | manager-docs |
| `/jikime:security` | security-auditor |
| `/jikime:test` | test-guide |
| `/jikime:e2e` | e2e-tester |
| `/jikime:learn` | explorer |
| `/jikime:refactor` | refactorer |
| `/jikime:friday` | (F.R.I.D.A.Y. orchestration) |
| `/jikime:migrate-*` | migrator |
| `/jikime:smart-rebuild` | migrator + specialist-nextjs |

### Tech Stack → Agent Mapping

| Tech Stack | Agent |
|------------|-------|
| Java/Spring Boot | specialist-java, specialist-spring |
| Next.js/React | specialist-nextjs |
| Go | specialist-go |
| PostgreSQL | specialist-postgres, manager-database |
| GraphQL | specialist-graphql, specialist-api |
| Microservices/K8s | specialist-microservices |
| React Native/Flutter | specialist-mobile |
| Electron | specialist-electron |
| WebSocket/Socket.IO | specialist-websocket |
| Data Pipeline | manager-data |
| Dependency Management | manager-dependency |
| Legacy Code | migrator |
| Design System | designer-ui |
| Full-stack Features | fullstack |

---

## Agent Collaboration Patterns

### Sequential Chaining

```
manager-spec → manager-strategy → manager-ddd → manager-quality → manager-git
    (SPEC)        (Planning)       (Implementation)  (Verification)   (Commit)
```

### Parallel Execution

```
backend ─┬─→ Result integration
frontend ─┘   (Concurrent work)
```

### Consultation Pattern

```
manager-ddd ─→ architect (Architecture consultation)
            ─→ security-auditor (Security review)
            ─→ test-guide (Test strategy)
```

### Multi-Agent Coordination

```
coordinator
    ├─→ backend (API implementation)
    ├─→ frontend (UI implementation)
    ├─→ test-guide (Test cases)
    └─→ [Integration] → manager-quality (Quality verification)
```

### Task Distribution

```
dispatcher
    ├─→ High priority → specialist-java (Fast response)
    ├─→ Medium priority → specialist-go (Normal processing)
    └─→ Low priority → (Batch processing)
```

---

Version: 5.1.0
Last Updated: 2026-02-06
