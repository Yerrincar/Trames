import { computed, ref, watch } from 'vue'
import type { Ref } from 'vue'
import { api, formBody, normalizeOperationList, normalizeOperationRow } from '../api/http'
import {
  defaultPriority,
  defaultProjectStatus,
  defaultTaskStatus,
  priorityOptions,
  projectStatusOptions,
  taskStatusOptions,
} from '../constants'
import type {
  ContextMenuState,
  DashboardFilterOption,
  EditableTarget,
  ModalType,
  OperationFormState,
  OperationItem,
  ProjectBoundFormState,
  ProjectDetails,
  SelectedEntity,
  TaskFormState,
  User,
} from '../types'
import type { NotifierController } from './useNotifier'

function emptyDetails(): ProjectDetails {
  return { subProjects: [], tasks: [] }
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

export function useOperations(user: Ref<User | null>, notifier: NotifierController) {
  const projects = ref<OperationItem[]>([])
  const projectDetails = ref<Record<number, ProjectDetails>>({})
  const loadingOperations = ref(false)
  const selectedProjectId = ref<number | null>(null)
  const expandedProjectIds = ref<number[]>([])
  const expandedSubProjectKeys = ref<string[]>([])
  const selectedEntity = ref<SelectedEntity>(null)
  const modal = ref<ModalType>(null)
  const contextMenu = ref<ContextMenuState | null>(null)
  const editingDescription = ref(false)
  const editingStatus = ref(false)
  const editingPriority = ref(false)
  const descriptionDraft = ref('')
  const draggedTask = ref<OperationItem | null>(null)
  const activeDropStatus = ref('')
  const dashboardTaskFilter = ref('all')

  const projectForm = ref<OperationFormState>({ name: '', description: '', status: defaultProjectStatus })
  const subProjectForm = ref<ProjectBoundFormState>({ projectId: null, name: '', description: '', status: defaultProjectStatus })
  const taskForm = ref<TaskFormState>({
    projectId: null,
    subProject: '',
    name: '',
    description: '',
    status: defaultTaskStatus,
    priority: defaultPriority,
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
  const projectDashboardFilterOptions = computed<DashboardFilterOption[]>(() => {
    if (!selectedProject.value) {
      return []
    }
    return [
      { value: 'all', label: 'All tasks' },
      { value: 'project', label: 'Project tasks' },
      ...detailsFor(selectedProject.value.id).subProjects.map((subProject) => ({
        value: `subProject:${subProject.id}`,
        label: subProject.name,
      })),
    ]
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
      const filter = dashboardTaskFilter.value
      if (filter === 'project') {
        return directTasks(selection.projectId)
      }
      if (filter.startsWith('subProject:')) {
        const subProjectId = Number(filter.replace('subProject:', ''))
        const subProject = detailsFor(selection.projectId).subProjects.find((item) => item.id === subProjectId)
        return subProject ? tasksForSubProject(selection.projectId, subProject.name) : []
      }
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
  const editableTarget = computed<EditableTarget | null>(() => {
    const selection = selectedEntity.value
    if (!selection) {
      return null
    }
    if (selection.type === 'project') {
      const item = projects.value.find((project) => project.id === selection.projectId)
      return item ? { type: 'project', projectId: selection.projectId, item } : null
    }
    if (selection.type === 'subProject') {
      const item = detailsFor(selection.projectId).subProjects.find((subProject) => subProject.id === selection.subProjectId)
      return item ? { type: 'subProject', projectId: selection.projectId, subProjectId: selection.subProjectId, item } : null
    }
    const item = detailsFor(selection.projectId).tasks.find((task) => task.id === selection.taskId)
    return item ? { type: 'task', projectId: selection.projectId, taskId: selection.taskId, item } : null
  })
  const editableDescription = computed(() => editableTarget.value?.item.description || 'No description yet.')
  const editableStatus = computed(() => editableTarget.value?.item.status || '')
  const editablePriority = computed(() => editableTarget.value?.type === 'task' ? editableTarget.value.item.priority : '')
  const editableStatusOptions = computed(() => editableTarget.value?.type === 'task' ? taskStatusOptions : projectStatusOptions)
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
    const subProject = task?.subProject ? detailsFor(projectId).subProjects.find((item) => item.name === task.subProject) : null
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

  function ensureDashboardTaskFilter() {
    if (selectedEntity.value?.type !== 'project') {
      return
    }
    const validFilter = projectDashboardFilterOptions.value.some((option) => option.value === dashboardTaskFilter.value)
    if (!validFilter) {
      dashboardTaskFilter.value = 'all'
    }
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
      ensureDashboardTaskFilter()

      const taskCount = Object.values(projectDetails.value).reduce((total, detail) => total + detail.tasks.length, 0)
      notifier.setMessage(projects.value.length === 0 ? 'No projects yet.' : `${projects.value.length} projects / ${taskCount} tasks loaded.`)
    } catch (error) {
      notifier.setError(error, 'Unable to load projects.')
    } finally {
      loadingOperations.value = false
    }
  }

  function resetProjectForm() {
    projectForm.value = { name: '', description: '', status: defaultProjectStatus }
  }

  function resetSubProjectForm(projectId: number | null) {
    subProjectForm.value = { projectId, name: '', description: '', status: defaultProjectStatus }
  }

  function resetTaskForm(projectId: number | null, subProject = '') {
    taskForm.value = { projectId, subProject, name: '', description: '', status: defaultTaskStatus, priority: defaultPriority }
  }

  function openModal(type: Exclude<ModalType, null>, projectId = selectedProjectId.value, subProject = '') {
    closeContextMenu()
    if (type === 'project') resetProjectForm()
    if (type === 'subProject') resetSubProjectForm(projectId)
    if (type === 'task') resetTaskForm(projectId, subProject)
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
      notifier.setMessage('Project created.')
    } catch (error) {
      notifier.setError(error, 'Unable to create project.')
    }
  }

  async function createSubProject() {
    if (subProjectForm.value.projectId === null) {
      notifier.setMessage('Create or select a project first.', 'error')
      return
    }

    try {
      const created = normalizeOperationRow(await api(`/sub_projects/create?projectId=${encodeURIComponent(String(subProjectForm.value.projectId))}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: formBody({ project: subProjectForm.value.name, Description: subProjectForm.value.description, Status: subProjectForm.value.status }),
      }))
      selectedProjectId.value = subProjectForm.value.projectId
      if (created.id > 0) {
        selectedEntity.value = { type: 'subProject', projectId: subProjectForm.value.projectId, subProjectId: created.id }
        expandedSubProjectKeys.value = [...expandedSubProjectKeys.value, subProjectKey(subProjectForm.value.projectId, created.id)]
      }
      closeModal()
      await loadOperations()
      notifier.setMessage('Sub-project created.')
    } catch (error) {
      notifier.setError(error, 'Unable to create sub-project.')
    }
  }

  async function createTask() {
    if (taskForm.value.projectId === null) {
      notifier.setMessage('Create or select a project first.', 'error')
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
      notifier.setMessage('Task created.')
    } catch (error) {
      notifier.setError(error, 'Unable to create task.')
    }
  }

  async function deleteProject(projectId: number, name: string) {
    if (!window.confirm(`Delete project "${name}"?`)) return

    try {
      await api(`/projects/delete?projectId=${encodeURIComponent(String(projectId))}`, { method: 'DELETE' })
      if (selectedProjectId.value === projectId) {
        selectedProjectId.value = null
        selectedEntity.value = null
      }
      await loadOperations()
      notifier.setMessage('Project deleted.')
    } catch (error) {
      notifier.setError(error, 'Unable to delete project.')
    }
  }

  async function deleteSubProject(projectId: number, subProjectId: number, name: string) {
    if (!window.confirm(`Delete sub-project "${name}"?`)) return

    try {
      await api(`/sub_projects/delete?projectId=${encodeURIComponent(String(projectId))}&subProjectId=${encodeURIComponent(String(subProjectId))}`, {
        method: 'DELETE',
      })
      if (selectedEntity.value?.type === 'subProject' && selectedEntity.value.subProjectId === subProjectId) {
        selectedEntity.value = { type: 'project', projectId }
      }
      await loadOperations()
      notifier.setMessage('Sub-project deleted.')
    } catch (error) {
      notifier.setError(error, 'Unable to delete sub-project.')
    }
  }

  async function deleteTask(projectId: number, taskId: number, name: string) {
    if (!window.confirm(`Delete task "${name}"?`)) return

    try {
      await api(`/tasks/delete?projectId=${encodeURIComponent(String(projectId))}&taskId=${encodeURIComponent(String(taskId))}`, { method: 'DELETE' })
      if (selectedEntity.value?.type === 'task' && selectedEntity.value.taskId === taskId) {
        selectedEntity.value = { type: 'project', projectId }
      }
      await loadOperations()
      notifier.setMessage('Task deleted.')
    } catch (error) {
      notifier.setError(error, 'Unable to delete task.')
    }
  }

  function resetInlineEditors() {
    editingDescription.value = false
    editingStatus.value = false
    editingPriority.value = false
    descriptionDraft.value = ''
  }

  function startDescriptionEdit() {
    const target = editableTarget.value
    if (!target) return
    descriptionDraft.value = target.item.description
    editingDescription.value = true
  }

  function cancelDescriptionEdit() {
    editingDescription.value = false
    descriptionDraft.value = ''
  }

  function eventValue(event: Event) {
    return event.target instanceof HTMLSelectElement ? event.target.value : ''
  }

  async function updateEditableTarget(fields: Partial<Pick<OperationItem, 'description' | 'status' | 'priority'>>) {
    const target = editableTarget.value
    if (!target) {
      notifier.setMessage('Select a project, sub-project, or task first.', 'error')
      return
    }

    const next = { ...target.item, ...fields }
    try {
      if (target.type === 'project') {
        await api(`/projects/update?projectId=${encodeURIComponent(String(target.projectId))}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: formBody({ project: next.name, Description: next.description, Status: next.status }),
        })
      }
      if (target.type === 'subProject') {
        await api(`/sub_projects/update?projectId=${encodeURIComponent(String(target.projectId))}&subProjectId=${encodeURIComponent(String(target.subProjectId))}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: formBody({ project: next.name, Description: next.description, Status: next.status }),
        })
      }
      if (target.type === 'task') {
        await api(`/tasks/update?projectId=${encodeURIComponent(String(target.projectId))}&taskId=${encodeURIComponent(String(target.taskId))}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: formBody({
            task: next.name,
            'Sub Project': next.subProject,
            Description: next.description,
            Status: next.status,
            Priority: next.priority,
          }),
        })
      }

      resetInlineEditors()
      await loadOperations()
      notifier.setMessage('Operation updated.')
    } catch (error) {
      notifier.setError(error, 'Unable to update operation.')
    }
  }

  async function saveDescriptionEdit() {
    await updateEditableTarget({ description: descriptionDraft.value })
  }

  async function changeStatus(status: string) {
    if (status === '') return
    editingStatus.value = false
    await updateEditableTarget({ status })
  }

  async function changePriority(priority: string) {
    if (priority === '') return
    editingPriority.value = false
    await updateEditableTarget({ priority })
  }

  function startTaskDrag(event: DragEvent, task: OperationItem) {
    draggedTask.value = task
    activeDropStatus.value = ''
    event.dataTransfer?.setData('text/plain', String(task.id))
    if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move'
  }

  function finishTaskDrag() {
    draggedTask.value = null
    activeDropStatus.value = ''
  }

  function allowStatusDrop(event: DragEvent, status: string) {
    if (!draggedTask.value) return
    event.preventDefault()
    activeDropStatus.value = status
    if (event.dataTransfer) event.dataTransfer.dropEffect = 'move'
  }

  function leaveStatusDrop(status: string) {
    if (activeDropStatus.value === status) activeDropStatus.value = ''
  }

  async function dropTaskOnStatus(status: string) {
    const task = draggedTask.value
    const projectId = selectedProjectId.value
    finishTaskDrag()
    if (!task || projectId === null || task.status === status) return

    try {
      await api(`/tasks/update?projectId=${encodeURIComponent(String(projectId))}&taskId=${encodeURIComponent(String(task.id))}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: formBody({ task: task.name, 'Sub Project': task.subProject, Description: task.description, Status: status, Priority: task.priority }),
      })
      await loadOperations()
      notifier.setMessage(`Task moved to ${status}.`)
    } catch (error) {
      notifier.setError(error, 'Unable to move task.')
    }
  }

  function openProjectMenu(event: MouseEvent, project: OperationItem) {
    event.preventDefault()
    selectedProjectId.value = project.id
    selectedEntity.value = { type: 'project', projectId: project.id }
    contextMenu.value = { type: 'project', x: event.clientX, y: event.clientY, projectId: project.id, itemId: project.id, name: project.name, subProjectName: '' }
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
    contextMenu.value = { type: 'task', x: event.clientX, y: event.clientY, projectId, itemId: task.id, name: task.name, subProjectName: task.subProject }
  }

  function closeContextMenu() {
    contextMenu.value = null
  }

  function contextCreateSubProject() {
    if (!contextMenu.value) return
    openModal('subProject', contextMenu.value.projectId)
  }

  function contextCreateTask() {
    if (!contextMenu.value) return
    openModal('task', contextMenu.value.projectId, contextMenu.value.subProjectName)
  }

  async function contextDelete() {
    const menu = contextMenu.value
    closeContextMenu()
    if (!menu) return
    if (menu.type === 'project') await deleteProject(menu.projectId, menu.name)
    if (menu.type === 'subProject') await deleteSubProject(menu.projectId, menu.itemId, menu.name)
    if (menu.type === 'task') await deleteTask(menu.projectId, menu.itemId, menu.name)
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

  watch(selectedEntity, (selection) => {
    resetInlineEditors()
    if (selection?.type === 'project') {
      dashboardTaskFilter.value = 'all'
    }
  })

  return {
    projects,
    loadingOperations,
    selectedProjectId,
    selectedEntity,
    modal,
    contextMenu,
    editingDescription,
    editingStatus,
    editingPriority,
    descriptionDraft,
    activeDropStatus,
    dashboardTaskFilter,
    projectForm,
    subProjectForm,
    taskForm,
    hasProjects,
    selectedProject,
    selectedSubProject,
    selectedTask,
    selectedTaskProject,
    selectedDetails,
    taskSubProjects,
    dashboardTitle,
    dashboardDescription,
    projectDashboardFilterOptions,
    dashboardGroups,
    dashboardTaskCount,
    dashboardSubProjectCount,
    editableDescription,
    editableStatus,
    editablePriority,
    editableStatusOptions,
    modalTitle,
    clearOperations,
    detailsFor,
    directTasks,
    tasksForSubProject,
    projectTaskCount,
    isProjectExpanded,
    isSubProjectExpanded,
    toggleProject,
    toggleSubProject,
    selectProject,
    selectSubProject,
    selectTask,
    loadOperations,
    openModal,
    closeModal,
    createProject,
    createSubProject,
    createTask,
    startDescriptionEdit,
    cancelDescriptionEdit,
    eventValue,
    saveDescriptionEdit,
    changeStatus,
    changePriority,
    startTaskDrag,
    finishTaskDrag,
    allowStatusDrop,
    leaveStatusDrop,
    dropTaskOnStatus,
    openProjectMenu,
    openSubProjectMenu,
    openTaskMenu,
    closeContextMenu,
    contextCreateSubProject,
    contextCreateTask,
    contextDelete,
  }
}

export type OperationsController = ReturnType<typeof useOperations>
