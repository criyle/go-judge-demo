<template>
  <div>
    <div style="display: flex">
      <md-field style="max-width: 240px">
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
const MonacoEditor = () => import("./MonacoEditor.vue");
import axios from "axios";
import router from "../routes.js";

const cSource = `#include <stdio.h>

int main() {
  int a, b;
  scanf("%d%d", &a, &b);
  printf("%d", a + b);
}
`;

const cppSource = `#include <iostream>
using namespace std;

int main() {
  int a, b;
  cin >> a >> b;
  cout << a + b;
}
`;

const goSource = `package main

import "fmt"

func main() {
  var a, b int
  fmt.Scanf("%d%d", &a, &b)
  fmt.Printf("%d", a + b)
}
`;
const nodeSource = `const fs = require('fs');
const [a, b] = fs.readFileSync(0, 'utf-8').split(' ').map(s => parseInt(s));
process.stdout.write(a + b + '\\n');
`;

const pascalSource = `program main;

var a, b: Integer;
begin
  Readln(a, b);
  Writeln(a + b);
end.
`;

const python3Source = `a, b = map(int, input().split())
print(a + b)
`;

const python2Source = `a, b = map(int, raw_input().split())
print a + b
`;

const javaSource = `import java.io.*;
import java.util.*;

public class Main
{
  public static void main(String args[]) throws Exception
  {
    Scanner cin = new Scanner(System.in);
    int a = cin.nextInt(), b = cin.nextInt();
    System.out.println(a + b);
  }
}
`;

const haskellSource = `main = do
  line <- getLine
  let a = (read (takeWhile (/= ' ') line) :: Int)
  let b = (read (drop 1 (dropWhile (/= ' ') line)) :: Int)
  putStrLn (show (a+b)) 
`;

const rustSource = `fn main() {
  let cin = std::io::stdin();
  let mut s = String::new();
  cin.read_line(&mut s).unwrap();
  let values = s
    .split_whitespace()
    .map(|x| x.parse::<i32>())
    .collect::<Result<Vec<i32>, _>>()
    .unwrap();
  println!("{}", values[0]+values[1]);
}
`;

const rubySource = `nums = gets.strip.split(/\\s+/).map(&:to_i)
print nums[0] + nums[1]
`;

const phpSource = `<?
fscanf(STDIN, "%d%d\n", $a, $b);
echo $a + $b;
`;

const csharpSource = `class Program{
  static void Main() {
    string[] input = System.Console.ReadLine().Split(new char[] {' '});
    System.Console.WriteLine(System.Convert.ToInt32(input[0]) + System.Convert.ToInt32(input[1]));
  }
}
`;

const perlSource = `($a, $b) = split(' ', <>);
print $a + $b;
`;

const perl6Source = `my ($a, $b) = split(' ', lines());
print $a + $b;
`;

const ocamlSource = `let input = read_line() in
let num = List.map int_of_string (Str.split (Str.regexp " ") input) in
let s = List.fold_left (+) 0 num in
print_int s
`;

