<template>
  <span>{{ parseDate }}</span>
</template>

<script>
import { defineComponent } from "vue";
import moment from "moment";

export default defineComponent({
  name: "Date",
  props: ["date"],
  computed: {
    parseDate() {
      const d = new Date(this.date.seconds * 1000 + this.date.nanos / 1e6);
      return moment(d).format("dddd, MMMM Do YYYY, kk:mm:ss.SSS ZZ");
    },
  },
});
</script>