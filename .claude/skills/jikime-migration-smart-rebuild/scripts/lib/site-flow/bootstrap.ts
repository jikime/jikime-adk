/**
 * site-flow CLI 부트스트랩 모듈
 *
 * NextAuth v5 credentials 로그인을 통해 site-flow 서버에 인증하고,
 * 사이트 생성/조회 및 API 키 발급까지 수행하는 CLI 전용 모듈입니다.
 *
 * SiteFlowClient(Bearer API 키 인증)와 달리, 이 모듈은 세션 쿠키 인증을
 * 사용하여 아직 API 키가 없는 초기 설정 단계에서 사용됩니다.
 *
 * 설계 원칙:
 * - Node.js 20+ 내장 fetch만 사용 (외부 의존성 없음)
 * - 수동 쿠키 관리 (Set-Cookie 헤더 파싱 + Cookie 헤더 전송)
 * - NextAuth v5 JWT 전략 호환 (credentials provider)
 * - 단계별 오류 처리 (csrf → login → site → apikey)
 *
 * @example
 * ```typescript
 * import { bootstrapSiteFlow } from '../lib/site-flow';
 *
 * const result = await bootstrapSiteFlow({
 *   apiUrl: 'http://localhost:3000',
 *   email: 'admin@example.com',
 *   password: 'password123',
 *   siteName: 'My Migration Site',
 *   siteUrl: 'https://legacy-site.com',
 * });
 *
 * console.log(result.apiKey);  // sf_xxx...
 * console.log(result.siteId);  // MongoDB ObjectId
 * ```
 */

import type { SiteResponse, ApiKeyCreateResponse } from './types';

// =============================================================================
// 타입 정의 (Bootstrap Types)
// =============================================================================

/**
 * 부트스트랩 옵션
 *
 * CLI에서 site-flow 서버에 로그인하고 사이트 + API 키를 생성하기 위한
 * 초기 설정 파라미터입니다.
 */
export interface BootstrapOptions {
  /** site-flow 서버 URL (예: http://localhost:3000) */
  apiUrl: string;
  /** 로그인 이메일 */
  email: string;
  /** 로그인 비밀번호 */
  password: string;
  /** 생성하거나 찾을 사이트 이름 */
  siteName: string;
  /** 기존 사이트 검색에 사용할 사이트 URL (선택) */
  siteUrl?: string;
  /** API 키 이름 (기본값: "migration-cli") */
  apiKeyName?: string;
}

/**
 * 부트스트랩 결과
 *
 * 부트스트랩 완료 후 반환되는 사이트 ID와 API 키 정보입니다.
 * 이 결과를 SiteFlowConfig에 설정하여 SiteFlowClient를 초기화할 수 있습니다.
 */
export interface BootstrapResult {
  /** 생성되거나 찾은 사이트 ID (MongoDB ObjectId) */
  siteId: string;
  /** 발급된 API 키 (sf_xxx 형식, 1회만 반환) */
  apiKey: string;
  /** 사이트 이름 */
  siteName: string;
  /** 기존 사이트를 찾은 경우 true, 새로 생성한 경우 false */
  isExistingSite: boolean;
}

/**
 * 부트스트랩 실패 단계
 */
export type BootstrapPhase = 'csrf' | 'login' | 'site' | 'apikey';

// =============================================================================
// 오류 클래스 (Error Class)
// =============================================================================

/**
 * 부트스트랩 오류
 *
 * 부트스트랩 과정의 특정 단계에서 실패한 경우 발생합니다.
 * `phase` 필드를 통해 어느 단계에서 실패했는지 확인할 수 있습니다.
 *
 * @example
 * ```typescript
 * try {
 *   await bootstrapSiteFlow(options);
 * } catch (err) {
 *   if (err instanceof BootstrapError) {
 *     console.error(`단계: ${err.phase}, 메시지: ${err.message}`);
 *   }
 * }
 * ```
 */
export class BootstrapError extends Error {
  constructor(
    message: string,
    /** 실패한 부트스트랩 단계 */
    public readonly phase: BootstrapPhase,
    /** 원인 오류 */
    public readonly cause?: Error,
  ) {
    super(message);
    this.name = 'BootstrapError';
  }
}

