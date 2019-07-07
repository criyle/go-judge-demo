<template>
<div class="code-editor-container">
  <md-toolbar>
    <h3 class="md-title" style="flex: 1">GO Judger</h3>
    <md-button @click="$emit('submit')">Submit</md-button>
  </md-toolbar>
  <monaco-editor :value="value" @input="$emit('input', $event)" class="code-editor-editor"></monaco-editor>
</div>
</template>

<script>
import MonacoEditor from './MonacoEditor.vue';

export default {
  name: "CodeEditor",
  props: {
    value: String,
  },
  components: {
    MonacoEditor,
  },
}
</script>

<style scoped>
.code-editor-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.code-editor-editor {
  margin-top: 5px;
  flex: 1 0;
}
</style>
