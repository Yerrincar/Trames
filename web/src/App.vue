<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

type View = 'home' | 'login' | 'register'
type ModalType = 'project' | 'subProject' | 'task' | null
type ContextType = 'project' | 'subProject' | 'task'
type MessageKind = 'info' | 'error'

type SelectedEntity =
  | { type: 'project'; projectId: number }
  | { type: 'subProject'; projectId: number; subProjectId: number }
  | { type: 'task'; projectId: number; taskId: number }
  | null

type User = {
  id: number
  username: string
  created_at: string
  updated_at: string
}

type OperationItem = {
  id: number
  name: string
  description: string
  status: string
  priority: string
  subProject: string
}

type ProjectDetails = {
  subProjects: OperationItem[]
  tasks: OperationItem[]
}

type OperationFormState = {
  name: string
  description: string
  status: string
}

type ProjectBoundFormState = OperationFormState & {
  projectId: number | null
}

type TaskFormState = ProjectBoundFormState & {
  priority: string
  subProject: string
}

type ContextMenuState = {
  type: ContextType
  x: number
  y: number
  projectId: number
  itemId: number
  name: string
  subProjectName: string
}

type OperationRow = Record<string, unknown>

const projectStatusOptions = ['IDEA', 'TO BE STARTED', 'PLANNING', 'IN PROGRESS', 'DONE']
const taskStatusOptions = ['EXPERIMENTAL', 'TO-DO', 'IN PROGRESS', 'BLOCKED', 'TEST', 'DONE']
const priorityOptions = ['IDEA', 'LOW', 'MEDIUM', 'HIGH', 'CRITICAL']

const user = ref<User | null>(null)
const username = ref('')
const password = ref('')
const message = ref('Checking session...')
const messageKind = ref<MessageKind>('info')
const view = ref<View>(viewFromPath())
const projects = ref<OperationItem[]>([])
const projectDetails = ref<Record<number, ProjectDetails>>({})
const loadingOperations = ref(false)
const selectedProjectId = ref<number | null>(null)
const expandedProjectIds = ref<number[]>([])
const expandedSubProjectKeys = ref<string[]>([])
const selectedEntity = ref<SelectedEntity>(null)
const modal = ref<ModalType>(null)
const contextMenu = ref<ContextMenuState | null>(null)

const projectForm = ref<OperationFormState>({ name: '', description: '', status: 'TO BE STARTED' })
const subProjectForm = ref<ProjectBoundFormState>({ projectId: null, name: '', description: '', status: 'TO BE STARTED' })
const taskForm = ref<TaskFormState>({
  projectId: null,
  subProject: '',
  name: '',
  description: '',
  status: 'TO-DO',
  priority: 'LOW',
})

const hasProjects = computed(() => projects.value.length > 0)
const selectedProject = computed(() => {
  if (selectedProjectId.value === null) {
    return null
  }
  return projects.value.find((project) => project.id === selectedProjectId.value) ?? null
})
const selectedSubProject = computed(() => {
  const selection = selectedEntity.value
  if (!selection || selection.type !== 'subProject') {
    return null
  }
  return detailsFor(selection.projectId).subProjects.find((subProject) => subProject.id === selection.subProjectId) ?? null
})
const selectedTask = computed(() => {
  const selection = selectedEntity.value
  if (!selection || selection.type !== 'task') {
    return null
  }
  return detailsFor(selection.projectId).tasks.find((task) => task.id === selection.taskId) ?? null
})
const selectedTaskProject = computed(() => {
  const selection = selectedEntity.value
  if (!selection || selection.type !== 'task') {
    return null
  }
  return projects.value.find((project) => project.id === selection.projectId) ?? null
})
const selectedDetails = computed(() => (selectedProjectId.value === null ? emptyDetails() : detailsFor(selectedProjectId.value)))
const taskSubProjects = computed(() => {
  if (taskForm.value.projectId === null) {
    return []
  }
  return detailsFor(taskForm.value.projectId).subProjects
})
const dashboardTitle = computed(() => {
  if (selectedEntity.value?.type === 'subProject') {
    return selectedSubProject.value?.name ?? 'Sub-project'
  }
  return selectedProject.value?.name ?? 'Home'
})
const dashboardDescription = computed(() => {
  if (selectedEntity.value?.type === 'subProject') {
    return selectedSubProject.value?.description || 'No description yet.'
  }
  return selectedProject.value?.description || 'No description yet.'
})
const dashboardTasks = computed(() => {
  const selection = selectedEntity.value
  if (!selection) {
    return []
  }
  if (selection.type === 'subProject') {
    const subProject = selectedSubProject.value
    return subProject ? tasksForSubProject(selection.projectId, subProject.name) : []
  }
  if (selection.type === 'project') {
    return detailsFor(selection.projectId).tasks
  }
  return []
})
const dashboardGroups = computed(() =>
  taskStatusOptions.map((status) => ({
    status,
    tasks: sortTasksByPriority(dashboardTasks.value.filter((task) => task.status === status)),
  })),
)
const dashboardTaskCount = computed(() => dashboardTasks.value.length)
const dashboardSubProjectCount = computed(() => {
  const selection = selectedEntity.value
  if (!selection || selection.type !== 'project') {
    return 0
  }
  return detailsFor(selection.projectId).subProjects.length
})
const modalTitle = computed(() => {
  if (modal.value === 'project') {
    return 'Create Project'
  }
  if (modal.value === 'subProject') {
    return 'Create Sub-Project'
  }
  if (modal.value === 'task') {
    return 'Create Task'
  }
  return ''
})

