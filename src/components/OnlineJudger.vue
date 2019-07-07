<template>
<div class="md-layout">
  <div class="md-layout-item md-size-50">
    <code-editor v-model="code" @submit="submit"></code-editor>
  </div>
  <div class="md-layout-item md-size-50 submission">
    <submission-list :submission="submission" @loadMore="loadMore"></submission-list>
  </div>
</div>
</template>

<script>
import CodeEditor from './CodeEditor.vue';
import SubmissionList from './SubmissionList.vue';
import axios from 'axios';

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
    code: defaultCode,
    submission: [],
  }),
  components: {
    CodeEditor,
    SubmissionList,
  },
  methods: {
    submit() {
      this.$ws.send(JSON.stringify({
        "language": "c++",
        "code": this.code,
      }));
    },
    loadMore() {
      const p = this.submission.length > 0 ? {
        id: this.submission[this.submission.length - 1].id
      } : {};
      axios.get('/api/submission', {
        params: p,
      }).then(r => {
        this.submission.push(...r.data);
      });
    }
  },
  mounted: function() {
    const url = (location.protocol == 'https:' ? 'wss' : 'ws') + '://' +
      document.domain + ':' + location.port + '/ws';
    const ws = new WebSocket(url)
    ws.addEventListener('message', event => {
      const data = JSON.parse(event.data);
      const sub = this.submission.find(s => s.id === data.id);
      if (sub) {
        if (sub.update) {
          sub.update.push(data.update);
        } else {
          sub.update = [data.update];
        }
        this.submission = [...this.submission];
      } else {
        this.submission.splice(0, 0, {
          ...data,
        });
      }
    });
    this.$ws = ws;
  },
  created: function() {
    this.loadMore();
  },
  beforeDestory: function() {
    this.$ws.close();
  },
}
</script>

<style scoped>
.md-layout {
  height: 100%;
  overflow-y: hidden;
}

.md-layout-item {
  height: 100%;
}

.submission {
  padding-left: 5px;
  padding-right: 5px;
}
</style>
