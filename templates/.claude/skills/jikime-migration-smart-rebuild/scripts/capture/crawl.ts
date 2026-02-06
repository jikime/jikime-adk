import { chromium, Browser, Page, BrowserContext } from 'playwright';
import * as fs from 'fs';
import * as path from 'path';
import * as readline from 'readline';

interface CaptureOptions {
  outputDir: string;
  maxPages?: number;
  concurrency?: number;
  authFile?: string;
  exclude?: string[];
  timeout?: number;
  login?: boolean;  // 로그인 필요 시 true
  dedupeByTemplate?: boolean;  // 템플릿 기준 중복 제거 (기본: true)
}

interface PageResult {
  id?: number;       // 페이지 번호 (1-based, 저장 시 할당)
  url: string;
  template?: string;  // 템플릿 패턴 (예: /customer/nt_list.php)
  urlPattern?: string;  // URL 패턴 (예: /customer/nt_list.php?page={page})
  title: string;
  h1: string;
  screenshot: string;
  html: string;
  links: string[];
  images: string[];  // 페이지 내 이미지 URL 목록
  skippedUrls?: string[];  // 같은 템플릿으로 스킵된 URL들
}

interface Sitemap {
  baseUrl: string;
  capturedAt: string;
  totalPages: number;
  totalTemplates: number;  // 고유 템플릿 수
  skippedUrls: number;     // 스킵된 URL 수
  dedupeByTemplate: boolean;
  pages: PageResult[];
}

/**
 * 사용자 입력 대기
 */
function waitForUserInput(prompt: string): Promise<void> {
  return new Promise((resolve) => {
    const rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout,
    });
    rl.question(prompt, () => {
      rl.close();
      resolve();
    });
  });
}

/**
 * URL을 안전한 파일명으로 변환
 */
