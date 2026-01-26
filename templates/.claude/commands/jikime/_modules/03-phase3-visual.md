
---

## Phase 3: Visual Regression Testing

**Activation**: `--visual` or `--full` flag

### Overview

Visual regression testing captures screenshots of each route on both source and target servers, then performs pixel-by-pixel comparison to detect unintended visual changes during migration.

### Step 3.1: Screenshot Capture Engine

**Capture Strategy:**

For each route in route_registry, capture screenshots on both servers across multiple viewports.

```
FUNCTION captureScreenshots(route_registry, source_url, target_url, viewports):
  screenshots = []

  FOR each route in route_registry.static_routes + route_registry.dynamic_routes:
    FOR each viewport in viewports:
      // Capture source
      source_screenshot = capture(source_url + route.path, viewport, "source")
      // Capture target
      target_screenshot = capture(target_url + route.path, viewport, "target")

      screenshots.add({
        route: route.path,
        viewport: viewport.name,
        source: source_screenshot,
        target: target_screenshot
      })

  RETURN screenshots
```

**Capture Function:**

```typescript
async function capturePageScreenshot(
  page: Page,
  url: string,
  viewport: { width: number; height: number },
  maskSelectors: string[]
): Promise<Buffer> {
  // Set viewport
  await page.setViewportSize(viewport)

  // Navigate and wait for stable state
  await page.goto(url, { waitUntil: 'networkidle', timeout: 30000 })

  // Wait for animations to settle
  await page.waitForTimeout(1000)

  // Hide dynamic content (timestamps, ads, avatars)
  for (const selector of maskSelectors) {
    await page.locator(selector).evaluateAll(elements => {
      elements.forEach(el => {
        (el as HTMLElement).style.visibility = 'hidden'
      })
    }).catch(() => {}) // Ignore if selector not found
  }

  // Capture full page screenshot
  return await page.screenshot({
    fullPage: true,
    type: 'png'
  })
}
```

**Screenshot Storage:**

```
{artifacts_dir}/screenshots/
├── source/
│   ├── desktop/
│   │   ├── home.png
│   │   ├── dashboard.png
│   │   └── settings.png
│   ├── tablet/
│   │   └── ...
│   └── mobile/
│       └── ...
├── target/
│   ├── desktop/
│   │   └── ...
│   ├── tablet/
│   │   └── ...
│   └── mobile/
│       └── ...
└── diff/
    ├── desktop/
    │   ├── home-diff.png
    │   ├── dashboard-diff.png
    │   └── settings-diff.png
    ├── tablet/
    │   └── ...
    └── mobile/
        └── ...
```

### Step 3.2: Pixel Comparison Engine

**Comparison Algorithm:**

```typescript
interface ComparisonResult {
  route: string
  viewport: string
  diffPercentage: number
  diffPixels: number
  totalPixels: number
  passed: boolean
  diffImagePath: string | null
}

async function compareScreenshots(
  sourceBuffer: Buffer,
  targetBuffer: Buffer,
  threshold: number  // percentage (default: 5)
): Promise<ComparisonResult> {
  // Decode PNG buffers to pixel arrays
  const sourcePixels = decodePNG(sourceBuffer)
  const targetPixels = decodePNG(targetBuffer)

  // Handle size differences
  const width = Math.max(sourcePixels.width, targetPixels.width)
  const height = Math.max(sourcePixels.height, targetPixels.height)
  const totalPixels = width * height

  let diffPixels = 0
  const diffImage = createEmptyImage(width, height)

  // Pixel-by-pixel comparison
  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      const sourceColor = getPixel(sourcePixels, x, y)
      const targetColor = getPixel(targetPixels, x, y)

      if (!colorsMatch(sourceColor, targetColor, colorThreshold: 10)) {
        diffPixels++
        setPixel(diffImage, x, y, RED)  // Mark diff in red
      } else {
        setPixel(diffImage, x, y, dimmed(targetColor))  // Dimmed original
      }
    }
  }

  const diffPercentage = (diffPixels / totalPixels) * 100
  const passed = diffPercentage <= threshold

  return {
    diffPercentage: round(diffPercentage, 2),
    diffPixels,
    totalPixels,
    passed,
    diffImagePath: passed ? null : saveDiffImage(diffImage)
  }
}
```

**Color Matching with Tolerance:**

```
FUNCTION colorsMatch(color1, color2, tolerance=10):
  // Allow slight color variations (anti-aliasing, rendering differences)
  RETURN abs(color1.r - color2.r) <= tolerance
     AND abs(color1.g - color2.g) <= tolerance
     AND abs(color1.b - color2.b) <= tolerance
     AND abs(color1.a - color2.a) <= tolerance
```

### Step 3.3: Responsive Viewport Matrix

**Default Viewports:**

| Name | Width | Height | Represents |
|------|-------|--------|------------|
| Desktop | 1920 | 1080 | Standard monitor |
| Laptop | 1366 | 768 | Common laptop |
| Tablet | 768 | 1024 | iPad portrait |
| Mobile | 375 | 812 | iPhone X/12/13 |
| Mobile Small | 320 | 568 | iPhone SE |

**Viewport Execution:**

```
IF --full:
  → Run all 5 viewports
ELIF --visual (no --cross-browser):
  → Run Desktop + Tablet + Mobile (3 viewports)
ELSE:
  → Run Desktop only (1 viewport)
```

