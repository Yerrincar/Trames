<script setup lang="ts">
import type { AuthController } from '../composables/useAuth'
import type { NotifierController } from '../composables/useNotifier'

const props = defineProps<{
  auth: AuthController
  notifier: NotifierController
}>()

const emit = defineEmits<{
  submit: [action: 'login' | 'register']
}>()

const { view, username, password, go } = props.auth
const { message, messageKind } = props.notifier
</script>

<template>
  <section v-if="view === 'home'" class="landing-screen">
    <p class="overline">Task manager</p>
    <h1>Trames</h1>
    <nav class="landing-actions" aria-label="Authentication">
      <a href="/login" @click.prevent="go('login')">Login</a>
      <a href="/register" @click.prevent="go('register')">Register</a>
    </nav>
  </section>

  <section v-else class="auth-screen">
    <form class="auth-box" @submit.prevent="emit('submit', view === 'register' ? 'register' : 'login')">
      <p class="overline">Trames</p>
      <h1>{{ view === 'register' ? 'Register' : 'Login' }}</h1>
      <label>
        Username
        <input v-model="username" autocomplete="username" required />
      </label>
      <label>
        Password
        <input v-model="password" type="password" :autocomplete="view === 'register' ? 'new-password' : 'current-password'" required />
      </label>
      <button type="submit" class="submit-button">{{ view === 'register' ? 'Register' : 'Login' }}</button>
      <nav class="auth-links" aria-label="Authentication links">
        <a href="/" @click.prevent="go('home')">Home</a>
        <a v-if="view === 'login'" href="/register" @click.prevent="go('register')">Register</a>
        <a v-else href="/login" @click.prevent="go('login')">Login</a>
      </nav>
      <p class="auth-message" :class="{ error: messageKind === 'error' }">{{ message }}</p>
    </form>
  </section>
</template>
