<script setup lang="ts">
import type { OperationsController } from '../composables/useOperations'
import type { NotifierController } from '../composables/useNotifier'
import DashboardView from './DashboardView.vue'
import TaskDetailView from './TaskDetailView.vue'

const props = defineProps<{
  ops: OperationsController
  notifier: NotifierController
}>()

const { message, messageKind } = props.notifier
const { selectedTask, selectedTaskProject, selectedProject, selectedEntity } = props.ops
</script>

<template>
  <section class="center-panel" aria-label="Workspace">
    <aside v-if="messageKind === 'error'" class="error-banner" role="alert">
      {{ message }}
    </aside>

    <TaskDetailView v-if="selectedTask && selectedTaskProject" :ops="ops" />
    <DashboardView v-else-if="selectedProject && selectedEntity" :ops="ops" />

    <section v-else class="home-focus" aria-label="Home">
      <p class="overline">Personal task manager</p>
      <h1>Trames</h1>
      <p>Select a project on the left or create one from the toolbar.</p>
    </section>
  </section>
</template>
