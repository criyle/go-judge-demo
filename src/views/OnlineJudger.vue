<template>
  <n-form label-placement="top">
    <n-grid :span="24" :x-gap="24">
      <n-form-item-gi :span="12" label="Language">
        <n-select placeholder="Language" v-model:value="selectedOption" :options="Object.entries(languageOptions).map(([k, v]) => ({
          label: k,
          value: k,
        }))
          " />
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
        <monaco-editor class="code-editor-editor" v-model="source" :language="language"
          :theme="themeVars"></monaco-editor>
      </n-gi>

      <n-gi :span="24">
        <input-answer-list v-model="inputAnswer" />
      </n-gi>
    </n-grid>
  </n-form>
</template>

<script setup lang="ts">
import axios from "axios";
import {
  NButton,
  NForm,
  NFormItemGi,
  NGi,
  NGrid,
  NInput,
  NSelect,
  useThemeVars,
} from "naive-ui";
import { defineAsyncComponent, ref, watch } from "vue";
import { useRouter } from "vue-router";
import InputAnswerList from "../components/InputAnswerList.vue";
import { languageOptions } from "../constants/languageConfig";

const MonacoEditor = defineAsyncComponent(() => import("../components/MonacoEditor.vue"));
const router = useRouter();

const source = ref(languageOptions["c++"].defaultSource);
const language = ref("cpp");
const sourceFileName = ref("a.cc");
const compileCmd = ref("/usr/bin/g++ -std=c++11 -o a a.cc");
const executables = ref("a");
const runCmd = ref("a");
const selectedOption = ref("c++");
const themeVars = useThemeVars();
const inputAnswer = ref(Array.from({ length: 4 }, (_, i) => ({
  input: (i + 1).toString() + " " + (i + 1).toString(),
  answer: (2 * i + 2).toString(),
})));

const submit = () => {
  axios
    .post("/api/submit", {
      source: source.value,
      language: {
        name: language.value,
        sourceFileName: sourceFileName.value,
        compileCmd: compileCmd.value,
        executables: executables.value,
        runCmd: runCmd.value,
      },
      inputAnswer: inputAnswer.value,
    })
    .then(() => {
      router.push("/submissions");
    });
}

watch(selectedOption, (v) => {
  const option = languageOptions[v];
  language.value = option.name;
  sourceFileName.value = option.sourceFileName;
  compileCmd.value = option.compileCmd;
  executables.value = option.executables;
  runCmd.value = option.runCmd;
  source.value = option.defaultSource;
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

.inputs>.input {
  flex: 0 0 33%;
  min-height: unset;
}
</style>
