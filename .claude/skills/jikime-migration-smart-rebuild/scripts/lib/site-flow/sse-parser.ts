/**
 * SSE (Server-Sent Events) 파서
 *
 * site-flow API의 `auto-capture-batch` 엔드포인트가 반환하는
 * SSE 스트리밍 응답을 파싱합니다.
 *
 * SSE 와이어 포맷:
 *   data: {"type":"progress","pageId":"abc123","current":1,"total":10,"message":"..."}\n\n
 *   data: {"type":"complete","message":"All pages captured"}\n\n
 *
 * 설계 원칙:
 * - 외부 의존성 없음 (Node.js 20+ 내장 API만 사용)
 * - 청크 경계에서의 부분 라인을 올바르게 버퍼링
 * - \n\n 및 \r\n\r\n 이벤트 구분자를 모두 처리
 * - 잘못된 JSON은 경고 로그 후 스킵 (오류 미전파)
 */

import type { SSEEventType, SSECaptureEvent } from './types';

// SSEEventType과 SSECaptureEvent를 외부에서도 사용할 수 있도록 재내보내기
export type { SSEEventType, SSECaptureEvent };

// =============================================================================
// 내부 유틸리티 타입
// =============================================================================

/**
 * SSE 라인 파싱 결과
 * SSE 이벤트를 구성하는 필드 한 줄의 파싱 결과를 나타냅니다.
 */
interface SSELine {
  /** 필드명 (예: "data", "event", "id") */
  field: string;
  /** 필드 값 */
  value: string;
}

// =============================================================================
// 내부 파싱 헬퍼 함수
// =============================================================================

/**
 * SSE 한 줄을 필드와 값으로 파싱합니다.
 *
 * SSE 명세(RFC 8895)에 따라:
 * - "field: value" 형식 파싱
 * - 빈 줄은 null 반환 (이벤트 경계)
 * - 콜론 없는 줄은 전체를 필드명으로 처리
 *
 * @param line - 파싱할 SSE 라인 (개행 문자 제거된 상태)
 * @returns 파싱된 SSE 라인 또는 null (빈 줄인 경우)
 */
function parseSSELine(line: string): SSELine | null {
  // 빈 줄은 이벤트 경계를 나타냄
  if (line === '') {
    return null;
  }

  // BOM(Byte Order Mark) 제거 - 일부 서버가 첫 줄에 포함시킬 수 있음
  const cleanLine = line.startsWith('\uFEFF') ? line.slice(1) : line;

  // 주석 라인 스킵 (콜론으로 시작하는 경우)
  if (cleanLine.startsWith(':')) {
    return { field: '', value: cleanLine.slice(1).trimStart() };
  }

  const colonIndex = cleanLine.indexOf(':');

  // 콜론이 없으면 전체가 필드명
  if (colonIndex === -1) {
    return { field: cleanLine, value: '' };
  }

  const field = cleanLine.slice(0, colonIndex);
  // 콜론 바로 뒤의 공백 하나만 제거 (SSE 명세)
  const rawValue = cleanLine.slice(colonIndex + 1);
  const value = rawValue.startsWith(' ') ? rawValue.slice(1) : rawValue;

  return { field, value };
}

/**
 * 누적된 SSE 필드들로부터 SSECaptureEvent를 조립합니다.
 *
 * "data" 필드의 값을 JSON으로 파싱하여 이벤트를 생성합니다.
 * 잘못된 JSON인 경우 경고를 출력하고 null을 반환합니다.
 *
 * @param dataLines - 수집된 "data" 필드 값 목록 (멀티라인 data 필드 지원)
 * @returns 파싱된 이벤트 또는 null (파싱 실패 시)
 */
function assembleEvent(dataLines: string[]): SSECaptureEvent | null {
  // data 필드가 없으면 이벤트가 아님
  if (dataLines.length === 0) {
    return null;
  }

  // 멀티라인 data 필드는 개행으로 이어 붙임 (SSE 명세)
  const rawData = dataLines.join('\n');

  // "[DONE]" 같은 특수 종료 마커는 무시
  if (rawData.trim() === '[DONE]') {
    return null;
  }

  try {
    const parsed = JSON.parse(rawData) as unknown;

    // 타입 안전성: 파싱된 객체가 SSECaptureEvent 형태인지 확인
    if (typeof parsed !== 'object' || parsed === null) {
      console.warn('[SSE 파서] 경고: data 필드가 객체가 아닙니다:', rawData);
      return null;
    }

    const obj = parsed as Record<string, unknown>;

    // "type" 필드는 필수
    if (typeof obj['type'] !== 'string') {
      console.warn('[SSE 파서] 경고: type 필드가 없거나 문자열이 아닙니다:', rawData);
      return null;
    }

    const event: SSECaptureEvent = {
      type: obj['type'] as SSEEventType,
    };

    // 선택적 필드 안전하게 할당
    if (typeof obj['pageId'] === 'string') {
      event.pageId = obj['pageId'];
    }
    if (typeof obj['current'] === 'number') {
      event.current = obj['current'];
    }
    if (typeof obj['total'] === 'number') {
      event.total = obj['total'];
    }
    if (typeof obj['message'] === 'string') {
      event.message = obj['message'];
    }

    return event;
  } catch {
    console.warn('[SSE 파서] 경고: JSON 파싱 실패, 이벤트를 스킵합니다:', rawData);
    return null;
  }
}

