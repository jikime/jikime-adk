# React CRA/Vite → Next.js Migration Patterns

This module provides detailed conversion patterns from React (Create React App or Vite) to Next.js 16 with App Router.

## Project Structure Migration

### CRA Structure → Next.js App Router

**Before (CRA)**:
```
src/
├── index.tsx           # Entry point
├── App.tsx             # Root component
├── components/
│   ├── Header.tsx
│   └── Footer.tsx
├── pages/
│   ├── Home.tsx
│   ├── About.tsx
│   └── UserDetail.tsx
├── hooks/
│   └── useAuth.ts
├── context/
│   └── AuthContext.tsx
├── services/
│   └── api.ts
└── styles/
    └── global.css
```

**After (Next.js)**:
```
src/
├── app/
│   ├── layout.tsx      # Root layout (replaces index.tsx + App.tsx)
│   ├── page.tsx        # Home page
│   ├── about/
│   │   └── page.tsx
│   ├── users/
│   │   └── [id]/
│   │       └── page.tsx
│   └── globals.css
├── components/
│   ├── header.tsx
│   └── footer.tsx
├── hooks/
│   └── use-auth.ts
├── lib/
│   ├── auth-context.tsx
│   └── api.ts
└── stores/             # If using Zustand
    └── auth-store.ts
```

## Entry Point Migration

### index.tsx + App.tsx → layout.tsx

**Before (CRA)**:
```tsx
// index.tsx
import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import App from './App'
import './styles/global.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <AuthProvider>
        <App />
      </AuthProvider>
    </BrowserRouter>
  </React.StrictMode>
)

// App.tsx
import { Routes, Route } from 'react-router-dom'
import { Header } from './components/Header'
import { Footer } from './components/Footer'
import { Home } from './pages/Home'
import { About } from './pages/About'

export default function App() {
  return (
    <div className="app">
      <Header />
      <main>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/about" element={<About />} />
        </Routes>
      </main>
      <Footer />
    </div>
  )
}
```

**After (Next.js)**:
```tsx
// app/layout.tsx
import type { Metadata } from 'next'
import { Header } from '@/components/header'
import { Footer } from '@/components/footer'
import { AuthProvider } from '@/lib/auth-context'
import './globals.css'

export const metadata: Metadata = {
  title: 'My App',
  description: 'App description',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>
        <AuthProvider>
          <div className="app">
            <Header />
            <main>{children}</main>
            <Footer />
          </div>
        </AuthProvider>
      </body>
    </html>
  )
}
```

## Routing Migration

### React Router → App Router

**Before (react-router-dom)**:
```tsx
// routes.tsx
import { Routes, Route, Navigate } from 'react-router-dom'

function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/about" element={<About />} />
      <Route path="/users" element={<UserList />} />
      <Route path="/users/:id" element={<UserDetail />} />
      <Route path="/dashboard/*" element={<DashboardRoutes />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

// Nested routes
function DashboardRoutes() {
  return (
    <Routes>
      <Route index element={<DashboardHome />} />
      <Route path="settings" element={<Settings />} />
      <Route path="profile" element={<Profile />} />
    </Routes>
  )
}
```

**After (Next.js File-based)**:
```
app/
├── page.tsx                    # /
├── about/
│   └── page.tsx                # /about
├── users/
│   ├── page.tsx                # /users
│   └── [id]/
│       └── page.tsx            # /users/:id
├── dashboard/
│   ├── layout.tsx              # Dashboard layout
│   ├── page.tsx                # /dashboard
│   ├── settings/
│   │   └── page.tsx            # /dashboard/settings
│   └── profile/
│       └── page.tsx            # /dashboard/profile
└── not-found.tsx               # 404 page
```

### Navigation Hooks

