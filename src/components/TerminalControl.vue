<template>
  <div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "@vue/runtime-core";
import { Terminal } from "xterm";
import "xterm/css/xterm.css";

export default defineComponent({
  name: "TerminalControl",
  mounted() {
    this.$terminal = new Terminal();
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

    this.$ws.onmessage = ev => {
      this.$terminal.write(new Uint8Array(ev.data));
    };

    this.$terminal.onData(data => {
      this.$ws.send(data);
    });

    // this.$terminal.onResize(data => {
    //   // console.log(data);
    // });
  },
  beforeUnmount() {
    this.$ws.close();
    this.$terminal.dispose();
  }
});
</script>
