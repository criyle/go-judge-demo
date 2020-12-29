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
    submission: []
  }),
  components: {
    SubmissionList
  },
  methods: {
    loadMore() {
      const p =
        this.submission.length > 0
          ? {
              id: this.submission[this.submission.length - 1].id
            }
          : {};
      axios
        .get("/api/submission", {
          params: p
        })
        .then(r => {
          this.submission.push(...r.data);
        });
    }
  },
  mounted: function() {
    const url =
      (location.protocol == "https:" ? "wss" : "ws") +
      "://" +
      document.domain +
      ":" +
      location.port +
      "/api/ws/judge";
    const ws = new WebSocket(url);
    ws.addEventListener("message", event => {
      const data = JSON.parse(event.data);
      const idx = this.submission.findIndex(s => s.id === data.id);
      if (idx >= 0) {
        this.$set(this.submission, idx, {
          ...this.submission[idx],
          status: data.status,
          results: data.results || this.submission[idx].results
        });
      } else {
        this.submission.unshift(data);
      }
    });
    this.$ws = ws;
  },
  created: function() {
    this.loadMore();
  },
  beforeDestory: function() {
    this.$ws.close();
  }
};
</script>
