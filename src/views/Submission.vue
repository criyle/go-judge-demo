<template>
  <submission-list :submissions="submissions" @loadMore="loadMore"></submission-list>
</template>

<script setup lang="ts">
import axios from "axios";
import { onBeforeUnmount, onMounted, ref } from "vue";
import SubmissionList from "../components/SubmissionList.vue";

const submissions = ref([]);
let websocket: WebSocket = null;

const loadMore = () => {
  const p =
    submissions.value.length > 0
      ? {
        id: submissions[submissions.value.length - 1].id,
      }
      : {};
  axios
    .get("/api/submission", {
      params: p,
    })
    .then((r) => {
      r.data.submissions.forEach((data) => {
        const idx = submissions.value.findIndex((s) => s.id === data.id);
        if (idx == -1) {
          submissions.value.push(data);
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
    const idx = submissions.value.findIndex((s) => s.id === data.id);
    if (idx >= 0) {
      submissions.value[idx] = {
        ...submissions.value[idx],
        status: data.status,
        results: data.results || submissions.value[idx].results,
      };
    } else {
      submissions.value.unshift(data);
    }
  });
  ws.addEventListener("close", () => {
    // reconnect after 1000 ms
    setTimeout(createWS, 1000);
  });
  ws.addEventListener("open", () => {
    if (submissions.value.length === 0) {
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
