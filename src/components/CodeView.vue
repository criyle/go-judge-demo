<template>
  <n-element>
    <template v-slot:default="{ themeVars }">
      <monaco-highlighter
        :value="value"
        :language="language"
        :theme="themeVars"
      ></monaco-highlighter>
    </template>
  </n-element>
</template>

<script>
import { defineAsyncComponent, defineComponent } from "vue";
import { NElement } from "naive-ui";

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
    NElement,
    MonacoHighlighter,
  },
});
</script>
