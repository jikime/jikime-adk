---
name: jikime-nextjs@14
description: Next.js 14 App Router baseline guide. Core patterns, conventions, and best practices for Next.js 14.x applications.
tags: ["framework", "nextjs", "version", "app-router", "server-components", "server-actions"]
triggers:
  keywords: ["nextjs", "next.js 14", "app-router", "server-components", "server-actions"]
  phases: ["run"]
  agents: ["expert-frontend"]
  languages: ["typescript"]
type: version
framework: nextjs
version: "14"
user-invocable: false
---

# Next.js 14 Baseline Guide

Next.js 14 App Router의 핵심 패턴과 규칙을 정의합니다. 버전 업그레이드 시 기준점으로 사용됩니다.

## Version Info

| 항목 | 값 |
|------|-----|
| Version | 14.0.0 ~ 14.2.x |
| Release Date | October 2023 |
| Node.js | 18.17+ |
| React | 18.2+ |

---

## Core Features (Next.js 14)

### 1. App Router (Stable)

```
app/
├── layout.tsx          # Root layout (required)
├── page.tsx            # Home page
├── loading.tsx         # Loading UI
├── error.tsx           # Error boundary
├── not-found.tsx       # 404 page
└── [slug]/
    └── page.tsx        # Dynamic route
```

### 2. Server Components (Default)

```tsx
// app/users/page.tsx - Server Component by default
async function getUsers() {
  const res = await fetch('https://api.example.com/users')
  return res.json()
}

export default async function UsersPage() {
  const users = await getUsers()
  return <UserList users={users} />
}
```

### 3. Client Components

```tsx
// components/counter.tsx
'use client'

import { useState } from 'react'

export function Counter() {
  const [count, setCount] = useState(0)
  return <button onClick={() => setCount(c => c + 1)}>Count: {count}</button>
}
```

### 4. Server Actions (Stable in 14.0)

```tsx
// app/actions.ts
'use server'

export async function createPost(formData: FormData) {
  const title = formData.get('title')
  await db.post.create({ data: { title } })
  revalidatePath('/posts')
}

// app/posts/new/page.tsx
import { createPost } from '../actions'

export default function NewPost() {
  return (
    <form action={createPost}>
      <input name="title" />
      <button type="submit">Create</button>
    </form>
  )
}
```

### 5. Metadata API

```tsx
// Static metadata
export const metadata = {
  title: 'My App',
  description: 'App description',
}

// Dynamic metadata
export async function generateMetadata({ params }) {
  const post = await getPost(params.slug)
  return { title: post.title }
}
```

### 6. Route Handlers

```tsx
// app/api/users/route.ts
import { NextRequest, NextResponse } from 'next/server'

export async function GET(request: NextRequest) {
  const users = await db.user.findMany()
  return NextResponse.json(users)
}

export async function POST(request: NextRequest) {
  const body = await request.json()
  const user = await db.user.create({ data: body })
  return NextResponse.json(user, { status: 201 })
}
```

---

## Data Fetching (Next.js 14)

### Fetch with Caching

```tsx
// Default: cached (equivalent to force-cache)
const data = await fetch('https://api.example.com/data')

// No caching
const data = await fetch('https://api.example.com/data', {
  cache: 'no-store'
})

// Time-based revalidation
const data = await fetch('https://api.example.com/data', {
  next: { revalidate: 3600 }  // Revalidate every hour
})

// Tag-based revalidation
const data = await fetch('https://api.example.com/data', {
  next: { tags: ['posts'] }
})
```

### Revalidation

```tsx
import { revalidatePath, revalidateTag } from 'next/cache'

// Path-based
revalidatePath('/posts')
revalidatePath('/posts/[slug]', 'page')

// Tag-based
revalidateTag('posts')
```

---

## Dynamic Routes (Next.js 14)

### Params Access (Synchronous)

```tsx
// app/posts/[slug]/page.tsx
type Props = {
  params: { slug: string }  // Direct access (synchronous)
  searchParams: { [key: string]: string | string[] | undefined }
}

export default function PostPage({ params, searchParams }: Props) {
  const { slug } = params  // Direct destructuring
  const { sort } = searchParams

  return <div>Post: {slug}</div>
}
```

### generateStaticParams

```tsx
export async function generateStaticParams() {
  const posts = await getPosts()
  return posts.map(post => ({ slug: post.slug }))
}
```

---

## Middleware (Next.js 14)

```tsx
// middleware.ts
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
  const token = request.cookies.get('token')

  if (!token && request.nextUrl.pathname.startsWith('/dashboard')) {
    return NextResponse.redirect(new URL('/login', request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/dashboard/:path*', '/api/:path*']
}
```

---

## Configuration (Next.js 14)

### next.config.js

```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    remotePatterns: [
      { protocol: 'https', hostname: 'example.com' }
    ]
  },
  experimental: {
    serverActions: true,  // Enabled by default in 14.0+
  }
}

module.exports = nextConfig
```

---

## Key Limitations (14.x)

| 기능 | 상태 |
|------|------|
| Server Actions | Stable |
| Partial Prerendering | Experimental |
| Turbopack | Dev only (unstable) |
| `params` access | Synchronous |
| `'use cache'` | Not available |

---

## Upgrade Path

**Next.js 14 → 15**: See `jikime-nextjs@15`
- `params`/`searchParams` become async (Promise)
- Turbopack becomes stable for dev
- fetch caching default changes

---

Version: 1.0.0
Last Updated: 2026-01-22
