<template>
<div class="submission-list-container md-content">
  <md-list>
    <md-list-item class="md-elevation-1" md-expand v-for="s in submission" :key="s.id">
      <div class="md-list-item-text list-item">
        <span>
          <span class="status">{{getStatus(s)}}</span>
          <span>{{s.date | date}}</span>
          <span>{{s.time | cpu}}</span>
          <span>{{s.memory | memory}}</span>
        </span>
      </div>
      <md-list slot="md-expand">
        <md-list-item>
          <div class="md-list-item-text">
            <span>_id: {{s.id}}</span>
            <span>
              <md-field>
                <label>Code</label>
                <md-textarea :value="s.code" md-autogrow disabled></md-textarea>
              </md-field>
            </span>
          </div>
        </md-list-item>
        <template v-for="u in s.update">
          <md-divider></md-divider>
          <md-list-item>
            <div class="md-list-item-text">
              <span>status: {{u.status}} </span>
              <span>date: {{u.date | date}} </span>
              <span>cpu: {{u.time | cpu}} </span>
              <span>memory: {{u.memory | memory}} </span>
              <span v-if="u.stdin">
                <md-field>
                  <label>Stdin</label>
                  <md-textarea :value="u.stdin" md-autogrow disabled></md-textarea>
                </md-field>
              </span>
              <span v-if="u.stdout">
                <md-field>
                  <label>Stdout</label>
                  <md-textarea :value="u.stdout" md-autogrow disabled></md-textarea>
                </md-field>
              </span>
              <span v-if="u.stderr">
                <md-field>
                  <label>Stderr</label>
                  <md-textarea :value="u.stderr" md-autogrow disabled></md-textarea>
                </md-field>
              </span>
              <span v-if="u.log">
                <md-field>
                  <label>Log</label>
                  <md-textarea :value="u.log" md-autogrow disabled></md-textarea>
                </md-field>
              </span>
            </div>
          </md-list-item>
        </template>
      </md-list>
    </md-list-item>
    <md-list-item>
      <md-button class="md-raised" @click="$emit('loadMore')">More</md-button>
    </md-list-item>
  </md-list>
</div>
</template>

<script>
export default {
  name: "SubmissionList",
  data: () => ({}),
  props: ['submission'],
  methods: {
    getStatus: function(s) {
      if (s.update && s.update.length > 0) {
        return s.update[s.update.length - 1].status;
      }
      return 'Submitted';
    },
  },
}
</script>

<style scoped>
.submission-list-container {
  height: 100%;
  overflow-y: scroll;
}

.list-item span.status {
  width: 120px;
}

.list-item>span>span {
  display: inline-block;
  width: auto;
  font-size: 14px;
  padding-right: 10px;
}
</style>
