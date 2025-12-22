<template>
  <div class="card floating otp-modal">
    <div class="card-title">
      <h2>{{ t("otp.name") }}</h2>
      <p>{{ t("otp.verifyInstructions") }}</p>
      <p class="recovery-hint">Or enter a recovery code (XXXX-XXXX-XXXX)</p>
    </div>

    <div class="card-content">
      <input
        v-model.trim="totpCode"
        :class="inputClassObject"
        :placeholder="isRecoveryCode ? 'XXXX-XXXX-XXXX' : t('otp.codeInputPlaceholder')"
        @keyup.enter="submit"
        id="focus-prompt"
        tabindex="1"
        class="input input--block"
        type="text"
        :maxlength="isRecoveryCode ? 14 : 6"
        required
        autocomplete="one-time-code"
        aria-describedby="totp-error"
      />
    </div>

    <div class="card-action">
      <button
        class="button button--flat button--grey"
        @click="layoutStore.closeHovers"
        :aria-label="t('buttons.cancel')"
        :title="t('buttons.cancel')"
        tabindex="3"
      >
        {{ t("buttons.cancel") }}
      </button>
      <button
        class="button button--flat"
        :aria-label="t('buttons.verify')"
        :title="t('buttons.verify')"
        @click="submit"
        tabindex="2"
      >
        {{ t("buttons.verify") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject, ref } from "vue";
import { useLayoutStore } from "@/stores/layout";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";
import { computed } from "vue";

const $showError = inject<IToastError>("$showError")!;
const layoutStore = useLayoutStore();
const { t } = useI18n();
const totpCode = ref<string>("");
const inputClassObject = computed(() => ({
  empty: totpCode.value === "",
}));

const isRecoveryCode = computed(() => {
  return totpCode.value.includes("-");
});

const submit = async (event: Event) => {
  event.preventDefault();
  event.stopPropagation();
  
  if (isRecoveryCode.value) {
    if (!/^[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$/i.test(totpCode.value)) {
      $showError("Invalid recovery code format. Use XXXX-XXXX-XXXX");
      return;
    }
  } else {
    if (totpCode.value.length !== 6 || !/^\d+$/.test(totpCode.value)) {
      $showError(t("otp.invalidCodeType"));
      return;
    }
  }

  try {
    await layoutStore.currentPrompt?.confirm(totpCode.value);
  } catch (e) {
    if (e instanceof StatusError) {
      console.error("TOTP Verification Error:", e);
      $showError(t("otp.verificationFailed"));
    } else if (e instanceof Error) {
      $showError(e);
    }
  }

  layoutStore.closeHovers();
};
</script>

<style scoped>
.recovery-hint {
  font-size: 0.9em;
  opacity: 0.7;
  margin-top: 0.5em;
}
</style>
