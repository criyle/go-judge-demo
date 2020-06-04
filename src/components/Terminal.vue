<template>
  <div>
    <terminal-control></terminal-control>
  </div>
</template>

<script>
// import TerminalControl from "./TerminalControl.vue";
const TerminalControl = () => import("./TerminalControl.vue");

export default {
  name: "Terminal",
  components: {
    TerminalControl
  }
};
</script>