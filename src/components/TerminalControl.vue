<template>
  <div>
  </div>
</template>

<script>
import { Terminal } from "xterm";
import "xterm/css/xterm.css";

export default {
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
  beforeDestroy() {
    this.$ws.close();
    this.$terminal.dispose();
  }
};
</script>