**Before**:
```tsx
import { useNavigate, useLocation, useParams, useSearchParams } from 'react-router-dom'

function Component() {
  const navigate = useNavigate()
  const location = useLocation()
  const { id } = useParams()
  const [searchParams, setSearchParams] = useSearchParams()

  const handleClick = () => {
    navigate('/users', { state: { from: location.pathname } })
  }

  const updateFilter = (filter: string) => {
    setSearchParams({ filter })
  }
}
```

**After**:
```tsx
'use client'

import { useRouter, usePathname, useParams, useSearchParams } from 'next/navigation'

function Component() {
  const router = useRouter()
  const pathname = usePathname()
  const params = useParams()
  const searchParams = useSearchParams()

  const handleClick = () => {
    // Note: Next.js doesn't have location.state
    // Use query params or cookies instead
    router.push('/users')
  }

  const updateFilter = (filter: string) => {
    const params = new URLSearchParams(searchParams)
    params.set('filter', filter)
    router.push(`${pathname}?${params.toString()}`)
  }
}
```

### Link Component

**Before**:
```tsx
import { Link, NavLink } from 'react-router-dom'

<Link to="/about">About</Link>
<Link to={`/users/${user.id}`}>View User</Link>
<NavLink
  to="/dashboard"
  className={({ isActive }) => isActive ? 'active' : ''}
>
  Dashboard
</NavLink>
```

**After**:
```tsx
'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'

<Link href="/about">About</Link>
<Link href={`/users/${user.id}`}>View User</Link>

// NavLink equivalent
function NavLink({ href, children }: { href: string; children: React.ReactNode }) {
  const pathname = usePathname()
  const isActive = pathname === href

  return (
    <Link href={href} className={isActive ? 'active' : ''}>
      {children}
    </Link>
  )
}
```

## Data Fetching

### useEffect + fetch → Server Components

**Before (CRA)**:
```tsx
import { useState, useEffect } from 'react'

function UserList() {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function fetchUsers() {
      try {
        const response = await fetch('/api/users')
        const data = await response.json()
        setUsers(data)
      } catch (err) {
        setError('Failed to fetch users')
      } finally {
        setLoading(false)
      }
    }

    fetchUsers()
  }, [])

  if (loading) return <div>Loading...</div>
  if (error) return <div>{error}</div>

  return (
    <ul>
      {users.map(user => (
        <li key={user.id}>{user.name}</li>
      ))}
    </ul>
  )
}
```

**After (Next.js Server Component)**:
```tsx
// app/users/page.tsx
// This is a Server Component by default
async function getUsers(): Promise<User[]> {
  const response = await fetch('https://api.example.com/users', {
    cache: 'no-store' // or 'force-cache' for caching
  })

  if (!response.ok) {
    throw new Error('Failed to fetch users')
  }

  return response.json()
}

export default async function UserListPage() {
  const users = await getUsers()

  return (
    <ul>
      {users.map(user => (
        <li key={user.id}>{user.name}</li>
      ))}
    </ul>
  )
}
```

### Client-Side Data Fetching (when needed)

**After (Next.js with React Query/SWR)**:
```tsx
'use client'

import useSWR from 'swr'

const fetcher = (url: string) => fetch(url).then(res => res.json())

function UserList() {
  const { data: users, error, isLoading } = useSWR<User[]>('/api/users', fetcher)

  if (isLoading) return <div>Loading...</div>
  if (error) return <div>Failed to load</div>

  return (
    <ul>
      {users?.map(user => (
        <li key={user.id}>{user.name}</li>
      ))}
    </ul>
  )
}
```

## Environment Variables

**Before (CRA)**:
```typescript
// Must prefix with REACT_APP_
const apiUrl = process.env.REACT_APP_API_URL
const apiKey = process.env.REACT_APP_API_KEY
```

**After (Next.js)**:
```typescript
// Client-side: prefix with NEXT_PUBLIC_
const apiUrl = process.env.NEXT_PUBLIC_API_URL

// Server-side only (no prefix needed)
const apiKey = process.env.API_KEY // Only accessible in Server Components/API routes
```

## API Routes

### Express/Custom Backend → Next.js API Routes

