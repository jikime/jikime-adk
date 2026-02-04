#!/usr/bin/env node

import { Command } from 'commander';
import { crawlAndCapture } from '../capture/crawl';
import { analyzeSource } from '../analyze/classify';
import { generateFrontend } from '../generate/frontend';
import { generateBackend } from '../generate/backend';
import { connectFrontendToBackend } from '../generate/connect';

const program = new Command();

program
  .name('smart-rebuild')
  .description('AI-powered legacy site rebuilding CLI')
  .version('1.2.0');

// Capture command
program
  .command('capture <url>')
  .description('Capture screenshots and HTML from a live site')
  .option('-o, --output <dir>', 'Output directory', './capture')
  .option('-m, --max-pages <n>', 'Maximum pages to capture', '100')
  .option('-c, --concurrency <n>', 'Concurrent page captures', '5')
  .option('-a, --auth <file>', 'Auth session file (JSON)')
  .option('-e, --exclude <patterns>', 'URL patterns to exclude', '/admin/*,/api/*')
  .option('-t, --timeout <ms>', 'Page load timeout', '30000')
  .option('--login', 'Open browser for login, then capture')
  .action(async (url, options) => {
    console.log('🚀 Smart Rebuild - Capture Phase');
    console.log(`📍 Target: ${url}`);
    console.log(`📁 Output: ${options.output}`);

    await crawlAndCapture(url, {
      outputDir: options.output,
      maxPages: parseInt(options.maxPages),
      concurrency: parseInt(options.concurrency),
      authFile: options.auth,
      exclude: options.exclude.split(','),
      timeout: parseInt(options.timeout),
      login: options.login || false,
    });
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
  .action(async (options) => {
    console.log('🔍 Smart Rebuild - Analyze Phase');
    console.log(`📂 Source: ${options.source}`);
    console.log(`📸 Capture: ${options.capture}`);

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
  .option('--style <type>', 'CSS framework', 'tailwind')
  .action(async (options) => {
    console.log('🎨 Smart Rebuild - Generate Frontend (Mock)');
    console.log(`📋 Mapping: ${options.mapping}`);
    console.log(`📁 Output: ${options.output}`);
    console.log(`🖼️ Framework: ${options.framework}`);

    await generateFrontend({
      mappingFile: options.mapping,
      outputDir: options.output,
      framework: options.framework,
      style: options.style,
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

// Full workflow command
program
  .command('run <url>')
  .description('Run full rebuild workflow (capture → analyze → generate)')
  .option('-s, --source <dir>', 'Legacy source directory', './source')
  .option('-o, --output <dir>', 'Output directory', './smart-rebuild-output')
  .option('-b, --backend <type>', 'Backend framework', 'java')
  .option('-f, --frontend <type>', 'Frontend framework', 'nextjs')
  .option('--login', 'Open browser for login before capture')
  .option('--db-schema <file>', 'Database schema file (prisma, sql, json)')
  .option('--db-from-env', 'Extract schema from DATABASE_URL in .env')
  .option('--env-path <file>', 'Path to .env file', '.env')
  .option('--frontend-only', 'Generate frontend only (skip backend)')
  .action(async (url, options) => {
    console.log('🚀 Smart Rebuild - Full Workflow');
    console.log(`📍 Target: ${url}`);
    console.log(`📂 Source: ${options.source}`);
    console.log(`📁 Output: ${options.output}`);

    // Phase 1: Capture
    console.log('\n📸 Phase 1: Capture');
    await crawlAndCapture(url, {
      outputDir: `${options.output}/capture`,
      maxPages: 100,
      concurrency: 5,
      login: options.login || false,
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
