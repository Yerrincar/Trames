export type View = 'home' | 'login' | 'register'
export type ModalType = 'project' | 'subProject' | 'task' | null
export type ContextType = 'project' | 'subProject' | 'task'
export type MessageKind = 'info' | 'error'

export type SelectedEntity =
  | { type: 'project'; projectId: number }
  | { type: 'subProject'; projectId: number; subProjectId: number }
  | { type: 'task'; projectId: number; taskId: number }
  | null

export type User = {
  id: number
  username: string
  created_at: string
  updated_at: string
}

export type OperationItem = {
  id: number
  name: string
  description: string
  status: string
  priority: string
  subProject: string
}

export type ProjectDetails = {
  subProjects: OperationItem[]
  tasks: OperationItem[]
}

export type OperationFormState = {
  name: string
  description: string
  status: string
}

export type ProjectBoundFormState = OperationFormState & {
  projectId: number | null
}

export type TaskFormState = ProjectBoundFormState & {
  priority: string
  subProject: string
}

export type ContextMenuState = {
  type: ContextType
  x: number
  y: number
  projectId: number
  itemId: number
  name: string
  subProjectName: string
}

export type EditableTarget =
  | { type: 'project'; projectId: number; item: OperationItem }
  | { type: 'subProject'; projectId: number; subProjectId: number; item: OperationItem }
  | { type: 'task'; projectId: number; taskId: number; item: OperationItem }

export type DashboardFilterOption = {
  value: string
  label: string
}

export type OperationRow = Record<string, unknown>
