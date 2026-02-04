#!/usr/bin/env node

import { Command } from 'commander';
import { crawlAndCapture, saveLoginSession } from '../capture/crawl';
import { analyzeSource } from '../analyze/classify';
import { generateCode } from '../generate/frontend';

const program = new Command();

program
  .name('smart-rebuild')
  .description('AI-powered legacy site rebuilding CLI')
  .version('1.0.0');

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
  .option('--login', 'Open browser for manual login')
  .option('--save-auth <file>', 'Save auth session to file')
  .action(async (url, options) => {
    console.log('🚀 Smart Rebuild - Capture Phase');
    console.log(`📍 Target: ${url}`);
    console.log(`📁 Output: ${options.output}`);

    if (options.login) {
      await saveLoginSession(url, options.saveAuth || 'auth.json');
      return;
    }

    await crawlAndCapture(url, {
      outputDir: options.output,
      maxPages: parseInt(options.maxPages),
      concurrency: parseInt(options.concurrency),
      authFile: options.auth,
      exclude: options.exclude.split(','),
      timeout: parseInt(options.timeout),
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

// Generate command
program
  .command('generate')
  .description('Generate code from mapping')
  .option('-m, --mapping <file>', 'Mapping file', './mapping.json')
  .option('-b, --backend <type>', 'Backend framework', 'java')
  .option('-f, --frontend <type>', 'Frontend framework', 'nextjs')
  .option('--output-backend <dir>', 'Backend output directory', './output/backend')
  .option('--output-frontend <dir>', 'Frontend output directory', './output/frontend')
  .option('--style <type>', 'CSS framework', 'tailwind')
  .action(async (options) => {
    console.log('⚡ Smart Rebuild - Generate Phase');
    console.log(`📋 Mapping: ${options.mapping}`);
    console.log(`🔧 Backend: ${options.backend}`);
    console.log(`🎨 Frontend: ${options.frontend}`);

    await generateCode({
      mappingFile: options.mapping,
      backend: options.backend,
      frontend: options.frontend,
      outputBackend: options.outputBackend,
      outputFrontend: options.outputFrontend,
      style: options.style,
    });
  });

// Full workflow command
program
  .command('run <url>')
  .description('Run full rebuild workflow')
  .option('-s, --source <dir>', 'Legacy source directory', './source')
  .option('-o, --output <dir>', 'Output directory', './smart-rebuild-output')
  .option('-b, --backend <type>', 'Backend framework', 'java')
  .option('-f, --frontend <type>', 'Frontend framework', 'nextjs')
  .option('--db-schema <file>', 'Database schema file (prisma, sql, json)')
  .option('--db-from-env', 'Extract schema from DATABASE_URL in .env')
  .option('--env-path <file>', 'Path to .env file', '.env')
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

    // Phase 3: Generate
    console.log('\n⚡ Phase 3: Generate');
    await generateCode({
      mappingFile: `${options.output}/mapping.json`,
      backend: options.backend,
      frontend: options.frontend,
      outputBackend: `${options.output}/backend`,
      outputFrontend: `${options.output}/frontend`,
    });

    console.log('\n✅ Smart Rebuild Complete!');
    console.log(`📁 Output: ${options.output}`);
  });

program.parse();
