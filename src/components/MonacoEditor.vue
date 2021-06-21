<template>
  <div ref="root" style="height: 100%"></div>
</template>

<script>
import {
  ref,
  onMounted,
  onBeforeUnmount,
  toRefs,
  watch,
  defineComponent,
} from "vue";
import * as monaco from "monaco-editor";

import { darkTheme } from "naive-ui";

export default defineComponent({
  name: "MonacoEditor",
  props: {
    modelValue: String,
    language: String,
    theme: Object,
  },
  inject: [],
  setup(props, { emit }) {
    const root = ref(null);
    let editor = null;
    const { modelValue, language, theme } = toRefs(props);

    const getTheme = () =>
      theme.value.baseColor === darkTheme.common.baseColor ? "vs-dark" : "vs";

    onMounted(() => {
      editor = monaco.editor.create(root.value, {
        value: modelValue.value,
        language: language.value,
        automaticLayout: true,
        theme: getTheme(),
      });
      editor.onDidChangeModelContent(() => {
        const val = editor.getValue();
        if (val !== modelValue.value) {
          emit("update:modelValue", editor.getValue());
        }
      });
    });

    onBeforeUnmount(() => {
      editor.dispose();
    });

    watch(language, (newVal) => {
      monaco.editor.setModelLanguage(editor.getModel(), newVal);
    });

    watch(modelValue, (newVal) => {
      const curVal = editor.getValue();
      if (newVal !== curVal) {
        editor.setValue(newVal);
      }
    });

    watch(theme, () => {
      monaco.editor.setTheme(getTheme());
    });

    return {
      root,
    };
  },
});
</script>
