<template>
  <errors v-if="error" :errorCode="error.status" />
  <div class="row" v-else-if="!layoutStore.loading">
    <div class="column">
      <div class="card">
        <div class="card-title">
          <h2>WebDAV Token Management</h2>
        </div>

        <div class="card-action">
          <button
            class="button button--flat"
            @click="showCreateDialog = true"
            :aria-label="t('buttons.new')"
            :title="t('buttons.new')"
          >
            <i class="material-icons">add</i>
            <span>Create Token</span>
          </button>
        </div>

        <div class="card-content full" v-if="tokens.length > 0">
          <table>
            <tr>
              <th>Name</th>
              <th>Path</th>
              <th>Permissions</th>
              <th>Status</th>
              <th>Created At</th>
              <th></th>
            </tr>

            <tr v-for="token in tokens" :key="token.id">
              <td>{{ token.name }}</td>
              <td>{{ token.path }}</td>
              <td>
                <span v-if="token.canRead" class="permission-badge">Read</span>
                <span v-if="token.canWrite" class="permission-badge">Write</span>
                <span v-if="token.canDelete" class="permission-badge">Delete</span>
              </td>
              <td>
                <span
                  :class="{
                    'status-active': token.status === 'active',
                    'status-suspended': token.status === 'suspended',
                  }"
                >
                  {{ token.status === "active" ? "Active" : "Suspended" }}
                </span>
              </td>
              <td>{{ formatDate(token.createdAt) }}</td>
              <td class="small">
                <button
                  class="action"
                  @click="viewToken(token)"
                  :aria-label="t('buttons.info')"
                  :title="t('buttons.info')"
                >
                  <i class="material-icons">info</i>
                </button>
                <button
                  class="action"
                  @click="editToken(token)"
                  :aria-label="t('buttons.edit')"
                  :title="t('buttons.edit')"
                >
                  <i class="material-icons">edit</i>
                </button>
                <button
                  v-if="token.status === 'active'"
                  class="action"
                  @click="suspendToken(token)"
                  aria-label="Suspend"
                  title="Suspend"
                >
                  <i class="material-icons">pause</i>
                </button>
                <button
                  v-else
                  class="action"
                  @click="activateToken(token)"
                  aria-label="Activate"
                  title="Activate"
                >
                  <i class="material-icons">play_arrow</i>
                </button>
                <button
                  class="action"
                  @click="deleteToken(token)"
                  :aria-label="t('buttons.delete')"
                  :title="t('buttons.delete')"
                >
                  <i class="material-icons">delete</i>
                </button>
              </td>
            </tr>
          </table>
        </div>
        <h2 class="message" v-else>
          <i class="material-icons">sentiment_dissatisfied</i>
          <span>{{ t("files.lonely") }}</span>
        </h2>
      </div>
    </div>
  </div>

  <!-- Create/Edit dialog -->
  <div class="overlay" v-if="showCreateDialog || showEditDialog" @click="closeDialog"></div>
  <div class="dialog" v-if="showCreateDialog || showEditDialog">
    <div class="card">
        <div class="card-title">
        <h2>{{ showCreateDialog ? "Create WebDAV Token" : "Edit WebDAV Token" }}</h2>
      </div>
      <div class="card-content">
        <form @submit.prevent="submitForm">
          <div class="input-group">
            <label for="token-name">Name *</label>
            <input
              id="token-name"
              v-model="formData.name"
              type="text"
              required
              placeholder="Enter token name"
            />
          </div>

          <div class="input-group">
            <label for="token-path">Path</label>
            <input
              id="token-path"
              v-model="formData.path"
              type="text"
              placeholder="/ (root)"
            />
          </div>

          <div class="input-group">
            <label>Permissions</label>
            <div class="checkbox-group">
              <label>
                <input type="checkbox" v-model="formData.canRead" />
                Read
              </label>
              <label>
                <input type="checkbox" v-model="formData.canWrite" />
                Write
              </label>
              <label>
                <input type="checkbox" v-model="formData.canDelete" />
                Delete
              </label>
            </div>
          </div>

          <div class="card-action">
            <button type="button" class="button button--flat" @click="closeDialog">
              {{ t("buttons.cancel") }}
            </button>
            <button type="submit" class="button button--flat button--blue">
              {{ showCreateDialog ? "Create" : "Save" }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>

  <!-- View Token dialog -->
  <div class="overlay" v-if="showViewDialog" @click="showViewDialog = false"></div>
  <div class="dialog" v-if="showViewDialog && viewingToken">
    <div class="card">
      <div class="card-title">
        <h2>Token Details</h2>
      </div>
      <div class="card-content">
        <div class="token-details">
          <div class="detail-item">
            <strong>Name:</strong>
            <span>{{ viewingToken.name }}</span>
          </div>
          <div class="detail-item">
            <strong>Username:</strong>
            <span>{{ authStore.user?.username }}</span>
          </div>
          <div class="detail-item">
            <strong>Path:</strong>
            <span>{{ viewingToken.path }}</span>
          </div>
          <div class="detail-item">
            <strong>Token:</strong>
            <div class="token-display">
              <input
                type="text"
                readonly
                :value="viewingToken.token"
                ref="tokenInput"
              />
              <button class="button button--flat" @click="copyToken">
                <i class="material-icons">content_copy</i>
              </button>
            </div>
          </div>
          <div class="detail-item">
            <strong>WebDAV URL:</strong>
            <div class="token-display">
              <input
                type="text"
                readonly
                :value="getWebDAVUrl()"
              />
              <button class="button button--flat" @click="copyWebDAVUrl">
                <i class="material-icons">content_copy</i>
              </button>
            </div>
          </div>
          <div class="detail-item">
            <strong>Permissions:</strong>
            <span>
              <span v-if="viewingToken.canRead" class="permission-badge">Read</span>
              <span v-if="viewingToken.canWrite" class="permission-badge">Write</span>
              <span v-if="viewingToken.canDelete" class="permission-badge">Delete</span>
            </span>
          </div>
          <div class="detail-item">
            <strong>Status:</strong>
            <span
              :class="{
                'status-active': viewingToken.status === 'active',
                'status-suspended': viewingToken.status === 'suspended',
              }"
            >
              {{ viewingToken.status === "active" ? "Active" : "Suspended" }}
            </span>
          </div>
          <div class="detail-item">
            <strong>Created At:</strong>
            <span>{{ formatDate(viewingToken.createdAt) }}</span>
          </div>
        </div>

        <div class="card-action">
          <button
            class="button button--flat"
            @click="showViewDialog = false"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";
