<template>
  <span>{{ parseDate }}</span>
</template>

<script setup lang="ts">
import moment from "moment";
import { computed } from "vue";

const { date } = defineProps<{ date: { seconds: number, nanos: number } }>();

const parseDate = computed(() => {
  const d = new Date(date.seconds * 1000 + date.nanos / 1e6);
  return moment(d).format("dddd, MMMM Do YYYY, kk:mm:ss.SSS ZZ");
});
</script>