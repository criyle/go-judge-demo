<template>
  <div>
    <div style="display: flex">
      <md-field
        style="max-width: 240px"
      >
        <label>Language</label>
        <md-select v-model="selectedOption">
          <md-option 
            v-for="(option, name) in languageOptions"
            :key="name"
            :value="name"
          >
            {{name}}
          </md-option>
        </md-select>
      </md-field>
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

const languageOptions = {
  "cpp": {
    name: "cpp",
    sourceFileName: "a.cc",
    compileCmd: "/usr/bin/g++ -std=c++11 -o a a.cc",
    executables: "a",
    runCmd: "a",
    defaultSource: `#include <iostream>
using namespace std;

int main() {
  int a, b;
  cin >> a >> b;
  cout << a + b;
}`,
  },
  "python3": {
    name: "python",
    sourceFileName: "a.py",
    compileCmd: "/bin/echo compile",
    executables: "a.py",
    runCmd: "/usr/bin/python3 a.py",
    defaultSource: `a, b = map(int, input().split())
print(a + b)`,
  },
  "python2": {
    name: "python",
    sourceFileName: "a.py",
    compileCmd: "/bin/echo compile",
    executables: "a.py",
    runCmd: "/usr/bin/python2 a.py",
    defaultSource: `a, b = map(int, raw_input().split())
print a + b`,
  },
  "c": {
    name: "c",
    sourceFileName: "a.c",
    compileCmd: "/usr/bin/gcc -o a a.c",
    executables: "a",
    runCmd: "a",
    defaultSource: `#include <stdio.h>

int main() {
  int a, b;
  scanf("%d%d", &a, &b);
  printf("%d", a + b);
}`,
  },
};

export default {
  name: "OnlineJudger",
  data: () => ({
    source: defaultCode,
    language: "cpp",
    sourceFileName: "a.cc",
    compileCmd: "/usr/bin/g++ -std=c++11 -o a a.cc",
    executables: "a",
    runCmd: "a",
    languageOptions: languageOptions,
    selectedOption: "cpp",
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
  watch: {
    selectedOption: function(v) {
      const option = languageOptions[v];
      this.language = option.name;
      this.sourceFileName = option.sourceFileName;
      this.compileCmd = option.compileCmd;
      this.executables = option.executables;
      this.runCmd = option.runCmd;
      this.source = option.defaultSource;
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
