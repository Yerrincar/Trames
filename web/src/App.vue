<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'

type View = 'home' | 'login' | 'register' | 'create-task'

type User = {
  id: number
  username: string
  created_at: string
  updated_at: string
}

const user = ref<User | null>(null)
const username = ref('')
const password = ref('')
const message = ref('Checking session...')
const view = ref<View>(viewFromPath())

const title = computed(() => {
  if (user.value) {
    return 'Trames'
  }
  return view.value === 'register' ? 'Register' : 'Login'
})

function viewFromPath(): View {
  switch (window.location.pathname) {
    case '/register':
      return 'register'
    case '/create-task':
      return 'create-task'
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

async function refreshCurrentUser() {
  const response = await fetch('/users/currentUser')
  if (!response.ok) {
    user.value = null
    message.value = 'Not logged in.'
    if (view.value === 'home' || view.value === 'create-task') {
      view.value = 'login'
    }
    return
  }
  user.value = (await response.json()) as User
  message.value = `Logged in as ${user.value.username}`
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
    return
  }

  message.value = 'User registered. You can log in now.'
  password.value = ''
  go('login')
}

async function logout() {
  const response = await fetch('/users/logout', { method: 'POST' })
  user.value = null
  message.value = response.ok ? 'Logged out.' : `${response.status} ${response.statusText}\n${await response.text()}`
  username.value = ''
  password.value = ''
  go('login')
}

function handlePopState() {
  view.value = viewFromPath()
}

onMounted(() => {
  window.addEventListener('popstate', handlePopState)
  refreshCurrentUser().catch((error) => {
    user.value = null
    message.value = error instanceof Error ? error.message : 'Unable to check current user.'
  })
})

onUnmounted(() => window.removeEventListener('popstate', handlePopState))
</script>

<template>
  <main>
    <section class="panel">
      <p class="eyebrow">Task manager</p>
      <h1>{{ title }}</h1>

      <div v-if="user" class="logged-in">
        <p class="intro">Logged in as {{ user.username }}</p>
        <div class="actions">
          <button type="button" @click="logout">Logout</button>
          <button type="button" class="secondary" @click="go('create-task')">CreateTask</button>
        </div>
        <p v-if="view === 'create-task'" class="notice">CreateTask selected.</p>
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
            <input v-model="password" type="password" autocomplete="current-password" required />
          </label>
          <button type="submit">{{ view === 'register' ? 'Register' : 'Login' }}</button>
        </form>
      </template>

      <pre>{{ message }}</pre>
    </section>
  </main>
</template>