// =============================================================================
// 상수 (Constants)
// =============================================================================

/** 기본 API 키 이름 */
const DEFAULT_API_KEY_NAME = 'migration-cli';

/** 기본 요청 타임아웃 (밀리초) */
const DEFAULT_TIMEOUT_MS = 15_000;

// =============================================================================
// 내부 유틸리티: 쿠키 관리 (Cookie Management)
// =============================================================================

/**
 * Response의 Set-Cookie 헤더에서 쿠키를 추출하여 Map에 병합합니다.
 *
 * 여러 Set-Cookie 헤더가 있을 수 있으므로 getSetCookie()를 사용합니다.
 * 각 Set-Cookie 값에서 `name=value` 부분만 추출하고 나머지 속성(Path, HttpOnly 등)은 무시합니다.
 *
 * @param response - fetch 응답 객체
 * @param existing - 기존 쿠키 Map (병합 대상)
 * @returns 병합된 쿠키 Map (기존 + 새로운 쿠키)
 */
function extractCookies(
  response: Response,
  existing: Map<string, string> = new Map(),
): Map<string, string> {
  const merged = new Map(existing);

  // getSetCookie()는 Node.js 20+에서 사용 가능
  const setCookieHeaders = response.headers.getSetCookie();

  for (const header of setCookieHeaders) {
    // Set-Cookie 형식: "name=value; Path=/; HttpOnly; ..."
    // 첫 번째 세미콜론 이전의 name=value 부분만 추출
    const nameValuePart = header.split(';')[0]?.trim();
    if (!nameValuePart) continue;

    const eqIndex = nameValuePart.indexOf('=');
    if (eqIndex === -1) continue;

    const name = nameValuePart.substring(0, eqIndex);
    const value = nameValuePart.substring(eqIndex + 1);
    merged.set(name, value);
  }

  return merged;
}

/**
 * 쿠키 Map을 Cookie 헤더 값 문자열로 변환합니다.
 *
 * @param cookies - 쿠키 Map (name → value)
 * @returns Cookie 헤더 값 (예: "csrfToken=abc; session-token=xyz")
 */
function formatCookieHeader(cookies: Map<string, string>): string {
  const parts: string[] = [];
  cookies.forEach((value, name) => {
    parts.push(`${name}=${value}`);
  });
  return parts.join('; ');
}

/**
 * 기본 URL에서 후행 슬래시를 제거합니다.
 *
 * @param url - 정규화할 URL 문자열
 * @returns 후행 슬래시가 제거된 URL
 */
function normalizeBaseUrl(url: string): string {
  return url.replace(/\/+$/, '');
}

// =============================================================================
// 내부 함수: 인증 단계 (Authentication Steps)
// =============================================================================

/**
 * Step 1: CSRF 토큰을 가져옵니다.
 *
 * NextAuth v5의 /api/auth/csrf 엔드포인트에 GET 요청을 보내
 * csrfToken 값과 Set-Cookie에 설정된 쿠키를 추출합니다.
 *
 * @param apiUrl - 정규화된 서버 URL
 * @returns csrfToken 문자열과 쿠키 Map
 * @throws {BootstrapError} CSRF 토큰 가져오기 실패 시 (phase: 'csrf')
 */
async function getCsrfToken(
  apiUrl: string,
): Promise<{ csrfToken: string; cookies: Map<string, string> }> {
  const url = `${apiUrl}/api/auth/csrf`;

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), DEFAULT_TIMEOUT_MS);

  try {
    const response = await fetch(url, {
      method: 'GET',
      signal: controller.signal,
    });

    if (!response.ok) {
      throw new BootstrapError(
        `[bootstrap] CSRF 토큰 가져오기 실패: ${response.status} ${response.statusText}`,
        'csrf',
      );
    }

    // 응답 JSON에서 csrfToken 추출
    const data = (await response.json()) as { csrfToken?: string };
    if (!data.csrfToken) {
      throw new BootstrapError(
        '[bootstrap] CSRF 응답에 csrfToken 필드가 없습니다.',
        'csrf',
      );
    }

    // Set-Cookie 헤더에서 쿠키 추출
    const cookies = extractCookies(response);

    return { csrfToken: data.csrfToken, cookies };
  } catch (err) {
    if (err instanceof BootstrapError) throw err;

    const cause = err instanceof Error ? err : new Error(String(err));
    throw new BootstrapError(
      `[bootstrap] CSRF 토큰 요청 실패: ${cause.message}`,
      'csrf',
      cause,
    );
  } finally {
    clearTimeout(timeoutId);
  }
}

