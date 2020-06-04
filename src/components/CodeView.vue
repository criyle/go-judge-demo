<template>
  <div>
    <div class="md-caption">{{label}}</div>
    <monaco-highlighter
      :value="value"
      :language="language"
    ></monaco-highlighter>
  </div>
</template>

<script>
const MonacoHighlighter = () => import("./MonacoHighlighter.vue");

export default {
  name: "CodeView",
  props: {
    value: String,
    language: String,
    label: String
  },
  components: {
    MonacoHighlighter
  }
};
</script>
