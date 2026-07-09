import type { OperationItem, OperationRow } from '../types'

export function formBody(fields: Record<string, string>) {
  const body = new URLSearchParams()
  for (const [key, value] of Object.entries(fields)) {
    body.set(key, value)
  }
  return body
}

export async function api(path: string, options: RequestInit = {}) {
  const response = await fetch(path, options)
  if (!response.ok) {
    const body = await response.text()
    throw new Error(`${response.status} ${response.statusText}${body ? `\n${body}` : ''}`)
  }

  const contentType = response.headers.get('Content-Type') ?? ''
  if (contentType.includes('application/json')) {
    return response.json() as Promise<unknown>
  }
  return null
}

function stringField(row: OperationRow, keys: string[]) {
  for (const key of keys) {
    const value = row[key]
    if (typeof value === 'string') {
      return value
    }
    if (typeof value === 'number') {
      return String(value)
    }
  }
  return ''
}

function numberField(row: OperationRow, keys: string[]) {
  for (const key of keys) {
    const value = row[key]
    if (typeof value === 'number') {
      return value
    }
    if (typeof value === 'string') {
      const parsed = Number(value)
      if (Number.isFinite(parsed)) {
        return parsed
      }
    }
  }
  return 0
}

function descriptionField(value: unknown) {
  if (typeof value === 'string') {
    return value
  }
  if (value && typeof value === 'object') {
    const description = value as { String?: unknown; Valid?: unknown }
    if (description.Valid === false) {
      return ''
    }
    return typeof description.String === 'string' ? description.String : ''
  }
  return ''
}

export function normalizeOperationRow(row: unknown): OperationItem {
  if (!row || typeof row !== 'object') {
    return { id: 0, name: '', description: '', status: '', priority: '', subProject: '' }
  }

  const record = row as OperationRow
  return {
    id: numberField(record, ['ID', 'id']),
    name: stringField(record, ['Operation', 'Project', 'SubProject', 'Task', 'operation', 'project', 'sub_project', 'task']),
    description: descriptionField(record.Description ?? record.description),
    status: stringField(record, ['Status', 'status']),
    priority: stringField(record, ['Priority', 'priority']),
    subProject: stringField(record, ['SubProject', 'subProject', 'sub_project']),
  }
}

export function normalizeOperationList(payload: unknown) {
  if (!Array.isArray(payload)) {
    return []
  }
  return payload.map(normalizeOperationRow).filter((item) => item.id > 0 && item.name !== '')
}

export function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback
}
