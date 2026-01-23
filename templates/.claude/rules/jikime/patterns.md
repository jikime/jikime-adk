# Common Patterns

Reusable code patterns for consistent implementation.

## API Patterns

### Response Format

```typescript
// Standard API Response
interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
    details?: unknown
  }
  meta?: {
    total: number
    page: number
    limit: number
    hasMore: boolean
  }
}

// Usage
function success<T>(data: T, meta?: Meta): ApiResponse<T> {
  return { success: true, data, meta }
}

function error(code: string, message: string): ApiResponse<never> {
  return { success: false, error: { code, message } }
}
```

### Error Handling

```typescript
// Custom Error Classes
class AppError extends Error {
  constructor(
    public code: string,
    message: string,
    public statusCode: number = 500
  ) {
    super(message)
    this.name = 'AppError'
  }
}

class ValidationError extends AppError {
  constructor(message: string, public fields: Record<string, string>) {
    super('VALIDATION_ERROR', message, 400)
  }
}

class NotFoundError extends AppError {
  constructor(resource: string) {
    super('NOT_FOUND', `${resource} not found`, 404)
  }
}
```

## Repository Pattern

```typescript
interface Repository<T, ID = string> {
  findAll(filters?: Filters): Promise<T[]>
  findById(id: ID): Promise<T | null>
  create(data: CreateDto<T>): Promise<T>
  update(id: ID, data: UpdateDto<T>): Promise<T>
  delete(id: ID): Promise<void>
  exists(id: ID): Promise<boolean>
}

// Implementation
class UserRepository implements Repository<User> {
  constructor(private db: Database) {}

  async findById(id: string): Promise<User | null> {
    return this.db.user.findUnique({ where: { id } })
  }

  // ... other methods
}
```

## React Patterns

### Custom Hooks

```typescript
// Debounce Hook
function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value)

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedValue(value), delay)
    return () => clearTimeout(timer)
  }, [value, delay])

  return debouncedValue
}

// Async State Hook
function useAsync<T>(asyncFn: () => Promise<T>, deps: unknown[] = []) {
  const [state, setState] = useState<{
    data: T | null
    loading: boolean
    error: Error | null
  }>({ data: null, loading: true, error: null })

  useEffect(() => {
    setState(s => ({ ...s, loading: true }))
    asyncFn()
      .then(data => setState({ data, loading: false, error: null }))
      .catch(error => setState({ data: null, loading: false, error }))
  }, deps)

  return state
}
```

### Component Patterns

```typescript
// Compound Component Pattern
const Tabs = ({ children, defaultTab }: TabsProps) => {
  const [activeTab, setActiveTab] = useState(defaultTab)
  return (
    <TabsContext.Provider value={{ activeTab, setActiveTab }}>
      {children}
    </TabsContext.Provider>
  )
}

Tabs.List = TabList
Tabs.Tab = Tab
Tabs.Panel = TabPanel

// Usage
<Tabs defaultTab="settings">
  <Tabs.List>
    <Tabs.Tab id="settings">Settings</Tabs.Tab>
    <Tabs.Tab id="profile">Profile</Tabs.Tab>
  </Tabs.List>
  <Tabs.Panel id="settings">...</Tabs.Panel>
  <Tabs.Panel id="profile">...</Tabs.Panel>
</Tabs>
```

## Service Pattern

```typescript
// Service Layer
class AuthService {
  constructor(
    private userRepo: UserRepository,
    private tokenService: TokenService
  ) {}

  async login(email: string, password: string): Promise<AuthResult> {
    const user = await this.userRepo.findByEmail(email)
    if (!user) throw new NotFoundError('User')

    const valid = await this.verifyPassword(password, user.passwordHash)
    if (!valid) throw new AppError('INVALID_CREDENTIALS', 'Invalid password', 401)

    const token = this.tokenService.generate(user)
    return { user, token }
  }
}
```

## Validation Pattern

```typescript
// Zod Schema Pattern
import { z } from 'zod'

const userSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8).max(100),
  name: z.string().min(1).max(50).optional()
})

type UserInput = z.infer<typeof userSchema>

// Usage with validation
function validateUser(input: unknown): UserInput {
  return userSchema.parse(input)
}
```

## Factory Pattern

```typescript
// Factory for creating instances
interface NotificationSender {
  send(message: string, recipient: string): Promise<void>
}

class NotificationFactory {
  static create(type: 'email' | 'sms' | 'push'): NotificationSender {
    switch (type) {
      case 'email': return new EmailSender()
      case 'sms': return new SmsSender()
      case 'push': return new PushSender()
      default: throw new Error(`Unknown notification type: ${type}`)
    }
  }
}
```

## Skeleton Project Approach

When implementing new features:

```markdown
1. Search for proven skeleton/boilerplate
2. Evaluate options:
   - Security assessment
   - Extensibility analysis
   - Relevance to requirements
   - Community support
3. Clone best match as foundation
4. Iterate within proven structure
5. Customize for specific needs
```

## Pattern Selection Guide

| Scenario | Pattern |
|----------|---------|
| Data access | Repository |
| Business logic | Service |
| Object creation | Factory |
| State management | Custom Hook |
| Complex components | Compound |
| Input validation | Zod Schema |

---

Version: 1.0.0
Source: JikiME-ADK pattern library
