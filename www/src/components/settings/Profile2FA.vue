<template>
  <div class="column">
    <!-- TOTP is verified and enabled -->
    <form v-if="authStore?.user?.otpEnabled" class="card">
      <div class="card-title">
        <h2>{{ t("otp.name") }}</h2>
      </div>

      <div v-if="otpSetupKey" class="card-content">
        <div class="qrcode-container">
          <qrcode-vue :value="otpSetupKey" :size="300" level="M" />
        </div>
        <div class="setup-key-container">
          <input
            :value="otpSecretB32"
            class="input input--block"
            type="text"
            name="otpSetupKey"
            disabled
          />
          <button class="action copy-clipboard" @click="copyOtpSetupKey">
            <i class="material-icons">content_paste_go</i>
          </button>
        </div>
        <div class="setup-key-container">
          <input
            v-model="otpCode"
            :placeholder="t('settings.otpCodeCheckPlaceholder')"
            type="text"
            pattern="[0-9]*"
            inputmode="numeric"
            maxlength="6"
            class="input input--block"
          />
          <button class="action copy-clipboard" @click="checkOtpCode">
            <i class="material-icons">send</i>
          </button>
        </div>
        <button class="button button--block button--red" @click="disableOtp">
          {{ t("buttons.disable") }}
        </button>
      </div>

      <div v-if="!otpSetupKey && !showRecoveryCodes" class="card-action">
        <button class="button button--flat" @click="showOtpInfo">
          {{ t("prompts.show") }}
        </button>
        <button class="button button--flat" @click="resetOtpKey">
          Reset TOTP
        </button>
        <button class="button button--flat" @click="generateRecovery">
          Generate Recovery Codes
        </button>
      </div>

      <!-- Recovery Codes Display -->
      <div v-if="showRecoveryCodes" class="card-content">
        <h3>Recovery Codes</h3>
        <p class="message warning">
          Save these recovery codes in a safe place. Each code can only be used once.
        </p>
        <div class="recovery-codes">
          <div v-for="code in recoveryCodes" :key="code" class="recovery-code">
            {{ code }}
          </div>
        </div>
        <div class="card-action">
          <button class="button button--flat" @click="copyRecoveryCodes">
            Copy All
          </button>
          <button class="button button--flat" @click="closeRecoveryCodes">
            Close
          </button>
        </div>
      </div>
    </form>

    <!-- TOTP is setup but pending verification -->
    <form v-else-if="authStore?.user?.otpPending" class="card">
      <div class="card-title">
        <h2>{{ t("otp.name") }} - Pending Verification</h2>
      </div>

      <div class="card-content">
        <p class="message" v-if="!otpSetupKey">
          TOTP has been setup but not yet verified. Please complete the verification to enable two-factor authentication.
        </p>

        <div v-if="otpSetupKey" class="card-content">
          <div class="qrcode-container">
            <qrcode-vue :value="otpSetupKey" :size="300" level="M" />
          </div>
          <div class="setup-key-container">
            <input
              :value="otpSecretB32"
              class="input input--block"
              type="text"
              name="otpSetupKey"
              disabled
            />
            <button class="action copy-clipboard" @click="copyOtpSetupKey">
              <i class="material-icons">content_paste_go</i>
            </button>
          </div>
          <div class="setup-key-container">
            <input
              v-model="otpCode"
              :placeholder="t('settings.otpCodeCheckPlaceholder')"
              type="text"
              pattern="[0-9]*"
              inputmode="numeric"
              maxlength="6"
              class="input input--block"
            />
            <button class="action copy-clipboard" @click="checkOtpCode">
              <i class="material-icons">send</i>
            </button>
          </div>
        </div>

        <div class="card-action">
          <button v-if="!otpSetupKey" class="button button--flat" @click="showOtpInfo">
            View Setup Key
          </button>
          <button class="button button--flat button--red" @click="cancelOtpSetup">
            Cancel Setup
          </button>
        </div>
      </div>
    </form>

    <!-- TOTP is not enabled -->
    <form v-else class="card" @submit="enable2FA">
      <div class="card-title">
        <h2>{{ t("otp.name") }}</h2>
      </div>

      <div class="card-content">
        <input
          v-if="!otpSetupKey"
          v-model="passwordForOTP"
          :placeholder="t('settings.password')"
          class="input input--block"
          type="password"
          name="password"
        />
        <template v-else>
          <div class="qrcode-container">
            <qrcode-vue :value="otpSetupKey" :size="300" level="M" />
          </div>
          <div class="setup-key-container">
            <input
              :value="otpSecretB32"
              class="input input--block"
              type="text"
              name="otpSetupKey"
              disabled
            />
            <button class="action copy-clipboard" @click="copyOtpSetupKey">
              <i class="material-icons">content_paste_go</i>
            </button>
          </div>
          <div class="setup-key-container">
            <input
              v-model="otpCode"
              :placeholder="t('settings.otpCodeCheckPlaceholder')"
              type="text"
              pattern="[0-9]*"
              inputmode="numeric"
              maxlength="6"
              class="input input--block"
            />
            <button class="action copy-clipboard" @click="checkOtpCode">
              <i class="material-icons">send</i>
            </button>
          </div>
        </template>
      </div>

      <div class="card-action">
        <input
          v-if="!otpSetupKey"
          :value="t('buttons.enable')"
          class="button button--flat"
          type="submit"
          name="submitEnableOTPForm"
        />
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { base32 } from "@scure/base";
import QrcodeVue from "qrcode.vue";
import { copy } from "@/utils/clipboard";
import { useLayoutStore } from "@/stores/layout";
import { useAuthStore } from "@/stores/auth";
import { useI18n } from "vue-i18n";
import { users as api } from "@/api";
import { inject, ref } from "vue";
import { computed } from "vue";

