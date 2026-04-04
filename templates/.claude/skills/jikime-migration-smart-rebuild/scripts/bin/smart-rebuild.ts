#!/usr/bin/env node

import { Command } from 'commander';
import { crawlAndCapture, captureSinglePage, getUrlFromMapping } from '../capture/crawl';
import { analyzeSource } from '../analyze/classify';
import { generateFrontend } from '../generate/frontend';
import { generateBackend } from '../generate/backend';
import { connectFrontendToBackend } from '../generate/connect';
import { spawn } from 'child_process';
import * as path from 'path';

const program = new Command();

program
  .name('smart-rebuild')
  .description('AI-powered legacy site rebuilding CLI')
  .version('1.2.0');

// Capture command
program
  .command('capture <url>')
  .description('Capture links from a live site (Lazy Capture mode by default)')
  .option('-o, --output <dir>', 'Output directory', './capture')
  .option('-m, --max-pages <n>', 'Maximum pages to capture', '100')
  .option('-c, --concurrency <n>', 'Concurrent page captures', '5')
  .option('-a, --auth <file>', 'Auth session file (JSON)')
  .option('-e, --exclude <patterns>', 'URL patterns to exclude', '/admin/*,/api/*')
  .option('-i, --include <patterns>', 'URL patterns to include (merge mode only)', '')
  .option('-t, --timeout <ms>', 'Page load timeout', '30000')
  .option('--login', 'Open browser for login, then capture')
  .option('--prefetch', '🔴 Capture all pages immediately (default: Lazy Capture - links only)')
  .option('--merge', 'Merge with existing sitemap.json (preserve completed pages)')
  .action(async (url, options) => {
    console.log('🚀 Smart Rebuild - Capture Phase');
    console.log(`📍 Target: ${url}`);
    console.log(`📁 Output: ${options.output}`);
    console.log(`📸 Mode: ${options.prefetch ? 'Prefetch (즉시 캡처)' : 'Lazy Capture (링크만 수집)'}`);
    if (options.merge) {
      console.log(`🔀 Merge: 기존 sitemap.json 병합 모드`);
    }

    await crawlAndCapture(url, {
      outputDir: options.output,
      maxPages: parseInt(options.maxPages),
      concurrency: parseInt(options.concurrency),
      authFile: options.auth,
      exclude: options.exclude.split(','),
      include: options.include ? options.include.split(',') : [],
      timeout: parseInt(options.timeout),
      login: options.login || false,
      prefetch: options.prefetch || false,
      merge: options.merge || false,
    });
  });

// 🔴 Capture single page command (for generate phase)
// URL 직접 지정 또는 --page ID + --mapping으로 URL 자동 조회
program
  .command('capture-page [url]')
  .description('Capture a single page and auto-update sitemap.json')
  .option('-o, --output <dir>', 'Output directory', './capture')
  .option('-p, --page <id>', 'Page ID from mapping.json (e.g. page_009)')
  .option('-m, --mapping <file>', 'Mapping file path (for --page lookup)')
  .option('-a, --auth <file>', 'Auth session file (JSON)')
  .option('-t, --timeout <ms>', 'Page load timeout', '30000')
  .action(async (url, options) => {
    // --page 옵션으로 mapping.json에서 URL 조회
    let targetUrl = url;
    if (!targetUrl && options.page) {
      const mappingFile = options.mapping || path.join(options.output, '..', 'mapping.json');
      console.log(`🔍 mapping.json에서 ${options.page} 조회 중...`);
      const resolved = getUrlFromMapping(mappingFile, options.page);
      if (!resolved) {
        console.error(`❌ ${options.page}를 찾을 수 없습니다: ${mappingFile}`);
        process.exit(1);
      }
      targetUrl = resolved;
      console.log(`📍 URL 확인: ${targetUrl}`);
    }

    if (!targetUrl) {
      console.error('❌ URL 또는 --page 옵션이 필요합니다.');
      console.error('   사용법: capture-page <url>');
      console.error('   또는:   capture-page --page page_009 --mapping ./mapping.json');
      process.exit(1);
    }

    console.log('📸 Smart Rebuild - Single Page Capture');
    console.log(`📍 URL: ${targetUrl}`);
    console.log(`📁 Output: ${options.output}`);

    const result = await captureSinglePage(
      targetUrl,
      options.output,
      options.auth,
      parseInt(options.timeout)
    );

    if (result) {
      console.log(`✅ 캡처 완료!`);
      console.log(`   스크린샷: ${result.screenshot}`);
      console.log(`   HTML: ${result.html}`);
      console.log(`   시간: ${result.capturedAt}`);
      console.log(`   sitemap.json: 자동 반영됨`);
      // JSON 결과 출력 (프로그래매틱 사용용)
      console.log(`\n<!-- CAPTURE_RESULT_JSON_START -->`);
      console.log(JSON.stringify(result, null, 2));
      console.log(`<!-- CAPTURE_RESULT_JSON_END -->`);
    } else {
      console.error('❌ 캡처 실패');
      process.exit(1);
    }
  });

