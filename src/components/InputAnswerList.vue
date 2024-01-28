<template>
  <n-dynamic-input :value="value" @update:value="$emit('update:value', $event)" :on-create="onCreate">
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

<script>
import { defineComponent, toRefs } from "vue";
import { NDynamicInput, NInput } from "naive-ui";

export default defineComponent({
  components: {
    NDynamicInput,
    NInput,
  },
  props: ["value"],
  emits: ["update:value"],
  setup(props) {
    const { value } = toRefs(props);
    const onCreate = () => {
      const v = value.value.length + 1;
      return {
        input: v.toString() + ' ' + v.toString(),
        answer: (v * 2).toString(),
      };
    };

    return {
      onCreate,
    };
  },
});
</script>