function viewFromPath(): View {
  switch (window.location.pathname) {
    case '/register':
      return 'register'
    case '/login':
      return 'login'
    default:
      return 'home'
  }
}

function go(nextView: View) {
  const path = nextView === 'home' ? '/' : `/${nextView}`
  window.history.pushState({}, '', path)
  view.value = nextView
}

function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback
}

function setMessage(text: string, kind: MessageKind = 'info') {
  message.value = text
  messageKind.value = kind
}

function setError(error: unknown, fallback: string) {
  setMessage(errorMessage(error, fallback), 'error')
}

function clearOperations() {
  projects.value = []
  projectDetails.value = {}
  selectedProjectId.value = null
  expandedProjectIds.value = []
  expandedSubProjectKeys.value = []
  selectedEntity.value = null
  subProjectForm.value.projectId = null
  taskForm.value.projectId = null
  taskForm.value.subProject = ''
}

function formBody(fields: Record<string, string>) {
  const body = new URLSearchParams()
  for (const [key, value] of Object.entries(fields)) {
    body.set(key, value)
  }
  return body
}

async function api(path: string, options: RequestInit = {}) {
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

function normalizeOperationRow(row: unknown): OperationItem {
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

function normalizeOperationList(payload: unknown) {
  if (!Array.isArray(payload)) {
    return []
  }
  return payload.map(normalizeOperationRow).filter((item) => item.id > 0 && item.name !== '')
}

function emptyDetails(): ProjectDetails {
  return { subProjects: [], tasks: [] }
}

function detailsFor(projectId: number) {
  return projectDetails.value[projectId] ?? emptyDetails()
}

function directTasks(projectId: number) {
  return detailsFor(projectId).tasks.filter((task) => task.subProject === '')
}

function tasksForSubProject(projectId: number, subProjectName: string) {
  return detailsFor(projectId).tasks.filter((task) => task.subProject === subProjectName)
}

function projectTaskCount(projectId: number) {
  return detailsFor(projectId).tasks.length
}

function priorityWeight(priority: string) {
  const order = ['CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'IDEA']
  const index = order.indexOf(priority)
  return index === -1 ? order.length : index
}

function sortTasksByPriority(tasks: OperationItem[]) {
  return [...tasks].sort((first, second) => {
    const priorityDelta = priorityWeight(first.priority) - priorityWeight(second.priority)
    if (priorityDelta !== 0) {
      return priorityDelta
    }
    return first.name.localeCompare(second.name)
  })
}

function subProjectKey(projectId: number, subProjectId: number) {
  return `${projectId}:${subProjectId}`
}

function isProjectExpanded(projectId: number) {
  return expandedProjectIds.value.includes(projectId)
}

function isSubProjectExpanded(projectId: number, subProjectId: number) {
  return expandedSubProjectKeys.value.includes(subProjectKey(projectId, subProjectId))
}

function toggleProject(projectId: number) {
  if (isProjectExpanded(projectId)) {
    expandedProjectIds.value = expandedProjectIds.value.filter((id) => id !== projectId)
    return
  }
  expandedProjectIds.value = [...expandedProjectIds.value, projectId]
}

function toggleSubProject(projectId: number, subProjectId: number) {
  const key = subProjectKey(projectId, subProjectId)
  if (expandedSubProjectKeys.value.includes(key)) {
    expandedSubProjectKeys.value = expandedSubProjectKeys.value.filter((existingKey) => existingKey !== key)
    return
  }
  expandedSubProjectKeys.value = [...expandedSubProjectKeys.value, key]
}

function selectProject(projectId: number) {
  selectedProjectId.value = projectId
  selectedEntity.value = { type: 'project', projectId }
  if (!isProjectExpanded(projectId)) {
    expandedProjectIds.value = [...expandedProjectIds.value, projectId]
  }
}

function selectSubProject(projectId: number, subProjectId: number) {
  selectedProjectId.value = projectId
  selectedEntity.value = { type: 'subProject', projectId, subProjectId }
  if (!isProjectExpanded(projectId)) {
    expandedProjectIds.value = [...expandedProjectIds.value, projectId]
  }
  if (!isSubProjectExpanded(projectId, subProjectId)) {
    expandedSubProjectKeys.value = [...expandedSubProjectKeys.value, subProjectKey(projectId, subProjectId)]
  }
}

function selectTask(projectId: number, taskId: number) {
  selectedProjectId.value = projectId
  selectedEntity.value = { type: 'task', projectId, taskId }
  if (!isProjectExpanded(projectId)) {
    expandedProjectIds.value = [...expandedProjectIds.value, projectId]
  }
  const task = detailsFor(projectId).tasks.find((item) => item.id === taskId)
  const subProject = task?.subProject
    ? detailsFor(projectId).subProjects.find((item) => item.name === task.subProject)
    : null
  if (subProject && !isSubProjectExpanded(projectId, subProject.id)) {
    expandedSubProjectKeys.value = [...expandedSubProjectKeys.value, subProjectKey(projectId, subProject.id)]
  }
}

function ensureProjectSelection() {
  const firstProjectId = projects.value[0]?.id ?? null
  const hasSelectedProject = projects.value.some((project) => project.id === selectedProjectId.value)
  if (!hasSelectedProject) {
    selectedProjectId.value = firstProjectId
  }

  const selectedId = selectedProjectId.value
  if (!projects.value.some((project) => project.id === subProjectForm.value.projectId)) {
    subProjectForm.value.projectId = selectedId
  }
  if (!projects.value.some((project) => project.id === taskForm.value.projectId)) {
    taskForm.value.projectId = selectedId
  }
}

function syncExpandedProjects() {
  const previous = new Set(expandedProjectIds.value)
  const selectedId = selectedProjectId.value
  expandedProjectIds.value = projects.value
    .map((project) => project.id)
    .filter((projectId) => previous.size === 0 || previous.has(projectId) || projectId === selectedId)
}

function syncExpandedSubProjects() {
  const existingKeys = new Set(
    projects.value.flatMap((project) => detailsFor(project.id).subProjects.map((subProject) => subProjectKey(project.id, subProject.id))),
  )
  expandedSubProjectKeys.value = expandedSubProjectKeys.value.filter((key) => existingKeys.has(key))
}

function selectionExists(selection: Exclude<SelectedEntity, null>) {
  if (selection.type === 'project') {
    return projects.value.some((project) => project.id === selection.projectId)
  }
  if (selection.type === 'subProject') {
    return detailsFor(selection.projectId).subProjects.some((subProject) => subProject.id === selection.subProjectId)
  }
  return detailsFor(selection.projectId).tasks.some((task) => task.id === selection.taskId)
}

function ensureSelectionEntity() {
  const selection = selectedEntity.value
  if (selection && selectionExists(selection)) {
    return
  }
  if (selectedProjectId.value !== null && projects.value.some((project) => project.id === selectedProjectId.value)) {
    selectedEntity.value = { type: 'project', projectId: selectedProjectId.value }
    return
  }
  selectedEntity.value = null
}

async function loadOperations() {
  if (!user.value) {
    clearOperations()
    return
  }

  loadingOperations.value = true
  try {
    projects.value = normalizeOperationList(await api('/projects'))
    ensureProjectSelection()
    syncExpandedProjects()

    const entries = await Promise.all(
      projects.value.map(async (project) => {
        const [subProjectRows, taskRows] = await Promise.all([
          api(`/sub_projects?projectId=${encodeURIComponent(String(project.id))}`),
          api(`/tasks?projectId=${encodeURIComponent(String(project.id))}`),
        ])
        return [
          project.id,
          {
            subProjects: normalizeOperationList(subProjectRows),
            tasks: normalizeOperationList(taskRows),
          },
        ] as const
      }),
    )

    projectDetails.value = Object.fromEntries(entries) as Record<number, ProjectDetails>
    ensureProjectSelection()
    syncExpandedProjects()
    syncExpandedSubProjects()
    ensureSelectionEntity()

    const taskCount = Object.values(projectDetails.value).reduce((total, detail) => total + detail.tasks.length, 0)
    setMessage(projects.value.length === 0 ? 'No projects yet.' : `${projects.value.length} projects / ${taskCount} tasks loaded.`)
  } catch (error) {
    setError(error, 'Unable to load projects.')
  } finally {
    loadingOperations.value = false
  }
}

async function refreshCurrentUser() {
  const response = await fetch('/users/currentUser')
  if (!response.ok) {
    user.value = null
    clearOperations()
    setMessage(view.value === 'home' ? 'Ready.' : 'Not logged in.')
    return
  }
  user.value = (await response.json()) as User
  setMessage(`Logged in as ${user.value.username}`)
  go('home')
  await loadOperations()
}

async function submit(action: 'login' | 'register') {
  const response = await fetch(`/users/${action}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: username.value, password: password.value }),
  })

  if (!response.ok) {
    setMessage(`${response.status} ${response.statusText}\n${await response.text()}`, 'error')
    return
  }

  if (action === 'login') {
    user.value = (await response.json()) as User
    setMessage(`Logged in as ${user.value.username}`)
    go('home')
    await loadOperations()
    return
  }

  setMessage('User registered. You can log in now.')
  password.value = ''
  go('login')
}

async function logout() {
  const response = await fetch('/users/logout', { method: 'POST' })
  user.value = null
  clearOperations()
  closeModal()
  closeContextMenu()
  setMessage(response.ok ? 'Logged out.' : `${response.status} ${response.statusText}\n${await response.text()}`, response.ok ? 'info' : 'error')
  username.value = ''
  password.value = ''
  go('home')
}

function resetProjectForm() {
  projectForm.value = { name: '', description: '', status: 'TO BE STARTED' }
}

function resetSubProjectForm(projectId: number | null) {
  subProjectForm.value = { projectId, name: '', description: '', status: 'TO BE STARTED' }
}

function resetTaskForm(projectId: number | null, subProject = '') {
  taskForm.value = {
    projectId,
    subProject,
    name: '',
    description: '',
    status: 'TO-DO',
    priority: 'LOW',
  }
}

function openModal(type: Exclude<ModalType, null>, projectId = selectedProjectId.value, subProject = '') {
  closeContextMenu()
  if (type === 'project') {
    resetProjectForm()
  }
  if (type === 'subProject') {
    resetSubProjectForm(projectId)
  }
  if (type === 'task') {
    resetTaskForm(projectId, subProject)
  }
  modal.value = type
}

function closeModal() {
  modal.value = null
}

async function createProject() {
  try {
    const created = normalizeOperationRow(
      await api('/projects/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: formBody({ project: projectForm.value.name, Description: projectForm.value.description, Status: projectForm.value.status }),
      }),
    )
    if (created.id > 0) {
      selectedProjectId.value = created.id
      selectedEntity.value = { type: 'project', projectId: created.id }
    }
    closeModal()
    await loadOperations()
    setMessage('Project created.')
  } catch (error) {
    setError(error, 'Unable to create project.')
  }
}

async function createSubProject() {
  if (subProjectForm.value.projectId === null) {
    setMessage('Create or select a project first.', 'error')
    return
  }

  try {
    const created = normalizeOperationRow(await api(`/sub_projects/create?projectId=${encodeURIComponent(String(subProjectForm.value.projectId))}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: formBody({
        project: subProjectForm.value.name,
        Description: subProjectForm.value.description,
        Status: subProjectForm.value.status,
      }),
    }))
    selectedProjectId.value = subProjectForm.value.projectId
    if (created.id > 0) {
      selectedEntity.value = { type: 'subProject', projectId: subProjectForm.value.projectId, subProjectId: created.id }
      expandedSubProjectKeys.value = [
        ...expandedSubProjectKeys.value,
        subProjectKey(subProjectForm.value.projectId, created.id),
      ]
    }
    closeModal()
    await loadOperations()
    setMessage('Sub-project created.')
  } catch (error) {
    setError(error, 'Unable to create sub-project.')
  }
}

