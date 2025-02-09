<template>
  <div>
    <n-collapse>
      <submission-list-item v-for="(submission, index) in submissions" :key="submission.id"
        :submission="submission"></submission-list-item>
    </n-collapse>

    <n-divider />

    <n-button @click="$emit('loadMore')">More</n-button>
  </div>
</template>

<script setup lang="ts">
import { NButton, NCollapse, NDivider } from "naive-ui";
import SubmissionListItem from "./SubmissionListItem.vue";

const { submissions } = defineProps<{ submissions: any[] }>();
</script>