// =============================================================================
// 공개 API
// =============================================================================

/**
 * fetch Response의 SSE 스트림을 파싱합니다.
 *
 * 스트림을 청크 단위로 읽으며 부분 라인을 버퍼링하고,
 * 완전한 SSE 이벤트가 구성될 때마다 onEvent 콜백을 호출합니다.
 *
 * 처리 로직:
 * 1. Response.body(ReadableStream)를 TextDecoderStream으로 디코딩
 * 2. 청크를 버퍼에 누적하며 라인 단위로 분리
 * 3. 빈 줄(이벤트 경계)을 만나면 수집된 data 필드로 이벤트 조립
 * 4. 조립된 이벤트를 onEvent 콜백으로 전달
 *
 * @param response - fetch()가 반환한 Response 객체 (SSE 스트림)
 * @param onEvent - 파싱된 이벤트를 처리할 콜백 함수
 * @returns 스트림이 완료되면 resolve되는 Promise
 * @throws 네트워크 오류 발생 시 reject (스트림 중단은 정상 종료로 처리)
 *
 * @example
 * const response = await fetch('/api/pages/auto-capture-batch', { ... });
 * await parseSSEStream(response, (event) => {
 *   console.log('이벤트:', event);
 * });
 */
export async function parseSSEStream(
  response: Response,
  onEvent: (event: SSECaptureEvent) => void
): Promise<void> {
  // Response body가 없는 경우 (예: HEAD 요청)
  if (!response.body) {
    console.warn('[SSE 파서] 경고: Response body가 없습니다.');
    return;
  }

  // TextDecoderStream을 통해 바이트 스트림을 문자열 스트림으로 변환 (UTF-8)
  const reader = response.body
    .pipeThrough(new TextDecoderStream('utf-8', { fatal: false }))
    .getReader();

  // 라인 경계에 걸친 부분 텍스트를 보관하는 버퍼
  let lineBuffer = '';

  // 현재 이벤트를 구성 중인 data 라인 목록
  let currentDataLines: string[] = [];

  try {
    while (true) {
      let done: boolean;
      let value: string | undefined;

      try {
        const result = await reader.read();
        done = result.done;
        value = result.value;
      } catch {
        // 스트림이 중단된 경우 (AbortController, 네트워크 끊김 등)
        // 정상 종료로 처리 (Promise resolve)
        break;
      }

      if (done) {
        // 스트림 종료 시 버퍼에 남은 내용 처리
        if (lineBuffer.trim().length > 0) {
          // 마지막 이벤트에 개행이 없어도 처리 시도
          const line = parseSSELine(lineBuffer);
          if (line && line.field === 'data') {
            currentDataLines.push(line.value);
          }
          if (currentDataLines.length > 0) {
            const event = assembleEvent(currentDataLines);
            if (event !== null) {
              onEvent(event);
            }
          }
        }
        break;
      }

      if (!value) {
        continue;
      }

      // 이번 청크를 버퍼에 추가
      lineBuffer += value;

      // 버퍼에서 완전한 라인들을 추출
      // \r\n\r\n 및 \n\n 이벤트 구분자를 모두 처리하기 위해
      // 라인 단위로 순차 처리합니다.
      while (true) {
        // 다음 개행 문자 위치 탐색 (\r\n 또는 \n)
        const lfIndex = lineBuffer.indexOf('\n');

        // 완전한 라인이 없으면 다음 청크 대기
        if (lfIndex === -1) {
          break;
        }

        // \r\n 처리: \n 앞의 \r을 제거
        const crIndex = lfIndex > 0 && lineBuffer[lfIndex - 1] === '\r' ? lfIndex - 1 : lfIndex;
        const rawLine = lineBuffer.slice(0, crIndex);

        // 버퍼에서 처리된 라인 제거 (\n 이후부터)
        lineBuffer = lineBuffer.slice(lfIndex + 1);

        // 빈 줄 = 이벤트 경계
        if (rawLine === '') {
          // 수집된 data 필드들로 이벤트 조립
          if (currentDataLines.length > 0) {
            const event = assembleEvent(currentDataLines);
            if (event !== null) {
              onEvent(event);
            }
            // 다음 이벤트를 위해 초기화
            currentDataLines = [];
          }
          continue;
        }

        // 라인 파싱
        const parsedLine = parseSSELine(rawLine);

        // null이면 빈 줄 (위에서 이미 처리됨) 또는 주석
        if (parsedLine === null) {
          continue;
        }

        // data 필드만 수집 (site-flow SSE는 data 필드만 사용)
        // 향후 event/id 필드 지원이 필요하면 여기에 추가
        if (parsedLine.field === 'data') {
          currentDataLines.push(parsedLine.value);
        }
        // 그 외 필드(event, id, retry 등)는 현재 무시
      }
    }
  } catch (error) {
    // 예상치 못한 네트워크 오류
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`[SSE 파서] 스트림 읽기 중 오류 발생: ${message}`);
  } finally {
    // 리더를 항상 해제하여 리소스 누수 방지
    reader.releaseLock();
  }
}

