<script setup lang="ts">
import { priorityOptions } from '../constants'
import type { OperationsController } from '../composables/useOperations'

const props = defineProps<{ ops: OperationsController }>()

const {
  selectedTask,
  selectedTaskProject,
  editingDescription,
  editingStatus,
  editingPriority,
  descriptionDraft,
  editableDescription,
  editableStatus,
  editablePriority,
  editableStatusOptions,
  startDescriptionEdit,
  cancelDescriptionEdit,
  eventValue,
  saveDescriptionEdit,
  changeStatus,
  changePriority,
} = props.ops
</script>

<template>
  <section v-if="selectedTask && selectedTaskProject" class="project-focus task-detail">
    <p class="overline">Task detail</p>
    <h1>{{ selectedTask.name }}</h1>
    <div v-if="editingDescription" class="description-editor">
      <textarea v-model="descriptionDraft" rows="6" @keydown.ctrl.enter.prevent="saveDescriptionEdit" />
      <div class="inline-actions">
        <button type="button" @click="saveDescriptionEdit">Save description</button>
        <button type="button" @click="cancelDescriptionEdit">Cancel</button>
      </div>
    </div>
    <p v-else class="project-description editable-description" title="Double-click to edit" @dblclick="startDescriptionEdit">
      {{ editableDescription }}
    </p>

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
        <select v-if="editingStatus" class="inline-select" :value="editableStatus" @change="changeStatus(eventValue($event))" @blur="editingStatus = false">
          <option v-for="status in editableStatusOptions" :key="status" :value="status">{{ status }}</option>
        </select>
        <button v-else type="button" class="meta-edit-button" @click="editingStatus = true">{{ editableStatus }}</button>
      </div>
      <div>
        <span>Priority</span>
        <select v-if="editingPriority" class="inline-select" :value="editablePriority" @change="changePriority(eventValue($event))" @blur="editingPriority = false">
          <option v-for="priority in priorityOptions" :key="priority" :value="priority">{{ priority }}</option>
        </select>
        <button v-else type="button" class="meta-edit-button" @click="editingPriority = true">{{ editablePriority }}</button>
      </div>
      <div>
        <span>ID</span>
        <strong>#{{ selectedTask.id }}</strong>
      </div>
    </div>
  </section>
</template>
