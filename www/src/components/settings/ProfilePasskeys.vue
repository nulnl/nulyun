<template>
  <div class="column">
    <form class="card">
      <div class="card-title">
        <h2>{{ t('passkey.name') }}</h2>
      </div>

      <div class="card-content">
        <p>{{ t('passkey.registerInstructions') }}</p>
        
        <div class="card-action">
          <button type="button" class="button button--flat" @click="registerPasskey" :disabled="!isLoggedIn">
            <i class="material-icons">add</i>
            <span>{{ t('passkey.add') }}</span>
          </button>
        </div>

        <div v-if="passkeys.length > 0" class="passkey-list">
          <h3>{{ t('passkey.list') }}</h3>
          <table>
            <thead>
              <tr>
                <th>{{ t('passkey.credentialName') }}</th>
                <th>{{ t('passkey.createdAt') }}</th>
                <th>{{ t('passkey.lastUsedAt') }}</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="passkey in passkeys" :key="passkey.id">
                <td>{{ passkey.name }}</td>
                <td>{{ formatDate(passkey.createdAt) }}</td>
                <td>{{ formatDate(passkey.lastUsedAt) }}</td>
                <td>
                  <button
                    class="action"
                    @click="deletePasskey(passkey.id)"
                    :aria-label="t('buttons.delete')"
                    :title="t('buttons.delete')"
                  >
                    <i class="material-icons">delete</i>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <p v-else class="message">{{ t('passkey.noPasskeys') }}</p>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { useI18n } from "vue-i18n";
import { inject } from "vue";
import * as api from "@/api/passkey";
import type { PasskeyCredential } from "@/api/passkey";
import { startRegistration } from "@simplewebauthn/browser";
import dayjs from "dayjs";
import { useAuthStore } from "@/stores/auth";

const { t } = useI18n();
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;
const $showError = inject<IToastError>("$showError")!;

const passkeys = ref<PasskeyCredential[]>([]);
const authStore = useAuthStore();
const isLoggedIn = computed(() => !!authStore.jwt);

onMounted(async () => {
  await loadPasskeys();
});

async function loadPasskeys() {
  try {
    passkeys.value = await api.listPasskeys();
  } catch (err: any) {
    $showError(err);
  }
}

async function registerPasskey() {
  if (!isLoggedIn.value) {
    $showError(t('passkey.loginRequired'));
    return;
  }

  try {
    // Get name from user
    const name = prompt(t('passkey.enterName'), t('passkey.defaultName'));
    if (!name) return;

    // Begin registration
    const options = await api.beginRegistration();
    
    // Use WebAuthn API to create credential
    const credential = await startRegistration(options.publicKey || options);
    
    // Finish registration
    await api.finishRegistration(credential, name);
    
    $showSuccess(t('passkey.registerSuccess'));
    await loadPasskeys();
  } catch (err: any) {
    console.error("Passkey registration error:", err);
    $showError(err.message || t('passkey.registerFailed'));
  }
}

async function deletePasskey(id: number) {
  if (!confirm(t('passkey.confirmDelete'))) {
    return;
  }

  try {
    await api.deletePasskey(id);
    $showSuccess(t('passkey.deleteSuccess'));
    await loadPasskeys();
  } catch (err: any) {
    $showError(err);
  }
}

function formatDate(dateStr: string): string {
  return dayjs(dateStr).format("YYYY-MM-DD HH:mm");
}
</script>

<style scoped>
.passkey-list {
  margin-top: 2em;
}

.passkey-list h3 {
  margin-bottom: 1em;
}

.passkey-list table {
  width: 100%;
}

.message {
  text-align: center;
  padding: 2em;
  color: #888;
}
</style>
