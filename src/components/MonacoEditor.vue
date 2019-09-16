<template>
  <div class="vue-monaco-editor">
  </div>
</template>

<script>
import * as monaco from "monaco-editor";

export default {
  name: "MonacoEditor",
  props: {
    value: String,
    language: String,
  },
  mounted: function() {
    this.$editor = monaco.editor.create(this.$el, {
      value: this.value,
      language: this.language,
      automaticLayout: true
    });
    this.$editor.onDidChangeModelContent(() => {
      this.$emit("input", this.$editor.getValue());
    });
  },
  beforeDestory: function() {
    this.$editor.dispose();
  },
  watch: {
    language: function (newVal, oldVal) {
      monaco.editor.setModelLanguage(this.$editor.getModel(), newVal);
    },
  }
};
</script>

<style scoped>
</style>
