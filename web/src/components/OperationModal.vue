<script setup lang="ts">
import { priorityOptions, projectStatusOptions, taskStatusOptions } from '../constants'
import type { OperationsController } from '../composables/useOperations'

const props = defineProps<{ ops: OperationsController }>()

const {
  modal,
  modalTitle,
  projects,
  projectForm,
  subProjectForm,
  taskForm,
  taskSubProjects,
  closeModal,
  createProject,
  createSubProject,
  createTask,
} = props.ops
</script>

<template>
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
