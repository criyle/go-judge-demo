<template>
  <n-tooltip trigger="hover">
    <template #trigger>
      <div>
        <n-ellipsis style="width: 100%; max-width: 1100px" :line-clamp="20" :tooltip="false" expand-trigger="click">
          <monaco-highlighter style="word-wrap: break-word" class="highlighter" :value="value" :language="language"
            :theme="themeVars"></monaco-highlighter>
        </n-ellipsis>
      </div>
    </template>

    <div v-once>
      Total character: {{ value.length }} <br />
      Total line: {{ value.split("\n").length }}
    </div>
  </n-tooltip>
</template>

<script>
import { defineAsyncComponent, defineComponent } from "vue";
import { NEllipsis, NTooltip, useThemeVars } from "naive-ui";

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
    NEllipsis,
    NTooltip,
    MonacoHighlighter,
  },
  data: () => ({
    themeVars: useThemeVars(),
  }),
});
</script>

<style scoped>
/* .highlighter {
  width: 100%;
  overflow-x: scroll;
} */
</style>