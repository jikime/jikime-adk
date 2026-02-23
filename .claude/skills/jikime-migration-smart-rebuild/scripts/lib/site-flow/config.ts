/**
 * site-flow 설정 로더 모듈
 *
 * CLI 마이그레이션 도구에서 사용하는 site-flow API 클라이언트 설정을
 * 다음 우선순위에 따라 로드합니다:
 *   1. 명시적 파라미터 (가장 높은 우선순위)
 *   2. 환경변수
 *   3. .migrate-config.yaml 파일 (가장 낮은 우선순위)
 *
 * 외부 의존성 없이 Node.js 기본 모듈만 사용합니다.
 */

import * as fs from 'fs';
import * as path from 'path';

import type { SiteFlowConfig } from './types';

// =============================================================================
// 내부 타입 (Internal Types)
// =============================================================================

/**
 * loadSiteFlowConfig 함수의 옵션 파라미터
 */
export interface LoadSiteFlowConfigOptions {
  /**
   * .migrate-config.yaml 파일의 디렉토리 경로
   * 지정하지 않으면 process.cwd() 사용
   */
  configPath?: string;
  /**
   * true로 설정 시 site-flow를 비활성화한 설정 반환
   * --skip-site-flow 플래그에 해당
   */
  skipSiteFlow?: boolean;
  /**
   * 로드된 값을 덮어쓸 부분 설정
   * 명시적 파라미터(최고 우선순위)에 해당
   */
  overrides?: Partial<SiteFlowConfig>;
}

/**
 * .migrate-config.yaml 파일의 site_flow 섹션 YAML 구조
 * (snake_case - YAML 관례)
 */
interface YamlSiteFlowSection {
  api_url?: string;
  api_key?: string;
  site_id?: string;
  enabled?: boolean;
}

// =============================================================================
// 기본값 상수 (Default Constants)
// =============================================================================

/**
 * site-flow 설정의 기본값
 */
const DEFAULT_CONFIG: SiteFlowConfig = {
  apiUrl: '',
  apiKey: '',
  siteId: '',
  enabled: false,
};

/**
 * .migrate-config.yaml 파일명
 */
const CONFIG_FILE_NAME = '.migrate-config.yaml';

// =============================================================================
// YAML 파서/직렬화 (Simple YAML Parser/Serializer)
// =============================================================================

/**
 * 단순 YAML 문자열에서 key: value 쌍을 파싱합니다.
 * 인덴트된 섹션 구조를 처리하며, 따옴표로 감싸인 문자열 값을 지원합니다.
 *
 * 지원 형식 예시:
 *   site_flow:
 *     api_url: "https://example.com"
 *     api_key: sf_xxxx
 *     enabled: true
 *
 * @param yamlText 파싱할 YAML 문자열
 * @returns 섹션명을 키로 하고 내부 키-값 맵을 값으로 하는 객체
 */
function parseSimpleYaml(yamlText: string): Record<string, Record<string, string>> {
  const result: Record<string, Record<string, string>> = {};
  const lines = yamlText.split('\n');

  let currentSection: string | null = null;

  for (const rawLine of lines) {
    // 빈 줄 및 주석 건너뜀
    if (rawLine.trim() === '' || rawLine.trim().startsWith('#')) {
      continue;
    }

    // 최상위 섹션 감지: 인덴트 없이 콜론으로 끝나는 키 (예: "site_flow:")
    const topLevelMatch = rawLine.match(/^([a-zA-Z_][a-zA-Z0-9_]*):\s*$/);
    if (topLevelMatch) {
      currentSection = topLevelMatch[1];
      result[currentSection] = {};
      continue;
    }

    // 최상위 인라인 값 감지 (예: "version: 1.0") - 섹션이 아닌 경우 무시
    const topLevelInlineMatch = rawLine.match(/^([a-zA-Z_][a-zA-Z0-9_]*):\s+(.+)$/);
    if (topLevelInlineMatch && !rawLine.startsWith(' ') && !rawLine.startsWith('\t')) {
      // 최상위 인라인 키-값 쌍은 섹션 없이 처리
      currentSection = null;
      continue;
    }

    // 인덴트된 키-값 쌍 파싱 (섹션 내부)
    if (currentSection !== null) {
      // 인덴트 감지 (공백 또는 탭으로 시작)
      const indentedMatch = rawLine.match(/^[\s\t]+([a-zA-Z_][a-zA-Z0-9_]*):\s*(.*)$/);
      if (indentedMatch) {
        const key = indentedMatch[1];
        const rawValue = indentedMatch[2].trim();

        // 따옴표 제거 (큰따옴표 또는 작은따옴표)
        let value = rawValue;
        if (
          (rawValue.startsWith('"') && rawValue.endsWith('"')) ||
          (rawValue.startsWith("'") && rawValue.endsWith("'"))
        ) {
          value = rawValue.slice(1, -1);
        }

        result[currentSection][key] = value;
      }
    }
  }

  return result;
}

