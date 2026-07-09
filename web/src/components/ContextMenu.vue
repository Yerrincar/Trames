<script setup lang="ts">
import type { OperationsController } from '../composables/useOperations'

const props = defineProps<{ ops: OperationsController }>()

const { contextMenu, contextCreateSubProject, contextCreateTask, contextDelete } = props.ops
</script>

<template>
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
</template>
