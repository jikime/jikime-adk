---
name: jikime-workflow-performance
description: Comprehensive code performance optimization workflow covering profiling, bottleneck analysis, and optimization patterns for frontend, backend, database, and algorithm layers.
version: 1.0.0
tags: ["workflow", "performance", "optimization", "profiling", "bottleneck", "web-vitals"]
triggers:
  keywords: ["performance", "optimize", "slow", "bottleneck", "profiling", "latency", "성능", "최적화", "병목", "느림", "프로파일링"]
  phases: ["run"]
  agents: ["optimizer", "backend", "frontend"]
  languages: ["typescript", "javascript", "python", "go"]
# Progressive Disclosure Configuration
progressive_disclosure:
  enabled: true
  level1_tokens: ~100
  level2_tokens: ~6000
user-invocable: false
context: fork
agent: general-purpose
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Grep
  - Glob
  - Task
  - TodoWrite
---

# Code Performance Optimization Workflow

## Quick Reference (30 seconds)

Performance Optimization Workflow - Systematic approach to identify, analyze, and resolve performance bottlenecks across all application layers.

Core Philosophy:
- **Measure First**: Never optimize without data. Always profile before and after changes.
- **Evidence-Based**: All optimizations must be validated with measurable metrics.
- **Systematic**: Follow the 5-step workflow for consistent results.

Key Capabilities:
- Multi-layer profiling (Frontend, Backend, Database, Algorithm)
- Bottleneck identification and categorization
- Targeted optimization patterns with code examples
- Performance budget management
- Regression prevention strategies

Quick Commands:
- Profile application: `/jikime:verify --focus performance`
- Analyze bottlenecks: Use optimizer agent
- Run benchmarks: `/jikime:test --benchmark`

---

## Implementation Guide (5 minutes)

### The 5-Step Performance Optimization Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│  Step 1: MEASURE          Step 2: IDENTIFY         Step 3: OPTIMIZE  │
│  ┌─────────┐             ┌─────────┐              ┌─────────┐       │
│  │Baseline │─────────────▶│Bottleneck│─────────────▶│ Apply   │       │
│  │Profiling│             │ Analysis │              │ Fixes   │       │
│  └─────────┘             └─────────┘              └─────────┘       │
│       │                                                  │          │
│       │              Step 5: PREVENT              Step 4: VALIDATE  │
│       │              ┌─────────┐              ┌─────────┐          │
│       └──────────────│Regression│◀─────────────│ Compare │          │
│                      │ Guards   │              │ Metrics │          │
│                      └─────────┘              └─────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

### Step 1: Measure Baseline

Before any optimization, capture current performance metrics.

**Profiling Commands by Language:**

```bash
# Node.js / JavaScript
node --prof app.js
node --prof-process isolate*.log > profile.txt

# Node.js with clinic.js (recommended)
npx clinic doctor -- node app.js
npx clinic flame -- node app.js
npx clinic bubbleprof -- node app.js

# Python
python -m cProfile -o profile.stats app.py
python -m pstats profile.stats
# or with py-spy (recommended)
py-spy record -o profile.svg -- python app.py

# Go
go test -cpuprofile cpu.prof -memprofile mem.prof -bench .
go tool pprof cpu.prof

# Web Performance (Lighthouse)
npx lighthouse https://example.com --output=json --output-path=./report.json
```

**Baseline Metrics to Capture:**

| Layer | Metrics | Tools |
|-------|---------|-------|
| Frontend | LCP, FID, CLS, TTFB, Bundle Size | Lighthouse, WebPageTest |
| Backend | P50/P95/P99 Latency, Throughput | APM, custom metrics |
| Database | Query time, Connection pool | EXPLAIN ANALYZE, pg_stat |
| Memory | Heap size, GC frequency | Profiler, heap snapshots |

### Step 2: Identify Bottlenecks

**Bottleneck Categories:**

| Category | Symptoms | Diagnostic Tools | Common Causes |
|----------|----------|------------------|---------------|
| **CPU** | High CPU%, slow computation | Profiler, flame graphs | O(n²) algorithms, blocking ops |
| **Memory** | High RAM, GC pauses, OOM | Heap snapshots, memory profiler | Memory leaks, large objects |
| **I/O** | Slow disk/network, waiting | strace, network inspector | Sync I/O, no connection pooling |
| **Database** | Slow queries, lock contention | EXPLAIN, query analyzer | Missing indexes, N+1 queries |
| **Rendering** | Low FPS, layout thrashing | DevTools Performance | Forced reflows, paint storms |

**Bottleneck Priority Matrix:**

```
Impact
  ▲
  │  ┌─────────────┬─────────────┐
  │  │   HIGH      │  CRITICAL   │
High│  │  Schedule   │  Fix Now    │
  │  │  Soon       │             │
  │  ├─────────────┼─────────────┤
  │  │   LOW       │   MEDIUM    │
Low │  │  Backlog    │  Plan Fix   │
  │  │             │             │
  │  └─────────────┴─────────────┘
  └──────────────────────────────▶ Frequency
       Rare            Frequent
```

