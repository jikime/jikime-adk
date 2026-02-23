/**
 * site-flow API HTTP 클라이언트 모듈
 *
 * site-flow REST API와의 모든 HTTP 통신을 담당하는 핵심 모듈입니다.
 * 모든 엔드포인트 모듈은 이 모듈에서 클라이언트 인스턴스를 가져와 사용합니다.
 *
 * 설계 원칙:
 * - Node.js 20+ 내장 fetch 사용 (외부 HTTP 라이브러리 없음)
 * - Bearer 토큰 인증: Authorization: Bearer {apiKey}
 * - 지수 백오프 재시도: 1s → 2s → 4s (최대 3회, SPEC EC-2)
 * - AbortController 기반 타임아웃 (기본 30초, 이미지 업로드 120초)
 * - 연결 실패 시 타입된 오류를 던져 호출자가 중단 여부를 결정
 */

import type { SiteFlowConfig, ClientOptions, ApiError } from './types';

// =============================================================================
// 상수 (Constants)
// =============================================================================

/** 기본 요청 타임아웃 (밀리초) */
const DEFAULT_TIMEOUT_MS = 30_000;

/** 기본 재시도 횟수 */
const DEFAULT_RETRY_COUNT = 3;

/** 기본 재시도 초기 지연 (밀리초) - 지수 백오프 기준값 */
const DEFAULT_RETRY_DELAY_MS = 1_000;

/** 재시도 가능한 HTTP 상태 코드 (5xx 서버 오류) */
const RETRYABLE_STATUS_MIN = 500;

// =============================================================================
// 오류 클래스 (Error Classes)
// =============================================================================

/**
 * site-flow API 응답 오류
 *
 * HTTP 4xx/5xx 응답을 받은 경우 발생합니다.
 * 호출자는 status를 확인하여 4xx(클라이언트 오류)와
 * 5xx(서버 오류)를 구분할 수 있습니다.
 */
export class SiteFlowApiError extends Error {
  constructor(
    message: string,
    /** HTTP 상태 코드 (예: 404, 500) */
    public readonly status: number,
    /** HTTP 상태 텍스트 (예: "Not Found", "Internal Server Error") */
    public readonly statusText: string,
    /** 서버가 반환한 오류 응답 본문 (파싱 성공 시) */
    public readonly body?: ApiError,
  ) {
    super(message);
    this.name = 'SiteFlowApiError';
  }
}

/**
 * site-flow API 연결 오류
 *
 * 네트워크 오류로 fetch 자체가 실패한 경우 발생합니다.
 * 서버가 응답하지 않거나, DNS 해석 실패, 연결 거부 등이 해당됩니다.
 * 호출자는 이 오류를 잡아 graceful degradation 여부를 결정합니다.
 */
export class SiteFlowConnectionError extends Error {
  constructor(
    message: string,
    /** 연결 오류의 원인 (원본 fetch 오류) */
    public readonly cause?: Error,
  ) {
    super(message);
    this.name = 'SiteFlowConnectionError';
  }
}

// =============================================================================
// 내부 유틸리티 (Internal Utilities)
// =============================================================================

/**
 * 지정된 시간(밀리초)만큼 대기합니다.
 *
 * @param ms - 대기 시간 (밀리초)
 */
function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * 경로 앞에 `/api` 프리픽스가 없으면 추가하여 완전한 API 경로를 반환합니다.
 *
 * @param path - API 경로 (예: '/sites', '/api/sites')
 * @returns 정규화된 API 경로 (예: '/api/sites')
 */