function urlToFilename(url: string): string {
  return url
    .replace(/https?:\/\//, '')
    .replace(/[^a-zA-Z0-9]/g, '_')
    .substring(0, 80);
}

/**
 * URL에서 템플릿 추출 (쿼리 파라미터 완전 무시)
 * 예: /customer/nt_view.php?idx=644&page=1 → /customer/nt_view.php
 * 예: /review/sty_list.php?s_orderby=sty_idx&page=7 → /review/sty_list.php
 */
function extractTemplate(url: string): string {
  try {
    const urlObj = new URL(url);
    return urlObj.pathname;  // 쿼리 파라미터 완전 무시
  } catch {
    // URL 파싱 실패 시 ? 이전 부분만 반환
    return url.split('?')[0];
  }
}

/**
 * URL 패턴 추출 (파라미터 키만 플레이스홀더로)
 * 예: /customer/nt_view.php?idx=644&page=1 → /customer/nt_view.php?idx={idx}&page={page}
 */
function extractUrlPattern(url: string): string {
  try {
    const urlObj = new URL(url);
    const params = new URLSearchParams(urlObj.search);
    const patternParams: string[] = [];

    params.forEach((_, key) => {
      patternParams.push(`${key}={${key}}`);
    });

    if (patternParams.length > 0) {
      return `${urlObj.pathname}?${patternParams.join('&')}`;
    }
    return urlObj.pathname;
  } catch {
    return url;
  }
}

/**
 * 쿼리 파라미터가 있는지 확인
 */
function hasQueryParams(url: string): boolean {
  try {
    const urlObj = new URL(url);
    return urlObj.search.length > 0;
  } catch {
    return url.includes('?');
  }
}

/**
 * Lazy loading 이미지를 로드하기 위한 자동 스크롤
 */
async function autoScroll(page: Page): Promise<void> {
  await page.evaluate(async () => {
    await new Promise<void>((resolve) => {
      let totalHeight = 0;
      const distance = 500;
      const maxHeight = 50000;

      const timer = setInterval(() => {
        window.scrollBy(0, distance);
        totalHeight += distance;

        if (totalHeight >= document.body.scrollHeight || totalHeight >= maxHeight) {
          clearInterval(timer);
          window.scrollTo(0, 0);
          resolve();
        }
      }, 100);
    });
  });
}

/**
 * 단일 페이지 캡처
 */
async function capturePage(
  context: BrowserContext,
  url: string,
  baseUrl: string,
  outputDir: string,
  timeout: number
): Promise<PageResult | null> {
  const page = await context.newPage();

  try {
    console.log(`📸 캡처 중: ${url}`);

    await page.goto(url, {
      waitUntil: 'domcontentloaded',
      timeout: timeout || 60000,
    });

    // 추가 로딩 대기
    await page.waitForTimeout(2000);

    // Lazy loading 해결
    await autoScroll(page);
    await page.waitForTimeout(500);

    const filename = urlToFilename(url);

    // 스크린샷 저장
    await page.screenshot({
      path: path.join(outputDir, `${filename}.png`),
      fullPage: true,
    });

    // HTML 저장
    const html = await page.content();
    fs.writeFileSync(path.join(outputDir, `${filename}.html`), html);

    // 페이지 정보 추출
    const pageInfo = await page.evaluate((base: string) => {
      const title = document.title;
      const h1 = document.querySelector('h1')?.innerText || '';

      // 링크 추출
      const links = [...document.querySelectorAll('a[href]')]
        .map((a) => (a as HTMLAnchorElement).href)
        .filter(
          (href) =>
            href.startsWith(base) &&
            !href.includes('#') &&
            !href.match(/\.(pdf|jpg|png|gif|zip|doc|docx)$/i) &&
            !href.includes('mailto:') &&
            !href.includes('tel:')
        );

      // 이미지 URL 추출 (절대 경로로 변환)
      const images = [...document.querySelectorAll('img[src]')]
        .map((img) => {
          const src = (img as HTMLImageElement).src;
          // 이미 절대 경로이거나 data URL인 경우 그대로 반환
          if (src.startsWith('http') || src.startsWith('data:')) {
            return src;
          }
          // 상대 경로를 절대 경로로 변환
          try {
            return new URL(src, base).href;
          } catch {
            return src;
          }
        })
        .filter((src) =>
          src.startsWith('http') &&
          !src.includes('data:') &&
          src.match(/\.(jpg|jpeg|png|gif|webp|svg)$/i)
        );

      return { title, h1, links: [...new Set(links)], images: [...new Set(images)] };
    }, baseUrl);

    return {
      url,
      title: pageInfo.title,
      h1: pageInfo.h1,
      screenshot: `${filename}.png`,
      html: `${filename}.html`,
      links: pageInfo.links,
      images: pageInfo.images,
    };
  } catch (error) {
    console.error(`❌ 에러: ${url} - ${(error as Error).message}`);
    return null;
  } finally {
    await page.close();
  }
}

/**
 * 사이트 크롤링 및 캡처
 */
export async function crawlAndCapture(
  startUrl: string,
  options: CaptureOptions
): Promise<Sitemap> {
  const {
    outputDir,
    maxPages = 100,
    concurrency = 5,
    authFile,
    exclude = [],
    timeout = 30000,
    login = false,
    dedupeByTemplate = true,  // 기본값: 템플릿 기준 중복 제거 활성화
  } = options;

  // 출력 디렉토리 생성
  if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
  }

  const baseUrl = new URL(startUrl).origin;
  const visited = new Set<string>();
  const visitedTemplates = new Set<string>();  // 캡처 완료된 템플릿
  const queuedTemplates = new Set<string>();   // 큐에 있는 템플릿 (아직 캡처 안 됨)
  const skippedByTemplate = new Map<string, string[]>();  // 템플릿별 스킵된 URL 기록
  const toVisit: string[] = [startUrl];
  const toVisitSet = new Set<string>([startUrl]);  // O(1) 조회용
  const results: PageResult[] = [];

  // 시작 URL의 템플릿을 큐에 등록
  const startTemplate = extractTemplate(startUrl);
  queuedTemplates.add(startTemplate);

  console.log(`\n🔧 [DEBUG] 템플릿 중복 제거 버전: v2.0`);
  console.log(`🔧 [DEBUG] 시작 URL: ${startUrl}`);
  console.log(`🔧 [DEBUG] 시작 템플릿: ${startTemplate}`);
  console.log(`🚀 크롤링 시작: ${baseUrl}`);
  console.log(`📁 출력 디렉토리: ${outputDir}`);
  console.log(`📄 최대 페이지: ${maxPages}`);
  console.log(`🔄 템플릿 중복 제거: ${dedupeByTemplate ? '활성화' : '비활성화'}`);

  let browser: Browser;
  let context: BrowserContext;
  const sessionFile = path.join(outputDir, 'auth.json');

  // 로그인 모드: 브라우저 열고 사용자 로그인 대기
  if (login) {
    console.log(`\n🔐 로그인 모드 활성화`);
    browser = await chromium.launch({ headless: false });
    context = await browser.newContext();
    const page = await context.newPage();

    await page.goto(startUrl);
    console.log(`📍 브라우저에서 로그인을 완료하세요.`);

    await waitForUserInput('✅ 로그인 완료 후 Enter를 누르세요...');

    // 세션 저장
    await context.storageState({ path: sessionFile });
    console.log(`💾 세션 저장 완료: ${sessionFile}`);
    await page.close();

    // headless 모드로 재시작하여 캡처 진행
    await context.close();
    await browser.close();

    browser = await chromium.launch({ headless: true });
    context = await browser.newContext({ storageState: sessionFile });
    console.log(`\n🚀 캡처 시작...`);
  } else {
    browser = await chromium.launch({ headless: true });

    // 기존 인증 세션 사용
    const contextOptions: { storageState?: string } = {};
    if (authFile && fs.existsSync(authFile)) {
      contextOptions.storageState = authFile;
      console.log(`🔐 인증 세션 사용: ${authFile}`);
    }
    context = await browser.newContext(contextOptions);
  }

  // URL 제외 패턴 체크
  const shouldExclude = (url: string): boolean => {
    return exclude.some((pattern) => {
      const regex = new RegExp(pattern.replace('*', '.*'));
      return regex.test(url);
    });
  };

  while (toVisit.length > 0 && results.length < maxPages) {
    // 배치 처리 전에 먼저 템플릿 중복 제거 (동기적으로!)
    const batch: string[] = [];
    while (toVisit.length > 0 && batch.length < concurrency) {
      const url = toVisit.shift()!;
      toVisitSet.delete(url);

      if (visited.has(url) || shouldExclude(url)) continue;

      const template = extractTemplate(url);

      // 템플릿 기준 중복 제거 (쿼리 파라미터 무시)
      if (dedupeByTemplate) {
        if (visitedTemplates.has(template)) {
          // 같은 템플릿의 다른 URL은 스킵하고 기록만
          console.log(`⏭️  [SKIP] 중복 템플릿: ${template} (URL: ${url})`);
          if (!skippedByTemplate.has(template)) {
            skippedByTemplate.set(template, []);
          }
          skippedByTemplate.get(template)!.push(url);
          continue;
        }
        console.log(`✅ [NEW] 신규 템플릿: ${template}`);
        visitedTemplates.add(template);
        queuedTemplates.delete(template);  // 큐에서 캡처됨으로 이동
      }

      visited.add(url);
      batch.push(url);
    }

    if (batch.length === 0) continue;

    const promises = batch.map(async (url) => {
      const template = extractTemplate(url);
      const result = await capturePage(context, url, baseUrl, outputDir, timeout);

      // 결과에 템플릿 정보 추가
      if (result) {
        result.template = template;
        result.urlPattern = hasQueryParams(url) ? extractUrlPattern(url) : template;
      }

      return result;
    });

    const batchResults = await Promise.all(promises);

    for (const result of batchResults) {
      if (!result) continue;

      // 스킵된 URL 정보 추가
      const skipped = skippedByTemplate.get(result.template);
      if (skipped && skipped.length > 0) {
        result.skippedUrls = [...skipped];
      }

      // 페이지 번호 할당 (1-based)
      result.id = results.length + 1;

      results.push(result);

      // 새로운 링크 추가 (템플릿 중복은 큐에 넣기 전에 필터링)
      for (const link of result.links) {
        if (visited.has(link) || toVisitSet.has(link) || shouldExclude(link)) continue;

        const linkTemplate = extractTemplate(link);

        // 템플릿 중복 제거: 이미 캡처했거나 큐에 있는 템플릿은 추가하지 않음
        if (dedupeByTemplate && (visitedTemplates.has(linkTemplate) || queuedTemplates.has(linkTemplate))) {
          // 스킵된 URL 기록
          if (!skippedByTemplate.has(linkTemplate)) {
            skippedByTemplate.set(linkTemplate, []);
          }
          skippedByTemplate.get(linkTemplate)!.push(link);
          // 첫 5개만 로그 출력 (너무 많으면 지저분해지므로)
          if (skippedByTemplate.get(linkTemplate)!.length <= 5) {
            console.log(`   ⏭️  큐 스킵: ${linkTemplate} (이미 ${visitedTemplates.has(linkTemplate) ? '캡처됨' : '큐에 있음'})`);
          }
          continue;
        }

        toVisit.push(link);
        toVisitSet.add(link);
        if (dedupeByTemplate) {
          queuedTemplates.add(linkTemplate);
        }
      }
    }

    console.log(`   진행: ${results.length}개 캡처, ${visitedTemplates.size}개 템플릿, ${toVisit.length}개 대기`);
  }

  await context.close();
  await browser.close();

  // 스킵된 URL 총 개수 계산
  let totalSkipped = 0;
  skippedByTemplate.forEach((urls) => {
    totalSkipped += urls.length;
  });

  // sitemap.json 저장 (페이지 번호 확정)
  const pagesWithId = results.map((page, index) => ({
    ...page,
    id: page.id ?? index + 1,  // 1-based 페이지 번호
  }));

  const sitemap: Sitemap = {
    baseUrl,
    capturedAt: new Date().toISOString(),
    totalPages: pagesWithId.length,
    totalTemplates: visitedTemplates.size,
    skippedUrls: totalSkipped,
    dedupeByTemplate,
    pages: pagesWithId,
  };

  fs.writeFileSync(
    path.join(outputDir, 'sitemap.json'),
    JSON.stringify(sitemap, null, 2)
  );

  console.log(`\n✅ 크롤링 완료!`);
  console.log(`📊 총 ${results.length}개 페이지 캡처 (${visitedTemplates.size}개 고유 템플릿)`);
  if (dedupeByTemplate && totalSkipped > 0) {
    console.log(`⏭️  ${totalSkipped}개 중복 URL 스킵 (템플릿 기준 중복 제거)`);
  }
  console.log(`📁 결과: ${outputDir}/sitemap.json`);

  return sitemap;
}

