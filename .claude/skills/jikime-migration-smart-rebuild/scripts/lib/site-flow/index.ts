/**
 * site-flow API 클라이언트 라이브러리
 *
 * jikime-adk 마이그레이션 워크플로우에서 site-flow 서버와 통신하기 위한
 * 통합 진입점입니다. 모든 API 타입, 클라이언트, 설정, 유틸리티 및
 * 엔드포인트 함수를 단일 모듈에서 제공합니다.
 *
 * @example
 * ```typescript
 * import {
 *   loadSiteFlowConfig,
 *   createSiteFlowClient,
 *   getSites,
 *   createPage,
 * } from '../lib/site-flow';
 *
 * const config = loadSiteFlowConfig();
 * const client = createSiteFlowClient(config);
 * if (client) {
 *   const sites = await getSites(client);
 * }
 * ```
 */

// ============================================================
// 타입 정의
// ============================================================
export type {
  // 공통
  ApiError,
  PaginationParams,
  SiteFlowConfig,
  ClientOptions,

  // 사이트
  CreateSiteRequest,
  UpdateSiteRequest,
  SiteResponse,

  // 페이지
  ISourcePath,
  IApiDependency,
  PageInspectionStatus,
  CreatePageRequest,
  UpdatePageRequest,
  BulkUpsertPagesRequest,
  PageResponse,
  PageListResponse,

  // 페이지 이미지
  PageImageSource,
  AddImageRequest,
  PageImageResponse,
  PageImageDetailResponse,

  // 기능 요구사항
  FeatureStatus,
  FeaturePriority,
  Testability,
  CreateFeatureRequest,
  UpdateFeatureRequest,
  FeatureResponse,

  // 테스트 케이스
  TestCaseStatus,
  TestCaseLastResult,
  CreateTestCaseRequest,
  UpdateTestCaseRequest,
  BulkExecuteRequest,
  TestCaseResponse,

  // 버그 리포트
  BugReportKind,
  BugSeverity,
  BugReportStatus,
  BugCategory,
  CreateBugReportRequest,
  UpdateBugReportRequest,
  BugReportResponse,

  // API 키
  CreateApiKeyRequest,
  ApiKeyResponse,
  ApiKeyCreateResponse,

  // 테스트 실행
  TestExecutionTrigger,
  TestExecutionStatus,
  TestExecutionScopeType,
  TestResultStatus,
  TestExecutionScopeFilter,
  TestExecutionScope,
  TestExecutionSummary,
  TestExecutionResult,
  PlaywrightReport,
  TestExecutionResponse,

  // 내보내기
  ExportResponse,

  // SSE
  SSEEventType,
  SSECaptureEvent,
} from './types';

// ============================================================
// 클라이언트
// ============================================================
export {
  SiteFlowClient,
  SiteFlowApiError,
  SiteFlowConnectionError,
  createSiteFlowClient,
} from './client';

// ============================================================
// 설정
// ============================================================
export {
  loadSiteFlowConfig,
  saveSiteFlowConfig,
  isSiteFlowEnabled,
} from './config';

export type { LoadSiteFlowConfigOptions } from './config';

// ============================================================
// SSE 파서
// ============================================================
export {
  parseSSEStream,
  createSSEProgressHandler,
} from './sse-parser';

export type { SSEProgressHandlerOptions } from './sse-parser';

// ============================================================
// 이미지 유틸리티
// ============================================================
export {
  IMAGE_SIZE_LIMIT_MB,
  getMimeType,
  stripDataUriPrefix,
  createDataUri,
  getBase64Size,
  isImageWithinSizeLimit,
  fileToBase64DataUri,
  chunkBase64Image,
  isValidBase64DataUri,
  isSupportedImageFormat,
  getImageSizeInfo,
} from './image-utils';

export type { ImageChunk, ImageSizeInfo } from './image-utils';

// ============================================================
// 엔드포인트: 사이트
// ============================================================
export {
  getSites,
  getSite,
  findSiteByUrl,
  createSite,
  updateSite,
  deleteSite,
} from './endpoints/sites';

// ============================================================
// 엔드포인트: API 키
// ============================================================
export {
  createApiKey,
  getApiKeys,
  deleteApiKey,
} from './endpoints/api-keys';

// ============================================================
// 엔드포인트: 페이지
// ============================================================
export {
  getPages,
  getPage,
  createPage,
  updatePage,
  deletePage,
  bulkUpsertPages,
  addPageImage,
  getPageImages,
  getPageImage,
  setPageThumbnail,
  autoCapturePages,
} from './endpoints/pages';

// ============================================================
// 엔드포인트: 기능 요구사항
// ============================================================
export {
  getFeatures,
  getFeature,
  createFeature,
  updateFeature,
  deleteFeature,
} from './endpoints/features';

// ============================================================
// 엔드포인트: 테스트 케이스
// ============================================================
export {
  getTestCases,
  getTestCase,
  createTestCase,
  updateTestCase,
  deleteTestCase,
  bulkExecuteTestCases,
} from './endpoints/test-cases';

// ============================================================
// 엔드포인트: 버그 리포트
// ============================================================
export {
  getBugReports,
  getBugReport,
  createBugReport,
  updateBugReport,
  deleteBugReport,
} from './endpoints/bug-reports';

// ============================================================
// 엔드포인트: 데이터 내보내기
// ============================================================
export { exportSiteData } from './endpoints/export';
