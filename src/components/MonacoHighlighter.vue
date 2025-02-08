<template>
  <div class="code" v-html="html"></div>
</template>

<script setup lang="ts">
import * as monaco from "monaco-editor";
import { darkTheme } from "naive-ui";
import { ref, toRefs, watch } from "vue";

const props = defineProps({
  value: String,
  language: String,
  theme: Object,
});

const html = ref("");
const { theme, value, language } = toRefs(props);

const getTheme = () =>
  props.theme?.baseColor === darkTheme.common.baseColor ? "vs-dark" : "vs";

const render = async () => {
  monaco.editor.setTheme(getTheme());
  html.value = await monaco.editor.colorize(props?.value || "", props?.language || "text", {});
};

render();

watch([theme, value, language], render);
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
