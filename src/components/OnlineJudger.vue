<template>
  <n-form label-placement="top">
    <n-grid :span="24" :x-gap="24">
      <n-form-item-gi :span="12" label="Language">
        <n-select
          placeholder="Language"
          v-model:value="selectedOption"
          :options="
            Object.entries(languageOptions).map(([k, v]) => ({
              label: k,
              value: k,
            }))
          "
        />
      </n-form-item-gi>
      <n-gi :span="12">
        <div style="display: flex; justify-content: flex-end">
          <n-button @click="submit" round type="primary">Submit</n-button>
        </div>
      </n-gi>

      <n-form-item-gi :span="4" label="Language Name">
        <n-input v-model:value="language" />
      </n-form-item-gi>

      <n-form-item-gi :span="4" label="Source File Name">
        <n-input v-model:value="sourceFileName" />
      </n-form-item-gi>

      <n-form-item-gi :span="8" label="Compile Cmd">
        <n-input v-model:value="compileCmd" />
      </n-form-item-gi>

      <n-form-item-gi :span="4" label="Executable File Name">
        <n-input v-model:value="executables" />
      </n-form-item-gi>

      <n-form-item-gi :span="4" label="Exec Cmd">
        <n-input v-model:value="runCmd" />
      </n-form-item-gi>

      <n-gi :span="24" class="editor">
        <monaco-editor
          class="code-editor-editor"
          v-model="source"
          :language="language"
          :theme="themeVars"
        ></monaco-editor>
      </n-gi>

      <n-gi :span="24">
        <input-answer-list v-model:value="inputAnswer" />
      </n-gi>
    </n-grid>
  </n-form>
</template>

<script>
import { defineAsyncComponent, defineComponent } from "vue";
import {
  NForm,
  NFormItemGi,
  NSelect,
  NGrid,
  NGi,
  NButton,
  NInput,
  useThemeVars,
} from "naive-ui";
import axios from "axios";
import router from "../routes";
import { languageOptions } from "../constants/languageConfig";
import InputAnswerList from "./InputAnswerList.vue";

const MonacoEditor = defineAsyncComponent(() => import("./MonacoEditor.vue"));

export default defineComponent({
  name: "OnlineJudger",
  data: () => ({
    source: languageOptions["c++"].defaultSource,
    language: "cpp",
    sourceFileName: "a.cc",
    compileCmd: "/usr/bin/g++ -std=c++11 -o a a.cc",
    executables: "a",
    runCmd: "a",
    languageOptions: languageOptions,
    selectedOption: "c++",
    themeVars: useThemeVars(),
    inputAnswer: Array.from({ length: 4 }, (_, i) => ({
      input: (i + 1).toString() + " " + (i + 1).toString(),
      answer: (2 * i + 2).toString(),
    })),
  }),
  components: {
    InputAnswerList,
    NForm,
    NFormItemGi,
    NSelect,
    NGrid,
    NGi,
    NButton,
    NInput,
    MonacoEditor,
  },
  methods: {
    submit() {
      axios
        .post("/api/submit", {
          source: this.source,
          language: {
            name: this.language,
            sourceFileName: this.sourceFileName,
            compileCmd: this.compileCmd,
            executables: this.executables,
            runCmd: this.runCmd,
          },
          inputAnswer: this.inputAnswer,
        })
        .then(() => {
          router.push("/submissions");
        });
    },
  },
  watch: {
    selectedOption: function (v) {
      const option = languageOptions[v];
      this.language = option.name;
      this.sourceFileName = option.sourceFileName;
      this.compileCmd = option.compileCmd;
      this.executables = option.executables;
      this.runCmd = option.runCmd;
      this.source = option.defaultSource;
    },
  },
});
</script>

<style scoped>
.editor {
  height: 500px;
  padding-bottom: 24px;
}

.inputs {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
}

.inputs > .input {
  flex: 0 0 33%;
  min-height: unset;
}
</style>
