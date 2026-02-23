/**
 * site-flow API 클라이언트 타입 정의
 *
 * Next.js + MongoDB 기반 site-flow REST API와의 통신에 사용되는
 * 모든 TypeScript 타입을 정의합니다.
 *
 * 설계 원칙:
 * - MongoDB ObjectId는 string으로 직렬화
 * - Date 필드는 ISO 8601 string으로 직렬화
 * - 요청 타입과 응답 타입은 별도로 정의
 * - 외부 의존성 없음
 */

// =============================================================================
// 공통 타입 (Common Types)
// =============================================================================

/**
 * API 오류 응답 구조
 */
export interface ApiError {
  /** 오류 메시지 */
  error: string;
  /** HTTP 상태 코드 */
  status?: number;
}

/**
 * 페이지네이션 파라미터
 */
export interface PaginationParams {
  /** 페이지 번호 (1부터 시작) */
  page?: number;
  /** 페이지당 항목 수 */
  limit?: number;
}

// =============================================================================
// 클라이언트 설정 타입 (Client Configuration Types)
// =============================================================================

/**
 * site-flow 클라이언트 설정
 */
export interface SiteFlowConfig {
  /** site-flow API 서버 URL (예: https://api.site-flow.io) */
  apiUrl: string;
  /** 인증에 사용할 API 키 */
  apiKey: string;
  /** 대상 사이트 ID (MongoDB ObjectId string) */
  siteId: string;
  /** 클라이언트 활성화 여부 */
  enabled: boolean;
}

/**
 * HTTP 클라이언트 옵션
 */
export interface ClientOptions {
  /** 요청 타임아웃 (밀리초, 기본값: 30000) */
  timeout?: number;
  /** 재시도 횟수 (기본값: 3) */
  retryCount?: number;
  /** 재시도 간격 (밀리초, 기본값: 1000) */
  retryDelay?: number;
}

// =============================================================================
// Site 타입 (사이트)
// =============================================================================

/**
 * 사이트 생성 요청
 */
export interface CreateSiteRequest {
  /** 사이트 이름 (필수, userId 범위 내 고유) */
  name: string;
  /** 사이트 URL */
  url?: string;
  /** 사이트 설명 */
  description?: string;
}

/**
 * 사이트 수정 요청
 */
export interface UpdateSiteRequest {
  /** 사이트 이름 */
  name?: string;
  /** 사이트 URL */
  url?: string;
  /** 사이트 설명 */
  description?: string;
}

/**
 * 사이트 응답 (MongoDB IUser 모델 직렬화)
 */
export interface SiteResponse {
  /** MongoDB ObjectId (string 직렬화) */
  _id: string;
  /** 소유 사용자 ID */
  userId: string;
  /** 사이트 이름 */
  name: string;
  /** 사이트 URL */
  url?: string;
  /** 사이트 설명 */
  description?: string;
  /** 생성 시각 (ISO 8601) */
  createdAt: string;
  /** 수정 시각 (ISO 8601) */
  updatedAt: string;
}

// =============================================================================
// Page 타입 (페이지)
// =============================================================================

/**
 * 소스 코드 경로 참조
 */
export interface ISourcePath {
  /** 소스 파일 경로 */
  filePath: string;
  /** 시작 라인 번호 */
  lineStart?: number;
  /** 종료 라인 번호 */
  lineEnd?: number;
  /** 해당 경로에 대한 설명 */
  description?: string;
}

/**
 * API 의존성 정보
 */
export interface IApiDependency {
  /** API 엔드포인트 경로 (예: /api/users) */
  endpoint: string;
  /** HTTP 메서드 */
  method: string;
  /** 해당 API 의존성에 대한 설명 */
  description?: string;
}

/**
 * 페이지 검수 상태
 */
export type PageInspectionStatus = 'not-started' | 'in-progress' | 'completed';

/**
 * 페이지 생성 요청
 */
export interface CreatePageRequest {
  /** 소속 사이트 ID */
  siteId: string;
  /** 페이지 경로 (예: /about, /products/list) */
  path: string;
  /** 페이지 이름 */
  name: string;
  /** 페이지 설명 */
  description?: string;
  /** 소스 코드 경로 참조 목록 */
  sourcePaths?: ISourcePath[];
  /** API 의존성 목록 */
  apiDependencies?: IApiDependency[];
  /** 개발 가이드 */
  developmentGuide?: string;
  /** 기술 스택 */
  techStack?: string;
  /** 검수 상태 */
  inspectionStatus?: PageInspectionStatus;
}

