/**
 * site-flow 이미지 유틸리티 모듈
 *
 * site-flow API 클라이언트에서 사용하는 이미지 처리 유틸리티.
 * base64 인코딩, 크기 검증, 데이터 URI 변환 기능을 제공합니다.
 *
 * 설계 원칙:
 * - 외부 의존성 없음 (Node.js 표준 라이브러리만 사용)
 * - TypeScript strict 모드 준수
 * - site-flow API 요청 크기 제한 (REQ-MIG-041: 10MB) 준수
 *
 * API 이미지 형식:
 * - site-flow API는 base64 데이터 URI를 사용합니다
 * - 형식: `data:{mimeType};base64,{base64data}`
 * - 예시: `data:image/png;base64,iVBORw0KGgo...`
 */

import * as fs from 'fs/promises';
import * as path from 'path';

// =============================================================================
// 상수 (Constants)
// =============================================================================

/**
 * site-flow API 요청 최대 크기 제한 (MB)
 * REQ-MIG-041: API 요청 본문은 10MB를 초과할 수 없습니다
 */
export const IMAGE_SIZE_LIMIT_MB = 10;

/**
 * 지원하는 이미지 MIME 타입 매핑
 * 확장자(소문자) → MIME 타입
 */
const MIME_TYPE_MAP: Readonly<Record<string, string>> = {
  png: 'image/png',
  jpg: 'image/jpeg',
  jpeg: 'image/jpeg',
  webp: 'image/webp',
  gif: 'image/gif',
  svg: 'image/svg+xml',
};

/**
 * 데이터 URI 프리픽스 패턴 (파싱에 사용)
 * 예: `data:image/png;base64,`
 */
const DATA_URI_PREFIX_PATTERN = /^data:[^;]+;base64,/;

// =============================================================================
// MIME 타입 유틸리티
// =============================================================================

/**
 * 파일 경로에서 MIME 타입을 반환합니다.
 *
 * 파일 확장자를 기반으로 MIME 타입을 결정합니다.
 * 지원하지 않는 확장자의 경우 'application/octet-stream'을 반환합니다.
 *
 * @param filePath - MIME 타입을 확인할 파일 경로
 * @returns MIME 타입 문자열
 *
 * @example
 * getMimeType('/path/to/screenshot.png')  // 'image/png'
 * getMimeType('/path/to/photo.jpg')       // 'image/jpeg'
 * getMimeType('/path/to/unknown.xyz')     // 'application/octet-stream'
 */
export function getMimeType(filePath: string): string {
  // 확장자 추출 (앞의 점 제거 후 소문자 변환)
  const ext = path.extname(filePath).toLowerCase().replace(/^\./, '');

  if (!ext) {
    return 'application/octet-stream';
  }

  return MIME_TYPE_MAP[ext] ?? 'application/octet-stream';
}

// =============================================================================
// 데이터 URI 유틸리티
// =============================================================================

/**
 * 데이터 URI에서 base64 프리픽스(`data:...;base64,`)를 제거하여
 * 순수한 base64 문자열을 반환합니다.
 *
 * 이미 프리픽스가 없는 순수 base64 문자열이 전달되면 그대로 반환합니다.
 *
 * @param dataUri - 데이터 URI 또는 순수 base64 문자열
 * @returns 프리픽스가 제거된 순수 base64 문자열
 *
 * @example
 * stripDataUriPrefix('data:image/png;base64,iVBORw0KGgo...')
 * // 'iVBORw0KGgo...'
 *
 * stripDataUriPrefix('iVBORw0KGgo...')
 * // 'iVBORw0KGgo...' (변경 없음)
 */
export function stripDataUriPrefix(dataUri: string): string {
  if (!dataUri) {
    return dataUri;
  }

  // 데이터 URI 프리픽스가 있으면 제거
  if (DATA_URI_PREFIX_PATTERN.test(dataUri)) {
    const commaIndex = dataUri.indexOf(',');
    if (commaIndex !== -1) {
      return dataUri.slice(commaIndex + 1);
    }
  }

  // 프리픽스가 없는 경우 원본 반환
  return dataUri;
}

