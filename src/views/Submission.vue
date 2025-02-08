<template>
  <submission-list :submission="submission" @loadMore="loadMore"></submission-list>
</template>

<script setup lang="ts">
import axios from "axios";
import { onBeforeUnmount, onMounted, ref } from "vue";
import SubmissionList from "../components/SubmissionList.vue";

const submission = ref([]);
let websocket: WebSocket = null;

const loadMore = () => {
  const p =
    submission.value.length > 0
      ? {
        id: submission[submission.value.length - 1].id,
      }
      : {};
  axios
    .get("/api/submission", {
      params: p,
    })
    .then((r) => {
      r.data.forEach((data) => {
        const idx = submission.value.findIndex((s) => s.id === data.id);
        if (idx == -1) {
          submission.value.push(data);
        }
      });
    });
}

const createWS = () => {
  const url =
    (location.protocol == "https:" ? "wss" : "ws") +
    "://" +
    location.host +
    "/api/ws/judge";
  const ws = new WebSocket(url);
  ws.addEventListener("message", (event) => {
    const data = JSON.parse(event.data);
    const idx = submission.value.findIndex((s) => s.id === data.id);
    if (idx >= 0) {
      submission.value[idx] = {
        ...submission.value[idx],
        status: data.status,
        results: data.results || submission.value[idx].results,
      };
    } else {
      submission.value.unshift(data);
    }
  });
  ws.addEventListener("close", () => {
    // reconnect after 1000 ms
    setTimeout(createWS, 1000);
  });
  ws.addEventListener("open", () => {
    if (submission.value.length === 0) {
      loadMore();
    }
  });
  websocket = ws;
};

onMounted(() => {
  createWS();
});

onBeforeUnmount(() => {
  websocket?.close();
});
</script>
