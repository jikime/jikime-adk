/**
 * site-flow 버그 리포트(BugReport) 엔드포인트 모듈
 *
 * GET /api/bug-reports, POST /api/bug-reports, PATCH /api/bug-reports/:id,
 * DELETE /api/bug-reports/:id 엔드포인트를 래핑하는 타입 안전 함수 모음입니다.
 */

import type { SiteFlowClient } from '../client';
import type {
  BugReportResponse,
  CreateBugReportRequest,
  UpdateBugReportRequest,
} from '../types';

// =============================================================================
// 버그 리포트 조회 함수 (Read)
// =============================================================================

/**
 * 조건에 맞는 버그 리포트 목록을 조회합니다.
 *
 * GET /api/bug-reports?siteId=X&...
 *
 * pageId, status, severity는 undefined인 경우 쿼리 파라미터에서 제외됩니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param params - 필터 파라미터 (siteId 필수, 나머지는 선택)
 * @returns 버그 리포트 목록
 */
export async function getBugReports(
  client: SiteFlowClient,
  params: {
    siteId: string;
    pageId?: string;
    status?: string;
    severity?: string;
  },
): Promise<BugReportResponse[]> {
  const queryParams: Record<string, string> = { siteId: params.siteId };

  if (params.pageId !== undefined) {
    queryParams.pageId = params.pageId;
  }
  if (params.status !== undefined) {
    queryParams.status = params.status;
  }
  if (params.severity !== undefined) {
    queryParams.severity = params.severity;
  }

  return client.get<BugReportResponse[]>('/bug-reports', queryParams);
}

/**
 * 지정된 ID의 버그 리포트 단건을 조회합니다.
 *
 * GET /api/bug-reports/{bugReportId}
 *
 * @param client - SiteFlowClient 인스턴스
 * @param bugReportId - 조회할 버그 리포트의 MongoDB ObjectId (string)
 * @returns 버그 리포트 정보
 */
export async function getBugReport(
  client: SiteFlowClient,
  bugReportId: string,
): Promise<BugReportResponse> {
  return client.get<BugReportResponse>(`/bug-reports/${bugReportId}`);
}

// =============================================================================
// 버그 리포트 생성/수정/삭제 함수 (Write)
// =============================================================================

/**
 * 새 버그 리포트를 생성합니다.
 *
 * POST /api/bug-reports
 *
 * @param client - SiteFlowClient 인스턴스
 * @param data - 버그 리포트 생성 요청 데이터
 * @returns 생성된 버그 리포트 정보
 */
export async function createBugReport(
  client: SiteFlowClient,
  data: CreateBugReportRequest,
): Promise<BugReportResponse> {
  return client.post<BugReportResponse>('/bug-reports', data);
}

/**
 * 기존 버그 리포트를 부분 업데이트합니다.
 *
 * PATCH /api/bug-reports/{bugReportId}
 *
 * @param client - SiteFlowClient 인스턴스
 * @param bugReportId - 수정할 버그 리포트의 MongoDB ObjectId (string)
 * @param data - 수정할 필드 (부분 업데이트)
 * @returns 업데이트된 버그 리포트 정보
 */
export async function updateBugReport(
  client: SiteFlowClient,
  bugReportId: string,
  data: UpdateBugReportRequest,
): Promise<BugReportResponse> {
  return client.patch<BugReportResponse>(`/bug-reports/${bugReportId}`, data);
}

/**
 * 지정된 버그 리포트를 삭제합니다.
 *
 * DELETE /api/bug-reports/{bugReportId}
 *
 * @param client - SiteFlowClient 인스턴스
 * @param bugReportId - 삭제할 버그 리포트의 MongoDB ObjectId (string)
 */
export async function deleteBugReport(
  client: SiteFlowClient,
  bugReportId: string,
): Promise<void> {
  await client.delete<void>(`/bug-reports/${bugReportId}`);
}