import { useAuthStore } from "@/stores/auth";
import { webdav as api } from "@/api";
import type { WebDAVToken } from "@/api/webdav";
import Errors from "@/views/Errors.vue";
import { inject, ref, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";
import { copy } from "@/utils/clipboard";
import dayjs from "dayjs";

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;
const { t } = useI18n();

const layoutStore = useLayoutStore();
const authStore = useAuthStore();

const error = ref<StatusError | null>(null);
const tokens = ref<WebDAVToken[]>([]);
const showCreateDialog = ref(false);
const showEditDialog = ref(false);
const showViewDialog = ref(false);
const viewingToken = ref<WebDAVToken | null>(null);
const editingToken = ref<WebDAVToken | null>(null);
const tokenInput = ref<HTMLInputElement | null>(null);

const formData = ref({
  name: "",
  path: "/",
  canRead: true,
  canWrite: true,
  canDelete: false,
});

onMounted(async () => {
  await loadTokens();
});

async function loadTokens() {
  layoutStore.loading = true;
  try {
    tokens.value = await api.listTokens();
  } catch (err) {
    error.value = err as StatusError;
    $showError(err as Error);
  } finally {
    layoutStore.loading = false;
  }
}

function viewToken(token: WebDAVToken) {
  api
    .getToken(token.id)
    .then((fullToken) => {
      viewingToken.value = fullToken;
      showViewDialog.value = true;
    })
    .catch((err) => {
      $showError(err as Error);
    });
}

function editToken(token: WebDAVToken) {
  editingToken.value = token;
  formData.value = {
    name: token.name,
    path: token.path,
    canRead: token.canRead,
    canWrite: token.canWrite,
    canDelete: token.canDelete,
  };
  showEditDialog.value = true;
}

async function submitForm() {
  try {
    if (showCreateDialog.value) {
      const newToken = await api.createToken(formData.value);
      viewingToken.value = newToken;
      showCreateDialog.value = false;
      showViewDialog.value = true;
      $showSuccess("Token created successfully! Please save this token; it will only be shown once.");
    } else if (showEditDialog.value && editingToken.value) {
      await api.updateToken(editingToken.value.id, formData.value);
      showEditDialog.value = false;
      $showSuccess("Token updated successfully!");
    }
    await loadTokens();
  } catch (err) {
    $showError(err as Error);
  }
}

function closeDialog() {
  showCreateDialog.value = false;
  showEditDialog.value = false;
  editingToken.value = null;
  formData.value = {
    name: "",
    path: "/",
    canRead: true,
    canWrite: true,
    canDelete: false,
  };
}

async function deleteToken(token: WebDAVToken) {
  if (!confirm(`Are you sure you want to delete token "${token.name}"?`)) {
    return;
  }

  try {
    await api.deleteToken(token.id);
    $showSuccess("Token deleted");
    await loadTokens();
  } catch (err) {
    $showError(err as Error);
  }
}

async function suspendToken(token: WebDAVToken) {
  try {
    await api.suspendToken(token.id);
    $showSuccess("Token suspended");
    await loadTokens();
  } catch (err) {
    $showError(err as Error);
  }
}

async function activateToken(token: WebDAVToken) {
  try {
    await api.activateToken(token.id);
    $showSuccess("Token activated");
    await loadTokens();
  } catch (err) {
    $showError(err as Error);
  }
}

function formatDate(dateStr: string): string {
  return dayjs(dateStr).format("YYYY-MM-DD HH:mm:ss");
}

function getWebDAVUrl(): string {
  const base = window.NulYun?.BaseURL || "";
  return `${window.location.origin}${base}/dav/`;
}

function copyToken() {
  if (viewingToken.value) {
    copy({ text: viewingToken.value.token });
    $showSuccess("Token copied to clipboard");
  }
}

function copyWebDAVUrl() {
  copy({ text: getWebDAVUrl() });
  $showSuccess("WebDAV URL copied to clipboard");
}
</script>

<style scoped>
.permission-badge {
  display: inline-block;
  padding: 2px 8px;
  margin: 0 2px;
  background-color: #2196f3;
  color: white;
  border-radius: 3px;
  font-size: 12px;
}

.status-active {
  color: #4caf50;
  font-weight: bold;
}

.status-suspended {
  color: #f44336;
  font-weight: bold;
}

.dialog {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 1000;
  max-width: 600px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
}

.overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 999;
}

.input-group {
  margin-bottom: 20px;
}

.input-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}

.input-group input[type="text"] {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.checkbox-group {
  display: flex;
  gap: 15px;
}

.checkbox-group label {
  display: flex;
  align-items: center;
  gap: 5px;
  font-weight: normal;
}

.token-details {
  margin-bottom: 20px;
}

.detail-item {
  margin-bottom: 15px;
}

.detail-item strong {
  display: block;
  margin-bottom: 5px;
}

.token-display {
  display: flex;
  gap: 10px;
  align-items: center;
}

.token-display input {
  flex: 1;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: monospace;
}

.token-display button {
  flex-shrink: 0;
}
</style>
