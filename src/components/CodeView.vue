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
      Total character: {{ value?.length }} <br />
      Total line: {{ value?.split("\n").length }}
    </div>
  </n-tooltip>
</template>

<script setup lang="ts">
import { NEllipsis, NTooltip, useThemeVars } from "naive-ui";
import { defineAsyncComponent } from "vue";

const MonacoHighlighter = defineAsyncComponent(() =>
  import("./MonacoHighlighter.vue")
);

const props = defineProps({
  value: String,
  language: String,
  label: String,
});
const themeVars = useThemeVars();
</script>

<style scoped>
/* .highlighter {
  width: 100%;
  overflow-x: scroll;
} */
</style>