const languageOptions = {
  c: {
    name: "c",
    sourceFileName: "a.c",
    compileCmd: "/usr/bin/gcc -O2 -o a a.c",
    executables: "a",
    runCmd: "a",
    defaultSource: cSource
  },
  "c++": {
    name: "cpp",
    sourceFileName: "a.cc",
    compileCmd: "/usr/bin/g++ -O2 -std=c++11 -o a a.cc",
    executables: "a",
    runCmd: "a",
    defaultSource: cppSource
  },
  go: {
    name: "go",
    sourceFileName: "a.go",
    compileCmd: "/usr/bin/go build -o a a.go",
    executables: "a",
    runCmd: "a",
    defaultSource: goSource
  },
  javascript: {
    name: "javascript",
    sourceFileName: "a.js",
    compileCmd: "/bin/echo compile",
    executables: "a.js",
    runCmd: "/usr/bin/node a.js",
    defaultSource: nodeSource
  },
  typescript: {
    name: "typescript",
    sourceFileName: "a.ts",
    compileCmd: "/usr/bin/tsc a.ts",
    executables: "a.js",
    runCmd: "/usr/bin/node a.js",
    defaultSource: nodeSource
  },
  java: {
    name: "java",
    sourceFileName: "Main.java",
    compileCmd: "/usr/bin/javac Main.java",
    executables: "Main.class",
    runCmd: "/usr/bin/java Main",
    defaultSource: javaSource
  },
  pascal: {
    name: "pascal",
    sourceFileName: "a.pas",
    compileCmd: "/usr/bin/fpc -O2 a.pas",
    executables: "a",
    runCmd: "a",
    defaultSource: pascalSource
  },
  python2: {
    name: "python",
    sourceFileName: "a.py",
    compileCmd:
      "/usr/bin/python -c \"import py_compile; py_compile.compile('a.py', 'a.pyc', doraise=True)\"",
    executables: "a.py a.pyc",
    runCmd: "/usr/bin/python2 a.py",
    defaultSource: python2Source
  },
  python: {
    name: "python",
    sourceFileName: "a.py",
    compileCmd:
      "/usr/bin/python3 -c \"import py_compile; py_compile.compile('a.py', 'a.pyc', doraise=True)\"",
    executables: "a.py a.pyc",
    runCmd: "/usr/bin/python3 a.py",
    defaultSource: python3Source
  },
  haskell: {
    name: "haskell",
    sourceFileName: "a.hs",
    compileCmd: "/usr/bin/ghc -o a a.hs",
    executables: "a",
    runCmd: "a",
    defaultSource: haskellSource
  },
  rust: {
    name: "rust",
    sourceFileName: "a.rs",
    compileCmd: "/usr/bin/rustc -o a a.rs",
    executables: "a",
    runCmd: "a",
    defaultSource: rustSource
  },
  ruby: {
    name: "ruby",
    sourceFileName: "a.rb",
    compileCmd: "/bin/echo compiled",
    executables: "a.rb",
    runCmd: "/usr/bin/ruby a.rb",
    defaultSource: rubySource
  },
  php: {
    name: "php",
    sourceFileName: "a.php",
    compileCmd: "/bin/echo compiled",
    executables: "a.php",
    runCmd: "/usr/bin/php a.php",
    defaultSource: phpSource
  },
  "c#": {
    name: "csharp",
    sourceFileName: "a.cs",
    compileCmd: "/usr/bin/mcs -optimize+ -out:a a.cs",
    executables: "a",
    runCmd: "/usr/bin/mono a",
    defaultSource: csharpSource
  },
  perl: {
    name: "perl",
    sourceFileName: "a.pl",
    compileCmd: "/bin/echo compiled",
    executables: "a.pl",
    runCmd: "/usr/bin/perl a.pl",
    defaultSource: perlSource
  },
  perl6: {
    name: "perl",
    sourceFileName: "a.pl",
    compileCmd: "/bin/echo compiled",
    executables: "a.pl",
    runCmd: "/usr/bin/perl6 a.pl",
    defaultSource: perl6Source
  },
  ocaml: {
    name: "ocaml",
    sourceFileName: "a.ml",
    compileCmd: "/usr/bin/ocamlc str.cma -o a a.ml",
    executables: "a",
    runCmd: "a",
    defaultSource: ocamlSource
  }
};

export default {
  name: "OnlineJudger",
  data: () => ({
    source: cppSource,
    language: "cpp",
    sourceFileName: "a.cc",
    compileCmd: "/usr/bin/g++ -std=c++11 -o a a.cc",
    executables: "a",
    runCmd: "a",
    languageOptions: languageOptions,
    selectedOption: "c++"
  }),
  components: {
    MonacoEditor
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
            runCmd: this.runCmd
          }
        })
        .then(() => {
          router.push("/submissions");
        });
    }
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
    }
  }
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
