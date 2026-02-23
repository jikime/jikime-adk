/**
 * site-flow 데이터 내보내기(Export) 엔드포인트 모듈
 *
 * GET /api/export 엔드포인트를 래핑하는 타입 안전 함수 모음입니다.
 *
 * 사이트 전체 데이터(페이지, 기능, 테스트 케이스, 버그 리포트,
 * 테스트 실행 이력 등)를 단일 JSON 응답으로 내보낼 때 사용합니다.
 */

import type { SiteFlowClient } from '../client';
import type { ExportResponse } from '../types';

// =============================================================================
// 데이터 내보내기 함수 (Export)
// =============================================================================

/**
 * 지정된 사이트의 전체 데이터를 내보냅니다.
 *
 * GET /api/export?siteId={siteId}
 *
 * 사이트에 속한 모든 페이지, 페이지 이미지(base64 제외), 기능,
 * 테스트 케이스, 버그 리포트, 테스트 실행 이력을 포함한
 * 전체 스냅샷을 반환합니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param siteId - 내보낼 사이트의 MongoDB ObjectId (string)
 * @returns 사이트 전체 데이터 스냅샷
 */
export async function exportSiteData(
  client: SiteFlowClient,
  siteId: string,
): Promise<ExportResponse> {
  return client.get<ExportResponse>('/export', { siteId });
}