/**
 * 순수 base64 문자열과 MIME 타입으로 데이터 URI를 생성합니다.
 *
 * @param base64 - 순수 base64 인코딩 문자열 (프리픽스 없음)
 * @param mimeType - MIME 타입 (예: 'image/png')
 * @returns 완성된 데이터 URI 문자열
 *
 * @example
 * createDataUri('iVBORw0KGgo...', 'image/png')
 * // 'data:image/png;base64,iVBORw0KGgo...'
 */
export function createDataUri(base64: string, mimeType: string): string {
  // 이미 데이터 URI 형식이라면 그대로 반환
  if (DATA_URI_PREFIX_PATTERN.test(base64)) {
    return base64;
  }

  return `data:${mimeType};base64,${base64}`;
}

// =============================================================================
// 크기 계산 및 검증
// =============================================================================

/**
 * base64 데이터 URI의 실제 바이트 크기를 계산합니다.
 *
 * base64 인코딩은 3바이트를 4문자로 표현하므로,
 * base64 문자열 길이에 3/4를 곱하여 원본 크기를 근사합니다.
 *
 * 데이터 URI 프리픽스(`data:...;base64,`)는 크기 계산에서 제외됩니다.
 *
 * @param base64DataUri - 크기를 계산할 base64 데이터 URI 또는 순수 base64 문자열
 * @returns 근사 바이트 크기
 *
 * @example
 * // 1MB 이미지의 경우 약 1,048,576 반환
 * getBase64Size('data:image/png;base64,iVBORw0KGgo...')
 */
export function getBase64Size(base64DataUri: string): number {
  if (!base64DataUri) {
    return 0;
  }

  // 프리픽스를 제거하여 순수 base64 문자열 추출
  const base64String = stripDataUriPrefix(base64DataUri);

  // base64 디코딩 후 바이트 크기 근사 계산
  // base64: 4문자 = 3바이트 (패딩 '=' 고려)
  const paddingCount = (base64String.match(/=+$/) ?? [])[0]?.length ?? 0;
  return Math.ceil((base64String.length * 3) / 4) - paddingCount;
}

/**
 * 이미지가 지정된 크기 제한 이내인지 확인합니다.
 *
 * site-flow API의 요청 크기 제한(REQ-MIG-041)을 준수하는지 검사할 때 사용합니다.
 * 전체 페이지 스크린샷은 수 MB에 달할 수 있으므로 업로드 전 반드시 확인해야 합니다.
 *
 * @param base64DataUri - 검사할 base64 데이터 URI
 * @param maxSizeMB - 최대 허용 크기 (MB 단위, 기본값: IMAGE_SIZE_LIMIT_MB = 10)
 * @returns 크기 제한 이내이면 true, 초과하면 false
 *
 * @example
 * isImageWithinSizeLimit('data:image/png;base64,...')           // 기본 10MB 제한 사용
 * isImageWithinSizeLimit('data:image/png;base64,...', 5)        // 5MB 제한으로 검사
 */
export function isImageWithinSizeLimit(
  base64DataUri: string,
  maxSizeMB: number = IMAGE_SIZE_LIMIT_MB
): boolean {
  if (!base64DataUri) {
    return true;
  }

  const byteSize = getBase64Size(base64DataUri);
  const maxBytes = maxSizeMB * 1024 * 1024;

  return byteSize <= maxBytes;
}

// =============================================================================
// 파일 읽기 및 인코딩
// =============================================================================

/**
 * 이미지 파일을 읽어서 base64 데이터 URI 문자열로 반환합니다.
 *
 * 파일 확장자를 기반으로 MIME 타입을 자동으로 감지합니다.
 * 반환값은 `data:{mimeType};base64,{base64data}` 형식입니다.
 *
 * 지원하는 형식: .png, .jpg, .jpeg, .webp, .gif, .svg
 *
 * @param filePath - 읽을 이미지 파일의 절대 경로 또는 상대 경로
 * @returns base64 데이터 URI 문자열
 * @throws {Error} 파일을 찾을 수 없는 경우
 * @throws {Error} 파일 읽기 권한이 없는 경우
 * @throws {Error} 지원하지 않는 이미지 형식인 경우
 *
 * @example
 * const dataUri = await fileToBase64DataUri('/captures/page_1_home.png');
 * // 'data:image/png;base64,iVBORw0KGgo...'
 *
 * // site-flow API 이미지 업로드에 바로 사용 가능
 * await client.addPageImage(pageId, {
 *   siteId: config.siteId,
 *   image: dataUri,
 *   source: 'migration',
 * });
 */