/**
 * SiteFlowConfig 객체를 .migrate-config.yaml의 site_flow 섹션 YAML 문자열로 직렬화합니다.
 *
 * @param config 직렬화할 SiteFlowConfig
 * @returns YAML 섹션 문자열
 */
function serializeSiteFlowSection(config: SiteFlowConfig): string {
  const lines: string[] = ['site_flow:'];

  // 값에 특수문자가 포함된 경우 큰따옴표로 감쌈
  const quoteIfNeeded = (value: string): string => {
    if (value === '') return '""';
    // 공백, 콜론, 해시, 대괄호, 중괄호, 쉼표 등이 포함된 경우 따옴표 사용
    if (/[\s:#\[\]{},]/.test(value)) {
      return `"${value.replace(/"/g, '\\"')}"`;
    }
    return value;
  };

  lines.push(`  api_url: ${quoteIfNeeded(config.apiUrl)}`);
  lines.push(`  api_key: ${quoteIfNeeded(config.apiKey)}`);
  lines.push(`  site_id: ${quoteIfNeeded(config.siteId)}`);
  lines.push(`  enabled: ${config.enabled}`);

  return lines.join('\n');
}

// =============================================================================
// YAML 파일 읽기/쓰기 (YAML File Read/Write)
// =============================================================================

/**
 * .migrate-config.yaml 파일에서 site_flow 섹션을 읽어 반환합니다.
 * 파일이 없거나 읽기 실패 시 빈 객체를 반환합니다.
 *
 * @param configFilePath .migrate-config.yaml 파일 경로
 * @returns site_flow 섹션의 키-값 맵
 */
function readSiteFlowFromYaml(configFilePath: string): YamlSiteFlowSection {
  if (!fs.existsSync(configFilePath)) {
    return {};
  }

  try {
    const content = fs.readFileSync(configFilePath, 'utf-8');
    const parsed = parseSimpleYaml(content);
    const siteFlowSection = parsed['site_flow'] ?? {};

    const result: YamlSiteFlowSection = {};

    if (typeof siteFlowSection['api_url'] === 'string') {
      result.api_url = siteFlowSection['api_url'];
    }
    if (typeof siteFlowSection['api_key'] === 'string') {
      result.api_key = siteFlowSection['api_key'];
    }
    if (typeof siteFlowSection['site_id'] === 'string') {
      result.site_id = siteFlowSection['site_id'];
    }
    if (typeof siteFlowSection['enabled'] === 'string') {
      // YAML에서 boolean은 문자열로 파싱됨
      result.enabled = siteFlowSection['enabled'].toLowerCase() === 'true';
    }

    return result;
  } catch (error) {
    // 파일 읽기 또는 파싱 실패 시 stderr에 경고 출력 후 빈 객체 반환
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(`[site-flow] 설정 파일 읽기 실패: ${configFilePath}\n  원인: ${message}\n`);
    return {};
  }
}

/**
 * .migrate-config.yaml 파일 전체 내용을 읽어 문자열로 반환합니다.
 * 파일이 없으면 빈 문자열을 반환합니다.
 *
 * @param configFilePath .migrate-config.yaml 파일 경로
 * @returns 파일 내용 문자열
 */
function readRawYaml(configFilePath: string): string {
  if (!fs.existsSync(configFilePath)) {
    return '';
  }

  try {
    return fs.readFileSync(configFilePath, 'utf-8');
  } catch {
    return '';
  }
}

// =============================================================================
// 환경변수 읽기 (Environment Variable Reader)
// =============================================================================

/**
 * 환경변수에서 site-flow 설정을 읽어 반환합니다.
 * 설정되지 않은 환경변수는 undefined로 반환됩니다.
 *
 * 지원 환경변수:
 *   - SITE_FLOW_API_URL
 *   - SITE_FLOW_API_KEY
 *   - SITE_FLOW_SITE_ID
 *   - SITE_FLOW_ENABLED ("true" / "false")
 *
 * @returns 환경변수에서 읽은 부분 SiteFlowConfig
 */
function readSiteFlowFromEnv(): Partial<SiteFlowConfig> {
  const env = process.env;
  const result: Partial<SiteFlowConfig> = {};

  if (env['SITE_FLOW_API_URL'] !== undefined) {
    result.apiUrl = env['SITE_FLOW_API_URL'];
  }
  if (env['SITE_FLOW_API_KEY'] !== undefined) {
    result.apiKey = env['SITE_FLOW_API_KEY'];
  }
  if (env['SITE_FLOW_SITE_ID'] !== undefined) {
    result.siteId = env['SITE_FLOW_SITE_ID'];
  }
  if (env['SITE_FLOW_ENABLED'] !== undefined) {
    result.enabled = env['SITE_FLOW_ENABLED'].toLowerCase() === 'true';
  }

  return result;
}

// =============================================================================
// 공개 API (Public API)
// =============================================================================

/**
 * site-flow 설정을 다음 우선순위로 로드합니다:
 *   1. options.overrides (가장 높은 우선순위 - 명시적 파라미터)
 *   2. 환경변수 (SITE_FLOW_*)
 *   3. .migrate-config.yaml 파일 (가장 낮은 우선순위)
 *
 * options.skipSiteFlow가 true이면 다른 설정과 관계없이
 * enabled: false인 설정을 즉시 반환합니다.
 *
 * @param options 설정 로드 옵션
 * @returns 로드된 SiteFlowConfig
 *
 * @example
 * // 기본 사용 (cwd의 .migrate-config.yaml + 환경변수)
 * const config = await loadSiteFlowConfig();
 *
 * @example
 * // --skip-site-flow 플래그 처리
 * const config = await loadSiteFlowConfig({ skipSiteFlow: true });
 * // => { enabled: false, apiUrl: '', apiKey: '', siteId: '' }
 *
 * @example
 * // 명시적 파라미터로 재정의
 * const config = await loadSiteFlowConfig({
 *   overrides: { apiKey: 'sf_override_key', siteId: 'abc123' }
 * });
 */
export function loadSiteFlowConfig(options: LoadSiteFlowConfigOptions = {}): SiteFlowConfig {
  const { configPath, skipSiteFlow = false, overrides } = options;

  // --skip-site-flow 플래그가 설정된 경우 즉시 비활성화 설정 반환
  if (skipSiteFlow) {
    return {
      ...DEFAULT_CONFIG,
      enabled: false,
    };
  }

  // 설정 파일 경로 결정
  const baseDir = configPath ?? process.cwd();
  const configFilePath = path.join(baseDir, CONFIG_FILE_NAME);

  // 우선순위 3 (최저): YAML 파일에서 로드
  const fromYaml = readSiteFlowFromYaml(configFilePath);

  // 우선순위 2: 환경변수에서 로드
  const fromEnv = readSiteFlowFromEnv();

  // 우선순위에 따라 병합: DEFAULT < YAML < ENV < overrides
  const merged: SiteFlowConfig = {
    apiUrl:
      overrides?.apiUrl ??
      fromEnv.apiUrl ??
      fromYaml.api_url ??
      DEFAULT_CONFIG.apiUrl,

    apiKey:
      overrides?.apiKey ??
      fromEnv.apiKey ??
      fromYaml.api_key ??
      DEFAULT_CONFIG.apiKey,

    siteId:
      overrides?.siteId ??
      fromEnv.siteId ??
      fromYaml.site_id ??
      DEFAULT_CONFIG.siteId,

    enabled:
      overrides?.enabled ??
      fromEnv.enabled ??
      fromYaml.enabled ??
      DEFAULT_CONFIG.enabled,
  };

  return merged;
}

/**
 * .migrate-config.yaml 파일의 site_flow 섹션을 저장(또는 업데이트)합니다.
 * 파일의 다른 섹션은 보존됩니다.
 *
 * 파일이 없으면 새로 생성하고, 있으면 site_flow 섹션만 교체합니다.
 *
 * @param configPath .migrate-config.yaml 파일이 위치한 디렉토리 경로
 * @param config 저장할 SiteFlowConfig
 * @throws 파일 쓰기 실패 시 Error를 던집니다
 *
 * @example
 * await saveSiteFlowConfig('/path/to/project', {
 *   apiUrl: 'https://api.site-flow.io',
 *   apiKey: 'sf_xxxx',
 *   siteId: '66a1b2c3d4e5f6789',
 *   enabled: true,
 * });
 */
export function saveSiteFlowConfig(configPath: string, config: SiteFlowConfig): void {
  const configFilePath = path.join(configPath, CONFIG_FILE_NAME);
  const newSiteFlowYaml = serializeSiteFlowSection(config);

  let existingContent = readRawYaml(configFilePath);

  if (existingContent === '') {
    // 파일이 없거나 비어있는 경우: 새로 생성
    const newContent = `${newSiteFlowYaml}\n`;
    fs.writeFileSync(configFilePath, newContent, 'utf-8');
    return;
  }

  // 기존 파일에서 site_flow 섹션을 찾아 교체
  // 패턴: "site_flow:" 로 시작하는 줄부터 다음 최상위 섹션(들여쓰기 없는 키:)까지
  const siteFlowSectionPattern = /^site_flow:\s*\n([ \t]+[^\n]*\n)*/m;

  if (siteFlowSectionPattern.test(existingContent)) {
    // 기존 site_flow 섹션을 새 내용으로 교체
    existingContent = existingContent.replace(
      siteFlowSectionPattern,
      `${newSiteFlowYaml}\n`
    );
  } else {
    // site_flow 섹션이 없으면 파일 끝에 추가
    // 파일이 개행으로 끝나지 않는 경우 개행 추가
    if (!existingContent.endsWith('\n')) {
      existingContent += '\n';
    }
    existingContent += `${newSiteFlowYaml}\n`;
  }

  try {
    fs.writeFileSync(configFilePath, existingContent, 'utf-8');
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`[site-flow] 설정 파일 저장 실패: ${configFilePath}\n  원인: ${message}`);
  }
}

/**
 * SiteFlowConfig가 유효하고 활성화되어 있는지 확인합니다.
 *
 * 다음 조건을 모두 만족해야 활성화 상태로 간주합니다:
 *   - enabled가 true
 *   - apiUrl이 비어있지 않음
 *   - apiKey가 비어있지 않음
 *   - siteId가 비어있지 않음
 *
 * @param config 확인할 SiteFlowConfig
 * @returns 유효하고 활성화된 경우 true, 그렇지 않으면 false
 *
 * @example
 * const config = loadSiteFlowConfig();
 * if (isSiteFlowEnabled(config)) {
 *   // site-flow API 호출
 * }
 */
export function isSiteFlowEnabled(config: SiteFlowConfig): boolean {
  return (
    config.enabled === true &&
    config.apiUrl.trim() !== '' &&
    config.apiKey.trim() !== '' &&
    config.siteId.trim() !== ''
  );
}

// =============================================================================
// 타입 재내보내기 (Re-export Types)
// =============================================================================

export type { SiteFlowConfig };
