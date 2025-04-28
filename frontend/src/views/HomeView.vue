<script setup lang="ts">
import Toast from 'primevue/toast'
import LoginUser from '../components/LoginUser.vue'
import FlightsInfo from '../components/FlightsInfo.vue'
import store from '../store'
import { useToast } from 'primevue/usetoast'

const toast = useToast()

const onLogout = () => {
  store.ClearUser()
  toast.add({
    severity: 'success',
    summary: 'Token invalid or expired!',
    detail: 'You were successfully logged out.',
  })
}
</script>

<template>
  <main>
    <Toast />
    <h1 class="green">Welcome to Flight Finder</h1>
    <!-- <LoginUser /> -->

    <FlightsInfo
      @logout="onLogout"
      :token="store.state.accessToken"
      v-if="store.state.accessToken"
    ></FlightsInfo>
    <LoginUser v-else></LoginUser>
  </main>
</template>

<style scoped>
h1 {
  font-weight: 500;
  font-size: 2.6rem;
}
</style>