export async function fileToBase64DataUri(filePath: string): Promise<string> {
  // 파일 확장자 확인
  const ext = path.extname(filePath).toLowerCase().replace(/^\./, '');

  // 지원하는 형식인지 검증
  if (ext && !(ext in MIME_TYPE_MAP)) {
    throw new Error(
      `지원하지 않는 이미지 형식입니다: .${ext}` +
      ` (지원 형식: ${Object.keys(MIME_TYPE_MAP).map(e => `.${e}`).join(', ')})`
    );
  }

  // 파일 읽기 (fs/promises 사용)
  let fileBuffer: Buffer;
  try {
    fileBuffer = await fs.readFile(filePath);
  } catch (err) {
    const nodeErr = err as NodeJS.ErrnoException;
    if (nodeErr.code === 'ENOENT') {
      throw new Error(`이미지 파일을 찾을 수 없습니다: ${filePath}`);
    }
    if (nodeErr.code === 'EACCES') {
      throw new Error(`이미지 파일 읽기 권한이 없습니다: ${filePath}`);
    }
    throw new Error(`이미지 파일 읽기 실패: ${filePath} - ${nodeErr.message}`);
  }

  // MIME 타입 결정
  const mimeType = getMimeType(filePath);

  // base64 인코딩 및 데이터 URI 생성
  const base64 = fileBuffer.toString('base64');
  return createDataUri(base64, mimeType);
}

// =============================================================================
// 청크(Chunk) 유틸리티
// =============================================================================

/**
 * 이미지 청크 정보
 */
export interface ImageChunk {
  /** 청크 인덱스 (0부터 시작) */
  index: number;
  /** 전체 청크 수 */
  total: number;
  /** 청크 데이터 (순수 base64, 프리픽스 없음) */
  data: string;
  /** 원본 MIME 타입 */
  mimeType: string;
  /** 이 청크가 마지막 청크인지 여부 */
  isLast: boolean;
}

/**
 * base64 데이터 URI를 지정된 크기 단위로 청크(조각)로 분할합니다.
 *
 * 대용량 이미지를 API 크기 제한에 맞게 분할하여 전송할 때 사용합니다.
 * 각 청크는 순수 base64 문자열로 반환됩니다 (데이터 URI 프리픽스 제외).
 *
 * 주의: site-flow API가 청크 업로드를 지원하는 경우에만 사용하세요.
 * 일반적으로는 이미지를 크기 제한 이내로 리사이즈하는 것을 권장합니다.
 *
 * @param base64DataUri - 분할할 base64 데이터 URI
 * @param chunkSizeMB - 각 청크의 최대 크기 (MB 단위, 기본값: 5)
 * @returns 청크 배열
 * @throws {Error} 유효하지 않은 base64 데이터 URI인 경우
 *
 * @example
 * const chunks = chunkBase64Image('data:image/png;base64,...', 5);
 * // [
 * //   { index: 0, total: 2, data: '...', mimeType: 'image/png', isLast: false },
 * //   { index: 1, total: 2, data: '...', mimeType: 'image/png', isLast: true },
 * // ]
 */
