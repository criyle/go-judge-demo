<template>
  <div
    class="code"
    :data-lang="language"
  >{{ value }}</div>
</template>

<script>
import * as monaco from "monaco-editor";

export default {
  name: "MonacoHighlighter",
  props: {
    value: String,
    language: String
  },
  mounted: function() {
    monaco.editor.colorizeElement(this.$el);
  }
};
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

