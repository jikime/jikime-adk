import * as fs from 'fs';
import * as path from 'path';

interface GenerateFrontendOptions {
  mappingFile: string;
  outputDir: string;
  framework: string;
  style?: string;
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
    queries: Array<{
      raw: string;
      table: string;
      type: string;
      columns?: string[];
    }>;
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
  pages: PageMapping[];
}

/**
 * 테이블 이름을 Entity 이름으로 변환
 */
function toEntityName(tableName: string): string {
  return tableName
    .split('_')
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join('');
}

/**
 * Mock 데이터 생성
 */
function generateMockData(entityName: string, count: number = 5): string {
  const varName = entityName.charAt(0).toLowerCase() + entityName.slice(1);

  const mockItems = Array.from({ length: count }, (_, i) => ({
    id: i + 1,
    name: `${entityName} ${i + 1}`,
    description: `Description for ${varName} ${i + 1}`,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  }));

  return JSON.stringify(mockItems, null, 2);
}

/**
 * Next.js 정적 페이지 생성
 */
function generateStaticPage(pagePath: string, url: string): string {
  const pageName = path.basename(path.dirname(pagePath)) || 'Home';
  const titleCase = pageName.charAt(0).toUpperCase() + pageName.slice(1);

  return `// ${pagePath}
// Generated from: ${url}
// Type: Static Page

export default function ${titleCase}Page() {
  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">${titleCase}</h1>

      <div className="prose max-w-none">
        {/* TODO: Extract content from captured HTML */}
        <p>
          Content goes here...
        </p>
      </div>
    </div>
  );
}
`;
}

/**
 * Next.js 동적 페이지 생성 (Mock 데이터 사용)
 */
function generateDynamicPageWithMock(
  pagePath: string,
  url: string,
  apiEndpoint: string,
  entityName: string
): string {
  const pageName = path.basename(path.dirname(pagePath)) || 'Items';
  const titleCase = pageName.charAt(0).toUpperCase() + pageName.slice(1);
  const varName = entityName.charAt(0).toLowerCase() + entityName.slice(1);

  return `// ${pagePath}
// Generated from: ${url}
// Type: Dynamic Page (Mock Data)
// TODO: Replace mock data with real API call after backend is ready
// API Endpoint: ${apiEndpoint}

interface ${entityName} {
  id: number;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
}

// ⚠️ MOCK DATA - Will be replaced by generate connect
const mock${entityName}s: ${entityName}[] = [
  { id: 1, name: '${entityName} 1', description: 'Description 1', createdAt: '${new Date().toISOString()}', updatedAt: '${new Date().toISOString()}' },
  { id: 2, name: '${entityName} 2', description: 'Description 2', createdAt: '${new Date().toISOString()}', updatedAt: '${new Date().toISOString()}' },
  { id: 3, name: '${entityName} 3', description: 'Description 3', createdAt: '${new Date().toISOString()}', updatedAt: '${new Date().toISOString()}' },
];

// ⚠️ MOCK FUNCTION - Will be replaced by real API call
async function get${entityName}s(): Promise<${entityName}[]> {
  // TODO: Replace with real API call
  // const res = await fetch(\`\${process.env.API_URL}${apiEndpoint}\`);
  // return res.json();
  return Promise.resolve(mock${entityName}s);
}

export default async function ${titleCase}Page() {
  const ${varName}s = await get${entityName}s();

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">${titleCase}</h1>

      {/* Mock Data Banner */}
      <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4 mb-6">
        <p className="text-yellow-700">
          ⚠️ 현재 Mock 데이터를 사용 중입니다. 백엔드 연동 후 실제 데이터로 교체됩니다.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {${varName}s.map((${varName}) => (
          <div
            key={${varName}.id}
            className="bg-white rounded-lg shadow-md p-4 hover:shadow-lg transition-shadow"
          >
            <h3 className="font-semibold text-lg">{${varName}.name}</h3>
            <p className="text-gray-600 mt-2">{${varName}.description}</p>
            <p className="text-sm text-gray-400 mt-2">
              Created: {new Date(${varName}.createdAt).toLocaleDateString()}
            </p>
          </div>
        ))}
      </div>

      {${varName}s.length === 0 && (
        <p className="text-gray-500 text-center py-8">
          No ${varName}s found.
        </p>
      )}
    </div>
  );
}
`;
}

/**
 * Frontend 생성 메인 함수 (Mock 데이터 사용)
 */