export function chunkBase64Image(
  base64DataUri: string,
  chunkSizeMB: number = 5
): ImageChunk[] {
  if (!base64DataUri) {
    throw new Error('유효하지 않은 base64 데이터 URI입니다: 빈 문자열');
  }

  if (!DATA_URI_PREFIX_PATTERN.test(base64DataUri)) {
    throw new Error(
      '유효하지 않은 base64 데이터 URI입니다: ' +
      '`data:{mimeType};base64,` 형식의 프리픽스가 필요합니다'
    );
  }

  // MIME 타입 추출
  const mimeTypeMatch = base64DataUri.match(/^data:([^;]+);base64,/);
  const mimeType = mimeTypeMatch?.[1] ?? 'application/octet-stream';

  // 순수 base64 문자열 추출
  const base64String = stripDataUriPrefix(base64DataUri);

  // 청크 크기 계산 (바이트 → base64 문자 수)
  // base64: 3바이트 = 4문자이므로 역산
  const chunkSizeBytes = chunkSizeMB * 1024 * 1024;
  const chunkSizeChars = Math.ceil((chunkSizeBytes * 4) / 3);

  // 청크 분할
  const chunks: ImageChunk[] = [];
  let offset = 0;

  while (offset < base64String.length) {
    const chunkData = base64String.slice(offset, offset + chunkSizeChars);
    chunks.push({
      index: chunks.length,
      total: 0, // 아래에서 총 수를 채움
      data: chunkData,
      mimeType,
      isLast: offset + chunkSizeChars >= base64String.length,
    });
    offset += chunkSizeChars;
  }

  // total 필드 업데이트
  const total = chunks.length;
  for (const chunk of chunks) {
    chunk.total = total;
  }

  return chunks;
}

// =============================================================================
// 유효성 검사 유틸리티
// =============================================================================

/**
 * 문자열이 유효한 base64 데이터 URI인지 검사합니다.
 *
 * `data:{mimeType};base64,{base64data}` 형식인지 확인합니다.
 *
 * @param value - 검사할 문자열
 * @returns 유효한 base64 데이터 URI이면 true
 *
 * @example
 * isValidBase64DataUri('data:image/png;base64,iVBORw0KGgo...')  // true
 * isValidBase64DataUri('iVBORw0KGgo...')                         // false
 * isValidBase64DataUri('')                                        // false
 */
export function isValidBase64DataUri(value: string): boolean {
  if (!value || typeof value !== 'string') {
    return false;
  }

  return DATA_URI_PREFIX_PATTERN.test(value);
}

/**
 * 이미지 파일 경로가 지원되는 형식인지 확인합니다.
 *
 * @param filePath - 확인할 파일 경로
 * @returns 지원하는 이미지 형식이면 true
 *
 * @example
 * isSupportedImageFormat('/path/to/image.png')   // true
 * isSupportedImageFormat('/path/to/image.jpg')   // true
 * isSupportedImageFormat('/path/to/file.pdf')    // false
 */
export function isSupportedImageFormat(filePath: string): boolean {
  const ext = path.extname(filePath).toLowerCase().replace(/^\./, '');
  return ext in MIME_TYPE_MAP;
}

/**
 * 이미지 크기 정보
 */
export interface ImageSizeInfo {
  /** 바이트 단위 크기 */
  bytes: number;
  /** KB 단위 크기 (소수점 2자리) */
  kilobytes: number;
  /** MB 단위 크기 (소수점 2자리) */
  megabytes: number;
  /** 크기 제한(10MB) 이내 여부 */
  withinLimit: boolean;
  /** 가독성 있는 크기 문자열 (예: "2.45 MB") */
  readable: string;
}

/**
 * base64 데이터 URI의 크기 정보를 상세하게 반환합니다.
 *
 * 디버깅과 로깅에 유용한 상세 크기 정보를 제공합니다.
 *
 * @param base64DataUri - 크기를 계산할 base64 데이터 URI
 * @returns 크기 정보 객체
 *
 * @example
 * const info = getImageSizeInfo('data:image/png;base64,...');
 * console.log(info.readable);  // '2.45 MB'
 * console.log(info.withinLimit);  // true
 */
export function getImageSizeInfo(base64DataUri: string): ImageSizeInfo {
  const bytes = getBase64Size(base64DataUri);
  const kilobytes = Math.round((bytes / 1024) * 100) / 100;
  const megabytes = Math.round((bytes / (1024 * 1024)) * 100) / 100;
  const withinLimit = isImageWithinSizeLimit(base64DataUri);

  // 가독성 있는 크기 문자열 생성
  let readable: string;
  if (bytes < 1024) {
    readable = `${bytes} B`;
  } else if (bytes < 1024 * 1024) {
    readable = `${kilobytes} KB`;
  } else {
    readable = `${megabytes} MB`;
  }

  return {
    bytes,
    kilobytes,
    megabytes,
    withinLimit,
    readable,
  };
}
