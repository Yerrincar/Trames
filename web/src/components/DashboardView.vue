<script setup lang="ts">
import type { OperationsController } from '../composables/useOperations'

const props = defineProps<{ ops: OperationsController }>()

const {
  selectedProject,
  selectedEntity,
  selectedSubProject,
  selectedProjectId,
  editingDescription,
  editingStatus,
  descriptionDraft,
  activeDropStatus,
  dashboardTaskFilter,
  dashboardTitle,
  dashboardDescription,
  dashboardTaskCount,
  dashboardSubProjectCount,
  dashboardGroups,
  projectDashboardFilterOptions,
  editableDescription,
  editableStatus,
  editableStatusOptions,
  startDescriptionEdit,
  cancelDescriptionEdit,
  eventValue,
  saveDescriptionEdit,
  changeStatus,
  selectTask,
  openTaskMenu,
  allowStatusDrop,
  leaveStatusDrop,
  dropTaskOnStatus,
  startTaskDrag,
  finishTaskDrag,
} = props.ops
</script>

<template>
  <section v-if="selectedProject && selectedEntity" class="project-focus">
    <p class="overline">{{ selectedEntity.type === 'subProject' ? 'Sub-project dashboard' : 'Project dashboard' }}</p>
    <h1>{{ dashboardTitle }}</h1>
    <div v-if="editingDescription" class="description-editor">
      <textarea v-model="descriptionDraft" rows="6" @keydown.ctrl.enter.prevent="saveDescriptionEdit" />
      <div class="inline-actions">
        <button type="button" @click="saveDescriptionEdit">Save description</button>
        <button type="button" @click="cancelDescriptionEdit">Cancel</button>
      </div>
    </div>
    <p v-else class="project-description editable-description" title="Double-click to edit" @dblclick="startDescriptionEdit">
      {{ dashboardDescription || editableDescription }}
    </p>

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
        <select v-if="editingStatus" class="inline-select" :value="editableStatus" @change="changeStatus(eventValue($event))" @blur="editingStatus = false">
          <option v-for="status in editableStatusOptions" :key="status" :value="status">{{ status }}</option>
        </select>
        <button v-else type="button" class="meta-edit-button" @click="editingStatus = true">{{ editableStatus || selectedSubProject?.status }}</button>
        <span>Status</span>
      </div>
    </div>

    <div class="dashboard-toolbar">
      <span>Status</span>
      <label v-if="selectedEntity.type === 'project'">
        Filter
        <select v-model="dashboardTaskFilter">
          <option v-for="option in projectDashboardFilterOptions" :key="option.value" :value="option.value">
            {{ option.label }}
          </option>
        </select>
      </label>
    </div>

    <section class="dashboard-grid" aria-label="Tasks by status">
      <article
        v-for="group in dashboardGroups"
        :key="group.status"
        class="status-column"
        :class="{ 'drop-active': activeDropStatus === group.status }"
        @dragover="allowStatusDrop($event, group.status)"
        @dragleave="leaveStatusDrop(group.status)"
        @drop.prevent="dropTaskOnStatus(group.status)"
      >
        <header>
          <span>{{ group.status }}</span>
          <span>{{ group.tasks.length }}</span>
        </header>
        <p v-if="group.tasks.length === 0" class="empty-center">No tasks.</p>
        <div
          v-for="task in group.tasks"
          :key="task.id"
          role="button"
          tabindex="0"
          class="dashboard-task"
          draggable="true"
          @click="selectedProjectId !== null && selectTask(selectedProjectId, task.id)"
          @keydown.enter="selectedProjectId !== null && selectTask(selectedProjectId, task.id)"
          @keydown.space.prevent="selectedProjectId !== null && selectTask(selectedProjectId, task.id)"
          @contextmenu="selectedProjectId !== null && openTaskMenu($event, selectedProjectId, task)"
          @dragstart.stop="startTaskDrag($event, task)"
          @dragend="finishTaskDrag"
        >
          <span>{{ task.name }}</span>
          <small>{{ task.subProject || 'Project' }}</small>
          <strong>{{ task.priority }}</strong>
        </div>
      </article>
    </section>
  </section>
</template>