export async function generateFrontend(options: GenerateFrontendOptions): Promise<void> {
  const { mappingFile, outputDir, framework } = options;

  console.log('🎨 Frontend 생성 시작 (Mock 데이터)');

  // 매핑 파일 로드
  if (!fs.existsSync(mappingFile)) {
    throw new Error(`Mapping file not found: ${mappingFile}`);
  }

  const mapping: Mapping = JSON.parse(fs.readFileSync(mappingFile, 'utf-8'));
  console.log(`📋 매핑 로드: ${mapping.pages.length}개 페이지`);

  // 출력 디렉토리 생성
  fs.mkdirSync(outputDir, { recursive: true });

  let staticCount = 0;
  let dynamicCount = 0;

  for (const page of mapping.pages) {
    if (framework === 'nextjs') {
      const frontendPath = page.output.frontend.path;
      const fullPath = path.join(outputDir, frontendPath);
      fs.mkdirSync(path.dirname(fullPath), { recursive: true });

      if (page.output.frontend.type === 'static-page') {
        fs.writeFileSync(fullPath, generateStaticPage(frontendPath, page.capture.url));
        staticCount++;
        console.log(`   ✓ Static: ${frontendPath}`);
      } else {
        const apiEndpoint = page.output.frontend.apiCalls?.[0] || '/api/items';
        const table = page.database?.queries?.[0]?.table || 'Item';
        const entityName = toEntityName(table);
        fs.writeFileSync(
          fullPath,
          generateDynamicPageWithMock(frontendPath, page.capture.url, apiEndpoint, entityName)
        );
        dynamicCount++;
        console.log(`   ✓ Dynamic (Mock): ${frontendPath}`);
      }
    }
  }

  // layout.tsx 생성
  const layoutPath = path.join(outputDir, 'app/layout.tsx');
  fs.mkdirSync(path.dirname(layoutPath), { recursive: true });
  fs.writeFileSync(layoutPath, generateLayout());

  // 정적 자산 복사 (이미지, 폰트 등)
  let assetCount = 0;
  if (mapping.project.sourcePath && fs.existsSync(mapping.project.sourcePath)) {
    console.log(`\n📦 정적 자산 복사 중...`);
    assetCount = copyStaticAssets(mapping.project.sourcePath, outputDir);
    console.log(`   ✓ ${assetCount}개 파일 복사 완료 → public/`);
  }

  console.log(`\n✅ Frontend 생성 완료!`);
  console.log(`📄 정적 페이지: ${staticCount}개`);
  console.log(`📄 동적 페이지 (Mock): ${dynamicCount}개`);
  console.log(`🖼️ 정적 자산: ${assetCount}개`);
  console.log(`📁 출력 경로: ${outputDir}`);
  console.log(`\n💡 다음 단계: UI 확인 후 'generate backend' 실행`);
}

/**
 * 정적 자산 복사 (이미지, 폰트 등)
 */
function copyStaticAssets(sourcePath: string, outputDir: string): number {
  const publicDir = path.join(outputDir, 'public');
  fs.mkdirSync(publicDir, { recursive: true });

  // 복사할 파일 확장자
  const assetExtensions = [
    // 이미지
    '.jpg', '.jpeg', '.png', '.gif', '.svg', '.webp', '.ico', '.bmp',
    // 폰트
    '.woff', '.woff2', '.ttf', '.eot', '.otf',
    // 기타
    '.pdf', '.mp4', '.mp3', '.webm',
  ];

  // 제외할 디렉토리
  const excludeDirs = ['node_modules', '.git', 'vendor', 'cache', '__pycache__'];

  let copiedCount = 0;

  function scanAndCopy(dir: string, relativePath: string = '') {
    if (!fs.existsSync(dir)) return;

    const items = fs.readdirSync(dir);

    for (const item of items) {
      const fullPath = path.join(dir, item);
      const relPath = path.join(relativePath, item);

      // 제외 디렉토리 스킵
      if (excludeDirs.includes(item)) continue;

      const stat = fs.statSync(fullPath);

      if (stat.isDirectory()) {
        scanAndCopy(fullPath, relPath);
      } else {
        const ext = path.extname(item).toLowerCase();
        if (assetExtensions.includes(ext)) {
          const destPath = path.join(publicDir, relPath);
          fs.mkdirSync(path.dirname(destPath), { recursive: true });
          fs.copyFileSync(fullPath, destPath);
          copiedCount++;
        }
      }
    }
  }

  scanAndCopy(sourcePath);
  return copiedCount;
}

/**
 * Next.js layout.tsx 생성
 */
function generateLayout(): string {
  return `import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Smart Rebuild App',
  description: 'Generated by Smart Rebuild',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ko">
      <body className="min-h-screen bg-gray-50">
        <header className="bg-white shadow-sm">
          <div className="container mx-auto px-4 py-4">
            <h1 className="text-xl font-bold text-gray-900">Smart Rebuild App</h1>
          </div>
        </header>
        <main>{children}</main>
      </body>
    </html>
  );
}
`;
}
