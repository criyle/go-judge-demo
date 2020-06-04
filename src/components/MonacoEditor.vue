<template>
  <div>
  </div>
</template>

<script>
import * as monaco from "monaco-editor";

export default {
  name: "MonacoEditor",
  props: {
    value: String,
    language: String
  },
  mounted() {
    this.$editor = monaco.editor.create(this.$el, {
      value: this.value,
      language: this.language,
      automaticLayout: true
    });
    this.$editor.onDidChangeModelContent(() => {
      this.$emit("input", this.$editor.getValue());
    });
  },
  beforeDestory() {
    this.$editor.dispose();
  },
  watch: {
    language: function(newVal) {
      monaco.editor.setModelLanguage(this.$editor.getModel(), newVal);
    },
    value: function(newVal) {
      const curVal = this.$editor.getValue();
      if (newVal !== curVal) {
        this.$editor.setValue(newVal);
      }
    }
  }
};
</script>
