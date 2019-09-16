<template>
  <div>
    <div style="display: flex">
      <div style="flex: 1 0 "></div>
      <md-button 
        class="md-raised" 
        @click="submit"
      >Submit</md-button>
    </div>
    <div class="inputs">
      <md-field class="input">
        <label>Language Name</label>
        <md-input v-model="language"></md-input>
      </md-field>
      <md-field class="input">
        <label>Source File Name</label>
        <md-input v-model="sourceFileName"></md-input>
      </md-field>
      <md-field class="input">
        <label>Compile Cmd</label>
        <md-input v-model="compileCmd"></md-input>
      </md-field>
      <md-field class="input">
        <label>Executable File Name</label>
        <md-input v-model="executables"></md-input>
      </md-field>
      <md-field class="input">
        <label>Exec Cmd</label>
        <md-input v-model="runCmd"></md-input>
      </md-field>
    </div>
    <monaco-editor
      class="code-editor-editor md-elevation-1 editor"
      v-model="source"
      :language="language"
    ></monaco-editor>
  </div>
</template>

<script>
import MonacoEditor from "./MonacoEditor.vue";
import axios from "axios";
import router from "../routes.js";

const defaultCode = `#include <iostream>
using namespace std;

int main() {
  int a, b;
  cin >> a >> b;
  cout << a + b;
}`;

export default {
  name: "OnlineJudger",
  data: () => ({
    source: defaultCode,
    language: "cpp",
    sourceFileName: "a.cc",
    compileCmd: "/usr/bin/g++ -o a a.cc",
    executables: "a",
    runCmd: "a",
  }),
  components: {
    MonacoEditor
  },
  methods: {
    submit() {
      axios.post('/api/submit', {
        source: this.source,
        language: {
          name: this.language,
          sourceFileName: this.sourceFileName,
          compileCmd: this.compileCmd,
          executables: this.executables,
          runCmd: this.runCmd,
        }
      }).then(() => {
        router.push("/submissions");
      })
    },
  },
};
</script>

<style scoped>
.editor {
  height: 500px;
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
