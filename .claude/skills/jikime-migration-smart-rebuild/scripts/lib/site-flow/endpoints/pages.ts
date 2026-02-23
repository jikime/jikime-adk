/**
 * 페이지(Page) 엔드포인트 모듈
 *
 * site-flow API의 페이지 관련 엔드포인트를 타입 안전하게 래핑합니다.
 * 페이지 CRUD, 이미지 관리, 자동 캡처(SSE 스트리밍) 기능을 제공합니다.
 */

import type { SiteFlowClient } from '../client';
import type {
  PaginationParams,
  PageResponse,
  PageListResponse,
  CreatePageRequest,
  UpdatePageRequest,
  BulkUpsertPagesRequest,
  AddImageRequest,
  PageImageResponse,
  PageImageDetailResponse,
} from '../types';

// =============================================================================
// 페이지 CRUD 함수
// =============================================================================

/**
 * 특정 사이트의 페이지 목록을 조회합니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param siteId - 조회할 사이트 ID
 * @param params - 페이지네이션 파라미터 (page, limit)
 * @returns 페이지 목록 (기능 수 포함)
 *
 * @example
 * const pages = await getPages(client, client.siteId, { page: 1, limit: 20 });
 */
export async function getPages(
  client: SiteFlowClient,
  siteId: string,
  params?: PaginationParams,
): Promise<PageListResponse[]> {
  // siteId는 항상 포함, 나머지 파라미터는 정의된 경우에만 포함
  const queryParams: Record<string, string> = { siteId };

  if (params?.page !== undefined) {
    queryParams['page'] = String(params.page);
  }
  if (params?.limit !== undefined) {
    queryParams['limit'] = String(params.limit);
  }

  return client.get<PageListResponse[]>('/pages', queryParams);
}

/**
 * 페이지 단건을 조회합니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param pageId - 조회할 페이지 ID (MongoDB ObjectId string)
 * @returns 페이지 상세 정보
 *
 * @example
 * const page = await getPage(client, '66a1b2c3d4e5f6789');
 */
export async function getPage(
  client: SiteFlowClient,
  pageId: string,
): Promise<PageResponse> {
  return client.get<PageResponse>(`/pages/${pageId}`);
}

/**
 * 새 페이지를 생성합니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param data - 페이지 생성 요청 데이터 (siteId, path, name 필수)
 * @returns 생성된 페이지 정보
 *
 * @example
 * const page = await createPage(client, {
 *   siteId: client.siteId,
 *   path: '/about',
 *   name: '소개 페이지',
 * });
 */
export async function createPage(
  client: SiteFlowClient,
  data: CreatePageRequest,
): Promise<PageResponse> {
  return client.post<PageResponse>('/pages', data);
}

/**
 * 페이지를 부분 업데이트합니다 (PATCH).
 *
 * @param client - SiteFlowClient 인스턴스
 * @param pageId - 수정할 페이지 ID
 * @param data - 변경할 필드만 포함한 업데이트 데이터
 * @returns 업데이트된 페이지 정보
 *
 * @example
 * const updated = await updatePage(client, pageId, {
 *   inspectionStatus: 'completed',
 * });
 */
export async function updatePage(
  client: SiteFlowClient,
  pageId: string,
  data: UpdatePageRequest,
): Promise<PageResponse> {
  return client.patch<PageResponse>(`/pages/${pageId}`, data);
}

/**
 * 페이지를 삭제합니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param pageId - 삭제할 페이지 ID
 *
 * @example
 * await deletePage(client, '66a1b2c3d4e5f6789');
 */
export async function deletePage(
  client: SiteFlowClient,
  pageId: string,
): Promise<void> {
  await client.delete<void>(`/pages/${pageId}`);
}

/**
 * 페이지를 일괄 upsert합니다 (PUT).
 *
 * {siteId, path} 조합을 기준으로 존재하면 업데이트, 없으면 생성합니다.
 * 사이트맵 일괄 동기화에 유용합니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param data - upsert 요청 데이터 (siteId와 pages 배열 필수)
 * @returns upsert된 페이지 목록
 *
 * @example
 * const pages = await bulkUpsertPages(client, {
 *   siteId: client.siteId,
 *   pages: [
 *     { path: '/', name: '홈' },
 *     { path: '/about', name: '소개' },
 *   ],
 * });
 */
export async function bulkUpsertPages(
  client: SiteFlowClient,
  data: BulkUpsertPagesRequest,
): Promise<PageResponse[]> {
  return client.put<PageResponse[]>('/pages', data);
}