function normalizeApiPath(path: string): string {
  if (path.startsWith('/api/') || path === '/api') {
    return path;
  }
  // 선행 슬래시 정규화 후 /api 프리픽스 추가
  const cleanPath = path.startsWith('/') ? path : `/${path}`;
  return `/api${cleanPath}`;
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
// SiteFlowClient 클래스
// =============================================================================

/**
 * site-flow API HTTP 클라이언트
 *
 * 설정(SiteFlowConfig)으로 초기화되며, 모든 API 엔드포인트 모듈이
 * 이 클라이언트를 통해 HTTP 요청을 수행합니다.
 *
 * @example
 * const client = new SiteFlowClient({
 *   apiUrl: 'https://api.site-flow.io',
 *   apiKey: 'sf_xxxx',
 *   siteId: '66a1b2c3d4e5f6789',
 *   enabled: true,
 * });
 *
 * const site = await client.get<SiteResponse>('/sites/66a1b2c3d4e5f6789');
 */
export class SiteFlowClient {
  private readonly _config: SiteFlowConfig;
  private readonly _normalizedBaseUrl: string;
  private readonly _timeout: number;
  private readonly _retryCount: number;
  private readonly _retryDelay: number;

  constructor(config: SiteFlowConfig, options?: ClientOptions) {
    this._config = config;
    this._normalizedBaseUrl = normalizeBaseUrl(config.apiUrl);
    this._timeout = options?.timeout ?? DEFAULT_TIMEOUT_MS;
    this._retryCount = options?.retryCount ?? DEFAULT_RETRY_COUNT;
    this._retryDelay = options?.retryDelay ?? DEFAULT_RETRY_DELAY_MS;
  }

  // ---------------------------------------------------------------------------
  // Getters
  // ---------------------------------------------------------------------------

  /** 설정된 사이트 ID */
  get siteId(): string {
    return this._config.siteId;
  }

  /** 정규화된 API 기본 URL (후행 슬래시 없음) */
  get baseUrl(): string {
    return this._normalizedBaseUrl;
  }

  /** 클라이언트 활성화 여부 */
  get isEnabled(): boolean {
    return this._config.enabled;
  }

  // ---------------------------------------------------------------------------
  // 공개 HTTP 메서드 (Public HTTP Methods)
  // ---------------------------------------------------------------------------

  /**
   * GET 요청을 수행합니다.
   *
   * @param path - API 경로 (예: '/sites/abc123')
   * @param params - URL 쿼리 파라미터
   * @returns 응답 데이터 (제네릭 타입 T)
   * @throws {SiteFlowApiError} 4xx/5xx 응답 시
   * @throws {SiteFlowConnectionError} 네트워크 오류 시
   *
   * @example
   * const site = await client.get<SiteResponse>('/sites/abc123');
   * const pages = await client.get<PageResponse[]>('/pages', { siteId: 'abc123' });
   */
  async get<T>(path: string, params?: Record<string, string>): Promise<T> {
    let fullPath = normalizeApiPath(path);

    // 쿼리 파라미터가 있으면 URL에 추가
    if (params && Object.keys(params).length > 0) {
      const searchParams = new URLSearchParams(params);
      fullPath = `${fullPath}?${searchParams.toString()}`;
    }

    return this._request<T>(fullPath, { method: 'GET' });
  }

  /**
   * POST 요청을 수행합니다.
   *
   * @param path - API 경로
   * @param body - 요청 본문 (JSON 직렬화됨)
   * @returns 응답 데이터 (제네릭 타입 T)
   * @throws {SiteFlowApiError} 4xx/5xx 응답 시
   * @throws {SiteFlowConnectionError} 네트워크 오류 시
   *
   * @example
   * const site = await client.post<SiteResponse>('/sites', { name: 'My Site' });
   */
  async post<T>(path: string, body?: unknown): Promise<T> {
    return this._request<T>(normalizeApiPath(path), {
      method: 'POST',
      ...(body !== undefined ? { body: JSON.stringify(body) } : {}),
    });
  }

  /**
   * PUT 요청을 수행합니다. (전체 교체)
   *
   * @param path - API 경로
   * @param body - 요청 본문 (JSON 직렬화됨)
   * @returns 응답 데이터 (제네릭 타입 T)
   * @throws {SiteFlowApiError} 4xx/5xx 응답 시
   * @throws {SiteFlowConnectionError} 네트워크 오류 시
   */
  async put<T>(path: string, body?: unknown): Promise<T> {
    return this._request<T>(normalizeApiPath(path), {
      method: 'PUT',
      ...(body !== undefined ? { body: JSON.stringify(body) } : {}),
    });
  }

  /**
   * PATCH 요청을 수행합니다. (부분 업데이트)
   *
   * @param path - API 경로
   * @param body - 요청 본문 (JSON 직렬화됨)
   * @returns 응답 데이터 (제네릭 타입 T)
   * @throws {SiteFlowApiError} 4xx/5xx 응답 시
   * @throws {SiteFlowConnectionError} 네트워크 오류 시
   *
   * @example
   * const page = await client.patch<PageResponse>('/pages/abc123', {
   *   inspectionStatus: 'completed',
   * });
   */
  async patch<T>(path: string, body?: unknown): Promise<T> {
    return this._request<T>(normalizeApiPath(path), {
      method: 'PATCH',
      ...(body !== undefined ? { body: JSON.stringify(body) } : {}),
    });
  }

  /**
   * DELETE 요청을 수행합니다.
   *
   * @param path - API 경로
   * @returns 응답 데이터 (제네릭 타입 T, 주로 빈 객체 또는 메시지)
   * @throws {SiteFlowApiError} 4xx/5xx 응답 시
   * @throws {SiteFlowConnectionError} 네트워크 오류 시
   *
   * @example
   * await client.delete('/sites/abc123');
   */
  async delete<T>(path: string): Promise<T> {
    return this._request<T>(normalizeApiPath(path), { method: 'DELETE' });
  }

  /**
   * 가공되지 않은 fetch 응답을 반환합니다.
   *
   * SSE 스트리밍, 파일 다운로드 등 JSON이 아닌 응답을 처리할 때 사용합니다.
   * 이 메서드는 재시도 로직을 적용하지 않습니다.
   *
   * @param path - API 경로
   * @param init - fetch RequestInit 옵션 (메서드, 헤더 등)
   * @returns 원시 Response 객체
   * @throws {SiteFlowConnectionError} 네트워크 오류 시
   *
   * @example
   * // SSE 스트리밍 처리
   * const response = await client.fetchRaw('/pages/auto-capture-batch', {
   *   method: 'POST',
   *   body: JSON.stringify({ siteId: client.siteId, pageIds: [...] }),
   * });
   * await parseSSEStream(response, onEvent);
   */
  async fetchRaw(path: string, init?: RequestInit): Promise<Response> {
    const url = `${this._normalizedBaseUrl}${normalizeApiPath(path)}`;
    const headers = this._buildHeaders(init?.headers);

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this._timeout);

    try {
      const response = await fetch(url, {
        ...init,
        headers,
        signal: controller.signal,
      });
      return response;
    } catch (err) {
      const cause = err instanceof Error ? err : new Error(String(err));

      if (cause.name === 'AbortError') {
        throw new SiteFlowConnectionError(
          `[site-flow] 요청 타임아웃 (${this._timeout}ms 초과): ${url}`,
          cause,
        );
      }

      throw new SiteFlowConnectionError(
        `[site-flow] 연결 실패: ${url} - ${cause.message}`,
        cause,
      );
    } finally {
      clearTimeout(timeoutId);
    }
  }

  // ---------------------------------------------------------------------------
  // 내부 메서드 (Private Methods)
  // ---------------------------------------------------------------------------

  /**
   * 공통 HTTP 요청 수행 메서드 (재시도 로직 포함)
   *
   * - 네트워크 오류 및 5xx 응답에 대해 지수 백오프 재시도 적용
   * - 4xx 클라이언트 오류는 즉시 SiteFlowApiError를 던짐 (재시도 없음)
   * - 타임아웃은 AbortController로 관리
   *
   * @param path - 정규화된 API 경로 (/api/... 형식)
   * @param init - fetch RequestInit (method, body 등)
   * @returns 파싱된 JSON 응답 (제네릭 타입 T)
   */
  private async _request<T>(path: string, init: RequestInit): Promise<T> {
    const url = `${this._normalizedBaseUrl}${path}`;
    const headers = this._buildHeaders(init.headers, init.body !== undefined);

    let lastError: SiteFlowApiError | SiteFlowConnectionError | undefined;

    for (let attempt = 0; attempt <= this._retryCount; attempt++) {
      // 재시도 전 대기 (첫 번째 시도는 대기 없음)
      if (attempt > 0) {
        const delayMs = this._retryDelay * Math.pow(2, attempt - 1);
        console.warn(
          `[site-flow] 재시도 ${attempt}/${this._retryCount}: ${init.method} ${url} ` +
          `(${delayMs}ms 후 재시도)`,
        );
        await sleep(delayMs);
      }

      // AbortController로 타임아웃 설정
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), this._timeout);

      try {
        const response = await fetch(url, {
          ...init,
          headers,
          signal: controller.signal,
        });

        clearTimeout(timeoutId);

        // 성공 응답 처리 (2xx)
        if (response.ok) {
          // 204 No Content: 빈 객체 반환
          if (response.status === 204) {
            return {} as T;
          }
          return (await response.json()) as T;
        }

        // 오류 응답 처리
        const errorBody = await this._parseErrorBody(response);
        const apiError = new SiteFlowApiError(
          `[site-flow] API 오류 ${response.status} ${response.statusText}: ${url}` +
          (errorBody?.error ? ` - ${errorBody.error}` : ''),
          response.status,
          response.statusText,
          errorBody,
        );

        // 4xx 클라이언트 오류: 재시도 없이 즉시 던짐
        if (response.status < RETRYABLE_STATUS_MIN) {
          throw apiError;
        }

        // 5xx 서버 오류: 재시도 대상으로 기록
        lastError = apiError;

        // 마지막 재시도인 경우 오류 로그 출력
        if (attempt === this._retryCount) {
          console.error(
            `[site-flow] 최종 실패 (${this._retryCount}회 재시도 소진): ` +
            `${init.method} ${url} - ${response.status} ${response.statusText}`,
          );
        }

      } catch (err) {
        clearTimeout(timeoutId);

        // SiteFlowApiError (4xx)는 재시도 없이 즉시 재던짐
        if (err instanceof SiteFlowApiError) {
          throw err;
        }

        // 네트워크/타임아웃 오류: 재시도 대상으로 기록
        const cause = err instanceof Error ? err : new Error(String(err));
        const isTimeout = cause.name === 'AbortError';

        const connectionError = new SiteFlowConnectionError(
          isTimeout
            ? `[site-flow] 요청 타임아웃 (${this._timeout}ms 초과): ${init.method} ${url}`
            : `[site-flow] 연결 오류: ${init.method} ${url} - ${cause.message}`,
          cause,
        );

        lastError = connectionError;

        // 마지막 재시도인 경우 오류 로그 출력
        if (attempt === this._retryCount) {
          console.error(
            `[site-flow] 최종 실패 (${this._retryCount}회 재시도 소진): ` +
            connectionError.message,
          );
        }
      }
    }

    // 모든 재시도가 소진된 경우 마지막 오류를 던짐
    throw lastError ?? new SiteFlowConnectionError(
      `[site-flow] 알 수 없는 오류: ${init.method} ${url}`,
    );
  }

  /**
   * 공통 요청 헤더를 생성합니다.
   *
   * - Authorization: Bearer {apiKey} 항상 주입
   * - Content-Type: application/json (본문이 있는 요청에만 추가)
   * - 기존 헤더와 병합 (기존 헤더가 우선)
   *
   * @param existingHeaders - 병합할 기존 헤더 (fetch RequestInit.headers)
   * @param hasBody - 요청 본문 존재 여부 (Content-Type 추가 여부 결정)
   * @returns 완성된 Headers 객체
   */
  private _buildHeaders(
    existingHeaders?: HeadersInit,
    hasBody: boolean = false,
  ): Headers {
    const headers = new Headers(existingHeaders);

    // Authorization 헤더 주입 (기존 값이 있어도 덮어씀)
    headers.set('Authorization', `Bearer ${this._config.apiKey}`);

    // 본문이 있는 경우 Content-Type 설정 (이미 설정된 경우는 유지)
    if (hasBody && !headers.has('Content-Type')) {
      headers.set('Content-Type', 'application/json');
    }

    return headers;
  }

  /**
   * 오류 응답 본문을 파싱합니다.
   *
   * JSON 파싱에 실패하면 null을 반환합니다 (오류 미전파).
   *
   * @param response - 오류 응답 객체
   * @returns 파싱된 ApiError 또는 undefined
   */
  private async _parseErrorBody(response: Response): Promise<ApiError | undefined> {
    try {
      const contentType = response.headers.get('Content-Type') ?? '';
      if (!contentType.includes('application/json')) {
        return undefined;
      }
      return (await response.json()) as ApiError;
    } catch {
      return undefined;
    }
  }
}

