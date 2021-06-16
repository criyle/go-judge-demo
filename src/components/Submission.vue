<template>
  <submission-list
    :submission="submission"
    @loadMore="loadMore"
  ></submission-list>
</template>

<script>
import SubmissionList from "./SubmissionList.vue";
import axios from "axios";

export default {
  name: "Submission",
  data: () => ({
    submission: [],
  }),
  components: {
    SubmissionList,
  },
  methods: {
    loadMore() {
      const p =
        this.submission.length > 0
          ? {
              id: this.submission[this.submission.length - 1].id,
            }
          : {};
      axios
        .get("/api/submission", {
          params: p,
        })
        .then((r) => {
          r.data.forEach((data) => {
            const idx = this.submission.findIndex((s) => s.id === data.id);
            if (idx == -1) {
              this.submission.push(data);
            }
          });
        });
    },
    createWS() {
      const url =
        (location.protocol == "https:" ? "wss" : "ws") +
        "://" +
        document.domain +
        ":" +
        location.port +
        "/api/ws/judge";
      const ws = new WebSocket(url);
      ws.addEventListener("message", (event) => {
        const data = JSON.parse(event.data);
        const idx = this.submission.findIndex((s) => s.id === data.id);
        if (idx >= 0) {
          this.submission[idx] = {
            ...this.submission[idx],
            status: data.status,
            results: data.results || this.submission[idx].results,
          };
        } else {
          this.submission.unshift(data);
        }
      });
      ws.addEventListener("close", () => {
        // reconnect after 1000 ms
        setTimeout(this.createWS, 1000);
      });
      this.$ws = ws;
    },
  },
  mounted: function () {
    this.createWS();
  },
  created: function () {
    this.loadMore();
  },
  beforeDestroy: function () {
    this.$ws.close();
  },
};
</script>
