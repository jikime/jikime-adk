# Security Guidelines

Security best practices based on OWASP Top 10 and industry standards.

## OWASP Top 10 Checklist

### 1. Injection (A01)

```typescript
// NEVER: String concatenation
const query = `SELECT * FROM users WHERE id = ${userId}`

// ALWAYS: Parameterized queries
const query = 'SELECT * FROM users WHERE id = $1'
const result = await db.query(query, [userId])
```

### 2. Broken Authentication (A02)

```typescript
// Password requirements
const passwordSchema = z.string()
  .min(8)
  .regex(/[A-Z]/, 'Requires uppercase')
  .regex(/[a-z]/, 'Requires lowercase')
  .regex(/[0-9]/, 'Requires number')
  .regex(/[^A-Za-z0-9]/, 'Requires special char')

// Session management
const sessionConfig = {
  secret: process.env.SESSION_SECRET,
  cookie: {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict',
    maxAge: 24 * 60 * 60 * 1000 // 24 hours
  }
}
```

### 3. Sensitive Data Exposure (A03)

```typescript
// NEVER: Log sensitive data
console.log('User login:', { email, password })

// ALWAYS: Sanitize logs
console.log('User login:', { email, password: '[REDACTED]' })

// NEVER: Return sensitive fields
return user

// ALWAYS: Exclude sensitive fields
const { passwordHash, ...safeUser } = user
return safeUser
```

### 4. XSS Prevention (A07)

```typescript
// React: Safe by default
return <div>{userInput}</div>

// DANGER: dangerouslySetInnerHTML
return <div dangerouslySetInnerHTML={{ __html: userInput }} />

// If HTML needed, sanitize first
import DOMPurify from 'dompurify'
const clean = DOMPurify.sanitize(userInput)
return <div dangerouslySetInnerHTML={{ __html: clean }} />
```

### 5. CSRF Protection (A05)

```typescript
// Use CSRF tokens
import csrf from 'csurf'
app.use(csrf({ cookie: true }))

// Include token in forms
<form>
  <input type="hidden" name="_csrf" value={csrfToken} />
</form>
```

## Secret Management

### Environment Variables

```typescript
// NEVER: Hardcoded secrets
const apiKey = 'sk-proj-xxxxx'

// ALWAYS: Environment variables
const apiKey = process.env.API_KEY

if (!apiKey) {
  throw new Error('API_KEY environment variable not set')
}
```

### Secret Detection

```markdown
Check for:
- API keys: sk-, pk-, api_
- Passwords in plain text
- Private keys: -----BEGIN
- Connection strings with credentials
- JWT secrets in code
```

### .gitignore Requirements

```gitignore
# Secrets
.env
.env.local
.env.*.local
*.pem
*.key
credentials.json
secrets/

# IDE
.idea/
.vscode/settings.json
```

## Input Validation

### Validate All Inputs

```typescript
// API endpoint validation
import { z } from 'zod'

const createUserSchema = z.object({
  email: z.string().email().max(255),
  password: z.string().min(8).max(100),
  name: z.string().min(1).max(50).trim()
})

app.post('/users', (req, res) => {
  const result = createUserSchema.safeParse(req.body)
  if (!result.success) {
    return res.status(400).json({ error: result.error })
  }
  // Process validated data
})
```

### File Upload Security

```typescript
const uploadConfig = {
  limits: {
    fileSize: 5 * 1024 * 1024, // 5MB max
    files: 1
  },
  fileFilter: (req, file, cb) => {
    const allowed = ['image/jpeg', 'image/png', 'image/gif']
    if (allowed.includes(file.mimetype)) {
      cb(null, true)
    } else {
      cb(new Error('Invalid file type'), false)
    }
  }
}
```

## Security Response Protocol

### If Security Issue Found

```markdown
1. STOP current work immediately
2. Assess severity (CRITICAL/HIGH/MEDIUM/LOW)
3. For CRITICAL/HIGH:
   - Fix before continuing
   - Rotate any exposed secrets
   - Review entire codebase for similar issues
4. Document the issue and fix
5. Add test to prevent regression
```

### Severity Definitions

| Level | Examples | Action |
|-------|----------|--------|
| **CRITICAL** | Exposed secrets, RCE | Fix immediately, rotate secrets |
| **HIGH** | SQL injection, auth bypass | Fix before merge |
| **MEDIUM** | XSS, CSRF | Should fix soon |
| **LOW** | Information disclosure | Plan to fix |

## Security Checklist

Before ANY commit:

- [ ] No hardcoded secrets
- [ ] All user inputs validated
- [ ] SQL injection prevention
- [ ] XSS prevention
- [ ] CSRF protection (if applicable)
- [ ] Authentication verified
- [ ] Authorization checked
- [ ] Sensitive data not logged
- [ ] Error messages don't leak info
- [ ] Dependencies up to date

## Dependency Security

```bash
# Check for vulnerabilities
npm audit
pnpm audit

# Fix automatically
npm audit fix
pnpm audit --fix

# Check outdated packages
npm outdated
```

---

Version: 1.0.0
Source: JikiME-ADK security rules
