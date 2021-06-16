<template>
  <div class="submission-list-container md-content">
    <div class="ui styled accordion fluid">
      <submission-list-item
        v-for="(s, index) in submission"
        :key="s.id"
        :index="index"
        :s="s"
      ></submission-list-item>
    </div>
    <sui-button @click="$emit('loadMore')">More</sui-button>
  </div>
</template>

<script>
import { defineComponent } from "@vue/runtime-core";
import SubmissionListItem from "./SubmissionListItem.vue";

export default defineComponent({
  name: "SubmissionList",
  props: ["submission"],
  components: {
    SubmissionListItem,
    Date,
  }
});
</script>
