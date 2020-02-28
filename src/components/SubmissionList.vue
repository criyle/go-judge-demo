<template>
  <div class="submission-list-container md-content">
    <md-list class="md-elevation-1">
      <md-list-item
        md-expand
        v-for="s in submission"
        :key="s.id"
        :md-expanded.sync="s.expanded"
      >
        <div class="md-list-item-text list-item">
          <span>
            <span class="status">{{(s.status)}}</span>
            <span>{{s.date | date}}</span>
          </span>
        </div>
        <div slot="md-expand">
          <div
            v-if="s.expanded"
            class="info properties"
          >
            <property-view
              label="_id"
              :value="s.id"
            ></property-view>
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
          <div v-if="s.expanded">
            <div
              v-for="(u, index) in s.results"
              :key="index"
            >
              <md-divider></md-divider>
              <div class="info properties">
                <property-view
                  label="cpu"
                  :value="u.time | cpu"
                ></property-view>
                <property-view
                  label="memory"
                  :value="u.memory | memory"
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
      </md-list-item>
    </md-list>
    <md-button
      class="md-raised"
      @click="$emit('loadMore')"
    >More</md-button>
  </div>
</template>

<script>
import CodeView from "./CodeView.vue";
import PropertyView from "./PropertyView";

export default {
  name: "SubmissionList",
  props: ["submission"],
  components: {
    CodeView,
    PropertyView
  }
};
</script>

<style scoped>
.list-item span.status {
  width: 180px;
}

.list-item > span > span {
  display: inline-block;
  width: auto;
  font-size: 14px;
  padding-right: 10px;
}

.info {
  padding: 4px 16px;
}

.properties {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
}

.properties > div {
  padding-right: 10px;
}
</style>