/**
 * 페이지 수정 요청 (PATCH - 부분 업데이트)
 */
export interface UpdatePageRequest {
  /** 페이지 경로 */
  path?: string;
  /** 페이지 이름 */
  name?: string;
  /** 페이지 설명 */
  description?: string;
  /** 소스 코드 경로 참조 목록 */
  sourcePaths?: ISourcePath[];
  /** API 의존성 목록 */
  apiDependencies?: IApiDependency[];
  /** 개발 가이드 */
  developmentGuide?: string;
  /** 기술 스택 */
  techStack?: string;
  /** 검수 상태 */
  inspectionStatus?: PageInspectionStatus;
  /** 검수 완료 시각 (ISO 8601) */
  inspectedAt?: string;
}

/**
 * 일괄 페이지 upsert 요청
 */
export interface BulkUpsertPagesRequest {
  /** 대상 사이트 ID */
  siteId: string;
  /** upsert할 페이지 목록 */
  pages: Array<{
    /** 페이지 경로 (upsert 기준 키) */
    path: string;
    /** 페이지 이름 */
    name: string;
    /** 페이지 설명 */
    description?: string;
    /** 소스 코드 경로 참조 목록 */
    sourcePaths?: ISourcePath[];
    /** API 의존성 목록 */
    apiDependencies?: IApiDependency[];
    /** 개발 가이드 */
    developmentGuide?: string;
    /** 기술 스택 */
    techStack?: string;
  }>;
}

/**
 * 페이지 응답 (MongoDB IPage 모델 직렬화)
 */
export interface PageResponse {
  /** MongoDB ObjectId (string 직렬화) */
  _id: string;
  /** 소속 사이트 ID */
  siteId: string;
  /** 페이지 경로 */
  path: string;
  /** 페이지 이름 */
  name: string;
  /** 썸네일 이미지 (base64) */
  image?: string;
  /** 페이지 설명 */
  description?: string;
  /** 소스 코드 경로 참조 목록 */
  sourcePaths?: ISourcePath[];
  /** API 의존성 목록 */
  apiDependencies?: IApiDependency[];
  /** 개발 가이드 */
  developmentGuide?: string;
  /** 기술 스택 */
  techStack?: string;
  /** 검수 상태 */
  inspectionStatus: PageInspectionStatus;
  /** 검수 완료 시각 (ISO 8601) */
  inspectedAt?: string;
  /** 생성 시각 (ISO 8601) */
  createdAt: string;
  /** 수정 시각 (ISO 8601) */
  updatedAt: string;
}

/**
 * 페이지 목록 응답 (기능 수 포함)
 */
export interface PageListResponse extends PageResponse {
  /** 연결된 기능 수 */
  featureCount?: number;
}

// =============================================================================
// PageImage 타입 (페이지 이미지)
// =============================================================================

/**
 * 페이지 이미지 출처
 */
export type PageImageSource =
  | 'manual'
  | 'test-execution'
  | 'clipboard'
  | 'auto-capture'
  | 'migration';

/**
 * 이미지 추가 요청
 */
export interface AddImageRequest {
  /** 소속 사이트 ID */
  siteId: string;
  /** 이미지 데이터 (base64 인코딩) */
  image: string;
  /** 이미지 제목 */
  title?: string;
  /** 이미지 출처 */
  source?: PageImageSource;
  /** 페이지 썸네일로 설정 여부 */
  setAsThumbnail?: boolean;
  /** 연결된 테스트 실행 ID */
  testExecutionId?: string;
}

/**
 * 페이지 이미지 응답
 * 주의: 목록 응답에서는 base64 image 필드가 제외됩니다
 */
export interface PageImageResponse {
  /** MongoDB ObjectId (string 직렬화) */
  _id: string;
  /** 소속 사이트 ID */
  siteId: string;
  /** 소속 페이지 ID */
  pageId: string;
  /** 썸네일 이미지 (리사이즈된 WebP, ~400px) */
  thumbnail?: string;
  /** 이미지 제목 */
  title?: string;
  /** 이미지 출처 */
  source: PageImageSource;
  /** 연결된 테스트 실행 ID */
  testExecutionId?: string;
  /** 표시 순서 */
  order: number;
  /** 썸네일 여부 */
  isThumbnail: boolean;
  /** 생성 시각 (ISO 8601) */
  createdAt: string;
  /** 수정 시각 (ISO 8601) */
  updatedAt: string;
}

