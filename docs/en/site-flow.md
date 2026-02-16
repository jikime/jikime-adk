# Site-Flow Integration Guide

> Project management and progress tracking through JiKiME-ADK and Site-Flow integration

## Overview

Site-Flow is a page flow and E2E test management platform. By integrating with JiKiME-ADK, you can automatically manage projects, record work activities, and track progress during the development process.

### Site-Flow Key Features

| Feature | Description |
|---------|-------------|
| **Page Management** | Site page structure management, screenshot gallery |
| **Feature Management** | Page-specific feature definitions, implementation status tracking |
| **Canvas (Flow)** | Page connection visualization, user flow management |
| **Test Cases** | E2E test creation, execution, deployment |
| **Bug Reports** | Bug/feature request/improvement registration and tracking |
| **Quality Analysis** | Page health, impact scope analysis |

---

## Integration Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        JiKiME-ADK                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  Commands    │  │   Skills     │  │    Hooks     │          │
│  │ /jikime:*    │  │ jikime-*     │  │ Pre/Post     │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
│         │                 │                 │                   │
│         └────────────┬────┴─────────────────┘                   │
│                      │                                          │
│              ┌───────▼───────┐                                  │
│              │ Site-Flow MCP │                                  │
│              │    Server     │                                  │
│              └───────┬───────┘                                  │
└──────────────────────┼──────────────────────────────────────────┘
                       │ HTTP API
                       ▼
┌──────────────────────────────────────────────────────────────────┐
│                        Site-Flow                                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐         │
│  │  Pages   │  │ Features │  │  Tests   │  │   Bugs   │         │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘         │
└──────────────────────────────────────────────────────────────────┘
```

---

## Configuration

### 1. Site-Flow MCP Server Setup

Add the Site-Flow MCP server to `.claude/settings.json`:

```json
{
  "mcpServers": {
    "site-flow": {
      "command": "node",
      "args": ["./path/to/site-flow/mcp-server/dist/index.js"],
      "env": {
        "SITE_FLOW_API_URL": "http://localhost:5174",
        "SITE_FLOW_API_KEY": "sf_your_api_key_here"
      }
    }
  }
}
```

### 2. API Key Generation

Generate an API key from Site-Flow:

1. Access Site-Flow admin page
2. Settings → API Key Management
3. Create new API key
4. Set the `SITE_FLOW_API_KEY` environment variable

---

## Commands

### `/jikime:project` - Project Management

Register and manage projects.

```bash
# Initialize project (analyze current directory)
/jikime:project init

# List pages
/jikime:project list

# Register page
/jikime:project add "/login" "Login Page"

# Get page details (including features, tests, bugs)
/jikime:project context "/login"
```

**Site-Flow MCP Tools Used:**
- `list_pages` - List pages
- `create_page` - Register page
- `update_page` - Update page
- `get_page_context` - Get page context

---

### `/jikime:progress` - Progress Check

Check overall project progress.

```bash
# Overall statistics
/jikime:progress

# Test statistics
/jikime:progress --test

# Bug statistics
/jikime:progress --bug

# Specific page health
/jikime:progress --page "/login"
```

**Site-Flow MCP Tools Used:**
- `get_test_stats` - Test statistics
- `get_bug_stats` - Bug statistics
- `get_page_health` - Page health
- `get_impact_scope` - Impact scope analysis

---

### `/jikime:record` - Work Recording

Record issues or changes that occur during development.

```bash
# Register bug
/jikime:record bug "No error message on login failure" --page "/login"

# Register feature request
/jikime:record feature "Add social login" --page "/login"

# Register improvement
/jikime:record improvement "Improve login form UI" --page "/login"

# Update feature status
/jikime:record update-feature FEATURE_ID --status completed
```

**Site-Flow MCP Tools Used:**
- `create_bug_report` - Register bug/feature request/improvement
- `add_bug_comment` - Add comment
- `update_bug_report` - Update status
- `update_feature` - Update feature status

---

## Skills

### `jikime-site-flow-project` - Project Management Skill

A comprehensive skill for managing pages, features, and canvases.

```bash
# Invoke skill
/jikime:skill site-flow-project

# Or use MCP tools directly
"Show me the list of pages registered in Site Flow"
"Register the login page (/login)"
"Create a main flow canvas"
```

**Included Features:**
- Page CRUD (Create, Read, Update, Delete)
- Feature CRUD
- Canvas management (page placement, link connections)
- Page image gallery management

---

### `jikime-site-flow-test` - Test Management Skill

A skill for creating and managing E2E test cases.

```bash
# Invoke skill
/jikime:skill site-flow-test

# Or use MCP tools directly
"Create E2E test cases for the login feature"
"Deploy the generated test cases to files"
"Run the tests"
```

**Included Features:**
- Test case creation/retrieval
- Feature-based auto test generation (`generate_test_cases`)
- Test file deployment (`deploy_test_cases`)
- Test execution (`run_tests`)
- Test execution history retrieval

---

### `jikime-site-flow-bug` - Bug Report Skill

A skill for managing bugs, feature requests, and improvements.

```bash
# Invoke skill
/jikime:skill site-flow-bug

