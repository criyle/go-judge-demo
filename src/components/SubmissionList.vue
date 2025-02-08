<template>
  <div>
    <n-collapse>
      <submission-list-item v-for="(s, index) in submission" :key="s.id" :index="index" :s="s"></submission-list-item>
    </n-collapse>

    <n-divider />

    <n-button @click="$emit('loadMore')">More</n-button>
  </div>
</template>

<script setup lang="ts">
import { NButton, NCollapse, NDivider } from "naive-ui";
import SubmissionListItem from "./SubmissionListItem.vue";

const { submission } = defineProps([
  "submission"
]);
</script>
