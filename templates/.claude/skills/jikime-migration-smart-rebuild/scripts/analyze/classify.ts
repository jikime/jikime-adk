import * as fs from 'fs';
import * as path from 'path';
import { glob } from 'glob';

interface AnalyzeOptions {
  sourcePath: string;
  capturePath: string;
  outputFile: string;
  dbSchemaFile?: string;
  manualMappingFile?: string;
}

interface PageAnalysis {
  path: string;
  type: 'static' | 'dynamic';
  reason: string[];
  dbQueries: ExtractedQuery[];
}

interface ExtractedQuery {
  raw: string;
  table: string;
  type: 'SELECT' | 'INSERT' | 'UPDATE' | 'DELETE';
  columns?: string[];
  conditions?: string;
}

interface CapturedPage {
  url: string;
  screenshot: string;
  html: string;
  title: string;
}

interface PageMapping {
  id: string;
  capture: {
    url: string;
    screenshot: string;
    html: string;
  };
  source: {
    file: string | null;
    type: 'static' | 'dynamic' | 'unknown';
    reason: string[];
  };
  database?: {
    queries: ExtractedQuery[];
  };
  output: {
    backend?: {
      entity?: string;
      repository?: string;
      controller?: string;
      endpoint?: string;
    };
    frontend: {
      path: string;
      type: 'static-page' | 'dynamic-page';
      apiCalls?: string[];
    };
  };
}

interface Mapping {
  project: {
    name: string;
    sourceUrl: string;
    sourcePath: string;
  };
  summary: {
    totalPages: number;
    static: number;
    dynamic: number;
    unknown: number;
  };
  pages: PageMapping[];
  database?: {
    tables: Array<{
      name: string;
      columns: Array<{
        name: string;
        type: string;
        primary?: boolean;
      }>;
    }>;
  };
}

/**
 * SQL 쿼리 추출
 */
