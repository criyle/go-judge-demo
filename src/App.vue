<template>
  <div id="app">
    <md-toolbar class="toolbar">
      <router-link to="/"><h3 class="md-title">GO Judger</h3></router-link>
      <md-tabs md-sync-route>
        <md-tab md-label="Submissions" to="/submissions"></md-tab>
      </md-tabs>
    </md-toolbar>
    <div class="container">
      <router-view></router-view> 
    </div>
  </div>
</template>

<style lang="scss">
@import "~vue-material/dist/theme/engine"; // Import the theme engine
@include md-register-theme(
  "default",
  (
    primary: md-get-palette-color(blue, A200),
    accent: md-get-palette-color(red, A200),
    theme: light
  )
);
@import "~vue-material/dist/theme/all"; // Apply the theme
#app {
  font-family: "Avenir", Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
}

.container {
  width: 100%;
  max-width: 1100px;
  margin: auto;
  margin-top: 10px;
  margin-bottom: 10px;
}

.toolbar {
  min-height: 48px !important;
}
</style>
