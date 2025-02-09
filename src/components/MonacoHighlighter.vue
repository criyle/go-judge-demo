<template>
  <div v-once ref="root" :data-lang="language" class="code">{{ value }}</div>
</template>

<script setup lang="ts">
import * as monaco from "monaco-editor";
import { darkTheme, type ThemeCommonVars } from "naive-ui";
import { onMounted, useTemplateRef, watchEffect } from "vue";

const { value, language, theme } = defineProps<{ value: string, language: string, theme: ThemeCommonVars }>();
const root = useTemplateRef("root");

const getTheme = () =>
  theme?.baseColor === darkTheme.common.baseColor ? "vs-dark" : "vs";

const render = async () => {
  root.value && await monaco.editor.colorizeElement(root.value, { theme: getTheme() })
};

watchEffect(() => value && language && theme && render());
onMounted(() => render());
</script>

<style scoped>
.code {
  font-family: Menlo, Monaco, "Courier New", monospace;
  font-weight: normal;
  font-size: 12px;
  line-height: 18px;
  letter-spacing: 0px;
}
</style>
