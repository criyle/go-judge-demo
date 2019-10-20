<template>
  <submission-list 
    :submission="submission"
    @loadMore="loadMore">
  </submission-list>
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
      "/ws";
    const ws = new WebSocket(url);
    ws.addEventListener("message", event => {
      const data = JSON.parse(event.data);
      const sub = this.submission.find(s => s.id === data.id);
      if (sub) {
        sub.status = data.status;
        sub.date = data.date || sub.date;
        sub.language = data.language || sub.language;
        sub.results = data.results || sub.results;
        this.submission = [...this.submission];
      } else {
        this.submission = [data, ...this.submission];
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
