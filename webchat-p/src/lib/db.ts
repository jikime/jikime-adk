/**
 * IndexedDB 래퍼 — webchat-p 채팅 메시지 영속화
 *
 * DB:    webchat_p  (v1)
 * Store: sessions
 *   key  : sessionId (string) — 빈 문자열 = 새 대화
 *   value: { sessionId, messages: StoredMessage[] }
 */

const DB_NAME = 'webchat_p'
const DB_VERSION = 1
const STORE = 'sessions'

export interface StoredMessage {
  id: string
  role: 'user' | 'assistant' | 'error'
  text: string
  streaming?: boolean          // 저장 시 항상 false
  progress: StoredProgress[]
}

export interface StoredProgress {
  type: 'tool_call' | 'tool_result'
  name?: string
  input?: string
  content?: string
}

// ─── DB 초기화 ───────────────────────────────────────────────────────────────

let _db: IDBDatabase | null = null

function openDB(): Promise<IDBDatabase> {
  if (_db) return Promise.resolve(_db)

  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, DB_VERSION)

    req.onupgradeneeded = () => {
      req.result.createObjectStore(STORE, { keyPath: 'sessionId' })
    }

    req.onsuccess = () => {
      _db = req.result
      resolve(_db)
    }

    req.onerror = () => reject(req.error)
  })
}

// ─── 퍼블릭 API ──────────────────────────────────────────────────────────────

/** 세션의 메시지 목록 로드 */
export async function loadMessages(sessionId: string): Promise<StoredMessage[]> {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE, 'readonly')
    const req = tx.objectStore(STORE).get(sessionId || '__new__')
    req.onsuccess = () => resolve((req.result?.messages ?? []) as StoredMessage[])
    req.onerror = () => reject(req.error)
  })
}

/** 세션의 메시지 목록 저장 (streaming 플래그 강제 false) */
export async function saveMessages(sessionId: string, messages: StoredMessage[]): Promise<void> {
  const db = await openDB()
  const sanitized = messages.map(m => ({ ...m, streaming: false }))
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE, 'readwrite')
    tx.objectStore(STORE).put({ sessionId: sessionId || '__new__', messages: sanitized })
    tx.oncomplete = () => resolve()
    tx.onerror = () => reject(tx.error)
  })
}

/** 세션 메시지 삭제 */
export async function clearMessages(sessionId: string): Promise<void> {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE, 'readwrite')
    tx.objectStore(STORE).delete(sessionId || '__new__')
    tx.oncomplete = () => resolve()
    tx.onerror = () => reject(tx.error)
  })
}