const layoutStore = useLayoutStore();
const authStore = useAuthStore();
const { t } = useI18n();

const $showSuccess = inject<IToastSuccess>("$showSuccess")!;
const $showError = inject<IToastError>("$showError")!;

const passwordForOTP = ref<string>("");
const otpSetupKey = ref<string>("");
const otpCode = ref<string>("");
const showRecoveryCodes = ref<boolean>(false);
const recoveryCodes = ref<string[]>([]);
const resetPasswordInput = ref<string>("");

const otpSecretB32 = computed(() => {
  if (!otpSetupKey.value) return "";
  const otpURI = new URL(otpSetupKey.value);
  // The `secret` query param in the otpauth URL is already a Base32 string.
  // Return it directly so users get the original, compact Base32 secret.
  return String(otpURI.searchParams.get("secret") || "");
});

const showOtpInfo = async (event: Event) => {
  event.preventDefault();
  
  if (authStore.user === null) {
    return;
  }
  const userId = authStore.user.id;

  // If TOTP is already verified, require OTP code for security
  if (authStore.user.otpEnabled) {
    layoutStore.showHover({
      prompt: "otp",
      confirm: async (code: string) => {
        try {
          const res = await api.getOtpInfo(userId, code);
          otpSetupKey.value = res.setupKey;
        } catch (err: any) {
          $showError(err);
        }
      },
    });
  } else {
    // If not verified yet, show directly without OTP code
    try {
      const res = await api.getOtpInfo(userId, "");
      otpSetupKey.value = res.setupKey;
    } catch (err: any) {
      $showError(err);
    }
  }
};
const disableOtp = async (event: Event) => {
  event.preventDefault();

  layoutStore.showHover({
    prompt: "otp",
    confirm: async (code: string) => {
      if (authStore.user === null) {
        return;
      }

      try {
        await api.disableOtp(authStore.user.id, code);
        otpSetupKey.value = "";
        authStore.user.otpEnabled = false;
        authStore.user.otpPending = false;
      } catch (err: any) {
        $showError(err);
      }
    },
  });
};
const enable2FA = async (event: Event) => {
  event.preventDefault();
  if (authStore.user === null || otpSetupKey.value) {
    return;
  }

  try {
    const res = await api.enableOTP(authStore.user.id, passwordForOTP.value);

    otpSetupKey.value = res.setupKey;
    // Set pending state, not enabled yet
    authStore.user.otpPending = true;
    $showSuccess(t("otp.enabledSuccessfully"));
  } catch (err: any) {
    $showError(err);
  } finally {
    passwordForOTP.value = "";
  }
};
const copyToClipboard = async (text: string) => {
  try {
    await copy({ text });
    $showSuccess(t("success.linkCopied"));
  } catch {
    try {
      await copy({ text }, { permission: true });
      $showSuccess(t("success.linkCopied"));
    } catch (e: any) {
      $showError(e);
    }
  }
};
const copyOtpSetupKey = async (event: Event) => {
  event.preventDefault();
  // Copy the full otpauth URL so users can paste the URL or scan it elsewhere.
  // Fallback to the Base32 secret if the URL is not available.
  const toCopy = otpSetupKey.value || otpSecretB32.value;
  await copyToClipboard(toCopy);
};
const checkOtpCode = async (event: Event) => {
  event.preventDefault();
  if (authStore.user === null) {
    return;
  }

  try {
    await api.checkOtp(authStore.user.id, otpCode.value);
    // Only after successful verification, mark TOTP as enabled and clear pending
    authStore.user.otpEnabled = true;
    authStore.user.otpPending = false;
    $showSuccess(t("otp.verificationSucceed"));
    // Clear the setup key after successful verification
    otpSetupKey.value = "";
    otpCode.value = "";
  } catch (err: any) {
    console.log(err);
    $showError(t("otp.verificationFailed"));
  }
};