# Or use MCP tools directly
"Register a bug on the login page"
"Add a fix comment to the bug"
"Show bug statistics"
```

**Included Features:**
- Bug/feature request/improvement registration
- Status management (open → in-progress → resolved → closed)
- Comment addition (source code attachment supported)
- Statistics and analysis

---

## Hooks Automation

Automatically record the development process to Site-Flow through JiKiME-ADK Hooks.

### PreCommit Hook

Identify related pages based on changed files before commit.

```yaml
# .claude/settings.local.yaml
hooks:
  PreCommit:
    - command: "jikime-adk hooks site-flow pre-commit"
      description: "Check Site-Flow page connections"
```

### PostCommit Hook

Automatically record work activities to Site-Flow after commit.

```yaml
hooks:
  PostCommit:
    - command: "jikime-adk hooks site-flow post-commit"
      description: "Record work to Site-Flow"
```

### PostSpecComplete Hook

Automatically generate test cases when SPEC is completed.

```yaml
hooks:
  PostSpecComplete:
    - command: "jikime-adk hooks site-flow generate-tests"
      description: "Generate Site-Flow test cases"
```

---

## Workflow Examples

### 1. Starting a New Project

```bash
# 1. Initialize project
/jikime:project init

# 2. Register page structure
/jikime:project add "/login" "Login Page"
/jikime:project add "/dashboard" "Dashboard"
/jikime:project add "/settings" "Settings Page"

# 3. Create flow canvas
"Create a main user flow canvas and connect the login → dashboard → settings flow"
```

### 2. SPEC-Based Development

```bash
# 1. SPEC planning
/jikime:1-plan "Login Feature"
  → Feature auto-registered in Site-Flow (status: planned)

# 2. SPEC implementation
/jikime:2-run SPEC-LOGIN-001
  → Feature status: in-progress

# 3. SPEC completion
/jikime:3-sync SPEC-LOGIN-001
  → Feature status: completed
  → Test cases auto-generated
```

### 3. Progress Monitoring

```bash
# Check overall progress
/jikime:progress

# Example output:
# ┌─────────────────────────────────────┐
# │ 📊 Project Progress                 │
# ├─────────────────────────────────────┤
# │ Pages: 15 (Done: 8, In Progress: 5) │
# │ Features: 45 (Done: 30, WIP: 10)    │
# │ Tests: 120 (Passed: 95, Failed: 5)  │
# │ Bugs: 12 (Resolved: 8, WIP: 3)      │
# └─────────────────────────────────────┘
```

### 4. Bug Discovery and Fix

```bash
# 1. Register bug
/jikime:record bug "Error message not displayed on login failure" --page "/login" --severity major

# 2. Add fix comment after fixing
"Add a fix comment to BUG_ID. Fix details: Added error handling in LoginForm.tsx"

# 3. Update bug status
/jikime:record update-bug BUG_ID --status resolved
```

---

## Complete Site-Flow MCP Tool List

### Page Management
| Tool | Description |
|------|-------------|
| `list_pages` | List pages |
| `create_page` | Register page |
| `update_page` | Update page |
| `update_pages_bulk` | Bulk register/update pages |
| `get_page_image` | Get page screenshot |
| `get_page_context` | Get page context (features, tests, bugs) |

### Page Image Gallery
| Tool | Description |
|------|-------------|
| `list_page_images` | List page images |
| `add_page_image` | Add image |
| `delete_page_image` | Delete image |

### Feature Management
| Tool | Description |
|------|-------------|
| `list_features` | List features |
| `create_feature` | Register feature |
| `update_feature` | Update feature |

### Canvas (Flow) Management
| Tool | Description |
|------|-------------|
| `list_canvases` | List canvases |
| `create_canvas` | Create canvas |
| `get_canvas_pages` | Get canvas pages/connections |
| `add_page_to_canvas` | Add page to canvas |
| `add_link_between_pages` | Add link between pages |

### Test Cases
| Tool | Description |
|------|-------------|
| `list_test_cases` | List test cases |
| `create_test_case` | Create test case |
| `generate_test_cases` | Auto-generate from features |
| `deploy_test_cases` | Deploy test files |

### Test Execution
| Tool | Description |
|------|-------------|
| `list_test_executions` | List execution history |
| `run_tests` | Run tests |

### Bug Reports
| Tool | Description |
|------|-------------|
| `list_bug_reports` | List bugs |
| `create_bug_report` | Register bug/feature/improvement |
| `update_bug_report` | Update bug |
| `add_bug_comment` | Add comment |

### Statistics/Analysis
| Tool | Description |
|------|-------------|
| `get_test_stats` | Test statistics |
| `get_bug_stats` | Bug statistics |
| `get_page_health` | Page health |
| `get_impact_scope` | Impact scope analysis |

---

## Related Documentation

- [Commands Reference](./commands.md)
- [Skills Catalog](./skills-catalog.md)
- [Hooks System](./hooks.md)
- [Memory System](./memory.md)
