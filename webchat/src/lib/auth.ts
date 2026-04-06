import { NextRequest, NextResponse } from 'next/server'

/**
 * Simple bearer token authentication for webchat API routes.
 *
 * Set WEBCHAT_AUTH_TOKEN env var to enable authentication.
 * When not set, all requests are allowed (local development mode).
 *
 * Usage in API routes:
 *   const authError = checkAuth(request)
 *   if (authError) return authError
 */

const AUTH_TOKEN = process.env.WEBCHAT_AUTH_TOKEN || ''

export function checkAuth(request: NextRequest): NextResponse | null {
  // No token configured → open access (local dev)
  if (!AUTH_TOKEN) return null

  const authHeader = request.headers.get('Authorization')
  if (!authHeader) {
    return NextResponse.json(
      { error: 'Authorization header required' },
      { status: 401 }
    )
  }

  const token = authHeader.replace(/^Bearer\s+/i, '')
  if (token !== AUTH_TOKEN) {
    return NextResponse.json(
      { error: 'Invalid token' },
      { status: 403 }
    )
  }

  return null
}
