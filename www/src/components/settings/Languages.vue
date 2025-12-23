<template>
  <select name="selectLanguage" v-on:change="change" :value="locale">
    <option v-for="(language, value) in locales" :key="value" :value="value">
      {{ language }}
    </option>
  </select>
</template>

<script>
import { markRaw } from "vue";

export default {
  name: "languages",
  props: ["locale"],
  data() {
    const dataObj = {};
    const locales = {
      de: "Deutsch",
      en: "English",
      es: "Español",
      fr: "Français",
      it: "Italiano",
      ja: "日本語",
      ko: "한국어",
      "pt-br": "Português (Brasil)",
      ru: "Русский",
      "zh-cn": "中文 (简体)",
      "zh-tw": "中文 (繁體)",
    };

    // Vue3 reactivity breaks with this configuration
    // so we need to use markRaw as a workaround
    // https://github.com/vuejs/core/issues/3024
    Object.defineProperty(dataObj, "locales", {
      value: markRaw(locales),
      configurable: false,
      writable: false,
    });

    return dataObj;
  },
  methods: {
    change(event) {
      this.$emit("update:locale", event.target.value);
    },
  },
};
</script>
