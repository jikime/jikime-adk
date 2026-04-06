/** Maximum name length for teams, agents, and tasks */
export const MAX_NAME_LEN = 80

/** Validates team/agent/task names -- alphanumeric, underscore, hyphen only */
export const NAME_RE = /^[a-zA-Z0-9_-]{1,80}$/

/** Validates budget values */
export const BUDGET_RE = /^\d{1,10}$/

/** Validate and sanitize a name, returns null if invalid */
export function validateName(name: string): string | null {
  if (!name || !NAME_RE.test(name)) return null
  return name
}
