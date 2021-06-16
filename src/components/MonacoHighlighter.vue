<template>
  <div class="code" :data-lang="language" ref="root">{{ value }}</div>
</template>

<script>
import { ref, onMounted } from "vue";
import { defineComponent } from "@vue/runtime-core";
import * as monaco from "monaco-editor";

export default defineComponent({
  name: "MonacoHighlighter",
  props: {
    value: String,
    language: String,
  },
  setup() {
    const root = ref(null);
    onMounted(() => {
      monaco.editor.colorizeElement(root.value);
    });

    return {
      root
    };
  },
});
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