**Before (Separate Express server or CRA proxy)**:
```typescript
// server.js
app.get('/api/users', async (req, res) => {
  const users = await db.users.findMany()
  res.json(users)
})

app.post('/api/users', async (req, res) => {
  const user = await db.users.create(req.body)
  res.json(user)
})
```

**After (Next.js Route Handlers)**:
```typescript
// app/api/users/route.ts
import { NextRequest, NextResponse } from 'next/server'

export async function GET() {
  const users = await db.users.findMany()
  return NextResponse.json(users)
}

export async function POST(request: NextRequest) {
  const body = await request.json()
  const user = await db.users.create({ data: body })
  return NextResponse.json(user, { status: 201 })
}
```

## Server Actions (Form Handling)

**Before (CRA)**:
```tsx
function ContactForm() {
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)

    const formData = new FormData(e.target as HTMLFormElement)
    await fetch('/api/contact', {
      method: 'POST',
      body: JSON.stringify(Object.fromEntries(formData)),
      headers: { 'Content-Type': 'application/json' }
    })

    setLoading(false)
  }

  return (
    <form onSubmit={handleSubmit}>
      <input name="email" type="email" required />
      <textarea name="message" required />
      <button type="submit" disabled={loading}>
        {loading ? 'Sending...' : 'Send'}
      </button>
    </form>
  )
}
```

**After (Next.js Server Actions)**:
```tsx
// app/contact/page.tsx
import { submitContact } from './actions'

export default function ContactPage() {
  return (
    <form action={submitContact}>
      <input name="email" type="email" required />
      <textarea name="message" required />
      <button type="submit">Send</button>
    </form>
  )
}

// app/contact/actions.ts
'use server'

import { revalidatePath } from 'next/cache'

export async function submitContact(formData: FormData) {
  const email = formData.get('email') as string
  const message = formData.get('message') as string

  await db.contacts.create({
    data: { email, message }
  })

  revalidatePath('/contact')
}
```

## Protected Routes

**Before (CRA with React Router)**:
```tsx
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, loading } = useAuth()
  const location = useLocation()

  if (loading) return <div>Loading...</div>

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  return <>{children}</>
}

// Usage
<Route
  path="/dashboard"
  element={
    <ProtectedRoute>
      <Dashboard />
    </ProtectedRoute>
  }
/>
```

**After (Next.js Middleware)**:
```typescript
// middleware.ts
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
  const token = request.cookies.get('auth-token')

  if (!token && request.nextUrl.pathname.startsWith('/dashboard')) {
    return NextResponse.redirect(new URL('/login', request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/dashboard/:path*']
}
```

## CSS/Styling

### CSS Modules (same in both)

```tsx
// Works the same way
import styles from './component.module.css'

<div className={styles.container}>Content</div>
```

### Tailwind Integration

**Before (CRA with manual setup)**:
```js
// postcss.config.js, tailwind.config.js needed
// Manual configuration in index.css
```

**After (Next.js built-in)**:
```typescript
// tailwind.config.ts - auto-generated by create-next-app
import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    './src/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}

export default config
```

## Image Optimization

**Before (CRA)**:
```tsx
<img src="/images/hero.jpg" alt="Hero" width={800} height={600} />
```

**After (Next.js Image)**:
```tsx
import Image from 'next/image'

<Image
  src="/images/hero.jpg"
  alt="Hero"
  width={800}
  height={600}
  priority // for LCP images
/>
```

## Head/Meta Tags

**Before (react-helmet)**:
```tsx
import { Helmet } from 'react-helmet'

function Page() {
  return (
    <>
      <Helmet>
        <title>Page Title</title>
        <meta name="description" content="Page description" />
      </Helmet>
      <div>Content</div>
    </>
  )
}
```

**After (Next.js Metadata)**:
```tsx
// app/page.tsx
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Page Title',
  description: 'Page description',
}

export default function Page() {
  return <div>Content</div>
}
```

---

Version: 1.0.0
