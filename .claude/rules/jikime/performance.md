# Performance Optimization

Guidelines for efficient Claude Code usage and code performance.

## Model Selection Strategy

### Haiku (Fast, Cost-Effective)

**Use for**:
- Simple code generation
- Formatting and linting
- Straightforward Q&A
- Worker agents in multi-agent workflows
- High-frequency, low-complexity tasks

**Characteristics**:
- 90% of Sonnet capability
- 3x cost savings
- Fastest response time

### Sonnet (Balanced, Recommended)

**Use for**:
- Main development work
- Complex coding tasks
- Orchestrating multi-agent workflows
- Code review and analysis
- Most everyday tasks

**Characteristics**:
- Best balance of capability and cost
- Strong coding performance
- Good reasoning ability

### Opus (Maximum Capability)

**Use for**:
- Complex architectural decisions
- Deep reasoning requirements
- Research and analysis
- Critical decision-making
- Ultrathink scenarios

**Characteristics**:
- Maximum reasoning depth
- Best for complex problems
- Higher cost and latency

## Context Window Management

### Critical Zone (80-100%)

Avoid these tasks when context is high:
- Large-scale refactoring
- Multi-file feature implementation
- Complex debugging
- Architectural changes

### Safe Zone (0-60%)

Lower context sensitivity tasks:
- Single-file edits
- Independent utility creation
- Documentation updates
- Simple bug fixes

### Management Strategies

```markdown
1. Start complex tasks with fresh context
2. Use /clear when context exceeds 70%
3. Break large tasks into smaller sessions
4. Summarize findings before context fills
```

## Tool Efficiency

### Parallel Execution

```markdown
DO: Execute independent operations in parallel
- Multiple file reads
- Independent searches
- Unrelated API calls

DON'T: Parallelize dependent operations
- Read then edit same file
- Create then modify directory
- Sequential workflow steps
```

### Tool Selection

| Task | Preferred Tool | Avoid |
|------|----------------|-------|
| Find files | Glob | Bash find |
| Search content | Grep | Bash grep |
| Read files | Read | Bash cat |
| Edit files | Edit | Bash sed |
| Create files | Write | Bash echo |

## Code Performance

### Algorithm Complexity

```markdown
Target:
- O(n) or better for common operations
- O(n log n) acceptable for sorting
- Avoid O(n²) in hot paths
- Never O(2^n) without explicit approval
```

### Common Optimizations

```typescript
// SLOW: O(n²) - Nested loops
items.forEach(item => {
  const match = others.find(o => o.id === item.id)
})

// FAST: O(n) - Map lookup
const othersMap = new Map(others.map(o => [o.id, o]))
items.forEach(item => {
  const match = othersMap.get(item.id)
})
```

### React Performance

```typescript
// Memoize expensive computations
const expensiveValue = useMemo(() =>
  computeExpensive(data), [data]
)

// Memoize callbacks
const handleClick = useCallback(() =>
  doSomething(id), [id]
)

// Avoid unnecessary re-renders
const MemoizedComponent = memo(Component)
```

### Database Performance

```markdown
1. Index frequently queried columns
2. Avoid N+1 queries (use includes/joins)
3. Paginate large result sets
4. Use connection pooling
5. Cache frequently accessed data
```

## Ultrathink + Plan Mode

For complex tasks requiring deep reasoning:

```markdown
1. Activate ultrathink for enhanced analysis
2. Enable Plan Mode for structured approach
3. Multiple critique rounds before implementation
4. Use sub-agents for diverse perspectives
5. Document decisions for future reference
```

### When to Use

- Architecture decisions affecting 3+ files
- Technology selection between options
- Performance vs maintainability trade-offs
- Breaking changes consideration
- Critical bug investigation

## Build Troubleshooting

### Incremental Fix Approach

```markdown
1. Run build, capture errors
2. Fix ONE error at a time
3. Re-run build after each fix
4. Repeat until success
5. Run tests to verify
```

### Common Build Issues

| Error Type | Likely Cause | Fix |
|------------|--------------|-----|
| Type error | Missing/wrong types | Check imports, add types |
| Module not found | Wrong path | Verify import path |
| Syntax error | Typo, missing bracket | Check recent changes |
| Circular dep | Import cycle | Restructure modules |

## Performance Checklist

Before committing:

- [ ] No unnecessary re-renders
- [ ] No N+1 queries
- [ ] Expensive operations memoized
- [ ] Large lists virtualized
- [ ] Images optimized
- [ ] Bundle size reasonable

---

Version: 1.0.0
Source: Adapted from everything-claude-code + expanded guidelines
