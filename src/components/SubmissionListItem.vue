<template>
  <div :class="{ active, title: true }" @click="active = !active">
    <i class="dropdown icon"></i>
    <span class="status">{{ s.status }}</span>
    <span class="date"><date :date="s.date"></date></span>
  </div>
  <div :class="{ active, content: true }">
    <div>
      <div class="info properties">
        <property-view label="_id" :value="s.id"></property-view>
        <property-view
          label="language name"
          :value="s.language.name"
        ></property-view>
        <property-view
          label="source file name"
          :value="s.language.sourceFileName"
        ></property-view>
        <property-view
          label="compile cmd"
          :value="s.language.compileCmd"
        ></property-view>
        <property-view
          label="executable file names"
          :value="s.language.executables"
        ></property-view>
        <property-view
          label="exec cmd"
          :value="s.language.runCmd"
        ></property-view>
      </div>
      <div class="info">
        <code-view
          label="code"
          :value="s.source"
          :language="s.language.name"
        ></code-view>
      </div>
      <div v-for="(u, index) in s.results" :key="index">
        <sui-divider />
        <div class="info properties">
          <property-view label="cpu" :value="cpu(u.time)"></property-view>
          <property-view
            label="memory"
            :value="memory(u.memory)"
          ></property-view>
        </div>
        <div class="info">
          <code-view
            v-if="u.stdin"
            label="stdin"
            :value="u.stdin"
            language="text"
          ></code-view>
          <code-view
            v-if="u.stdout"
            label="stdout"
            :value="u.stdout"
            language="text"
          ></code-view>
          <code-view
            v-if="u.stderr"
            label="stderr"
            :value="u.stderr"
            language="text"
          ></code-view>
          <code-view
            v-if="u.log"
            label="log"
            :value="u.log"
            language="text"
          ></code-view>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { defineAsyncComponent, defineComponent } from "@vue/runtime-core";
import CodeView from "./CodeView.vue";
import PropertyView from "./PropertyView.vue";
const Date = defineAsyncComponent(() => import("./Date.vue"));

export default defineComponent({
  name: "SubmissionListItem",
  props: ["index", "s"],
  data: () => ({
    active: false,
  }),
  components: {
    CodeView,
    PropertyView,
    Date,
  },
  methods: {
    cpu: (value) => {
      if (value) {
        return value + " ms";
      } else {
        return "0 ms";
      }
    },
    memory: (value) => {
      if (value) {
        return value + " kB";
      } else {
        return "0 kB";
      }
    },
  },
});
</script>

<style scoped>
/* .info {
  padding: 4px 16px;
} */

.properties {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
}

.properties > div {
  padding-right: 10px;
}

.status {
  min-width: 180px;
}

.date {
  width: auto;
  font-size: 14px;
  padding-left: 10px;
}
</style>