### Step 3: Apply Targeted Optimizations

#### Frontend Optimizations

**Bundle Size Reduction:**

```typescript
// ❌ Import entire library (adds 70KB+)
import _ from 'lodash';
import { format } from 'date-fns';

// ✅ Tree-shakeable imports (adds ~2KB each)
import debounce from 'lodash/debounce';
import { format } from 'date-fns/format';

// ✅ Dynamic imports for code splitting
const HeavyComponent = lazy(() => import('./HeavyComponent'));
const ChartLibrary = lazy(() => import('recharts'));
```

**Rendering Optimization:**

```typescript
// ❌ Re-renders on every parent update
function ProductList({ products, filter }) {
  const filtered = products.filter(p => p.category === filter);
  return filtered.map(p => <ProductCard key={p.id} product={p} />);
}

// ✅ Memoized computation + component
const ProductList = memo(function ProductList({ products, filter }) {
  const filtered = useMemo(
    () => products.filter(p => p.category === filter),
    [products, filter]
  );

  return filtered.map(p => <ProductCard key={p.id} product={p} />);
});

// ✅ Virtualized list for 1000+ items
import { FixedSizeList } from 'react-window';

function VirtualizedList({ items }) {
  return (
    <FixedSizeList
      height={600}
      itemCount={items.length}
      itemSize={50}
      width="100%"
    >
      {({ index, style }) => (
        <div style={style}>{items[index].name}</div>
      )}
    </FixedSizeList>
  );
}
```

**Image Optimization:**

```html
<!-- ❌ Unoptimized -->
<img src="hero.jpg" />

<!-- ✅ Fully optimized -->
<img
  src="hero.webp"
  srcset="hero-400.webp 400w, hero-800.webp 800w, hero-1200.webp 1200w"
  sizes="(max-width: 600px) 400px, (max-width: 1200px) 800px, 1200px"
  loading="lazy"
  decoding="async"
  alt="Hero image"
  width="1200"
  height="600"
/>

<!-- ✅ Next.js Image component -->
<Image
  src="/hero.jpg"
  alt="Hero"
  width={1200}
  height={600}
  priority={false}
  placeholder="blur"
/>
```

#### Backend Optimizations

**N+1 Query Resolution:**

```typescript
// ❌ N+1 Problem - 101 queries for 100 users
async function getUsers() {
  const users = await db.user.findMany();
  for (const user of users) {
    user.orders = await db.order.findMany({ where: { userId: user.id } });
  }
  return users;
}

// ✅ Eager loading - 1 query with JOIN
async function getUsers() {
  return db.user.findMany({
    include: { orders: true }
  });
}

// ✅ DataLoader for GraphQL (batching)
const userLoader = new DataLoader(async (userIds) => {
  const users = await db.user.findMany({
    where: { id: { in: userIds } }
  });
  return userIds.map(id => users.find(u => u.id === id));
});
```

**Multi-Layer Caching:**

```typescript
// 3-tier caching strategy
class CacheService {
  private memoryCache = new Map<string, { data: any; expiry: number }>();

  async get<T>(key: string): Promise<T | null> {
    // L1: In-memory (fastest, ~0.1ms)
    const memory = this.memoryCache.get(key);
    if (memory && memory.expiry > Date.now()) {
      return memory.data;
    }

    // L2: Redis (fast, ~1-5ms)
    const redis = await this.redis.get(key);
    if (redis) {
      const data = JSON.parse(redis);
      this.memoryCache.set(key, { data, expiry: Date.now() + 60000 });
      return data;
    }

    // L3: Database (slow, ~10-100ms)
    return null;
  }

  async set<T>(key: string, data: T, ttl: number): Promise<void> {
    // Write to all layers
    this.memoryCache.set(key, { data, expiry: Date.now() + 60000 });
    await this.redis.setex(key, ttl, JSON.stringify(data));
  }
}
```

**Async Processing with Queue:**

```typescript
// ❌ Blocking operation (user waits 5+ minutes)
app.post('/upload', async (req, res) => {
  await processVideo(req.file);  // Takes 5 minutes
  res.send('Done');
});

// ✅ Queue for background processing
import { Queue, Worker } from 'bullmq';

const videoQueue = new Queue('video-processing');

app.post('/upload', async (req, res) => {
  const job = await videoQueue.add('process', {
    fileId: req.file.id,
    userId: req.user.id
  });
  res.json({ jobId: job.id, status: 'processing' });
});

// Separate worker process
new Worker('video-processing', async (job) => {
  await processVideo(job.data.fileId);
  await notifyUser(job.data.userId, 'Video ready!');
});
```

#### Algorithm Optimizations

