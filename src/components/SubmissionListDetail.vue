<template>
  <div>
    <md-dialog :md-active="showDetail" @update:mdActive="$emit('update:showDetail',false)" class="dialog">
      <md-dialog-title>Details</md-dialog-title>
      <md-tabs v-if="item" md-dynamic-height>
        <md-tab md-label="Details">
          <div>_id: {{item.id}} </div>
          <div>date: {{item.date | date}} </div>
          <div>cpu: {{item.time | cpu}} </div>
          <div>memory: {{item.memory | memory}} </div>
        </md-tab>
        <md-tab md-label="Code">
          <md-field>
            <label>Code</label>
            <md-textarea :value="item.code" md-autogrow disabled></md-textarea>
          </md-field>
        </md-tab>
        <md-tab md-label="Stdin">
          <md-field>
            <label>Stdin</label>
            <md-textarea :value="item.stdin" md-autogrow disabled></md-textarea>
          </md-field>
        </md-tab>
        <md-tab md-label="Stdout">
          <md-field>
            <label>Stdout</label>
            <md-textarea :value="item.stdout" md-autogrow disabled></md-textarea>
          </md-field>
        </md-tab>
        <md-tab md-label="Stderr">
          <md-field>
            <label>Stderr</label>
            <md-textarea :value="item.stderr" md-autogrow disabled></md-textarea>
          </md-field>
        </md-tab>
      </md-tabs>
      <md-dialog-actions>
        <md-button @click="$emit('update:showDetail',false)">Close</md-button>
      </md-dialog-actions>
    </md-dialog>
  </div>
</template>

<script>

export default {
  name: "SubmissionListDetail",
  props: [
    "showDetail", "item",
  ]
}
</script>

<style scoped>
  .dialog {
    background-color: white;
  }
</style>
