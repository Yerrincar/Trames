<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

type View = 'home' | 'login' | 'register'

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

type OperationRow = Record<string, unknown>

const projectStatusOptions = ['IDEA', 'TO BE STARTED', 'PLANNING', 'IN PROGRESS', 'DONE']
const taskStatusOptions = ['EXPERIMENTAL', 'TO-DO', 'IN PROGRESS', 'BLOCKED', 'TEST', 'DONE']
const priorityOptions = ['LOW', 'MEDIUM', 'HIGH']

const user = ref<User | null>(null)
const username = ref('')
const password = ref('')
const message = ref('Checking session...')
const view = ref<View>(viewFromPath())
const projects = ref<OperationItem[]>([])
const projectDetails = ref<Record<number, ProjectDetails>>({})
const loadingOperations = ref(false)
const selectedProjectId = ref<number | null>(null)

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

const title = computed(() => {
  if (user.value) {
    return 'Trames'
  }
  return view.value === 'register' ? 'Register' : 'Login'
})

const hasProjects = computed(() => projects.value.length > 0)
const taskSubProjects = computed(() => {
  if (taskForm.value.projectId === null) {
    return []
  }
  return detailsFor(taskForm.value.projectId).subProjects
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

function clearOperations() {
  projects.value = []
  projectDetails.value = {}
  selectedProjectId.value = null
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

function ensureProjectSelection() {
  const firstProjectId = projects.value[0]?.id ?? null
  const hasSelectedProject = projects.value.some((project) => project.id === selectedProjectId.value)
  if (!hasSelectedProject) {
    selectedProjectId.value = firstProjectId
  }

  const formProjectId = selectedProjectId.value
  if (!projects.value.some((project) => project.id === subProjectForm.value.projectId)) {
    subProjectForm.value.projectId = formProjectId
  }
  if (!projects.value.some((project) => project.id === taskForm.value.projectId)) {
    taskForm.value.projectId = formProjectId
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

    const taskCount = Object.values(projectDetails.value).reduce((total, detail) => total + detail.tasks.length, 0)
    message.value = projects.value.length === 0 ? 'No projects yet.' : `Loaded ${projects.value.length} project(s), ${taskCount} task(s).`
  } catch (error) {
    message.value = errorMessage(error, 'Unable to load projects.')
  } finally {
    loadingOperations.value = false
  }
}

async function refreshCurrentUser() {
  const response = await fetch('/users/currentUser')
  if (!response.ok) {
    user.value = null
    clearOperations()
    message.value = 'Not logged in.'
    if (view.value === 'home') {
      view.value = 'login'
    }
    return
  }
  user.value = (await response.json()) as User
  message.value = `Logged in as ${user.value.username}`
  await loadOperations()
}

async function submit(action: 'login' | 'register') {
  const response = await fetch(`/users/${action}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: username.value, password: password.value }),
  })

  if (!response.ok) {
    message.value = `${response.status} ${response.statusText}\n${await response.text()}`
    return
  }

  if (action === 'login') {
    user.value = (await response.json()) as User
    message.value = `Logged in as ${user.value.username}`
    go('home')
    await loadOperations()
    return
  }

  message.value = 'User registered. You can log in now.'
  password.value = ''
  go('login')
}

async function logout() {
  const response = await fetch('/users/logout', { method: 'POST' })
  user.value = null
  clearOperations()
  message.value = response.ok ? 'Logged out.' : `${response.status} ${response.statusText}\n${await response.text()}`
  username.value = ''
  password.value = ''
  go('login')
}

async function createProject() {
  try {
    await api('/projects/create', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: formBody({ project: projectForm.value.name, Description: projectForm.value.description, Status: projectForm.value.status }),
    })
    projectForm.value.name = ''
    projectForm.value.description = ''
    await loadOperations()
    message.value = 'Project created.'
  } catch (error) {
    message.value = errorMessage(error, 'Unable to create project.')
  }
}

async function createSubProject() {
  if (subProjectForm.value.projectId === null) {
    message.value = 'Create or select a project first.'
    return
  }

  try {
    await api(`/sub_projects/create?projectId=${encodeURIComponent(String(subProjectForm.value.projectId))}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: formBody({
        project: subProjectForm.value.name,
        Description: subProjectForm.value.description,
        Status: subProjectForm.value.status,
      }),
    })
    subProjectForm.value.name = ''
    subProjectForm.value.description = ''
    await loadOperations()
    message.value = 'Sub-project created.'
  } catch (error) {
    message.value = errorMessage(error, 'Unable to create sub-project.')
  }
}

async function createTask() {
  if (taskForm.value.projectId === null) {
    message.value = 'Create or select a project first.'
    return
  }

  try {
    await api(`/tasks/create?projectId=${encodeURIComponent(String(taskForm.value.projectId))}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: formBody({
        task: taskForm.value.name,
        'Sub Project': taskForm.value.subProject,
        Description: taskForm.value.description,
        Status: taskForm.value.status,
        Priority: taskForm.value.priority,
      }),
    })
    taskForm.value.name = ''
    taskForm.value.description = ''
    taskForm.value.subProject = ''
    await loadOperations()
    message.value = 'Task created.'
  } catch (error) {
    message.value = errorMessage(error, 'Unable to create task.')
  }
}

function handlePopState() {
  view.value = viewFromPath()
}

watch(selectedProjectId, (projectId) => {
  subProjectForm.value.projectId = projectId
  taskForm.value.projectId = projectId
  taskForm.value.subProject = ''
})

watch(
  () => taskForm.value.projectId,
  () => {
    taskForm.value.subProject = ''
  },
)

onMounted(() => {
  window.addEventListener('popstate', handlePopState)
  refreshCurrentUser().catch((error) => {
    user.value = null
    clearOperations()
    message.value = errorMessage(error, 'Unable to check current user.')
  })
})

onUnmounted(() => window.removeEventListener('popstate', handlePopState))
</script>

<template>
  <main>
    <section class="panel" :class="{ wide: user }">
      <p class="eyebrow">Task manager</p>
      <h1>{{ title }}</h1>

      <div v-if="user" class="workspace">
        <header class="workspace-header">
          <p class="intro">Logged in as {{ user.username }}</p>
          <div class="actions">
            <button type="button" @click="loadOperations" :disabled="loadingOperations">Refresh</button>
            <button type="button" class="secondary" @click="logout">Logout</button>
          </div>
        </header>

        <section class="forms-grid" aria-label="Create operations">
          <form @submit.prevent="createProject">
            <h2>Create project</h2>
            <label>
              Project
              <input v-model="projectForm.name" required />
            </label>
            <label>
              Description
              <textarea v-model="projectForm.description" rows="3" />
            </label>
            <label>
              Status
              <select v-model="projectForm.status" required>
                <option v-for="status in projectStatusOptions" :key="status" :value="status">{{ status }}</option>
              </select>
            </label>
            <button type="submit">Create project</button>
          </form>

          <form @submit.prevent="createSubProject">
            <h2>Create sub-project</h2>
            <label>
              Project
              <select v-model="subProjectForm.projectId" :disabled="!hasProjects" required>
                <option v-for="project in projects" :key="project.id" :value="project.id">{{ project.name }}</option>
              </select>
            </label>
            <label>
              Sub-project
              <input v-model="subProjectForm.name" :disabled="!hasProjects" required />
            </label>
            <label>
              Description
              <textarea v-model="subProjectForm.description" :disabled="!hasProjects" rows="3" />
            </label>
            <label>
              Status
              <select v-model="subProjectForm.status" :disabled="!hasProjects" required>
                <option v-for="status in projectStatusOptions" :key="status" :value="status">{{ status }}</option>
              </select>
            </label>
            <button type="submit" :disabled="!hasProjects">Create sub-project</button>
          </form>

          <form @submit.prevent="createTask">
            <h2>Create task</h2>
            <label>
              Project
              <select v-model="taskForm.projectId" :disabled="!hasProjects" required>
                <option v-for="project in projects" :key="project.id" :value="project.id">{{ project.name }}</option>
              </select>
            </label>
            <label>
              Sub-project
              <select v-model="taskForm.subProject" :disabled="!hasProjects">
                <option value="">Direct project task</option>
                <option v-for="subProject in taskSubProjects" :key="subProject.id" :value="subProject.name">
                  {{ subProject.name }}
                </option>
              </select>
            </label>
            <label>
              Task
              <input v-model="taskForm.name" :disabled="!hasProjects" required />
            </label>
            <label>
              Description
              <textarea v-model="taskForm.description" :disabled="!hasProjects" rows="3" />
            </label>
            <div class="split-fields">
              <label>
                Status
                <select v-model="taskForm.status" :disabled="!hasProjects" required>
                  <option v-for="status in taskStatusOptions" :key="status" :value="status">{{ status }}</option>
                </select>
              </label>
              <label>
                Priority
                <select v-model="taskForm.priority" :disabled="!hasProjects" required>
                  <option v-for="priority in priorityOptions" :key="priority" :value="priority">{{ priority }}</option>
                </select>
              </label>
            </div>
            <button type="submit" :disabled="!hasProjects">Create task</button>
          </form>
        </section>

        <section class="project-board" aria-label="Project overview">
          <div class="section-title">
            <div>
              <p class="eyebrow">Overview</p>
              <h2>All projects</h2>
            </div>
            <label v-if="hasProjects" class="compact-label">
              Active project
              <select v-model="selectedProjectId">
                <option v-for="project in projects" :key="project.id" :value="project.id">{{ project.name }}</option>
              </select>
            </label>
          </div>

          <p v-if="loadingOperations" class="notice">Loading projects...</p>
          <p v-else-if="!projects.length" class="empty-state">Create a project to start testing sub-projects and tasks.</p>

          <article v-for="project in projects" :key="project.id" class="project-card" :class="{ active: project.id === selectedProjectId }">
            <header class="project-card-header">
              <div>
                <p class="eyebrow">Project #{{ project.id }}</p>
                <h3>{{ project.name }}</h3>
                <p v-if="project.description" class="description">{{ project.description }}</p>
                <p class="meta">{{ project.status }}</p>
              </div>
              <button type="button" class="secondary" @click="selectedProjectId = project.id">Use for forms</button>
            </header>

            <div class="project-columns">
              <section>
                <h4>Sub-projects</h4>
                <p v-if="detailsFor(project.id).subProjects.length === 0" class="empty-state">No sub-projects.</p>
                <article v-for="subProject in detailsFor(project.id).subProjects" :key="subProject.id" class="mini-card">
                  <div class="mini-card-header">
                    <strong>{{ subProject.name }}</strong>
                    <span>{{ subProject.status }}</span>
                  </div>
                  <p v-if="subProject.description" class="description">{{ subProject.description }}</p>
                  <ul v-if="tasksForSubProject(project.id, subProject.name).length" class="task-list">
                    <li v-for="task in tasksForSubProject(project.id, subProject.name)" :key="task.id">
                      <span>{{ task.name }}</span>
                      <small>{{ task.status }} / {{ task.priority }}</small>
                    </li>
                  </ul>
                  <p v-else class="empty-state">No tasks in this sub-project.</p>
                </article>
              </section>

              <section>
                <h4>Direct tasks</h4>
                <ul v-if="directTasks(project.id).length" class="task-list direct">
                  <li v-for="task in directTasks(project.id)" :key="task.id">
                    <span>{{ task.name }}</span>
                    <small>{{ task.status }} / {{ task.priority }}</small>
                  </li>
                </ul>
                <p v-else class="empty-state">No direct tasks.</p>
              </section>
            </div>
          </article>
        </section>
      </div>

      <template v-else>
        <nav class="auth-links" aria-label="Authentication">
          <a href="/login" :aria-current="view === 'login' ? 'page' : undefined" @click.prevent="go('login')">Login</a>
          <a href="/register" :aria-current="view === 'register' ? 'page' : undefined" @click.prevent="go('register')">Register</a>
        </nav>

        <p class="intro">Use your username and password to access Trames.</p>
        <form @submit.prevent="submit(view === 'register' ? 'register' : 'login')">
          <label>
            Username
            <input v-model="username" autocomplete="username" required />
          </label>
          <label>
            Password
            <input
              v-model="password"
              type="password"
              :autocomplete="view === 'register' ? 'new-password' : 'current-password'"
              required
            />
          </label>
          <button type="submit">{{ view === 'register' ? 'Register' : 'Login' }}</button>
        </form>
      </template>

      <pre>{{ message }}</pre>
    </section>
  </main>
</template>