**Responsive-Specific Checks:**

```typescript
// Check for horizontal overflow (common migration issue)
async function checkHorizontalOverflow(page: Page): Promise<boolean> {
  return await page.evaluate(() => {
    return document.documentElement.scrollWidth > document.documentElement.clientWidth
  })
}

// Check for overlapping elements
async function checkElementOverlap(page: Page): Promise<string[]> {
  return await page.evaluate(() => {
    const elements = document.querySelectorAll('*')
    const overlaps: string[] = []
    // ... bounding rect comparison logic
    return overlaps
  })
}
```

### Step 3.4: Threshold & Masking Configuration

**Threshold Levels:**

| Level | Percentage | Use Case |
|-------|-----------|----------|
| Strict | 1% | Pixel-perfect migrations (same CSS framework) |
| Normal | 5% | Default - allows minor rendering differences |
| Relaxed | 10% | Framework changes with known visual differences |
| Loose | 20% | Major redesigns where layout is similar |

**Configuration from `.migrate-config.yaml`:**

```yaml
verification:
  visual_threshold: 5          # Default threshold
  mask_selectors:              # Elements to hide before comparison
    - "[data-testid='timestamp']"
    - "[data-testid='random-id']"
    - ".ad-banner"
    - ".user-avatar"
    - "[class*='skeleton']"
    - ".loading-spinner"
    - "time"                   # All <time> elements
    - "[datetime]"             # Elements with datetime attr
```

**Per-Route Threshold Override (advanced):**

```yaml
verification:
  visual_threshold: 5          # Global default
  route_thresholds:            # Per-route overrides
    "/": 2                     # Homepage: stricter
    "/dashboard": 10           # Dashboard: has dynamic charts
    "/settings": 3             # Settings: static content
```

### Step 3.5: Visual Report Generation

**Report Format: HTML (viewable in browser)**

```
{artifacts_dir}/visual-report.html
```

**Report Structure:**

```html
<!DOCTYPE html>
<html>
<head>
  <title>Visual Regression Report - Migration Verification</title>
  <style>
    .comparison { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 8px; }
    .pass { border: 2px solid green; }
    .fail { border: 2px solid red; }
    .warn { border: 2px solid orange; }
    img { width: 100%; height: auto; }
  </style>
</head>
<body>
  <h1>Visual Regression Report</h1>
  <p>Generated: {timestamp} | Threshold: {threshold}%</p>

  <!-- Summary -->
  <table>
    <tr><th>Route</th><th>Viewport</th><th>Diff %</th><th>Status</th></tr>
    <!-- Per-route results -->
  </table>

  <!-- Detailed Comparisons -->
  <div class="route-detail" id="route-{encoded_path}">
    <h2>{route} ({viewport})</h2>
    <p>Diff: {diff_pct}% | Pixels: {diff_pixels}/{total_pixels}</p>
    <div class="comparison">
      <div>
        <h3>Source</h3>
        <img src="screenshots/source/{viewport}/{route}.png" />
      </div>
      <div>
        <h3>Target</h3>
        <img src="screenshots/target/{viewport}/{route}.png" />
      </div>
      <div>
        <h3>Diff</h3>
        <img src="screenshots/diff/{viewport}/{route}-diff.png" />
      </div>
    </div>
  </div>
</body>
</html>
```

**Report Summary Section:**

```
Visual Regression Summary
=========================
Total Routes Tested: 25
Total Comparisons: 75 (25 routes x 3 viewports)

Results:
  PASS: 70 (93.3%)
  WARN: 3 (4.0%) - within 80-100% of threshold
  FAIL: 2 (2.7%) - exceeded threshold

Failed Routes:
  /profile (Desktop) - diff: 7.2% (threshold: 5%)
  /profile (Mobile) - diff: 8.1% (threshold: 5%)

Warning Routes:
  /settings (Desktop) - diff: 4.9% (threshold: 5%)
  /dashboard (Tablet) - diff: 4.2% (threshold: 5%)
  /dashboard (Mobile) - diff: 4.5% (threshold: 5%)
```

### Step 3.6: Diff Analysis & Categorization

**Diff Categories (for actionable feedback):**

| Category | Detection Method | Severity |
|----------|-----------------|----------|
| Layout Shift | Large contiguous diff regions | High |
| Color Change | Scattered pixel diffs with consistent color offset | Medium |
| Font Rendering | Text-area-only diffs with similar shapes | Low |
| Missing Element | One side has content, other is blank | Critical |
| Extra Element | Target has content not in source | Medium |
| Size Difference | Page height/width mismatch | High |

**Categorization Logic:**

```
FUNCTION categorizeDiff(diffImage, sourceImage, targetImage):
  IF targetImage.height != sourceImage.height:
    → "Size Difference" (page height changed)

  IF hasLargeContiguousRegion(diffImage, minSize: 100x100):
    → "Layout Shift" (major positioning change)

  IF diffOnlyInTextAreas(diffImage, sourceImage):
    → "Font Rendering" (acceptable in most cases)

  IF hasBlankRegionInTarget(targetImage, diffRegions):
    → "Missing Element" (critical - content lost)

  IF hasNewContentInTarget(targetImage, sourceImage, diffRegions):
    → "Extra Element" (review needed)

  DEFAULT:
    → "Color Change" (minor styling difference)
```

