<template>
  <span>{{ parseDate }}</span>
</template>

<script setup lang="ts">
import moment from "moment";
import { computed } from "vue";

const { date } = defineProps<{ date: { seconds: number, nanos: number } }>();

const parseDate = computed(() => {
  const d = date;
  return moment(d).format("dddd, MMMM Do YYYY, kk:mm:ss.SSS ZZ");
});
</script>