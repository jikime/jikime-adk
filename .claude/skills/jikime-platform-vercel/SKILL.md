---
name: jikime-platform-vercel
description: Vercel edge deployment specialist covering Edge Functions, Next.js optimization, preview deployments, ISR, and storage solutions. Use when deploying Next.js applications, implementing edge computing, or configuring Vercel platform features.
user-invocable: false
---

# Vercel Deployment Guide

Vercel + Next.js 배포를 위한 간결한 가이드.

## Quick Reference

| 기능 | 설명 |
|------|------|
| **Edge Functions** | 글로벌 엣지에서 실행, 50ms 이하 콜드스타트 |
| **ISR** | 점진적 정적 재생성 |
| **Preview** | PR별 자동 프리뷰 URL |
| **KV** | Redis 호환 키-값 스토어 |
| **Blob** | 파일 스토리지 |

## Configuration

### vercel.json

```json
{
  "$schema": "https://openapi.vercel.sh/vercel.json",
  "framework": "nextjs",
  "regions": ["icn1", "hnd1"],
  "functions": {
    "app/api/**/*.ts": {
      "memory": 1024,
      "maxDuration": 30
    }
  },
  "headers": [
    {
      "source": "/api/(.*)",
      "headers": [
        { "key": "Cache-Control", "value": "s-maxage=60, stale-while-revalidate" }
      ]
    }
  ]
}
```

## Edge Functions

### Basic Edge Route

```typescript
// app/api/geo/route.ts
export const runtime = 'edge';
export const preferredRegion = ['icn1', 'hnd1'];

export async function GET(request: Request) {
  const country = request.geo?.country ?? 'Unknown';
  const city = request.geo?.city ?? 'Unknown';

  return Response.json({
    country,
    city,
    timestamp: new Date().toISOString(),
  });
}
```

### Edge Middleware

```typescript
// middleware.ts
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  // Geo-based redirect
  const country = request.geo?.country;
  if (country === 'KR' && !request.nextUrl.pathname.startsWith('/ko')) {
    return NextResponse.redirect(new URL('/ko', request.url));
  }

  // Add custom header
  const response = NextResponse.next();
  response.headers.set('x-custom-header', 'value');

  return response;
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
};
```

## ISR (Incremental Static Regeneration)

### Time-based Revalidation

```typescript
// app/posts/page.tsx
export const revalidate = 60; // 60초마다 재생성

export default async function PostsPage() {
  const posts = await fetchPosts();
  return <PostList posts={posts} />;
}
```

### On-demand Revalidation

```typescript
// app/api/revalidate/route.ts
import { revalidateTag, revalidatePath } from 'next/cache';
import { NextRequest } from 'next/server';

export async function POST(request: NextRequest) {
  const { tag, path, secret } = await request.json();

  if (secret !== process.env.REVALIDATION_SECRET) {
    return Response.json({ error: 'Invalid secret' }, { status: 401 });
  }

  if (tag) {
    revalidateTag(tag);
  }
  if (path) {
    revalidatePath(path);
  }

  return Response.json({ revalidated: true, now: Date.now() });
}
```

### Cache Tags

```typescript
// lib/api.ts
export async function getPosts() {
  const res = await fetch('https://api.example.com/posts', {
    next: { tags: ['posts'] },
  });
  return res.json();
}

export async function getPost(id: string) {
  const res = await fetch(`https://api.example.com/posts/${id}`, {
    next: { tags: ['posts', `post-${id}`] },
  });
  return res.json();
}
```

## Vercel KV

```typescript
// lib/kv.ts
import { kv } from '@vercel/kv';

// Set
await kv.set('user:123', { name: 'John', email: 'john@example.com' });

// Get
const user = await kv.get('user:123');

// Set with expiry (TTL)
await kv.set('session:abc', { userId: '123' }, { ex: 3600 }); // 1시간

// Delete
await kv.del('user:123');

// Increment
await kv.incr('page:views');

// List operations
await kv.lpush('queue:tasks', task);
const task = await kv.rpop('queue:tasks');
```

## Vercel Blob

```typescript
// app/api/upload/route.ts
import { put, del } from '@vercel/blob';

export async function POST(request: Request) {
  const formData = await request.formData();
  const file = formData.get('file') as File;

  const blob = await put(file.name, file, {
    access: 'public',
  });

  return Response.json(blob);
}

export async function DELETE(request: Request) {
  const { url } = await request.json();
  await del(url);
  return Response.json({ deleted: true });
}
```

## Analytics & Speed Insights

```typescript
// app/layout.tsx
import { Analytics } from '@vercel/analytics/react';
import { SpeedInsights } from '@vercel/speed-insights/next';

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        {children}
        <Analytics />
        <SpeedInsights />
      </body>
    </html>
  );
}
```

## Environment Variables

```bash
# Vercel CLI
vercel env add NEXT_PUBLIC_API_URL production
vercel env pull .env.local

# Preview-specific
vercel env add DATABASE_URL preview
```

## Deployment

```bash
# CLI 배포
vercel              # Preview
vercel --prod       # Production

# Git 연동
# main → Production
# PR → Preview (자동)
```

## Best Practices

- **Edge First**: 가능하면 Edge Runtime 사용
- **ISR**: 정적 페이지에 ISR 활용
- **Cache Tags**: 세밀한 캐시 무효화
- **Region**: 사용자 위치에 가까운 리전 선택
- **Middleware**: 인증, 리다이렉트에 활용

---

Last Updated: 2026-01-21
Version: 2.0.0