/**
 * Step 2: credentials 로그인을 수행합니다.
 *
 * NextAuth v5의 /api/auth/callback/credentials 엔드포인트에
 * URL-encoded form 데이터를 POST합니다.
 *
 * redirect: 'manual'을 사용하여 302 리다이렉트 응답의 Set-Cookie 헤더에서
 * 세션 쿠키를 직접 추출합니다.
 *
 * @param apiUrl - 정규화된 서버 URL
 * @param email - 로그인 이메일
 * @param password - 로그인 비밀번호
 * @param csrfToken - Step 1에서 받은 CSRF 토큰
 * @param cookies - Step 1에서 받은 쿠키 (CSRF 쿠키 포함)
 * @returns 세션 쿠키가 병합된 쿠키 Map
 * @throws {BootstrapError} 로그인 실패 시 (phase: 'login')
 */
async function credentialsLogin(
  apiUrl: string,
  email: string,
  password: string,
  csrfToken: string,
  cookies: Map<string, string>,
): Promise<Map<string, string>> {
  const url = `${apiUrl}/api/auth/callback/credentials`;

  // URL-encoded form 데이터 생성
  const formData = new URLSearchParams();
  formData.set('email', email);
  formData.set('password', password);
  formData.set('csrfToken', csrfToken);
  formData.set('redirect', 'false');
  formData.set('callbackUrl', '/');
  formData.set('json', 'true');

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), DEFAULT_TIMEOUT_MS);

  try {
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
        Cookie: formatCookieHeader(cookies),
      },
      body: formData.toString(),
      redirect: 'manual',
      signal: controller.signal,
    });

    // 302 리다이렉트 또는 200 OK가 성공 응답
    // NextAuth는 redirect: false + json: true 설정 시 200 JSON 응답을 반환하기도 함
    if (response.status !== 302 && response.status !== 200) {
      throw new BootstrapError(
        `[bootstrap] 로그인 실패: ${response.status} ${response.statusText} - ` +
        '잘못된 인증 정보이거나 사용자가 승인되지 않았습니다.',
        'login',
      );
    }

    // Set-Cookie에서 세션 쿠키 추출 및 기존 쿠키와 병합
    const sessionCookies = extractCookies(response, cookies);

    // 세션 쿠키 존재 확인
    // HTTP: authjs.session-token / HTTPS: __Secure-authjs.session-token
    const hasSessionCookie =
      sessionCookies.has('authjs.session-token') ||
      sessionCookies.has('__Secure-authjs.session-token');

    if (!hasSessionCookie) {
      throw new BootstrapError(
        '[bootstrap] 로그인 응답에 세션 쿠키가 없습니다. ' +
        '잘못된 인증 정보이거나 사용자가 승인되지 않았습니다.',
        'login',
      );
    }

    return sessionCookies;
  } catch (err) {
    if (err instanceof BootstrapError) throw err;

    const cause = err instanceof Error ? err : new Error(String(err));
    throw new BootstrapError(
      `[bootstrap] 로그인 요청 실패: ${cause.message}`,
      'login',
      cause,
    );
  } finally {
    clearTimeout(timeoutId);
  }
}

// =============================================================================
// 내부 함수: 리소스 관리 (Resource Management)
// =============================================================================

/**
 * 세션 쿠키 인증으로 기존 사이트를 URL로 검색합니다.
 *
 * GET /api/sites 엔드포인트는 현재 사용자의 사이트 배열을 반환합니다.
 * 해당 배열에서 url 필드가 일치하는 사이트를 찾습니다.
 *
 * @param apiUrl - 정규화된 서버 URL
 * @param siteUrl - 검색할 사이트 URL
 * @param cookies - 세션 쿠키가 포함된 쿠키 Map
 * @returns 찾은 사이트 응답 또는 null
 * @throws {BootstrapError} API 호출 실패 시 (phase: 'site')
 */