const cancelOtpSetup = async (event: Event) => {
  event.preventDefault();
  
  if (!confirm("Are you sure you want to cancel TOTP setup? You will need to start over.")) {
    return;
  }

  if (authStore.user === null) {
    return;
  }

  try {
    // Disable OTP without requiring a code since it's not verified yet
    await api.disableOtp(authStore.user.id, "");
    otpSetupKey.value = "";
    authStore.user.otpPending = false;
    $showSuccess("TOTP setup cancelled");
  } catch (err: any) {
    $showError(err);
  }
};

const resetOtpKey = async (event: Event) => {
  event.preventDefault();

  const password = prompt("Enter your password to reset TOTP:");
  if (!password) {
    return;
  }

  if (authStore.user === null) {
    return;
  }

  try {
    const res = await api.resetOtp(authStore.user.id, password);
    otpSetupKey.value = res.setupKey;
    $showSuccess("TOTP key has been reset. Please reconfigure your authenticator app.");
  } catch (err: any) {
    $showError(err);
  }
};

const generateRecovery = async (event: Event) => {
  event.preventDefault();

  layoutStore.showHover({
    prompt: "otp",
    confirm: async (code: string) => {
      if (authStore.user === null) {
        return;
      }

      try {
        const res = await api.generateRecoveryCodes(authStore.user.id, code);
        recoveryCodes.value = res.codes;
        showRecoveryCodes.value = true;
        $showSuccess("Recovery codes generated successfully");
      } catch (err: any) {
        $showError(err);
      }
    },
  });
};

const copyRecoveryCodes = async (event: Event) => {
  event.preventDefault();
  const text = recoveryCodes.value.join("\n");
  await copyToClipboard(text);
};

const closeRecoveryCodes = (event: Event) => {
  event.preventDefault();
  showRecoveryCodes.value = false;
  recoveryCodes.value = [];
};
</script>

<style lang="css" scoped>
.qrcode-container,
.setup-key-container {
  display: flex;
  justify-content: center;
  align-items: center;
  margin: 1em 0;
}

.setup-key-container {
  justify-content: space-between;
}

.setup-key-container > * {
  margin: 0 0.5em;
}

.recovery-codes {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.5em;
  margin: 1em 0;
  padding: 1em;
  background: var(--surfaceSecondary);
  border-radius: 0.5em;
}

.recovery-code {
  font-family: monospace;
  font-size: 1.1em;
  padding: 0.5em;
  background: var(--surfacePrimary);
  border-radius: 0.25em;
  text-align: center;
}

.message.warning {
  color: var(--colorWarning);
  background: var(--surfaceWarning);
  padding: 0.75em;
  border-radius: 0.25em;
  margin-bottom: 1em;
}
</style>
