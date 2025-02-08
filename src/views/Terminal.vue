<template>
  <div ref="root" class="terminal"></div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, VNodeRef } from "vue";
import { Terminal } from "xterm";
import "xterm/css/xterm.css";

const root = ref<VNodeRef | null>(null);
const terminal: Terminal = new Terminal({
  allowTransparency: true,
  theme: {
    background: "rgba(0,0,0,60%)",
  },
});
const url =
  (location.protocol == "https:" ? "wss" : "ws") +
  "://" +
  location.host +
  "/api/ws/shell";
const ws: WebSocket = new WebSocket(url);
ws.binaryType = "arraybuffer";
ws.onmessage = (ev) => {
  terminal.write(new Uint8Array(ev.data));
};
terminal.onData((data) => {
  ws.send(data);
});

onMounted(() => {
  terminal.open(root.value);

  // this.$terminal.onResize(data => {
  //   // console.log(data);
  // });
});
onBeforeUnmount(() => {
  ws.close();
  terminal.dispose();
});
</script>

<style>
.terminal .xterm-viewport {
  overflow: hidden !important;
}
</style>