// =============================================================================
// 팩토리 함수 (Factory Function)
// =============================================================================

/**
 * SiteFlowClient 인스턴스를 생성하는 팩토리 함수입니다.
 *
 * `config.enabled`가 false인 경우 null을 반환합니다.
 * (`--skip-site-flow` 플래그가 설정된 경우에 해당)
 *
 * 호출자는 반환값이 null인지 확인한 후 API 호출을 수행해야 합니다.
 *
 * @param config - site-flow 클라이언트 설정
 * @param options - HTTP 클라이언트 옵션 (타임아웃, 재시도 등)
 * @returns SiteFlowClient 인스턴스 또는 null (비활성화 시)
 *
 * @example
 * const client = createSiteFlowClient(config);
 * if (client === null) {
 *   console.log('[site-flow] 비활성화됨, 스킵합니다.');
 *   return;
 * }
 * const site = await client.get<SiteResponse>(`/sites/${client.siteId}`);
 *
 * @example
 * // 이미지 업로드용 타임아웃 연장
 * const imageClient = createSiteFlowClient(config, { timeout: 120_000 });
 */
export function createSiteFlowClient(
  config: SiteFlowConfig,
  options?: ClientOptions,
): SiteFlowClient | null {
  if (!config.enabled) {
    return null;
  }

  return new SiteFlowClient(config, options);
}
