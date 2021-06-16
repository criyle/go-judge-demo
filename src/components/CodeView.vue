<template>
  <div>
    <div class="md-caption">{{ label }}</div>
    <monaco-highlighter
      :value="value"
      :language="language"
    ></monaco-highlighter>
  </div>
</template>

<script>
import { defineAsyncComponent, defineComponent } from "@vue/runtime-core";
const MonacoHighlighter = defineAsyncComponent(() =>
  import("./MonacoHighlighter.vue")
);

export default defineComponent({
  name: "CodeView",
  props: {
    value: String,
    language: String,
    label: String,
  },
  components: {
    MonacoHighlighter,
  },
});
</script>