// =============================================================================
// 페이지 이미지 관리 함수
// =============================================================================

/**
 * 페이지에 이미지를 추가합니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param pageId - 이미지를 추가할 페이지 ID
 * @param data - 이미지 추가 요청 데이터 (siteId, image base64 필수)
 * @returns 추가된 이미지 정보
 *
 * @example
 * const image = await addPageImage(client, pageId, {
 *   siteId: client.siteId,
 *   image: base64EncodedImageData,
 *   source: 'auto-capture',
 *   setAsThumbnail: true,
 * });
 */
export async function addPageImage(
  client: SiteFlowClient,
  pageId: string,
  data: AddImageRequest,
): Promise<PageImageResponse> {
  return client.post<PageImageResponse>(`/pages/${pageId}/images`, data);
}

/**
 * 페이지의 이미지 목록을 조회합니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param pageId - 조회할 페이지 ID
 * @param params - 필터 파라미터 (source로 이미지 출처 필터링 가능)
 * @returns 이미지 목록 (base64 원본 이미지 제외, 썸네일만 포함)
 *
 * @example
 * // 자동 캡처 이미지만 조회
 * const images = await getPageImages(client, pageId, { source: 'auto-capture' });
 */
export async function getPageImages(
  client: SiteFlowClient,
  pageId: string,
  params?: { source?: string },
): Promise<PageImageResponse[]> {
  const queryParams: Record<string, string> = {};

  if (params?.source !== undefined) {
    queryParams['source'] = params.source;
  }

  return client.get<PageImageResponse[]>(
    `/pages/${pageId}/images`,
    Object.keys(queryParams).length > 0 ? queryParams : undefined,
  );
}

/**
 * 페이지 이미지 단건을 상세 조회합니다 (원본 base64 이미지 포함).
 *
 * @param client - SiteFlowClient 인스턴스
 * @param pageId - 페이지 ID
 * @param imageId - 조회할 이미지 ID
 * @returns 이미지 상세 정보 (원본 base64 이미지 포함)
 *
 * @example
 * const detail = await getPageImage(client, pageId, imageId);
 * const base64Data = detail.image; // 원본 이미지 데이터
 */
export async function getPageImage(
  client: SiteFlowClient,
  pageId: string,
  imageId: string,
): Promise<PageImageDetailResponse> {
  return client.get<PageImageDetailResponse>(`/pages/${pageId}/images/${imageId}`);
}

/**
 * 특정 이미지를 페이지 썸네일로 지정합니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param pageId - 페이지 ID
 * @param imageId - 썸네일로 지정할 이미지 ID
 * @returns 업데이트된 이미지 정보 (isThumbnail: true)
 *
 * @example
 * const updated = await setPageThumbnail(client, pageId, imageId);
 * console.log(updated.isThumbnail); // true
 */
export async function setPageThumbnail(
  client: SiteFlowClient,
  pageId: string,
  imageId: string,
): Promise<PageImageResponse> {
  return client.patch<PageImageResponse>(`/pages/${pageId}/images/${imageId}`, {
    isThumbnail: true,
  });
}

// =============================================================================
// 자동 캡처 (SSE 스트리밍)
// =============================================================================

/**
 * 여러 페이지를 일괄 자동 캡처합니다 (SSE 스트리밍).
 *
 * 이 함수는 Server-Sent Events(SSE) 스트리밍 응답을 반환합니다.
 * 재시도 로직이 없는 fetchRaw를 사용하므로 호출자가 SSE 스트림을
 * 직접 파싱해야 합니다 (parseSSEStream 유틸리티 참조).
 *
 * @param client - SiteFlowClient 인스턴스
 * @param siteId - 캡처 대상 사이트 ID
 * @param pageIds - 캡처할 페이지 ID 목록
 * @returns SSE 스트리밍 원시 Response 객체
 * @throws {SiteFlowConnectionError} 네트워크 오류 시
 *
 * @example
 * const response = await autoCapturePages(client, client.siteId, pageIds);
 * await parseSSEStream(response, (event) => {
 *   if (event.type === 'progress') {
 *     console.log(`진행: ${event.current}/${event.total}`);
 *   }
 * });
 */
export async function autoCapturePages(
  client: SiteFlowClient,
  siteId: string,
  pageIds: string[],
): Promise<Response> {
  return client.fetchRaw('/pages/auto-capture-batch', {
    method: 'POST',
    body: JSON.stringify({ siteId, pageIds }),
    headers: { 'Content-Type': 'application/json' },
  });
}
