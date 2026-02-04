import { chromium, Browser, Page, BrowserContext } from 'playwright';
import * as fs from 'fs';
import * as path from 'path';

interface CaptureOptions {
  outputDir: string;
  maxPages?: number;
  concurrency?: number;
  authFile?: string;
  exclude?: string[];
  timeout?: number;
}

interface PageResult {
  url: string;
  title: string;
  h1: string;
  screenshot: string;
  html: string;
  links: string[];
}

interface Sitemap {
  baseUrl: string;
  capturedAt: string;
  totalPages: number;
  pages: PageResult[];
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
      waitUntil: 'networkidle',
      timeout,
    });

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

      return { title, h1, links: [...new Set(links)] };
    }, baseUrl);

    return {
      url,
      title: pageInfo.title,
      h1: pageInfo.h1,
      screenshot: `${filename}.png`,
      html: `${filename}.html`,
      links: pageInfo.links,
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
  } = options;

  // 출력 디렉토리 생성
  if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
  }

  const baseUrl = new URL(startUrl).origin;
  const visited = new Set<string>();
  const toVisit: string[] = [startUrl];
  const results: PageResult[] = [];

  console.log(`🚀 크롤링 시작: ${baseUrl}`);
  console.log(`📁 출력 디렉토리: ${outputDir}`);
  console.log(`📄 최대 페이지: ${maxPages}`);

  const browser = await chromium.launch({ headless: true });

  // 인증 세션 사용
  const contextOptions: { storageState?: string } = {};
  if (authFile && fs.existsSync(authFile)) {
    contextOptions.storageState = authFile;
    console.log(`🔐 인증 세션 사용: ${authFile}`);
  }

  const context = await browser.newContext(contextOptions);

  // URL 제외 패턴 체크
  const shouldExclude = (url: string): boolean => {
    return exclude.some((pattern) => {
      const regex = new RegExp(pattern.replace('*', '.*'));
      return regex.test(url);
    });
  };

  while (toVisit.length > 0 && results.length < maxPages) {
    const batch = toVisit.splice(0, concurrency);

    const promises = batch.map(async (url) => {
      if (visited.has(url) || shouldExclude(url)) return null;
      visited.add(url);
      return await capturePage(context, url, baseUrl, outputDir, timeout);
    });

    const batchResults = await Promise.all(promises);

    for (const result of batchResults) {
      if (!result) continue;
      results.push(result);

      // 새로운 링크 추가
      for (const link of result.links) {
        if (!visited.has(link) && !toVisit.includes(link) && !shouldExclude(link)) {
          toVisit.push(link);
        }
      }
    }

    console.log(`   진행: ${results.length}개 완료, ${toVisit.length}개 대기`);
  }

  await context.close();
  await browser.close();

  // sitemap.json 저장
  const sitemap: Sitemap = {
    baseUrl,
    capturedAt: new Date().toISOString(),
    totalPages: results.length,
    pages: results,
  };

  fs.writeFileSync(
    path.join(outputDir, 'sitemap.json'),
    JSON.stringify(sitemap, null, 2)
  );

  console.log(`\n✅ 크롤링 완료!`);
  console.log(`📊 총 ${results.length}개 페이지 캡처`);
  console.log(`📁 결과: ${outputDir}/sitemap.json`);

  return sitemap;
}

/**
 * 수동 로그인 후 세션 저장
 */
export async function saveLoginSession(
  loginUrl: string,
  outputFile: string
): Promise<void> {
  console.log('🔐 로그인 세션 저장 모드');
  console.log(`📍 로그인 URL: ${loginUrl}`);

  const browser = await chromium.launch({ headless: false });
  const context = await browser.newContext();
  const page = await context.newPage();

  await page.goto(loginUrl);

  console.log('⏳ 브라우저에서 로그인하세요 (60초 대기)...');
  await page.waitForTimeout(60000);

  // 세션 저장
  await context.storageState({ path: outputFile });

  await browser.close();

  console.log(`✅ 세션 저장 완료: ${outputFile}`);
}
