import { ref } from 'vue'
import { errorMessage } from '../api/http'
import type { MessageKind } from '../types'

export function useNotifier() {
  const message = ref('Checking session...')
  const messageKind = ref<MessageKind>('info')

  function setMessage(text: string, kind: MessageKind = 'info') {
    message.value = text
    messageKind.value = kind
  }

  function setError(error: unknown, fallback: string) {
    setMessage(errorMessage(error, fallback), 'error')
  }

  return { message, messageKind, setMessage, setError }
}

export type NotifierController = ReturnType<typeof useNotifier>