async function findSiteByUrlWithSession(
  apiUrl: string,
  siteUrl: string,
  cookies: Map<string, string>,
): Promise<SiteResponse | null> {
  const url = `${apiUrl}/api/sites`;

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), DEFAULT_TIMEOUT_MS);

  try {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        Cookie: formatCookieHeader(cookies),
      },
      signal: controller.signal,
    });

    if (!response.ok) {
      throw new BootstrapError(
        `[bootstrap] 사이트 목록 조회 실패: ${response.status} ${response.statusText}`,
        'site',
      );
    }

    const sites = (await response.json()) as SiteResponse[];

    // URL이 일치하는 사이트 검색
    const found = sites.find((site) => site.url === siteUrl);
    return found ?? null;
  } catch (err) {
    if (err instanceof BootstrapError) throw err;

    const cause = err instanceof Error ? err : new Error(String(err));
    throw new BootstrapError(
      `[bootstrap] 사이트 검색 실패: ${cause.message}`,
      'site',
      cause,
    );
  } finally {
    clearTimeout(timeoutId);
  }
}

/**
 * 세션 쿠키 인증으로 새 사이트를 생성합니다.
 *
 * @param apiUrl - 정규화된 서버 URL
 * @param data - 사이트 생성 데이터 (name, url, description)
 * @param cookies - 세션 쿠키가 포함된 쿠키 Map
 * @returns 생성된 사이트 응답
 * @throws {BootstrapError} 사이트 생성 실패 시 (phase: 'site')
 */
async function createSiteWithSession(
  apiUrl: string,
  data: { name: string; url?: string; description?: string },
  cookies: Map<string, string>,
): Promise<SiteResponse> {
  const url = `${apiUrl}/api/sites`;

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), DEFAULT_TIMEOUT_MS);

  try {
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Cookie: formatCookieHeader(cookies),
      },
      body: JSON.stringify(data),
      signal: controller.signal,
    });

    if (!response.ok) {
      const errorText = await response.text().catch(() => '');
      throw new BootstrapError(
        `[bootstrap] 사이트 생성 실패: ${response.status} ${response.statusText}` +
        (errorText ? ` - ${errorText}` : ''),
        'site',
      );
    }

    return (await response.json()) as SiteResponse;
  } catch (err) {
    if (err instanceof BootstrapError) throw err;

    const cause = err instanceof Error ? err : new Error(String(err));
    throw new BootstrapError(
      `[bootstrap] 사이트 생성 요청 실패: ${cause.message}`,
      'site',
      cause,
    );
  } finally {
    clearTimeout(timeoutId);
  }
}

/**
 * 세션 쿠키 인증으로 API 키를 생성합니다.
 *
 * POST /api/api-keys 엔드포인트를 호출하여 새 API 키를 발급받습니다.
 * 반환되는 `key` 필드에는 전체 API 키(sf_xxx...)가 포함되며,
 * 이 값은 생성 직후 1회만 반환됩니다.
 *
 * @param apiUrl - 정규화된 서버 URL
 * @param siteId - API 키를 연결할 사이트 ID
 * @param name - API 키 이름 (식별용)
 * @param cookies - 세션 쿠키가 포함된 쿠키 Map
 * @returns 전체 API 키가 포함된 생성 응답
 * @throws {BootstrapError} API 키 생성 실패 시 (phase: 'apikey')
 */
async function createApiKeyWithSession(
  apiUrl: string,
  siteId: string,
  name: string,
  cookies: Map<string, string>,
): Promise<ApiKeyCreateResponse> {
  const url = `${apiUrl}/api/api-keys`;

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), DEFAULT_TIMEOUT_MS);

  try {
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Cookie: formatCookieHeader(cookies),
      },
      body: JSON.stringify({ siteId, name }),
      signal: controller.signal,
    });

    if (!response.ok) {
      const errorText = await response.text().catch(() => '');
      throw new BootstrapError(
        `[bootstrap] API 키 생성 실패: ${response.status} ${response.statusText}` +
        (errorText ? ` - ${errorText}` : ''),
        'apikey',
      );
    }

    const result = (await response.json()) as ApiKeyCreateResponse;

    if (!result.key) {
      throw new BootstrapError(
        '[bootstrap] API 키 응답에 key 필드가 없습니다.',
        'apikey',
      );
    }

    return result;
  } catch (err) {
    if (err instanceof BootstrapError) throw err;

    const cause = err instanceof Error ? err : new Error(String(err));
    throw new BootstrapError(
      `[bootstrap] API 키 생성 요청 실패: ${cause.message}`,
      'apikey',
      cause,
    );
  } finally {
    clearTimeout(timeoutId);
  }
}

