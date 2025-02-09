<script setup lang="ts">
import * as monaco from "monaco-editor";
import { darkTheme, type ThemeCommonVars } from "naive-ui";
import {
  onBeforeUnmount,
  onMounted,
  useTemplateRef,
  watch,
  watchEffect
} from "vue";

const modelValue = defineModel<string>();
const { language, theme } = defineProps<{ modelValue: string, language: string, theme: ThemeCommonVars }>();

const root = useTemplateRef("root");
let editor: monaco.editor.IStandaloneCodeEditor | null = null;

const getTheme = () =>
  theme?.baseColor === darkTheme.common.baseColor ? "vs-dark" : "vs";

onMounted(() => {
  editor = monaco.editor.create(root.value!!, {
    value: modelValue.value,
    language: language,
    automaticLayout: true,
    theme: getTheme(),
  });
  editor.onDidChangeModelContent(() => {
    const val = editor?.getValue();
    if (val !== modelValue.value) {
      modelValue.value = val;
    }
  });
});

onBeforeUnmount(() => {
  editor?.dispose();
});

watchEffect(() => {
  language && editor && monaco.editor.setModelLanguage(editor.getModel()!!, language);
});

watch(modelValue, (v) => {
  const curVal = editor?.getValue();
  if (v != curVal) {
    editor?.setValue(v!!);
  }
})

watchEffect(() => {
  monaco.editor.setTheme(getTheme());
});
</script>

<template>
  <div ref="root" style="height: 100%"></div>
</template>