```typescript
// ❌ O(n²) - Nested loops
function findDuplicates(arr: number[]): number[] {
  const duplicates: number[] = [];
  for (let i = 0; i < arr.length; i++) {
    for (let j = i + 1; j < arr.length; j++) {
      if (arr[i] === arr[j] && !duplicates.includes(arr[i])) {
        duplicates.push(arr[i]);
      }
    }
  }
  return duplicates;
}

// ✅ O(n) - Hash map
function findDuplicates(arr: number[]): number[] {
  const seen = new Set<number>();
  const duplicates = new Set<number>();
  for (const item of arr) {
    if (seen.has(item)) duplicates.add(item);
    seen.add(item);
  }
  return [...duplicates];
}

// ❌ O(n) lookup in array
const users = [{ id: 1, name: 'A' }, { id: 2, name: 'B' }, ...];
const user = users.find(u => u.id === targetId);

// ✅ O(1) lookup with Map
const userMap = new Map(users.map(u => [u.id, u]));
const user = userMap.get(targetId);
```

### Step 4: Validate Improvements

**Performance Targets:**

| Metric | Good | Needs Work | Poor |
|--------|------|------------|------|
| **LCP** (Largest Contentful Paint) | < 2.5s | 2.5-4s | > 4s |
| **FID** (First Input Delay) | < 100ms | 100-300ms | > 300ms |
| **CLS** (Cumulative Layout Shift) | < 0.1 | 0.1-0.25 | > 0.25 |
| **TTFB** (Time to First Byte) | < 800ms | 800ms-1.8s | > 1.8s |
| **INP** (Interaction to Next Paint) | < 200ms | 200-500ms | > 500ms |

**API Performance Targets:**

| Metric | Target | Alert Threshold |
|--------|--------|-----------------|
| P50 Latency | < 100ms | > 200ms |
| P95 Latency | < 500ms | > 1s |
| P99 Latency | < 1s | > 2s |
| Error Rate | < 0.1% | > 1% |
| Throughput | Baseline +10% | Baseline -10% |

### Step 5: Prevent Regressions

**Performance Budget Configuration:**

```json
// .performance-budget.json
{
  "budgets": [
    {
      "resourceType": "script",
      "budget": 300000
    },
    {
      "resourceType": "total",
      "budget": 1000000
    },
    {
      "timingMetric": "largest-contentful-paint",
      "budget": 2500
    },
    {
      "timingMetric": "cumulative-layout-shift",
      "budget": 0.1
    }
  ]
}
```

**CI Integration:**

```yaml
# .github/workflows/performance.yml
performance-check:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Lighthouse CI
      uses: treosh/lighthouse-ci-action@v11
      with:
        configPath: ./lighthouserc.json
        uploadArtifacts: true
        temporaryPublicStorage: true
    - name: Bundle Size Check
      run: npx bundlesize
```

---

## Advanced Implementation (10+ minutes)

### Real-Time Performance Monitoring

```typescript
// Custom performance observer
const observer = new PerformanceObserver((list) => {
  for (const entry of list.getEntries()) {
    if (entry.entryType === 'largest-contentful-paint') {
      analytics.track('LCP', { value: entry.startTime });
    }
    if (entry.entryType === 'first-input') {
      analytics.track('FID', { value: entry.processingStart - entry.startTime });
    }
  }
});

observer.observe({ entryTypes: ['largest-contentful-paint', 'first-input', 'layout-shift'] });
```

### Database Query Optimization

```sql
-- Analyze slow queries
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT u.*, COUNT(o.id) as order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.created_at > '2024-01-01'
GROUP BY u.id;

-- Add missing index
CREATE INDEX CONCURRENTLY idx_orders_user_id ON orders(user_id);
CREATE INDEX CONCURRENTLY idx_users_created_at ON users(created_at);

-- Partial index for frequent queries
CREATE INDEX idx_active_users ON users(email) WHERE status = 'active';
```

---

## Works Well With

Commands:
- `/jikime:verify --focus performance` - Performance-focused verification
- `/jikime:test --benchmark` - Run performance benchmarks

Skills:
- `jikime-domain-backend` - Backend optimization patterns
- `jikime-domain-frontend` - Frontend optimization patterns
- `jikime-domain-database` - Database query optimization

Agents:
- `optimizer` - Performance optimization specialist
- `backend` - Server-side performance
- `frontend` - Client-side performance

---

## Quick Decision Guide

| Symptom | Likely Cause | First Action |
|---------|--------------|--------------|
| Slow page load | Large bundle, unoptimized images | Check bundle size, add lazy loading |
| Laggy interactions | Main thread blocking | Profile with DevTools, add virtualization |
| API timeouts | N+1 queries, no caching | Check query count, add Redis cache |
| Memory growth | Leaks, large objects | Take heap snapshots, check event listeners |
| High CPU | O(n²) algorithms, sync ops | Profile hotspots, optimize algorithms |

---

## Validation Checklist

```
Performance Optimization Validation:
- [ ] Baseline metrics captured before changes
- [ ] Bottleneck identified with profiling data
- [ ] Optimization applied to correct layer
- [ ] Metrics improved from baseline (quantified)
- [ ] No functionality regressions introduced
- [ ] Performance budget not exceeded
- [ ] CI checks pass (Lighthouse, bundle size)
- [ ] Documentation updated with findings
```

---

Version: 1.0.0
Last Updated: 2026-01-25
Integration Status: Complete - Full performance optimization workflow
