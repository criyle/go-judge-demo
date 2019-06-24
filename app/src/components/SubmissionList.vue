<template>
  <div class="submission-list-container">
    <md-table class="submission-list" v-model="submission" md-fixed-header
      :md-selected-value="item" @update:mdSelectedValue="onSelect">
      <md-table-toolbar>
        <div class="md-toolbar-section-start"><h1 class="md-title">Submissions</h1></div>
        <div class="md-toolbar-section-end">
          <md-button class="md-icon-button" @click="$emit('loadMore')">
            <md-icon>refresh</md-icon>
          </md-button>
        </div>
      </md-table-toolbar>
      <md-table-row slot="md-table-row" slot-scope="{ item }" md-selectable="single">
        <md-table-cell md-label="Status">{{item.status}}</md-table-cell>
        <md-table-cell md-label="Date">{{item.date | date}}</md-table-cell>
        <md-table-cell md-label="CPU">{{item.time | cpu}}</md-table-cell>
        <md-table-cell md-label="Memory">{{item.memory | memory}}</md-table-cell>
      </md-table-row>
    </md-table>
    <submission-list-detail :showDetail="showDetail" @update:showDetail="closeDetail" :item="item"></submission-list-detail>
  </div>
</template>

<script>
import SubmissionListDetail from './SubmissionListDetail.vue'

export default {
  name: "SubmissionList",
  data: () => {
    return {
      showDetail: false,
      item: null,
    };
  },
  props: ['submission'],
  methods: {
    onSelect(event) {
      this.item = event;
      if (event != null) {
        this.showDetail = true;
      }
    },
    closeDetail(event) {
      this.showDetail = event;
      this.item = null;
    }
  },
  components: {
    SubmissionListDetail,
  }
}
</script>

<style scoped>
  .submission-list-container {
    height: 100%;
  }

  .submission-list {
    height: 100%;
  }
</style>
