<template>
  <div>
    <n-collapse>
      <submission-list-item
        v-for="(s, index) in submission"
        :key="s.id"
        :index="index"
        :s="s"
      ></submission-list-item>
    </n-collapse>

    <n-divider />

    <n-button @click="$emit('loadMore')">More</n-button>
  </div>
</template>

<script>
import { defineComponent } from "vue";
import { NCollapse, NButton, NDivider } from "naive-ui";
import SubmissionListItem from "./SubmissionListItem.vue";

export default defineComponent({
  name: "SubmissionList",
  props: ["submission"],
  components: {
    NButton,
    NCollapse,
    NDivider,
    SubmissionListItem,
    Date,
  },
});
</script>
