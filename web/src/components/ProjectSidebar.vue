<script setup lang="ts">
import type { OperationsController } from '../composables/useOperations'
import type { User } from '../types'

const props = defineProps<{
  user: User
  ops: OperationsController
}>()

const emit = defineEmits<{
  logout: []
}>()

const {
  projects,
  loadingOperations,
  selectedProjectId,
  selectedEntity,
  hasProjects,
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
  openModal,
  openProjectMenu,
  openSubProjectMenu,
  openTaskMenu,
} = props.ops
</script>

<template>
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
      <button type="button" class="logout-button" @click="emit('logout')">Logout</button>
    </footer>
  </aside>
</template>