// Analyze command
program
  .command('analyze')
  .description('Analyze legacy source and create mapping')
  .option('-s, --source <dir>', 'Legacy source directory', './source')
  .option('-c, --capture <dir>', 'Capture directory', './capture')
  .option('-o, --output <file>', 'Output mapping file', './mapping.json')
  .option('--db-schema <file>', 'Database schema file (prisma, sql, json)')
  .option('--db-from-env', 'Extract schema from DATABASE_URL in .env')
  .option('--env-path <file>', 'Path to .env file', '.env')
  .option('--manual-mapping <file>', 'Manual URL to source mapping file')
  .option('--framework <type>', 'Source framework (auto-detect if not specified)', '')
  .action(async (options) => {
    console.log('🔍 Smart Rebuild - Analyze Phase');
    console.log(`📂 Source: ${options.source}`);
    console.log(`📸 Capture: ${options.capture}`);

    if (options.framework) {
      console.log(`📦 Framework: ${options.framework} (수동 지정)`);
    } else {
      console.log(`📦 Framework: 자동 감지`);
    }

    if (options.dbFromEnv) {
      console.log(`🔌 DB: DATABASE_URL에서 스키마 추출`);
    } else if (options.dbSchema) {
      console.log(`📄 DB: ${options.dbSchema}`);
    }

    await analyzeSource({
      sourcePath: options.source,
      capturePath: options.capture,
      outputFile: options.output,
      dbSchemaFile: options.dbSchema,
      dbFromEnv: options.dbFromEnv,
      envPath: options.envPath,
      manualMappingFile: options.manualMapping,
      framework: options.framework || undefined,
    });
  });

// Generate command with subcommands
const generateCmd = program
  .command('generate')
  .description('Generate code from mapping (frontend → backend → connect)');

// Generate Frontend (with mock data)
generateCmd
  .command('frontend')
  .description('Generate frontend pages with mock data')
  .option('-m, --mapping <file>', 'Mapping file', './mapping.json')
  .option('-o, --output <dir>', 'Output directory', './output/frontend')
  .option('-f, --framework <type>', 'Frontend framework', 'nextjs')
  .option('-c, --capture <dir>', 'Capture directory (for HTML extraction)')
  .option('--style <type>', 'CSS framework', 'tailwind')
  .action(async (options) => {
    console.log('🎨 Smart Rebuild - Generate Frontend (Mock)');
    console.log(`📋 Mapping: ${options.mapping}`);
    console.log(`📁 Output: ${options.output}`);
    console.log(`🖼️ Framework: ${options.framework}`);
    if (options.capture) {
      console.log(`📸 Capture: ${options.capture}`);
    }

    await generateFrontend({
      mappingFile: options.mapping,
      outputDir: options.output,
      framework: options.framework,
      style: options.style,
      captureDir: options.capture,
    });
  });

// Generate Backend
generateCmd
  .command('backend')
  .description('Generate backend API from mapping')
  .option('-m, --mapping <file>', 'Mapping file', './mapping.json')
  .option('-o, --output <dir>', 'Output directory', './output/backend')
  .option('-b, --framework <type>', 'Backend framework', 'java')
  .action(async (options) => {
    console.log('🔧 Smart Rebuild - Generate Backend');
    console.log(`📋 Mapping: ${options.mapping}`);
    console.log(`📁 Output: ${options.output}`);
    console.log(`⚙️ Framework: ${options.framework}`);

    await generateBackend({
      mappingFile: options.mapping,
      outputDir: options.output,
      framework: options.framework,
    });
  });

// Connect Frontend to Backend
generateCmd
  .command('connect')
  .description('Replace mock data with real API calls')
  .option('-m, --mapping <file>', 'Mapping file', './mapping.json')
  .option('-f, --frontend-dir <dir>', 'Frontend directory', './output/frontend')
  .option('--api-base <url>', 'API base URL', 'http://localhost:8080')
  .action(async (options) => {
    console.log('🔗 Smart Rebuild - Connect Frontend to Backend');
    console.log(`📋 Mapping: ${options.mapping}`);
    console.log(`📁 Frontend: ${options.frontendDir}`);
    console.log(`🌐 API Base: ${options.apiBase}`);

    await connectFrontendToBackend({
      mappingFile: options.mapping,
      frontendDir: options.frontendDir,
      apiBaseUrl: options.apiBase,
    });
  });

