<template>
  <n-dynamic-input :value="model" @update:value="model = $event" :on-create="onCreate">
    <template #default="{ value }">
      <div style="display: flex; width: 100%">
        <n-input type="textarea" v-model:value="value.input" placeholder="Input" :autosize="{ minRows: 1 }"
          style="margin-right: 12px" />
        <n-input type="textarea" v-model:value="value.answer" placeholder="Answer" :autosize="{ minRows: 1 }">
        </n-input>
      </div>
    </template>
  </n-dynamic-input>
</template>

<script setup lang="ts">
import { NDynamicInput, NInput } from "naive-ui";

const model = defineModel<any[]>();

const onCreate = () => {
  const v = model.value.length + 1;
  return {
    input: v.toString() + ' ' + v.toString(),
    answer: (v * 2).toString(),
  };
};
</script>
