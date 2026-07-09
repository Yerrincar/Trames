import { ref } from 'vue'
import type { NotifierController } from './useNotifier'
import type { User, View } from '../types'

export function viewFromPath(): View {
  switch (window.location.pathname) {
    case '/register':
      return 'register'
    case '/login':
      return 'login'
    default:
      return 'home'
  }
}

export function useAuth(notifier: NotifierController) {
  const user = ref<User | null>(null)
  const username = ref('')
  const password = ref('')
  const view = ref<View>(viewFromPath())

  function go(nextView: View) {
    const path = nextView === 'home' ? '/' : `/${nextView}`
    window.history.pushState({}, '', path)
    view.value = nextView
  }

  function handlePopState() {
    view.value = viewFromPath()
  }

  async function refreshCurrentUser() {
    const response = await fetch('/users/currentUser')
    if (!response.ok) {
      user.value = null
      notifier.setMessage(view.value === 'home' ? 'Ready.' : 'Not logged in.')
      return null
    }

    user.value = (await response.json()) as User
    notifier.setMessage(`Logged in as ${user.value.username}`)
    go('home')
    return user.value
  }

  async function submit(action: 'login' | 'register') {
    const response = await fetch(`/users/${action}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: username.value, password: password.value }),
    })

    if (!response.ok) {
      notifier.setMessage(`${response.status} ${response.statusText}\n${await response.text()}`, 'error')
      return false
    }

    if (action === 'login') {
      user.value = (await response.json()) as User
      notifier.setMessage(`Logged in as ${user.value.username}`)
      go('home')
      return true
    }

    notifier.setMessage('User registered. You can log in now.')
    password.value = ''
    go('login')
    return false
  }

  async function logout() {
    const response = await fetch('/users/logout', { method: 'POST' })
    user.value = null
    notifier.setMessage(response.ok ? 'Logged out.' : `${response.status} ${response.statusText}\n${await response.text()}`, response.ok ? 'info' : 'error')
    username.value = ''
    password.value = ''
    go('home')
  }

  return { user, username, password, view, go, handlePopState, refreshCurrentUser, submit, logout }
}

export type AuthController = ReturnType<typeof useAuth>