// =============================================================================
// 진행 상황 핸들러 팩토리
// =============================================================================

/**
 * createSSEProgressHandler 옵션
 */
export interface SSEProgressHandlerOptions {
  /**
   * 커스텀 진행 상황 콜백
   * 이벤트를 직접 처리하고 싶을 때 사용합니다.
   */
  onProgress?: (event: SSECaptureEvent) => void;

  /**
   * 콘솔 출력 억제 여부 (기본값: false)
   * true로 설정하면 콘솔에 아무것도 출력하지 않습니다.
   */
  silent?: boolean;
}

/**
 * SSE 진행 상황 핸들러를 생성하는 팩토리 함수입니다.
 *
 * parseSSEStream의 onEvent 콜백으로 전달할 수 있는 함수를 반환합니다.
 * 기본 동작으로 진행 상황을 콘솔에 출력하며,
 * 커스텀 onProgress 콜백을 통해 추가 처리를 할 수 있습니다.
 *
 * 콘솔 출력 포맷:
 * - progress: "[2/10] Capturing page 2/10"
 * - complete: "[완료] All pages captured"
 * - error: "[오류] Error message"
 *
 * @param options - 핸들러 설정 옵션
 * @returns SSECaptureEvent를 처리하는 콜백 함수
 *
 * @example
 * // 기본 사용법 (콘솔 출력)
 * const handler = createSSEProgressHandler();
 * await parseSSEStream(response, handler);
 *
 * @example
 * // 커스텀 핸들러와 함께 사용
 * const handler = createSSEProgressHandler({
 *   onProgress: (event) => updateProgressBar(event.current, event.total),
 *   silent: true,
 * });
 * await parseSSEStream(response, handler);
 */
export function createSSEProgressHandler(
  options: SSEProgressHandlerOptions = {}
): (event: SSECaptureEvent) => void {
  const { onProgress, silent = false } = options;

  return (event: SSECaptureEvent): void => {
    // 커스텀 콜백이 있으면 먼저 호출
    if (onProgress) {
      onProgress(event);
    }

    // silent 모드이면 콘솔 출력 생략
    if (silent) {
      return;
    }

    // 이벤트 유형별 콘솔 출력
    switch (event.type) {
      case 'progress': {
        // "[2/10] Capturing page 2/10" 형식으로 출력
        const position =
          event.current !== undefined && event.total !== undefined
            ? `[${event.current}/${event.total}] `
            : '';
        const message = event.message ?? '처리 중...';
        console.log(`${position}${message}`);
        break;
      }

      case 'complete': {
        const message = event.message ?? '완료되었습니다.';
        console.log(`[완료] ${message}`);
        break;
      }

      case 'error': {
        const message = event.message ?? '알 수 없는 오류가 발생했습니다.';
        console.error(`[오류] ${message}`);
        break;
      }

      default: {
        // 알 수 없는 이벤트 유형은 경고 출력
        // TypeScript의 exhaustive check: event.type이 never여야 하지만
        // 런타임에서 새로운 타입이 올 수 있으므로 안전하게 처리
        const unknownType = (event as SSECaptureEvent).type;
        console.warn(`[SSE 파서] 경고: 알 수 없는 이벤트 유형: ${unknownType}`);
        break;
      }
    }
  };
}
