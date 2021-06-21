<template>
  <n-collapse-item>
    <template #header>
      <span class="status">{{ s.status }}</span>
      <span class="date"><date :date="s.date"></date></span>
    </template>
    <n-descriptions label-placement="top" :column="6">
      <n-descriptions-item>
        <template #label>_id</template>
        {{ s.id }}
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label>language name</template>
        {{ s.language.name }}
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label>source file name</template>
        {{ s.language.sourceFileName }}
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label>compile cmd</template>
        {{ s.language.compileCmd }}
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label>executable file names</template>
        {{ s.language.executables }}
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label>exec cmd</template>
        {{ s.language.runCmd }}
      </n-descriptions-item>

      <n-descriptions-item :span="6">
        <template #label>code</template>
        <code-view
          label="code"
          :value="s.source"
          :language="s.language.name"
        ></code-view>
      </n-descriptions-item>
    </n-descriptions>
    <template v-for="(u, index) in s.results" :key="index">
      <n-divider />
      <n-descriptions :column="2">
        <n-descriptions-item>
          <template #label>cpu</template>
          {{ cpu(u.time) }}
        </n-descriptions-item>
        <n-descriptions-item>
          <template #label>memory</template>
          {{ memory(u.memory) }}
        </n-descriptions-item>
      </n-descriptions>
      <template v-for="name in ['stdin', 'stdout', 'stderr', 'log']">
        <n-descriptions v-if="u[name]">
          <n-descriptions-item>
            <template #label>{{ name }}</template>
            <code-view
              :label="name"
              :value="u[name]"
              language="text"
            ></code-view>
          </n-descriptions-item>
        </n-descriptions>
      </template>
    </template>
  </n-collapse-item>
</template>

<script>
import { defineAsyncComponent, defineComponent } from "vue";
import {
  NCollapseItem,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
} from "naive-ui";
import CodeView from "./CodeView.vue";
const Date = defineAsyncComponent(() => import("./Date.vue"));

export default defineComponent({
  name: "SubmissionListItem",
  props: ["index", "s"],
  data: () => ({
    active: false,
  }),
  components: {
    NCollapseItem,
    NDescriptions,
    NDescriptionsItem,
    NDivider,
    CodeView,
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