<template>
  <div class="vue-monaco-editor">
  </div>
</template>

<script>
import * as monaco from 'monaco-editor';

export default {
  name: "MonacoEditor",
  props: {
    value: String,
  },
  mounted: function() {
    this.$editor = monaco.editor.create(this.$el, {
      value: this.value,
      language: "cpp",
      automaticLayout: true,
    });
    this.$editor.onDidChangeModelContent(() => {
      this.$emit('input', this.$editor.getValue());
    });
  },
  beforeDestory: function() {
    this.$editor.dispose();
  }
}
</script>

<style scoped>
.vue-monaco-editor {
}
</style>
