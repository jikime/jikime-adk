/**
 * IndexedDB 래퍼 — webchat-p 채팅 메시지 영속화
 *
 * DB:    webchat_p  (v2)
 * Store: sessions
 *   key  : sessionId (string)
 *   value: { sessionId, messages, savedAt }
 */

const DB_NAME = 'webchat_p'
const DB_VERSION = 2
const STORE = 'sessions'

export interface StoredProgress {
  type: 'tool_call' | 'tool_result'
  name?: string
  input?: string
  content?: string
  images?: { data: string; mediaType: string }[]
}

export interface StoredMessage {
  id: string
  role: 'user' | 'assistant' | 'error'
  text: string
  streaming?: boolean   // 저장 시 항상 false
  progress: StoredProgress[]
}

export interface SessionMeta {
  sessionId: string
  firstMessage: string  // 첫 사용자 메시지 요약 (60자)
  messageCount: number
  savedAt: number       // 마지막 저장 타임스탬프
}

// ─── DB 초기화 ───────────────────────────────────────────────────────────────

let _db: IDBDatabase | null = null

function openDB(): Promise<IDBDatabase> {
  if (_db) return Promise.resolve(_db)

  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, DB_VERSION)

    req.onupgradeneeded = () => {
      const db = req.result
      if (!db.objectStoreNames.contains(STORE)) {
        db.createObjectStore(STORE, { keyPath: 'sessionId' })
      }
    }

    req.onsuccess = () => { _db = req.result; resolve(_db) }
    req.onerror = () => reject(req.error)
  })
}

// ─── 퍼블릭 API ──────────────────────────────────────────────────────────────

/** 세션 메시지 로드 */
export async function loadMessages(sessionId: string): Promise<StoredMessage[]> {
  if (!sessionId) return []
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE, 'readonly')
    const req = tx.objectStore(STORE).get(sessionId)
    req.onsuccess = () => resolve((req.result?.messages ?? []) as StoredMessage[])
    req.onerror = () => reject(req.error)
  })
}

/** 세션 메시지 저장 (streaming 플래그 강제 false, savedAt 갱신) */
export async function saveMessages(sessionId: string, messages: StoredMessage[]): Promise<void> {
  if (!sessionId) return
  const db = await openDB()
  const sanitized = messages.map(m => ({ ...m, streaming: false }))
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE, 'readwrite')
    tx.objectStore(STORE).put({ sessionId, messages: sanitized, savedAt: Date.now() })
    tx.oncomplete = () => resolve()
    tx.onerror = () => reject(tx.error)
  })
}

/** 세션 메시지 삭제 */
export async function clearMessages(sessionId: string): Promise<void> {
  if (!sessionId) return
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE, 'readwrite')
    tx.objectStore(STORE).delete(sessionId)
    tx.oncomplete = () => resolve()
    tx.onerror = () => reject(tx.error)
  })
}

/** IndexedDB에 저장된 세션 목록 반환 (최신순) */
export async function listSessions(): Promise<SessionMeta[]> {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE, 'readonly')
    const req = tx.objectStore(STORE).getAll()
    req.onsuccess = () => {
      const rows = (req.result ?? []) as Array<{
        sessionId: string
        messages: StoredMessage[]
        savedAt: number
      }>
      const metas: SessionMeta[] = rows
        .filter(r => r.messages.length > 0)
        .map(r => {
          const first = r.messages.find(m => m.role === 'user')
          return {
            sessionId: r.sessionId,
            firstMessage: (first?.text ?? '').slice(0, 60),
            messageCount: r.messages.length,
            savedAt: r.savedAt ?? 0,
          }
        })
        .sort((a, b) => b.savedAt - a.savedAt)
      resolve(metas)
    }
    req.onerror = () => reject(req.error)
  })
}