/**
 * 전체 이미지 데이터 포함 응답 (단건 조회 시)
 */
export interface PageImageDetailResponse extends PageImageResponse {
  /** 원본 이미지 데이터 (base64) */
  image: string;
}

// =============================================================================
// Feature 타입 (기능)
// =============================================================================

/**
 * 기능 개발 상태
 */
export type FeatureStatus = 'planned' | 'in-progress' | 'completed' | 'deprecated';

/**
 * 기능 우선순위
 */
export type FeaturePriority = 'low' | 'medium' | 'high';

/**
 * 테스트 용이성
 */
export type Testability = 'easy' | 'medium' | 'hard' | 'not-testable';

/**
 * 기능 생성 요청
 */
export interface CreateFeatureRequest {
  /** 소속 사이트 ID */
  siteId: string;
  /** 소속 페이지 ID */
  pageId: string;
  /** 기능 이름 (필수) */
  name: string;
  /** 기능 설명 */
  description?: string;
  /** 기능 카테고리 */
  category?: string;
  /** 개발 상태 */
  status?: FeatureStatus;
  /** 우선순위 */
  priority?: FeaturePriority;
  /** 테스트 용이성 */
  testability?: Testability;
  /** 테스트 용이성 판단 이유 */
  testabilityReason?: string;
  /** 구현 가이드 */
  implementationGuide?: string;
}

/**
 * 기능 수정 요청 (PATCH - 부분 업데이트)
 */
export interface UpdateFeatureRequest {
  /** 기능 이름 */
  name?: string;
  /** 기능 설명 */
  description?: string;
  /** 기능 카테고리 */
  category?: string;
  /** 개발 상태 */
  status?: FeatureStatus;
  /** 우선순위 */
  priority?: FeaturePriority;
  /** 테스트 용이성 */
  testability?: Testability;
  /** 테스트 용이성 판단 이유 */
  testabilityReason?: string;
  /** 구현 가이드 */
  implementationGuide?: string;
}

/**
 * 기능 응답 (MongoDB IFeature 모델 직렬화)
 */
export interface FeatureResponse {
  /** MongoDB ObjectId (string 직렬화) */
  _id: string;
  /** 소속 사이트 ID */
  siteId: string;
  /** 소속 페이지 ID */
  pageId: string;
  /** 기능 이름 */
  name: string;
  /** 기능 설명 */
  description?: string;
  /** 기능 카테고리 */
  category?: string;
  /** 개발 상태 */
  status?: FeatureStatus;
  /** 우선순위 */
  priority?: FeaturePriority;
  /** 테스트 용이성 */
  testability?: Testability;
  /** 테스트 용이성 판단 이유 */
  testabilityReason?: string;
  /** 구현 가이드 */
  implementationGuide?: string;
  /** 생성 시각 (ISO 8601) */
  createdAt: string;
  /** 수정 시각 (ISO 8601) */
  updatedAt: string;
}

// =============================================================================
// TestCase 타입 (테스트 케이스)
// =============================================================================

/**
 * 테스트 케이스 상태
 */
export type TestCaseStatus = 'draft' | 'ready' | 'passed' | 'failed' | 'skipped';

/**
 * 테스트 실행 결과
 */
export interface TestCaseLastResult {
  /** 통과 여부 */
  passed: boolean;
  /** 실행 시간 (밀리초) */
  duration?: number;
  /** 오류 메시지 */
  errorMessage?: string;
  /** 스크린샷 (base64 또는 URL) */
  screenshot?: string;
}

/**
 * 테스트 케이스 생성 요청
 */
export interface CreateTestCaseRequest {
  /** 소속 사이트 ID */
  siteId: string;
  /** 소속 기능 ID */
  featureId: string;
  /** 소속 페이지 ID */
  pageId: string;
  /** 테스트 케이스 이름 */
  name: string;
  /** 테스트 케이스 설명 */
  description?: string;
  /** Playwright 테스트 코드 (원본 사이트용) */
  testCode: string;
  /** 테스트 spec 파일명 */
  specFileName: string;
  /** 테스트 카테고리 */
  category: string;
  /** 테스트 용이성 */
  testability: Testability;
}

/**
 * 테스트 케이스 수정 요청 (PATCH - 부분 업데이트)
 */
