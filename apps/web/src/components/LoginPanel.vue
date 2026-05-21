<script setup lang="ts">
import { computed, ref } from 'vue'
import { Activity, LogIn } from '@lucide/vue'
import UiButton from './UiButton.vue'

defineProps<{
  error?: string
  busy?: boolean
}>()

const emit = defineEmits<{
  login: [payload: { mode: 'local' | 'ldap'; bankId: string; username: string; password: string }]
}>()

const mode = ref<'local' | 'ldap'>('local')
const bankId = ref('bank-demo')
const username = ref('admin')
const password = ref('admin123')

const title = computed(() => (mode.value === 'ldap' ? 'LDAP / AD login' : 'Local login'))

async function submit() {
  emit('login', { mode: mode.value, bankId: bankId.value, username: username.value, password: password.value })
}
</script>

<template>
  <main class="login-shell">
    <section class="login-card">
      <span class="brand__mark" aria-hidden="true">
        <Activity :size="18" />
      </span>
      <div>
        <p class="eyebrow">imsi-rails access</p>
        <h1>{{ title }}</h1>
        <p>Use local credentials for product administrators or LDAP/AD for bank operators.</p>
      </div>
      <div class="segmented-group" aria-label="Login mode">
        <button type="button" :class="{ 'is-selected': mode === 'local' }" @click="mode = 'local'">
          <strong>Local</strong>
          <small>Password</small>
        </button>
        <button type="button" :class="{ 'is-selected': mode === 'ldap' }" @click="mode = 'ldap'">
          <strong>LDAP</strong>
          <small>AD group map</small>
        </button>
      </div>
      <form class="login-form" @submit.prevent="submit">
        <label>
          <span>Bank ID</span>
          <input v-model="bankId" autocomplete="organization" required />
        </label>
        <label>
          <span>Username</span>
          <input v-model="username" autocomplete="username" required />
        </label>
        <label>
          <span>Password</span>
          <input v-model="password" type="password" autocomplete="current-password" required />
        </label>
        <p v-if="error" class="form-error">{{ error }}</p>
        <UiButton type="submit" :disabled="busy">
          <LogIn :size="15" aria-hidden="true" />
          Sign in
        </UiButton>
      </form>
    </section>
  </main>
</template>
