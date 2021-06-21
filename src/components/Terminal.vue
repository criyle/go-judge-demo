<template>
  <div class="terminal"></div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { Terminal } from "xterm";
import "xterm/css/xterm.css";

export default defineComponent({
  name: "Terminal",
  mounted() {
    this.$terminal = new Terminal({
      allowTransparency: true,
      theme: {
        background: "rgba(0,0,0,60%)",
      },
    });
    this.$terminal.open(this.$el);

    const url =
      (location.protocol == "https:" ? "wss" : "ws") +
      "://" +
      document.domain +
      ":" +
      location.port +
      "/api/ws/shell";
    this.$ws = new WebSocket(url);

    this.$ws.binaryType = "arraybuffer";

    this.$ws.onmessage = (ev) => {
      this.$terminal.write(new Uint8Array(ev.data));
    };

    this.$terminal.onData((data) => {
      this.$ws.send(data);
    });

    // this.$terminal.onResize(data => {
    //   // console.log(data);
    // });
  },
  beforeUnmount() {
    this.$ws.close();
    this.$terminal.dispose();
  },
});
</script>

<style>
.terminal .xterm-viewport {
  overflow: hidden !important;
}
</style>