export interface UpdateTestCaseRequest {
  /** 테스트 케이스 이름 */
  name?: string;
  /** 테스트 케이스 설명 */
  description?: string;
  /** Playwright 테스트 코드 (원본 사이트용) */
  testCode?: string;
  /** 테스트 spec 파일명 */
  specFileName?: string;
  /** Playwright 테스트 코드 (마이그레이션된 사이트용) */
  testCodeNext?: string;
  /** 마이그레이션 사이트용 spec 파일명 */
  specFileNameNext?: string;
  /** 검수 완료 여부 */
  inspected?: boolean;
  /** 테스트 카테고리 */
  category?: string;
  /** 테스트 용이성 */
  testability?: Testability;
  /** 테스트 상태 */
  status?: TestCaseStatus;
}

/**
 * 일괄 테스트 실행 요청
 */
export interface BulkExecuteRequest {
  /** 대상 사이트 ID */
  siteId: string;
  /** 실행할 테스트 케이스 ID 목록 */
  testCaseIds: string[];
}

/**
 * 테스트 케이스 응답 (MongoDB ITestCase 모델 직렬화)
 */
export interface TestCaseResponse {
  /** MongoDB ObjectId (string 직렬화) */
  _id: string;
  /** 소속 사이트 ID */
  siteId: string;
  /** 소속 기능 ID */
  featureId: string;
  /** 소속 페이지 ID */
  pageId: string;
  /** 테스트 케이스 이름 */
  name: string;
  /** 테스트 케이스 설명 */
  description?: string;
  /** Playwright 테스트 코드 (원본 사이트용) */
  testCode: string;
  /** 테스트 spec 파일명 */
  specFileName: string;
  /** Playwright 테스트 코드 (마이그레이션된 사이트용) */
  testCodeNext?: string;
  /** 마이그레이션 사이트용 spec 파일명 */
  specFileNameNext?: string;
  /** 검수 완료 여부 */
  inspected: boolean;
  /** 검수 완료 시각 (ISO 8601) */
  inspectedAt?: string;
  /** 테스트 카테고리 */
  category: string;
  /** 테스트 용이성 */
  testability: Testability;
  /** 테스트 상태 */
  status: TestCaseStatus;
  /** 마지막 실행 시각 (ISO 8601) */
  lastRunAt?: string;
  /** 마지막 실행 결과 */
  lastResult?: TestCaseLastResult;
  /** 전체 실행 횟수 */
  runCount: number;
  /** 통과 횟수 */
  passCount: number;
  /** 실패 횟수 */
  failCount: number;
  /** 생성 시각 (ISO 8601) */
  createdAt: string;
  /** 수정 시각 (ISO 8601) */
  updatedAt: string;
}

// =============================================================================
// BugReport 타입 (버그 리포트)
// =============================================================================

/**
 * 버그 리포트 종류
 */
export type BugReportKind = 'bug' | 'feature' | 'improvement';

/**
 * 버그 심각도
 */
export type BugSeverity = 'critical' | 'major' | 'minor' | 'trivial';

/**
 * 버그 리포트 상태
 */
export type BugReportStatus = 'open' | 'in-progress' | 'resolved' | 'closed' | 'rejected';

/**
 * 버그 카테고리
 */
export type BugCategory = 'ui' | 'functional' | 'performance' | 'data' | 'security' | 'other';

/**
 * 버그 리포트 생성 요청
 */
export interface CreateBugReportRequest {
  /** 소속 사이트 ID */
  siteId: string;
  /** 소속 페이지 ID */
  pageId: string;
  /** 리포트 제목 */
  title: string;
  /** 상세 설명 */
  description: string;
  /** 심각도 */
  severity: BugSeverity;
  /** 카테고리 */
  category: BugCategory;
  /** 리포트 종류 */
  kind?: BugReportKind;
  /** 연결된 기능 ID */
  featureId?: string;
  /** 연결된 테스트 케이스 ID */
  testCaseId?: string;
  /** 재현 단계 */
  stepsToReproduce?: string;
  /** 예상 동작 */
  expectedBehavior?: string;
  /** 실제 동작 */
  actualBehavior?: string;
  /** 리포터 이름 또는 ID */
  reporter?: string;
  /** 스크린샷 목록 (base64 또는 URL) */
  screenshots?: string[];
  /** 발생 환경 */
  environment?: string;
}

/**
 * 버그 리포트 수정 요청 (PATCH - 부분 업데이트)
 */
