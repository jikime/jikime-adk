/**
 * site-flow 테스트 케이스(TestCase) 엔드포인트 모듈
 *
 * GET /api/test-cases, POST /api/test-cases, PATCH /api/test-cases/:id,
 * DELETE /api/test-cases/:id, POST /api/test-cases/bulk-execute
 * 엔드포인트를 래핑하는 타입 안전 함수 모음입니다.
 */

import type { SiteFlowClient } from '../client';
import type {
  TestCaseResponse,
  CreateTestCaseRequest,
  UpdateTestCaseRequest,
  BulkExecuteRequest,
  TestExecutionResponse,
} from '../types';

// =============================================================================
// 테스트 케이스 조회 함수 (Read)
// =============================================================================

/**
 * 조건에 맞는 테스트 케이스 목록을 조회합니다.
 *
 * GET /api/test-cases?siteId=X&...
 *
 * pageId, featureId, status는 undefined인 경우 쿼리 파라미터에서 제외됩니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param params - 필터 파라미터 (siteId 필수, 나머지는 선택)
 * @returns 테스트 케이스 목록
 */
export async function getTestCases(
  client: SiteFlowClient,
  params: {
    siteId: string;
    pageId?: string;
    featureId?: string;
    status?: string;
  },
): Promise<TestCaseResponse[]> {
  const queryParams: Record<string, string> = { siteId: params.siteId };

  if (params.pageId !== undefined) {
    queryParams.pageId = params.pageId;
  }
  if (params.featureId !== undefined) {
    queryParams.featureId = params.featureId;
  }
  if (params.status !== undefined) {
    queryParams.status = params.status;
  }

  return client.get<TestCaseResponse[]>('/test-cases', queryParams);
}

/**
 * 지정된 ID의 테스트 케이스 단건을 조회합니다.
 *
 * GET /api/test-cases/{testCaseId}
 *
 * @param client - SiteFlowClient 인스턴스
 * @param testCaseId - 조회할 테스트 케이스의 MongoDB ObjectId (string)
 * @returns 테스트 케이스 정보
 */
export async function getTestCase(
  client: SiteFlowClient,
  testCaseId: string,
): Promise<TestCaseResponse> {
  return client.get<TestCaseResponse>(`/test-cases/${testCaseId}`);
}

// =============================================================================
// 테스트 케이스 생성/수정/삭제 함수 (Write)
// =============================================================================

/**
 * 새 테스트 케이스를 생성합니다.
 *
 * POST /api/test-cases
 *
 * @param client - SiteFlowClient 인스턴스
 * @param data - 테스트 케이스 생성 요청 데이터
 * @returns 생성된 테스트 케이스 정보
 */
export async function createTestCase(
  client: SiteFlowClient,
  data: CreateTestCaseRequest,
): Promise<TestCaseResponse> {
  return client.post<TestCaseResponse>('/test-cases', data);
}

/**
 * 기존 테스트 케이스를 부분 업데이트합니다.
 *
 * PATCH /api/test-cases/{testCaseId}
 *
 * @param client - SiteFlowClient 인스턴스
 * @param testCaseId - 수정할 테스트 케이스의 MongoDB ObjectId (string)
 * @param data - 수정할 필드 (부분 업데이트)
 * @returns 업데이트된 테스트 케이스 정보
 */
export async function updateTestCase(
  client: SiteFlowClient,
  testCaseId: string,
  data: UpdateTestCaseRequest,
): Promise<TestCaseResponse> {
  return client.patch<TestCaseResponse>(`/test-cases/${testCaseId}`, data);
}

/**
 * 지정된 테스트 케이스를 삭제합니다.
 *
 * DELETE /api/test-cases/{testCaseId}
 *
 * @param client - SiteFlowClient 인스턴스
 * @param testCaseId - 삭제할 테스트 케이스의 MongoDB ObjectId (string)
 */
export async function deleteTestCase(
  client: SiteFlowClient,
  testCaseId: string,
): Promise<void> {
  await client.delete<void>(`/test-cases/${testCaseId}`);
}

// =============================================================================
// 테스트 케이스 일괄 실행 함수 (Bulk Execute)
// =============================================================================

/**
 * 여러 테스트 케이스를 일괄 실행합니다.
 *
 * POST /api/test-cases/bulk-execute
 *
 * 지정된 테스트 케이스 ID 목록에 해당하는 Playwright 테스트를
 * 서버 측에서 일괄 실행하고 테스트 실행 결과를 반환합니다.
 *
 * @param client - SiteFlowClient 인스턴스
 * @param data - 일괄 실행 요청 데이터 (siteId, testCaseIds 포함)
 * @returns 테스트 실행 결과 (TestExecutionResponse)
 */
export async function bulkExecuteTestCases(
  client: SiteFlowClient,
  data: BulkExecuteRequest,
): Promise<TestExecutionResponse> {
  return client.post<TestExecutionResponse>('/test-cases/bulk-execute', data);
}
