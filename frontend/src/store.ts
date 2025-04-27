import { reactive, readonly } from 'vue'
import type { AuthState } from './types/auth_state'

const state = reactive<AuthState>({
  accessToken: null,
})

function setUser(accessToken: string) {
  state.accessToken = accessToken
}

function clearUser() {
  state.accessToken = null
}

export default {
  state: readonly(state),
  setUser,
  clearUser,
}