export interface UpdateBugReportRequest {
  /** 리포트 제목 */
  title?: string;
  /** 상세 설명 */
  description?: string;
  /** 심각도 */
  severity?: BugSeverity;
  /** 카테고리 */
  category?: BugCategory;
  /** 리포트 종류 */
  kind?: BugReportKind;
  /** 리포트 상태 */
  status?: BugReportStatus;
  /** 담당자 */
  assignee?: string;
  /** 재현 단계 */
  stepsToReproduce?: string;
  /** 예상 동작 */
  expectedBehavior?: string;
  /** 실제 동작 */
  actualBehavior?: string;
  /** 스크린샷 목록 */
  screenshots?: string[];
  /** 발생 환경 */
  environment?: string;
}

/**
 * 버그 리포트 응답 (MongoDB IBugReport 모델 직렬화)
 */
export interface BugReportResponse {
  /** MongoDB ObjectId (string 직렬화) */
  _id: string;
  /** 소속 사이트 ID */
  siteId: string;
  /** 소속 페이지 ID */
  pageId: string;
  /** 리포트 종류 */
  kind?: BugReportKind;
  /** 연결된 기능 ID */
  featureId?: string;
  /** 연결된 테스트 케이스 ID */
  testCaseId?: string;
  /** 리포트 제목 */
  title: string;
  /** 상세 설명 */
  description: string;
  /** 재현 단계 */
  stepsToReproduce?: string;
  /** 예상 동작 */
  expectedBehavior?: string;
  /** 실제 동작 */
  actualBehavior?: string;
  /** 심각도 */
  severity: BugSeverity;
  /** 리포트 상태 */
  status: BugReportStatus;
  /** 카테고리 */
  category: BugCategory;
  /** 리포터 이름 또는 ID */
  reporter: string;
  /** 담당자 */
  assignee?: string;
  /** 스크린샷 목록 */
  screenshots?: string[];
  /** 발생 환경 */
  environment?: string;
  /** 해결 시각 (ISO 8601) */
  resolvedAt?: string;
  /** 종료 시각 (ISO 8601) */
  closedAt?: string;
  /** 댓글 수 */
  commentCount: number;
  /** 생성 시각 (ISO 8601) */
  createdAt: string;
  /** 수정 시각 (ISO 8601) */
  updatedAt: string;
}

// =============================================================================
// ApiKey 타입 (API 키)
// =============================================================================

/**
 * API 키 생성 요청
 */
export interface CreateApiKeyRequest {
  /** 소속 사이트 ID */
  siteId: string;
  /** API 키 이름 (식별용) */
  name?: string;
}

/**
 * API 키 응답
 * 주의: keyHash는 절대 클라이언트에 반환되지 않습니다
 */
export interface ApiKeyResponse {
  /** MongoDB ObjectId (string 직렬화) */
  _id: string;
  /** 소속 사이트 ID */
  siteId: string;
  /** API 키 이름 */
  name: string;
  /** 키 프리픽스 (앞 11자, 예: "sf_a1b2c3d4e") */
  keyPrefix: string;
  /** 마지막 사용 시각 (ISO 8601) */
  lastUsedAt?: string;
  /** 만료 시각 (ISO 8601) */
  expiresAt?: string;
  /** 생성 시각 (ISO 8601) */
  createdAt: string;
}

/**
 * API 키 생성 응답 (전체 키 포함, 생성 시에만 반환)
 */
export interface ApiKeyCreateResponse extends ApiKeyResponse {
  /** 전체 API 키 (생성 직후 1회만 반환, 이후 조회 불가) */
  key: string;
}

// =============================================================================
// TestExecution 타입 (테스트 실행)
// =============================================================================

/**
 * 테스트 실행 트리거 유형
 */
export type TestExecutionTrigger = 'manual' | 'scheduled' | 'ci';

/**
 * 테스트 실행 상태
 */
export type TestExecutionStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

/**
 * 테스트 실행 범위 유형
 */
export type TestExecutionScopeType = 'all' | 'category' | 'testability' | 'page' | 'specific';

/**
 * 개별 테스트 결과 상태
 */
export type TestResultStatus = 'passed' | 'failed' | 'skipped';

/**
 * 테스트 실행 범위 필터
 */
export interface TestExecutionScopeFilter {
  /** 실행할 카테고리 목록 */
  categories?: string[];
  /** 실행할 테스트 용이성 목록 */
  testabilities?: Testability[];
  /** 실행할 페이지 ID 목록 */
  pageIds?: string[];
  /** 실행할 테스트 케이스 ID 목록 */
  testCaseIds?: string[];
}

