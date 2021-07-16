<template>
  <monaco-highlighter
    :value="value"
    :language="language"
    :theme="themeVars"
  ></monaco-highlighter>
</template>

<script>
import { defineAsyncComponent, defineComponent } from "vue";
import { useThemeVars } from "naive-ui";

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
  data: () => ({
    themeVars: useThemeVars(),
  })
});
</script>
