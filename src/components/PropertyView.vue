<template>
  <div>
    <div class="md-caption">{{label}}</div>
    <div>{{value}}</div>
  </div>
</template>

<script>
export default {
  name: "PropertyView",
  props: {
    label: String,
    value: String
  }
};
</script>
