/**
 * 기능(Feature) 엔드포인트 모듈
 *
 * site-flow API의 기능 관련 엔드포인트를 타입 안전하게 래핑합니다.
 * 기능 CRUD 및 사이트/페이지 기준 목록 조회 기능을 제공합니다.
 */

import type { SiteFlowClient } from '../client';
import type {
  FeatureResponse,
  CreateFeatureRequest,
  UpdateFeatureRequest,
} from '../types';

// =============================================================================
// 기능 CRUD 함수
// =============================================================================

/**
 * 기능 목록을 조회합니다.
 *
 * siteId는 필수이며, pageId를 함께 전달하면 해당 페이지의 기능만 필터링합니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param params - 조회 파라미터 (siteId 필수, pageId 선택)
 * @returns 기능 목록
 *
 * @example
 * // 사이트 전체 기능 조회
 * const allFeatures = await getFeatures(client, { siteId: client.siteId });
 *
 * // 특정 페이지 기능만 조회
 * const pageFeatures = await getFeatures(client, {
 *   siteId: client.siteId,
 *   pageId: '66a1b2c3d4e5f6789',
 * });
 */
export async function getFeatures(
  client: SiteFlowClient,
  params: { siteId: string; pageId?: string },
): Promise<FeatureResponse[]> {
  const queryParams: Record<string, string> = { siteId: params.siteId };

  if (params.pageId !== undefined) {
    queryParams['pageId'] = params.pageId;
  }

  return client.get<FeatureResponse[]>('/features', queryParams);
}

/**
 * 기능 단건을 조회합니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param featureId - 조회할 기능 ID (MongoDB ObjectId string)
 * @returns 기능 상세 정보
 *
 * @example
 * const feature = await getFeature(client, '66a1b2c3d4e5f6789');
 */
export async function getFeature(
  client: SiteFlowClient,
  featureId: string,
): Promise<FeatureResponse> {
  return client.get<FeatureResponse>(`/features/${featureId}`);
}

/**
 * 새 기능을 생성합니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param data - 기능 생성 요청 데이터 (siteId, pageId, name 필수)
 * @returns 생성된 기능 정보
 *
 * @example
 * const feature = await createFeature(client, {
 *   siteId: client.siteId,
 *   pageId: '66a1b2c3d4e5f6789',
 *   name: '사용자 로그인',
 *   description: '이메일과 비밀번호로 로그인하는 기능',
 *   priority: 'high',
 *   testability: 'easy',
 * });
 */
export async function createFeature(
  client: SiteFlowClient,
  data: CreateFeatureRequest,
): Promise<FeatureResponse> {
  return client.post<FeatureResponse>('/features', data);
}

/**
 * 기능을 부분 업데이트합니다 (PATCH).
 *
 * @param client - SiteFlowClient 인스턴스
 * @param featureId - 수정할 기능 ID
 * @param data - 변경할 필드만 포함한 업데이트 데이터
 * @returns 업데이트된 기능 정보
 *
 * @example
 * const updated = await updateFeature(client, featureId, {
 *   status: 'completed',
 *   priority: 'medium',
 * });
 */
export async function updateFeature(
  client: SiteFlowClient,
  featureId: string,
  data: UpdateFeatureRequest,
): Promise<FeatureResponse> {
  return client.patch<FeatureResponse>(`/features/${featureId}`, data);
}

/**
 * 기능을 삭제합니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param featureId - 삭제할 기능 ID
 *
 * @example
 * await deleteFeature(client, '66a1b2c3d4e5f6789');
 */
export async function deleteFeature(
  client: SiteFlowClient,
  featureId: string,
): Promise<void> {
  await client.delete<void>(`/features/${featureId}`);
}
