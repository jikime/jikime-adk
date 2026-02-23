/**
 * site-flow API 키(ApiKey) 엔드포인트 모듈
 *
 * POST /api/api-keys, GET /api/api-keys, DELETE /api/api-keys/:id
 * 엔드포인트를 래핑하는 타입 안전 함수 모음입니다.
 *
 * 주의: 생성 응답(ApiKeyCreateResponse)의 key 필드는 생성 직후 1회만
 * 반환됩니다. 이후 조회에서는 keyPrefix만 확인 가능합니다.
 */

import type { SiteFlowClient } from '../client';
import type { ApiKeyResponse, ApiKeyCreateResponse, CreateApiKeyRequest } from '../types';

// =============================================================================
// API 키 생성 함수 (Create)
// =============================================================================

/**
 * 지정된 사이트에 새 API 키를 발급합니다.
 *
 * POST /api/api-keys
 *
 * 응답의 key 필드(전체 API 키 문자열)는 생성 직후 이 응답에서만
 * 반환됩니다. 반드시 안전한 장소에 저장하세요.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param data - API 키 생성 요청 데이터 (siteId 필수, name 선택)
 * @returns 생성된 API 키 정보 (전체 key 포함)
 */
export async function createApiKey(
  client: SiteFlowClient,
  data: CreateApiKeyRequest,
): Promise<ApiKeyCreateResponse> {
  return client.post<ApiKeyCreateResponse>('/api-keys', data);
}

// =============================================================================
// API 키 조회 함수 (Read)
// =============================================================================

/**
 * 지정된 사이트의 API 키 목록을 조회합니다.
 *
 * GET /api/api-keys?siteId={siteId}
 *
 * 반환되는 목록에는 keyPrefix만 포함되며, 전체 키 문자열은
 * 생성 시점 이후 다시 조회할 수 없습니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param siteId - 조회할 사이트의 MongoDB ObjectId (string)
 * @returns API 키 목록 (keyHash 제외, keyPrefix 포함)
 */
export async function getApiKeys(
  client: SiteFlowClient,
  siteId: string,
): Promise<ApiKeyResponse[]> {
  return client.get<ApiKeyResponse[]>('/api-keys', { siteId });
}

// =============================================================================
// API 키 삭제 함수 (Delete)
// =============================================================================

/**
 * 지정된 API 키를 폐기합니다.
 *
 * DELETE /api/api-keys/{keyId}
 *
 * 삭제된 키는 즉시 인증에 사용할 수 없게 됩니다.
 * 이 작업은 되돌릴 수 없습니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param keyId - 삭제할 API 키의 MongoDB ObjectId (string)
 */
export async function deleteApiKey(
  client: SiteFlowClient,
  keyId: string,
): Promise<void> {
  await client.delete<void>(`/api-keys/${keyId}`);
}
