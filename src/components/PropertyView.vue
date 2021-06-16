<template>
  <div>
    <div class="ui sub header">{{ label }}</div>
    <span>{{ value }}</span>
  </div>
</template>

<script>
export default {
  name: "PropertyView",
  props: {
    label: String,
    value: String,
  },
};
</script>
