<template>
  <div>
    <terminal-control></terminal-control>
  </div>
</template>

<script>
import { defineAsyncComponent } from "@vue/runtime-core";
// import TerminalControl from "./TerminalControl.vue";
const TerminalControl = defineAsyncComponent(() =>
  import("./TerminalControl.vue")
);

export default {
  name: "Terminal",
  components: {
    TerminalControl,
  },
};
</script>