/**
 * 테스트 실행 범위
 */
export interface TestExecutionScope {
  /** 범위 유형 */
  type: TestExecutionScopeType;
  /** 필터 조건 (type이 'all'이 아닌 경우) */
  filter?: TestExecutionScopeFilter;
}

/**
 * 테스트 실행 요약
 */
export interface TestExecutionSummary {
  /** 전체 테스트 수 */
  total: number;
  /** 통과 수 */
  passed: number;
  /** 실패 수 */
  failed: number;
  /** 건너뜀 수 */
  skipped: number;
  /** 전체 실행 시간 (밀리초) */
  duration: number;
}

/**
 * 개별 테스트 케이스 실행 결과
 */
export interface TestExecutionResult {
  /** 테스트 케이스 ID */
  testCaseId: string;
  /** 결과 상태 */
  status: TestResultStatus;
  /** 실행 시간 (밀리초) */
  duration?: number;
  /** 오류 메시지 */
  errorMessage?: string;
  /** 스크린샷 (base64 또는 URL) */
  screenshot?: string;
}

/**
 * Playwright 리포트 경로
 */
export interface PlaywrightReport {
  /** HTML 리포트 파일 경로 */
  htmlReportPath?: string;
  /** JSON 리포트 파일 경로 */
  jsonReportPath?: string;
}

/**
 * 테스트 실행 응답 (MongoDB ITestExecution 모델 직렬화)
 */
export interface TestExecutionResponse {
  /** MongoDB ObjectId (string 직렬화) */
  _id: string;
  /** 소속 사이트 ID */
  siteId: string;
  /** 실행 이름 */
  name: string;
  /** 실행 트리거 유형 */
  triggeredBy: TestExecutionTrigger;
  /** 실행 범위 */
  scope: TestExecutionScope;
  /** 실행 상태 */
  status: TestExecutionStatus;
  /** 실행 요약 */
  summary: TestExecutionSummary;
  /** 개별 테스트 결과 목록 */
  results: TestExecutionResult[];
  /** Playwright 리포트 경로 */
  playwrightReport?: PlaywrightReport;
  /** 실행 시작 시각 (ISO 8601) */
  startedAt?: string;
  /** 실행 완료 시각 (ISO 8601) */
  completedAt?: string;
  /** 오류 메시지 (실행 실패 시) */
  errorMessage?: string;
  /** 생성 시각 (ISO 8601) */
  createdAt: string;
  /** 수정 시각 (ISO 8601) */
  updatedAt: string;
}

// =============================================================================
// Export 타입 (내보내기)
// =============================================================================

/**
 * 전체 데이터 내보내기 응답
 * site-flow에서 사이트 전체 데이터를 JSON으로 내보낼 때 사용
 */
export interface ExportResponse {
  /** 내보낸 사이트 정보 */
  site: SiteResponse;
  /** 사이트 내 모든 페이지 */
  pages: PageResponse[];
  /** 사이트 내 모든 페이지 이미지 (base64 제외) */
  pageImages: PageImageResponse[];
  /** 사이트 내 모든 기능 */
  features: FeatureResponse[];
  /** 사이트 내 모든 테스트 케이스 */
  testCases: TestCaseResponse[];
  /** 사이트 내 모든 버그 리포트 */
  bugReports: BugReportResponse[];
  /** 사이트 내 모든 테스트 실행 이력 */
  testExecutions: TestExecutionResponse[];
  /** 내보내기 생성 시각 (ISO 8601) */
  exportedAt: string;
  /** site-flow API 버전 */
  version: string;
}

// =============================================================================
// SSE 이벤트 타입 (Server-Sent Events)
// =============================================================================

/**
 * SSE 이벤트 유형
 */
export type SSEEventType = 'progress' | 'complete' | 'error';

/**
 * 캡처 진행 SSE 이벤트
 * 자동 캡처(auto-capture) 작업의 실시간 진행 상황 스트리밍에 사용
 */
export interface SSECaptureEvent {
  /** 이벤트 유형 */
  type: SSEEventType;
  /** 처리 중인 페이지 ID */
  pageId?: string;
  /** 현재 처리 수 */
  current?: number;
  /** 전체 처리 대상 수 */
  total?: number;
  /** 상태 메시지 */
  message?: string;
}