async function createTask() {
  if (taskForm.value.projectId === null) {
    setMessage('Create or select a project first.', 'error')
    return
  }

  try {
    const created = normalizeOperationRow(await api(`/tasks/create?projectId=${encodeURIComponent(String(taskForm.value.projectId))}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: formBody({
        task: taskForm.value.name,
        'Sub Project': taskForm.value.subProject,
        Description: taskForm.value.description,
        Status: taskForm.value.status,
        Priority: taskForm.value.priority,
      }),
    }))
    selectedProjectId.value = taskForm.value.projectId
    if (created.id > 0) {
      selectedEntity.value = { type: 'task', projectId: taskForm.value.projectId, taskId: created.id }
    }
    closeModal()
    await loadOperations()
    setMessage('Task created.')
  } catch (error) {
    setError(error, 'Unable to create task.')
  }
}

async function deleteProject(projectId: number, name: string) {
  if (!window.confirm(`Delete project "${name}"?`)) {
    return
  }

  try {
    await api(`/projects/delete?projectId=${encodeURIComponent(String(projectId))}`, { method: 'DELETE' })
    if (selectedProjectId.value === projectId) {
      selectedProjectId.value = null
      selectedEntity.value = null
    }
    await loadOperations()
    setMessage('Project deleted.')
  } catch (error) {
    setError(error, 'Unable to delete project.')
  }
}

async function deleteSubProject(projectId: number, subProjectId: number, name: string) {
  if (!window.confirm(`Delete sub-project "${name}"?`)) {
    return
  }

  try {
    await api(
      `/sub_projects/delete?projectId=${encodeURIComponent(String(projectId))}&subProjectId=${encodeURIComponent(String(subProjectId))}`,
      { method: 'DELETE' },
    )
    if (selectedEntity.value?.type === 'subProject' && selectedEntity.value.subProjectId === subProjectId) {
      selectedEntity.value = { type: 'project', projectId }
    }
    await loadOperations()
    setMessage('Sub-project deleted.')
  } catch (error) {
    setError(error, 'Unable to delete sub-project.')
  }
}

async function deleteTask(projectId: number, taskId: number, name: string) {
  if (!window.confirm(`Delete task "${name}"?`)) {
    return
  }

  try {
    await api(`/tasks/delete?projectId=${encodeURIComponent(String(projectId))}&taskId=${encodeURIComponent(String(taskId))}`, { method: 'DELETE' })
    if (selectedEntity.value?.type === 'task' && selectedEntity.value.taskId === taskId) {
      selectedEntity.value = { type: 'project', projectId }
    }
    await loadOperations()
    setMessage('Task deleted.')
  } catch (error) {
    setError(error, 'Unable to delete task.')
  }
}

function openProjectMenu(event: MouseEvent, project: OperationItem) {
  event.preventDefault()
  selectedProjectId.value = project.id
  selectedEntity.value = { type: 'project', projectId: project.id }
  contextMenu.value = {
    type: 'project',
    x: event.clientX,
    y: event.clientY,
    projectId: project.id,
    itemId: project.id,
    name: project.name,
    subProjectName: '',
  }
}

function openSubProjectMenu(event: MouseEvent, projectId: number, subProject: OperationItem) {
  event.preventDefault()
  selectedProjectId.value = projectId
  selectedEntity.value = { type: 'subProject', projectId, subProjectId: subProject.id }
  contextMenu.value = {
    type: 'subProject',
    x: event.clientX,
    y: event.clientY,
    projectId,
    itemId: subProject.id,
    name: subProject.name,
    subProjectName: subProject.name,
  }
}

function openTaskMenu(event: MouseEvent, projectId: number, task: OperationItem) {
  event.preventDefault()
  selectedProjectId.value = projectId
  selectedEntity.value = { type: 'task', projectId, taskId: task.id }
  contextMenu.value = {
    type: 'task',
    x: event.clientX,
    y: event.clientY,
    projectId,
    itemId: task.id,
    name: task.name,
    subProjectName: task.subProject,
  }
}

function closeContextMenu() {
  contextMenu.value = null
}

function contextCreateSubProject() {
  if (!contextMenu.value) {
    return
  }
  openModal('subProject', contextMenu.value.projectId)
}

function contextCreateTask() {
  if (!contextMenu.value) {
    return
  }
  openModal('task', contextMenu.value.projectId, contextMenu.value.subProjectName)
}

async function contextDelete() {
  const menu = contextMenu.value
  closeContextMenu()
  if (!menu) {
    return
  }
  if (menu.type === 'project') {
    await deleteProject(menu.projectId, menu.name)
  }
  if (menu.type === 'subProject') {
    await deleteSubProject(menu.projectId, menu.itemId, menu.name)
  }
  if (menu.type === 'task') {
    await deleteTask(menu.projectId, menu.itemId, menu.name)
  }
}

function handlePopState() {
  view.value = viewFromPath()
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    closeModal()
    closeContextMenu()
  }
}

watch(selectedProjectId, (projectId) => {
  subProjectForm.value.projectId = projectId
  taskForm.value.projectId = projectId
  taskForm.value.subProject = ''
  if (projectId !== null && !isProjectExpanded(projectId)) {
    expandedProjectIds.value = [...expandedProjectIds.value, projectId]
  }
})

watch(
  () => taskForm.value.projectId,
  () => {
    taskForm.value.subProject = ''
  },
)

onMounted(() => {
  window.addEventListener('popstate', handlePopState)
  window.addEventListener('click', closeContextMenu)
  window.addEventListener('keydown', handleKeydown)
  refreshCurrentUser().catch((error) => {
    user.value = null
    clearOperations()
    setError(error, 'Unable to check current user.')
  })
})

onUnmounted(() => {
  window.removeEventListener('popstate', handlePopState)
  window.removeEventListener('click', closeContextMenu)
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <main class="app-root" :class="{ authenticated: user }">
    <template v-if="user">
      <section class="workspace-layout">
        <aside class="left-panel" aria-label="Projects">
          <header class="left-toolbar">
            <button type="button" class="icon-button" aria-label="Create project" @click="openModal('project')">
              <svg viewBox="0 0 16 16" aria-hidden="true">
                <path d="M2.5 3.5h5l1.4 1.8h4.6v7.2h-11z" />
                <path d="M8 7v4M6 9h4" />
              </svg>
            </button>
            <button type="button" class="icon-button" aria-label="Create sub-project" :disabled="!hasProjects" @click="openModal('subProject')">
              <svg viewBox="0 0 16 16" aria-hidden="true">
                <path d="M2.5 4h4l1.1 1.4h5.9v6.6h-11z" />
                <path d="M5 8h6M8 6v4" />
              </svg>
            </button>
            <button type="button" class="icon-button" aria-label="Create task" :disabled="!hasProjects" @click="openModal('task')">
              <svg viewBox="0 0 16 16" aria-hidden="true">
                <path d="M4 3h8M4 8h8M4 13h8" />
                <path d="M2 3h.5M2 8h.5M2 13h.5" />
              </svg>
            </button>
          </header>

          <section class="user-strip" aria-label="Current user">
            <span class="strip-label">User</span>
            <span>{{ user.username }}</span>
          </section>

          <nav class="project-tree" aria-label="Project tree">
            <div class="tree-title">
              <span>Projects</span>
              <span>{{ projects.length }}</span>
            </div>

            <p v-if="loadingOperations" class="tree-empty">Loading...</p>
            <p v-else-if="projects.length === 0" class="tree-empty">No projects. Create one from the toolbar.</p>

            <div v-for="project in projects" :key="project.id" class="tree-group">
              <button
                type="button"
                class="tree-row project-row"
                :class="{ active: project.id === selectedProjectId }"
                @click="selectProject(project.id)"
                @contextmenu="openProjectMenu($event, project)"
              >
                <span class="disclosure" :class="{ open: isProjectExpanded(project.id) }" @click.stop="toggleProject(project.id)">
                  <svg viewBox="0 0 8 8" aria-hidden="true"><path d="M2 1.5 5.5 4 2 6.5" /></svg>
                </span>
                <svg class="project-icon" viewBox="0 0 16 16" aria-hidden="true">
                  <path d="M2.5 3.5h4.2l1.4 1.7h5.4v7.3h-11z" />
                  <path d="M4.5 7.5h7" />
                </svg>
                <span class="tree-name">{{ project.name }}</span>
                <span class="tree-count">{{ projectTaskCount(project.id) }}</span>
              </button>

              <div v-if="isProjectExpanded(project.id)" class="tree-children">
                <div v-for="subProject in detailsFor(project.id).subProjects" :key="subProject.id" class="sub-tree">
                  <button
                    type="button"
                    class="tree-row subproject-row"
                    :class="{ active: selectedEntity?.type === 'subProject' && selectedEntity.subProjectId === subProject.id }"
                    @click="selectSubProject(project.id, subProject.id)"
                    @contextmenu="openSubProjectMenu($event, project.id, subProject)"
                  >
                    <span class="disclosure" :class="{ open: isSubProjectExpanded(project.id, subProject.id) }" @click.stop="toggleSubProject(project.id, subProject.id)">
                      <svg viewBox="0 0 8 8" aria-hidden="true"><path d="M2 1.5 5.5 4 2 6.5" /></svg>
                    </span>
                    <span class="branch-line"></span>
                    <span class="tree-name">{{ subProject.name }}</span>
                    <span class="tree-count">{{ tasksForSubProject(project.id, subProject.name).length }}</span>
                  </button>

                  <button
                    v-if="isSubProjectExpanded(project.id, subProject.id)"
                    v-for="task in tasksForSubProject(project.id, subProject.name)"
                    :key="task.id"
                    type="button"
                    class="tree-row task-row nested"
                    :class="{ active: selectedEntity?.type === 'task' && selectedEntity.taskId === task.id }"
                    @click="selectTask(project.id, task.id)"
                    @contextmenu="openTaskMenu($event, project.id, task)"
                  >
                    <span class="tree-indent deep"></span>
                    <span class="task-marker"></span>
                    <span class="tree-name">{{ task.name }}</span>
                    <span class="tree-status">{{ task.priority }}</span>
                  </button>
                </div>

                <button
                  v-for="task in directTasks(project.id)"
                  :key="task.id"
                  type="button"
                  class="tree-row task-row"
                  :class="{ active: selectedEntity?.type === 'task' && selectedEntity.taskId === task.id }"
                  @click="selectTask(project.id, task.id)"
                  @contextmenu="openTaskMenu($event, project.id, task)"
                >
                  <span class="tree-indent"></span>
                  <span class="task-marker"></span>
                  <span class="tree-name">{{ task.name }}</span>
                  <span class="tree-status">{{ task.status }}</span>
                </button>
              </div>
            </div>
          </nav>

          <footer class="left-footer">
            <button type="button" class="logout-button" @click="logout">Logout</button>
          </footer>
        </aside>

        <section class="center-panel" aria-label="Workspace">
          <aside v-if="messageKind === 'error'" class="error-banner" role="alert">
            {{ message }}
          </aside>

          <section v-if="selectedTask && selectedTaskProject" class="project-focus task-detail">
            <p class="overline">Task detail</p>
            <h1>{{ selectedTask.name }}</h1>
            <p class="project-description">{{ selectedTask.description || 'No description yet.' }}</p>

            <div class="detail-grid" aria-label="Task information">
              <div>
                <span>Project</span>
                <strong>{{ selectedTaskProject.name }}</strong>
              </div>
              <div>
                <span>Sub-project</span>
                <strong>{{ selectedTask.subProject || 'Direct project task' }}</strong>
              </div>
              <div>
                <span>Status</span>
                <strong>{{ selectedTask.status }}</strong>
              </div>
              <div>
                <span>Priority</span>
                <strong>{{ selectedTask.priority }}</strong>
              </div>
              <div>
                <span>ID</span>
                <strong>#{{ selectedTask.id }}</strong>
              </div>
            </div>
          </section>

          <section v-else-if="selectedProject && selectedEntity" class="project-focus">
            <p class="overline">{{ selectedEntity.type === 'subProject' ? 'Sub-project dashboard' : 'Project dashboard' }}</p>
            <h1>{{ dashboardTitle }}</h1>
            <p class="project-description">{{ dashboardDescription }}</p>

            <div class="project-stats" aria-label="Dashboard stats">
              <div>
                <strong>{{ selectedEntity.type === 'subProject' ? selectedProject.name : dashboardSubProjectCount }}</strong>
                <span>{{ selectedEntity.type === 'subProject' ? 'Parent project' : 'Sub-projects' }}</span>
              </div>
              <div>
                <strong>{{ dashboardTaskCount }}</strong>
                <span>Tasks</span>
              </div>
              <div>
                <strong>{{ selectedEntity.type === 'subProject' ? selectedSubProject?.status : selectedProject.status }}</strong>
                <span>Status</span>
              </div>
            </div>

            <section class="dashboard-grid" aria-label="Tasks by status">
              <article v-for="group in dashboardGroups" :key="group.status" class="status-column">
                <header>
                  <span>{{ group.status }}</span>
                  <span>{{ group.tasks.length }}</span>
                </header>
                <p v-if="group.tasks.length === 0" class="empty-center">No tasks.</p>
                <button
                  v-for="task in group.tasks"
                  :key="task.id"
                  type="button"
                  class="dashboard-task"
                  @click="selectedProjectId !== null && selectTask(selectedProjectId, task.id)"
                  @contextmenu="selectedProjectId !== null && openTaskMenu($event, selectedProjectId, task)"
                >
                  <span>{{ task.name }}</span>
                  <small>{{ task.subProject || 'Project' }}</small>
                  <strong>{{ task.priority }}</strong>
                </button>
              </article>
            </section>
          </section>

          <section v-else class="home-focus" aria-label="Home">
            <p class="overline">Personal task manager</p>
            <h1>Trames</h1>
            <p>Select a project on the left or create one from the toolbar.</p>
          </section>
        </section>
      </section>

      <div
        v-if="contextMenu"
        class="context-menu"
        :style="{ left: `${contextMenu.x}px`, top: `${contextMenu.y}px` }"
        @click.stop
      >
        <button v-if="contextMenu.type === 'project'" type="button" @click="contextCreateSubProject">Create sub-project</button>
        <button v-if="contextMenu.type === 'project' || contextMenu.type === 'subProject'" type="button" @click="contextCreateTask">Create task</button>
        <button type="button" class="danger" @click="contextDelete">Delete {{ contextMenu.type }}</button>
      </div>

      <div v-if="modal" class="modal-backdrop" @click.self="closeModal">
        <section class="modal-window" role="dialog" aria-modal="true" :aria-label="modalTitle">
          <header class="modal-header">
            <div>
              <p class="overline">Operation</p>
              <h2>{{ modalTitle }}</h2>
            </div>
            <button type="button" class="icon-button" aria-label="Close" @click="closeModal">
              <svg viewBox="0 0 16 16" aria-hidden="true"><path d="M4 4l8 8M12 4l-8 8" /></svg>
            </button>
          </header>

          <form v-if="modal === 'project'" class="operation-form" @submit.prevent="createProject">
            <label>
              Project
              <input v-model="projectForm.name" required autofocus />
            </label>
            <label>
              Description
              <textarea v-model="projectForm.description" rows="4" />
            </label>
            <label>
              Status
              <select v-model="projectForm.status" required>
                <option v-for="status in projectStatusOptions" :key="status" :value="status">{{ status }}</option>
              </select>
            </label>
            <button type="submit" class="submit-button">Create project</button>
          </form>

          <form v-if="modal === 'subProject'" class="operation-form" @submit.prevent="createSubProject">
            <label>
              Project
              <select v-model="subProjectForm.projectId" required>
                <option v-for="project in projects" :key="project.id" :value="project.id">{{ project.name }}</option>
              </select>
            </label>
            <label>
              Sub-project
              <input v-model="subProjectForm.name" required autofocus />
            </label>
            <label>
              Description
              <textarea v-model="subProjectForm.description" rows="4" />
            </label>
            <label>
              Status
              <select v-model="subProjectForm.status" required>
                <option v-for="status in projectStatusOptions" :key="status" :value="status">{{ status }}</option>
              </select>
            </label>
            <button type="submit" class="submit-button">Create sub-project</button>
          </form>

          <form v-if="modal === 'task'" class="operation-form" @submit.prevent="createTask">
            <label>
              Project
              <select v-model="taskForm.projectId" required>
                <option v-for="project in projects" :key="project.id" :value="project.id">{{ project.name }}</option>
              </select>
            </label>
            <label>
              Sub-project
              <select v-model="taskForm.subProject">
                <option value="">Direct project task</option>
                <option v-for="subProject in taskSubProjects" :key="subProject.id" :value="subProject.name">{{ subProject.name }}</option>
              </select>
            </label>
            <label>
              Task
              <input v-model="taskForm.name" required autofocus />
            </label>
            <label>
              Description
              <textarea v-model="taskForm.description" rows="4" />
            </label>
            <div class="form-pair">
              <label>
                Status
                <select v-model="taskForm.status" required>
                  <option v-for="status in taskStatusOptions" :key="status" :value="status">{{ status }}</option>
                </select>
              </label>
              <label>
                Priority
                <select v-model="taskForm.priority" required>
                  <option v-for="priority in priorityOptions" :key="priority" :value="priority">{{ priority }}</option>
                </select>
              </label>
            </div>
            <button type="submit" class="submit-button">Create task</button>
          </form>
        </section>
      </div>
    </template>

    <template v-else>
      <section v-if="view === 'home'" class="landing-screen">
        <p class="overline">Task manager</p>
        <h1>Trames</h1>
        <nav class="landing-actions" aria-label="Authentication">
          <a href="/login" @click.prevent="go('login')">Login</a>
          <a href="/register" @click.prevent="go('register')">Register</a>
        </nav>
      </section>

      <section v-else class="auth-screen">
        <form class="auth-box" @submit.prevent="submit(view === 'register' ? 'register' : 'login')">
          <p class="overline">Trames</p>
          <h1>{{ view === 'register' ? 'Register' : 'Login' }}</h1>
          <label>
            Username
            <input v-model="username" autocomplete="username" required />
          </label>
          <label>
            Password
            <input v-model="password" type="password" :autocomplete="view === 'register' ? 'new-password' : 'current-password'" required />
          </label>
          <button type="submit" class="submit-button">{{ view === 'register' ? 'Register' : 'Login' }}</button>
          <nav class="auth-links" aria-label="Authentication links">
            <a href="/" @click.prevent="go('home')">Home</a>
            <a v-if="view === 'login'" href="/register" @click.prevent="go('register')">Register</a>
            <a v-else href="/login" @click.prevent="go('login')">Login</a>
          </nav>
          <p class="auth-message" :class="{ error: messageKind === 'error' }">{{ message }}</p>
        </form>
      </section>
    </template>
  </main>
</template>
