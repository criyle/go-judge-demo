<script setup lang="ts">
import * as monaco from "monaco-editor";
import { darkTheme } from "naive-ui";
import {
  onBeforeUnmount,
  onMounted,
  ref,
  type VNodeRef,
  watchEffect,
} from "vue";

const props = defineProps({
  modelValue: String,
  language: String,
  theme: Object,
});

const emit = defineEmits(['update:modelValue']);

const root = ref<VNodeRef | null>(null);
let editor: monaco.editor.IStandaloneCodeEditor | null = null;

const getTheme = () =>
  props.theme?.baseColor === darkTheme.common.baseColor ? "vs-dark" : "vs";

onMounted(() => {
  editor = monaco.editor.create(root.value, {
    value: props.modelValue,
    language: props.language,
    automaticLayout: true,
    theme: getTheme(),
  });
  editor.onDidChangeModelContent(() => {
    const val = editor?.getValue();
    if (val !== props.modelValue) {
      emit("update:modelValue", editor?.getValue());
    }
  });
});

onBeforeUnmount(() => {
  editor?.dispose();
});

watchEffect(() => {
  editor && monaco.editor.setModelLanguage(editor.getModel()!!, props.language!!);
});

watchEffect(() => {
  const curVal = editor?.getValue();
  if (props.modelValue !== curVal) {
    editor?.setValue(props.modelValue as string);
  }
});

watchEffect(() => {
  monaco.editor.setTheme(getTheme());
});
</script>

<template>
  <div ref="root" style="height: 100%"></div>
</template>
