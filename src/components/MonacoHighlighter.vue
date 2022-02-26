<template>
  <div class="code" v-html="html"></div>
</template>

<script>
import { ref, toRefs } from "vue";
import { darkTheme } from "naive-ui";
import { defineComponent, watch } from "vue";
import * as monaco from "monaco-editor";

export default defineComponent({
  name: "MonacoHighlighter",
  props: {
    value: String,
    language: String,
    theme: Object,
  },
  setup(props) {
    const html = ref("");
    const { theme, value, language } = toRefs(props);

    const getTheme = () =>
      theme.value.baseColor === darkTheme.common.baseColor ? "vs-dark" : "vs";

    const render = async () => {
      monaco.editor.setTheme(getTheme());
      html.value = await monaco.editor.colorize(value.value, language.value);
    };

    render();

    watch([theme, value, language], render);

    return {
      html,
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