function extractQueries(content: string): ExtractedQuery[] {
  const queries: ExtractedQuery[] = [];

  // SELECT 쿼리
  const selectPattern = /SELECT\s+([\w\s,*`]+)\s+FROM\s+[`']?(\w+)[`']?(?:\s+WHERE\s+(.+?))?(?:;|$|ORDER|LIMIT|GROUP)/gi;
  let match;

  while ((match = selectPattern.exec(content)) !== null) {
    queries.push({
      raw: match[0].trim(),
      type: 'SELECT',
      columns: match[1].split(',').map((c) => c.trim()),
      table: match[2],
      conditions: match[3]?.trim(),
    });
  }

  // INSERT 쿼리
  const insertPattern = /INSERT\s+INTO\s+[`']?(\w+)[`']?/gi;
  while ((match = insertPattern.exec(content)) !== null) {
    queries.push({
      raw: match[0].trim(),
      type: 'INSERT',
      table: match[1],
    });
  }

  // UPDATE 쿼리
  const updatePattern = /UPDATE\s+[`']?(\w+)[`']?\s+SET/gi;
  while ((match = updatePattern.exec(content)) !== null) {
    queries.push({
      raw: match[0].trim(),
      type: 'UPDATE',
      table: match[1],
    });
  }

  // DELETE 쿼리
  const deletePattern = /DELETE\s+FROM\s+[`']?(\w+)[`']?/gi;
  while ((match = deletePattern.exec(content)) !== null) {
    queries.push({
      raw: match[0].trim(),
      type: 'DELETE',
      table: match[1],
    });
  }

  return queries;
}

/**
 * 페이지 분류 (정적/동적)
 */
function classifyPage(filePath: string): PageAnalysis {
  const content = fs.readFileSync(filePath, 'utf-8');
  const reasons: string[] = [];
  const dbQueries = extractQueries(content);

  // 1. SQL 쿼리 체크
  if (dbQueries.length > 0) {
    reasons.push(`SQL 쿼리 ${dbQueries.length}개 발견`);
  }

  // 2. DB 연결 함수 체크
  const dbPatterns = [
    { pattern: /mysqli_query|mysqli_fetch/g, name: 'mysqli' },
    { pattern: /\$pdo->query|\$pdo->prepare/g, name: 'PDO' },
    { pattern: /\$wpdb->/g, name: 'WordPress DB' },
    { pattern: /\$this->db->get|\$this->db->query/g, name: 'CodeIgniter' },
    { pattern: /DB::table|DB::select/g, name: 'Laravel' },
  ];

  for (const { pattern, name } of dbPatterns) {
    if (pattern.test(content)) {
      reasons.push(`${name} 사용`);
    }
  }

  // 3. 세션 체크
  if (/\$_SESSION|session_start\s*\(/g.test(content)) {
    reasons.push('세션 사용');
  }

  // 4. POST 처리 체크
  if (/\$_POST\s*\[|\$_REQUEST\s*\[/g.test(content)) {
    reasons.push('POST 데이터 처리');
  }

  // 5. 동적 파라미터 체크
  if (/\$_GET\s*\[/g.test(content)) {
    reasons.push('GET 파라미터 사용');
  }

  return {
    path: filePath,
    type: reasons.length > 0 ? 'dynamic' : 'static',
    reason: reasons,
    dbQueries,
  };
}

/**
 * URL과 소스 파일 매칭
 */
function matchUrlToSource(
  url: string,
  sourcePath: string,
  manualMapping?: Record<string, string>
): string | null {
  const urlObj = new URL(url);
  let urlPath = urlObj.pathname;

  // 수동 매핑 체크
  if (manualMapping && manualMapping[url]) {
    return manualMapping[url];
  }

  // 경로 정규화
  if (urlPath === '/') urlPath = '/index';
  if (urlPath.endsWith('/')) urlPath = urlPath.slice(0, -1);

  // 1. 직접 매칭 (path.php)
  const directMatch = path.join(sourcePath, `${urlPath}.php`);
  if (fs.existsSync(directMatch)) return directMatch;

  // 2. index.php 매칭 (path/index.php)
  const indexMatch = path.join(sourcePath, urlPath, 'index.php');
  if (fs.existsSync(indexMatch)) return indexMatch;

  // 3. 쿼리 파라미터 기반 (index.php?page=about → about.php)
  const pageParam = urlObj.searchParams.get('page');
  if (pageParam) {
    const pageMatch = path.join(sourcePath, `${pageParam}.php`);
    if (fs.existsSync(pageMatch)) return pageMatch;
  }

  return null;
}

/**
 * 소스 분석 및 매핑 생성
 */
export async function analyzeSource(options: AnalyzeOptions): Promise<Mapping> {
  const { sourcePath, capturePath, outputFile, dbSchemaFile, manualMappingFile } = options;

  console.log('🔍 소스 분석 시작');

  // sitemap.json 로드
  const sitemapPath = path.join(capturePath, 'sitemap.json');
  if (!fs.existsSync(sitemapPath)) {
    throw new Error(`sitemap.json not found at ${sitemapPath}`);
  }

  const sitemap = JSON.parse(fs.readFileSync(sitemapPath, 'utf-8'));
  const capturedPages: CapturedPage[] = sitemap.pages;

  // 수동 매핑 로드
  let manualMapping: Record<string, string> | undefined;
  if (manualMappingFile && fs.existsSync(manualMappingFile)) {
    manualMapping = JSON.parse(fs.readFileSync(manualMappingFile, 'utf-8'));
    console.log(`📋 수동 매핑 로드: ${Object.keys(manualMapping).length}개`);
  }

  // PHP 파일 분석
  const phpFiles = await glob('**/*.php', { cwd: sourcePath });
  console.log(`📂 소스 파일: ${phpFiles.length}개`);

  const pageAnalyses = new Map<string, PageAnalysis>();
  for (const file of phpFiles) {
    const fullPath = path.join(sourcePath, file);
    pageAnalyses.set(fullPath, classifyPage(fullPath));
  }

  // 매핑 생성
  const pages: PageMapping[] = [];
  let staticCount = 0;
  let dynamicCount = 0;
  let unknownCount = 0;

  for (let i = 0; i < capturedPages.length; i++) {
    const captured = capturedPages[i];
    const sourceFile = matchUrlToSource(captured.url, sourcePath, manualMapping);

    let pageType: 'static' | 'dynamic' | 'unknown' = 'unknown';
    let reasons: string[] = [];
    let queries: ExtractedQuery[] = [];

    if (sourceFile && pageAnalyses.has(sourceFile)) {
      const analysis = pageAnalyses.get(sourceFile)!;
      pageType = analysis.type;
      reasons = analysis.reason;
      queries = analysis.dbQueries;
    }

    // 카운트 업데이트
    if (pageType === 'static') staticCount++;
    else if (pageType === 'dynamic') dynamicCount++;
    else unknownCount++;

    // 출력 경로 생성
    const urlPath = new URL(captured.url).pathname || '/';
    const frontendPath = urlPath === '/' ? '/app/page.tsx' : `/app${urlPath}/page.tsx`;

    const pageMapping: PageMapping = {
      id: `page_${String(i + 1).padStart(3, '0')}`,
      capture: {
        url: captured.url,
        screenshot: captured.screenshot,
        html: captured.html,
      },
      source: {
        file: sourceFile ? path.relative(sourcePath, sourceFile) : null,
        type: pageType,
        reason: reasons,
      },
      output: {
        frontend: {
          path: frontendPath,
          type: pageType === 'dynamic' ? 'dynamic-page' : 'static-page',
        },
      },
    };

    // 동적 페이지인 경우 백엔드 정보 추가
    if (pageType === 'dynamic' && queries.length > 0) {
      const tables = [...new Set(queries.map((q) => q.table))];
      const mainTable = tables[0];
      const entityName = mainTable.charAt(0).toUpperCase() + mainTable.slice(1);

      pageMapping.database = { queries };
      pageMapping.output.backend = {
        entity: `${entityName}.java`,
        repository: `${entityName}Repository.java`,
        controller: `${entityName}Controller.java`,
        endpoint: `GET /api/${mainTable}`,
      };
      pageMapping.output.frontend.apiCalls = [`GET /api/${mainTable}`];
    }

    pages.push(pageMapping);
  }

  // 매핑 결과 생성
  const mapping: Mapping = {
    project: {
      name: path.basename(sourcePath),
      sourceUrl: sitemap.baseUrl,
      sourcePath,
    },
    summary: {
      totalPages: pages.length,
      static: staticCount,
      dynamic: dynamicCount,
      unknown: unknownCount,
    },
    pages,
  };

  // DB 스키마 추가
  if (dbSchemaFile && fs.existsSync(dbSchemaFile)) {
    mapping.database = JSON.parse(fs.readFileSync(dbSchemaFile, 'utf-8'));
  }

  // 결과 저장
  fs.writeFileSync(outputFile, JSON.stringify(mapping, null, 2));

  console.log(`\n✅ 분석 완료!`);
  console.log(`📊 정적: ${staticCount}, 동적: ${dynamicCount}, 미확인: ${unknownCount}`);
  console.log(`📁 결과: ${outputFile}`);

  return mapping;
}