// =============================================================================
// 메인 함수 (Main Function)
// =============================================================================

/**
 * site-flow CLI 부트스트랩을 수행합니다.
 *
 * 다음 단계를 순차적으로 실행합니다:
 * 1. CSRF 토큰 가져오기 (GET /api/auth/csrf)
 * 2. credentials 로그인 (POST /api/auth/callback/credentials)
 * 3. 사이트 찾기 또는 생성 (GET/POST /api/sites)
 * 4. API 키 생성 (POST /api/api-keys)
 *
 * 반환된 API 키와 사이트 ID를 SiteFlowConfig에 설정하면
 * 이후 SiteFlowClient(Bearer 인증)를 사용하여 API를 호출할 수 있습니다.
 *
 * @param options - 부트스트랩 옵션 (서버 URL, 인증 정보, 사이트 정보)
 * @returns 사이트 ID, API 키, 사이트 이름, 기존 사이트 여부
 * @throws {BootstrapError} 각 단계에서 실패 시 (phase 필드로 단계 구분)
 *
 * @example
 * ```typescript
 * const result = await bootstrapSiteFlow({
 *   apiUrl: 'http://localhost:3000',
 *   email: 'admin@example.com',
 *   password: 'password123',
 *   siteName: 'Legacy Site Migration',
 *   siteUrl: 'https://legacy.example.com',
 *   apiKeyName: 'cli-key',
 * });
 *
 * // SiteFlowConfig로 변환
 * const config: SiteFlowConfig = {
 *   apiUrl: 'http://localhost:3000',
 *   apiKey: result.apiKey,
 *   siteId: result.siteId,
 *   enabled: true,
 * };
 * ```
 */
export async function bootstrapSiteFlow(
  options: BootstrapOptions,
): Promise<BootstrapResult> {
  const apiUrl = normalizeBaseUrl(options.apiUrl);
  const apiKeyName = options.apiKeyName ?? DEFAULT_API_KEY_NAME;

  // -------------------------------------------------------------------------
  // Step 1: CSRF 토큰 가져오기
  // -------------------------------------------------------------------------
  const { csrfToken, cookies: csrfCookies } = await getCsrfToken(apiUrl);

  // -------------------------------------------------------------------------
  // Step 2: credentials 로그인
  // -------------------------------------------------------------------------
  const sessionCookies = await credentialsLogin(
    apiUrl,
    options.email,
    options.password,
    csrfToken,
    csrfCookies,
  );

  // -------------------------------------------------------------------------
  // Step 3: 사이트 찾기 또는 생성
  // -------------------------------------------------------------------------
  let site: SiteResponse | null = null;
  let isExistingSite = false;

  // siteUrl이 지정된 경우 기존 사이트 검색 시도
  if (options.siteUrl) {
    site = await findSiteByUrlWithSession(
      apiUrl,
      options.siteUrl,
      sessionCookies,
    );

    if (site) {
      isExistingSite = true;
    }
  }

  // 기존 사이트를 찾지 못한 경우 새로 생성
  if (!site) {
    site = await createSiteWithSession(
      apiUrl,
      {
        name: options.siteName,
        url: options.siteUrl,
      },
      sessionCookies,
    );
  }

  // -------------------------------------------------------------------------
  // Step 4: API 키 생성
  // -------------------------------------------------------------------------
  const apiKeyResponse = await createApiKeyWithSession(
    apiUrl,
    site._id,
    apiKeyName,
    sessionCookies,
  );

  return {
    siteId: site._id,
    apiKey: apiKeyResponse.key,
    siteName: site.name,
    isExistingSite,
  };
}
