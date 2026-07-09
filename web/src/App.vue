<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import AuthView from './components/AuthView.vue'
import ContextMenu from './components/ContextMenu.vue'
import OperationModal from './components/OperationModal.vue'
import ProjectSidebar from './components/ProjectSidebar.vue'
import WorkspacePanel from './components/WorkspacePanel.vue'
import { useAuth } from './composables/useAuth'
import { useNotifier } from './composables/useNotifier'
import { useOperations } from './composables/useOperations'

const notifier = useNotifier()
const auth = useAuth(notifier)
const ops = useOperations(auth.user, notifier)

async function refreshSession() {
  try {
    const currentUser = await auth.refreshCurrentUser()
    if (currentUser) {
      await ops.loadOperations()
    } else {
      ops.clearOperations()
    }
  } catch (error) {
    auth.user.value = null
    ops.clearOperations()
    notifier.setError(error, 'Unable to check current user.')
  }
}

async function submitAuth(action: 'login' | 'register') {
  const loggedIn = await auth.submit(action)
  if (loggedIn) {
    await ops.loadOperations()
  }
}

async function logout() {
  await auth.logout()
  ops.clearOperations()
  ops.closeModal()
  ops.closeContextMenu()
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    ops.closeModal()
    ops.closeContextMenu()
  }
}

onMounted(() => {
  window.addEventListener('popstate', auth.handlePopState)
  window.addEventListener('click', ops.closeContextMenu)
  window.addEventListener('keydown', handleKeydown)
  refreshSession()
})

onUnmounted(() => {
  window.removeEventListener('popstate', auth.handlePopState)
  window.removeEventListener('click', ops.closeContextMenu)
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <main class="app-root" :class="{ authenticated: auth.user.value }">
    <template v-if="auth.user.value">
      <section class="workspace-layout">
        <ProjectSidebar :user="auth.user.value" :ops="ops" @logout="logout" />
        <WorkspacePanel :ops="ops" :notifier="notifier" />
      </section>

      <ContextMenu :ops="ops" />
      <OperationModal :ops="ops" />
    </template>

    <AuthView v-else :auth="auth" :notifier="notifier" @submit="submitAuth" />
  </main>
</template>
