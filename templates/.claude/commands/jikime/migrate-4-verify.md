---
description: "[Step 4/4] Migration verification. Runs /jikime:verify pre-pr on migrated project."
argument-hint: '[--skip-browser] [--headed]'
type: workflow
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Edit, Glob, Grep
model: inherit
---

# Migration Step 4: Verify

**Final Step**: Verify the migrated project works correctly.

## What This Command Does

1. Read `output_dir` from `.migrate-config.yaml`
2. Run `/jikime:verify pre-pr` on the migrated project
3. Update `progress.yaml` with verification status

## Usage

```bash
# Verify migrated project (full verification including browser)
/jikime:migrate-4-verify

# Skip browser verification (static analysis only)
/jikime:migrate-4-verify --skip-browser

# Run with visible browser window (headed mode for manual verification)
/jikime:migrate-4-verify --headed
```

## Options

| Option | Description |
|--------|-------------|
| `--skip-browser` | Skip browser verification (pass `--no-browser` to verify) |
| `--headed` | Run browser verification in headed mode (visible browser window) |

---

## Execution Flow

### Step 1: Load Configuration

```
Read .migrate-config.yaml:
  - output_dir: Path to migrated project
  - artifacts_dir: Path to migration artifacts
```

If `.migrate-config.yaml` not found:
- Inform user that previous steps must be completed first
- Suggest running `/jikime:migrate-3-execute` before this command

### Step 2: Run Verification

```bash
cd {output_dir}

# If --skip-browser:
/jikime:verify pre-pr --no-browser

# If --headed (visible browser for manual verification):
/jikime:verify --browser-only --headed

# Otherwise (headless browser):
/jikime:verify pre-pr
```

**What verify pre-pr checks:**
- Build compilation
- Type checking
- Lint validation
- Test execution
- Security scanning
- Browser runtime errors (unless --skip-browser)

**Browser Modes:**
- **headless** (default): No visible browser, automated verification
- **headed** (`--headed`): Visible browser window for manual verification and debugging

### Step 3: Update Progress

On success, update `{artifacts_dir}/progress.yaml`:

```yaml
status: verified
verified_at: "2026-01-26T12:00:00Z"
verification:
  profile: pre-pr
  browser_check: true  # or false if --skip-browser
  result: passed
```

On failure:
```yaml
status: verification_failed
verification:
  profile: pre-pr
  result: failed
  errors: [list of errors from verify]
```

---

## Workflow (Data Flow)

```
/jikime:migrate-0-discover
        ↓ (.migrate-config.yaml)
/jikime:migrate-1-analyze
        ↓ (as_is_spec.md)
/jikime:migrate-2-plan
        ↓ (migration_plan.md)
/jikime:migrate-3-execute
        ↓ (output_dir/ + progress.yaml)
/jikime:migrate-4-verify  ← current
        ↓
        └─ Runs: /jikime:verify pre-pr
        └─ Updates: progress.yaml (status: verified)
```

---

## EXECUTION DIRECTIVE

Arguments: $ARGUMENTS

1. **Parse flags from $ARGUMENTS**:
   - `--skip-browser`: Skip browser verification
   - `--headed`: Run browser in headed mode (visible window)

2. **Load configuration**:
   ```bash
   # Read .migrate-config.yaml
   cat .migrate-config.yaml
   ```
   - Extract `output_dir` and `artifacts_dir`
   - IF file not found: Inform user to run `/jikime:migrate-3-execute` first

3. **Change to output directory**:
   ```bash
   cd {output_dir}
   ```

4. **Run static analysis verification**:
   ```bash
   # Build check
   npm run build 2>&1 | tail -30

   # Type check
   npx tsc --noEmit 2>&1

   # Lint check
   npm run lint 2>&1

   # Test
   npm test 2>&1
   ```

5. **Run browser verification** (unless `--skip-browser`):

   **IF `--headed` flag**:
   ```bash
   # Start dev server in background
   npm run dev &
   DEV_PID=$!
   sleep 5  # Wait for server to start

   # Open browser in headed mode (visible window)
   npx playwright open http://localhost:3000

   # After manual verification, kill dev server
   kill $DEV_PID
   ```

   **ELSE (headless mode)**:
   ```bash
   # Start dev server in background
   npm run dev &
   DEV_PID=$!
   sleep 5

   # Take screenshot for verification
   npx playwright screenshot http://localhost:3000 verification-screenshot.png

   # Kill dev server
   kill $DEV_PID
   ```

6. **Update progress.yaml**:
   - On success: Set `status: verified`
   - On failure: Set `status: verification_failed` with error details

7. **Report results** to user in F.R.I.D.A.Y. format

Execute NOW. Do NOT just describe.

---

## Migration Complete!

When verification passes, migration is complete.

**Next Steps:**
1. Review the migrated code
2. Deploy to staging environment
3. User Acceptance Testing (UAT)
4. Production deployment

---

Version: 5.0.0
Changelog:
- v5.0.0: Simplified to wrapper for /jikime:verify pre-pr. Removed source↔target comparison.
- v4.0.0: Playwright-based verification with visual regression, cross-browser, accessibility
- v3.0.0: Config-first approach
