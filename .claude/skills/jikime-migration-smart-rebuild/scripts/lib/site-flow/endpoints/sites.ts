/**
 * site-flow 사이트(Site) 엔드포인트 모듈
 *
 * GET /api/sites, POST /api/sites, PATCH /api/sites/:id,
 * DELETE /api/sites/:id 엔드포인트를 래핑하는 타입 안전 함수 모음입니다.
 *
 * 주의: createSite는 NextAuth 세션 인증이 필요합니다.
 * 나머지 함수는 Bearer API 키 인증을 사용합니다.
 */

import type { SiteFlowClient } from '../client';
import type { SiteResponse, CreateSiteRequest, UpdateSiteRequest } from '../types';

// =============================================================================
// 사이트 조회 함수 (Read)
// =============================================================================

/**
 * 현재 클라이언트에 설정된 siteId로 필터링된 사이트 목록을 조회합니다.
 *
 * GET /api/sites?siteId={siteId}
 *
 * @param client - SiteFlowClient 인스턴스
 * @returns 사이트 목록
 */
export async function getSites(client: SiteFlowClient): Promise<SiteResponse[]> {
  return client.get<SiteResponse[]>('/sites', { siteId: client.siteId });
}

/**
 * 지정된 ID의 사이트 단건을 조회합니다.
 *
 * GET /api/sites/{siteId}
 *
 * @param client - SiteFlowClient 인스턴스
 * @param siteId - 조회할 사이트의 MongoDB ObjectId (string)
 * @returns 사이트 정보
 */
export async function getSite(
  client: SiteFlowClient,
  siteId: string,
): Promise<SiteResponse> {
  return client.get<SiteResponse>(`/sites/${siteId}`);
}

/**
 * URL로 사이트를 검색하여 첫 번째 일치 항목을 반환합니다.
 *
 * GET /api/sites?url={url}
 *
 * 일치하는 사이트가 없거나 404 응답 시 null을 반환합니다.
 * Phase 0 discover 단계에서 기존 사이트 존재 여부를 확인할 때 사용합니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param url - 검색할 사이트 URL
 * @returns 첫 번째 일치 사이트 또는 null (없는 경우)
 */
export async function findSiteByUrl(
  client: SiteFlowClient,
  url: string,
): Promise<SiteResponse | null> {
  try {
    const sites = await client.get<SiteResponse[]>('/sites', { url });
    return sites.length > 0 ? sites[0] : null;
  } catch {
    // 404 포함 모든 오류는 미발견으로 처리
    return null;
  }
}

// =============================================================================
// 사이트 생성/수정/삭제 함수 (Write)
// =============================================================================

/**
 * 새 사이트를 생성합니다.
 *
 * POST /api/sites
 *
 * 주의: 이 엔드포인트는 API 키 인증이 아닌 NextAuth 세션 인증을 사용합니다.
 * smart-rebuild Phase 0 discover 통합 시 반드시 세션 쿠키가 포함된
 * 브라우저 컨텍스트 또는 별도 인증 흐름을 사용해야 합니다.
 * API 키로 호출할 경우 401 Unauthorized 응답이 반환됩니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param data - 사이트 생성 요청 데이터 (name 필수)
 * @returns 생성된 사이트 정보
 */
export async function createSite(
  client: SiteFlowClient,
  data: CreateSiteRequest,
): Promise<SiteResponse> {
  return client.post<SiteResponse>('/sites', data);
}

/**
 * 기존 사이트의 정보를 부분 업데이트합니다.
 *
 * PATCH /api/sites/{siteId}
 *
 * @param client - SiteFlowClient 인스턴스
 * @param siteId - 수정할 사이트의 MongoDB ObjectId (string)
 * @param data - 수정할 필드 (부분 업데이트)
 * @returns 업데이트된 사이트 정보
 */
export async function updateSite(
  client: SiteFlowClient,
  siteId: string,
  data: UpdateSiteRequest,
): Promise<SiteResponse> {
  return client.patch<SiteResponse>(`/sites/${siteId}`, data);
}

/**
 * 지정된 사이트를 삭제합니다.
 *
 * DELETE /api/sites/{siteId}
 *
 * 삭제된 사이트에 연결된 모든 페이지, 기능, 테스트 케이스 등도
 * 함께 삭제될 수 있습니다. 신중하게 사용하세요.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param siteId - 삭제할 사이트의 MongoDB ObjectId (string)
 */
export async function deleteSite(
  client: SiteFlowClient,
  siteId: string,
): Promise<void> {
  await client.delete<void>(`/sites/${siteId}`);
}
