<template>
  <div ref="root"></div>
</template>

<script>
import { ref, onMounted, onBeforeUnmount, toRefs, watch } from "vue";
import * as monaco from "monaco-editor";

export default {
  name: "MonacoEditor",
  props: {
    modelValue: String,
    language: String,
  },
  setup(props, { emit }) {
    const root = ref(null);
    let editor = null;
    const { modelValue, language } = toRefs(props);

    onMounted(() => {
      editor = monaco.editor.create(root.value, {
        value: modelValue.value,
        language: language.value,
        automaticLayout: true,
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

    return {
      root,
    };
  },
};
</script>