// HITL (Human-in-the-Loop) Visual Refinement
generateCmd
  .command('hitl')
  .description('HITL visual refinement - capture and compare original vs local')
  .option('-c, --capture <dir>', 'Capture directory (with sitemap.json)', './capture')
  .option('-p, --page <id>', 'Page ID to process')
  .option('-s, --section <id>', 'Section ID to process')
  .option('--responsive', 'Capture all viewports (desktop, tablet, mobile)')
  .option('--status', 'Show progress status')
  .option('--approve <id>', 'Approve section')
  .option('--skip <id>', 'Skip section')
  .option('--reset', 'Reset HITL state')
  .action(async (options) => {
    console.log('👁️ Smart Rebuild - HITL Visual Refinement');

    // Build args for hitl-refine.ts
    const args: string[] = [];
    args.push('--capture', options.capture);

    if (options.page) args.push('--page=' + options.page);
    if (options.section) args.push('--section=' + options.section);
    if (options.responsive) args.push('--responsive');
    if (options.status) args.push('--status');
    if (options.approve) args.push('--approve=' + options.approve);
    if (options.skip) args.push('--skip=' + options.skip);
    if (options.reset) args.push('--reset');

    // Run hitl-refine.ts
    const hitlScript = path.join(__dirname, '../generate/hitl-refine.ts');

    const child = spawn('npx', ['ts-node', hitlScript, ...args], {
      stdio: 'inherit',
      shell: true,
    });

    child.on('close', (code) => {
      if (code !== 0) {
        console.error(`❌ HITL exited with code ${code}`);
        process.exit(code || 1);
      }
    });
  });

// Full workflow command
program
  .command('run <url>')
  .description('Run full rebuild workflow (capture → analyze → generate)')
  .option('-s, --source <dir>', 'Legacy source directory', './source')
  .option('-o, --output <dir>', 'Output directory', './smart-rebuild-output')
  .option('-b, --backend <type>', 'Backend framework', 'java')
  .option('-f, --frontend <type>', 'Frontend framework', 'nextjs')
  .option('--login', 'Open browser for login before capture')
  .option('--prefetch', '🔴 Capture all pages immediately (default: Lazy Capture)')
  .option('--db-schema <file>', 'Database schema file (prisma, sql, json)')
  .option('--db-from-env', 'Extract schema from DATABASE_URL in .env')
  .option('--env-path <file>', 'Path to .env file', '.env')
  .option('--frontend-only', 'Generate frontend only (skip backend)')
  .action(async (url, options) => {
    console.log('🚀 Smart Rebuild - Full Workflow');
    console.log(`📍 Target: ${url}`);
    console.log(`📂 Source: ${options.source}`);
    console.log(`📁 Output: ${options.output}`);
    console.log(`📸 Capture Mode: ${options.prefetch ? 'Prefetch (즉시 캡처)' : 'Lazy Capture (링크만 수집)'}`);

    // Phase 1: Capture (🔴 Lazy Capture by default)
    console.log('\n📸 Phase 1: Capture');
    await crawlAndCapture(url, {
      outputDir: `${options.output}/capture`,
      maxPages: 100,
      concurrency: 5,
      login: options.login || false,
      prefetch: options.prefetch || false,  // 🔴 Lazy Capture: 기본값 false
    });

    // Phase 2: Analyze
    console.log('\n🔍 Phase 2: Analyze');
    await analyzeSource({
      sourcePath: options.source,
      capturePath: `${options.output}/capture`,
      outputFile: `${options.output}/mapping.json`,
      dbSchemaFile: options.dbSchema,
      dbFromEnv: options.dbFromEnv,
      envPath: options.envPath,
    });

    // Phase 3a: Generate Frontend (with mock data)
    console.log('\n🎨 Phase 3a: Generate Frontend (Mock)');
    await generateFrontend({
      mappingFile: `${options.output}/mapping.json`,
      outputDir: `${options.output}/frontend`,
      framework: options.frontend,
      style: 'tailwind',
      captureDir: `${options.output}/capture`,
    });

    console.log('\n✅ Frontend 생성 완료!');
    console.log(`📁 Frontend: ${options.output}/frontend`);
    console.log('💡 UI를 확인하고, 백엔드 생성을 진행하세요:');
    console.log(`   smart-rebuild generate backend -m ${options.output}/mapping.json -o ${options.output}/backend`);

    if (!options.frontendOnly) {
      // Phase 3b: Generate Backend
      console.log('\n🔧 Phase 3b: Generate Backend');
      await generateBackend({
        mappingFile: `${options.output}/mapping.json`,
        outputDir: `${options.output}/backend`,
        framework: options.backend,
      });

      // Phase 3c: Connect
      console.log('\n🔗 Phase 3c: Connect Frontend to Backend');
      await connectFrontendToBackend({
        mappingFile: `${options.output}/mapping.json`,
        frontendDir: `${options.output}/frontend`,
        apiBaseUrl: 'http://localhost:8080',
      });

      console.log('\n✅ Smart Rebuild Complete!');
      console.log(`📁 Frontend: ${options.output}/frontend`);
      console.log(`📁 Backend: ${options.output}/backend`);
    }
  });

program.parse();
