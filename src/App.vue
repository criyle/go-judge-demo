<template>
<div id="app">
  <OnlineJudger />
</div>
</template>

<script>
import OnlineJudger from './components/OnlineJudger.vue'

export default {
  name: 'app',
  components: {
    OnlineJudger
  }
}
</script>

<style lang="scss">
@import "~vue-material/dist/theme/engine"; // Import the theme engine

@include md-register-theme("default", (primary: md-get-palette-color(blue, A200), accent: md-get-palette-color(red, A200), theme: light));

@import "~vue-material/dist/theme/all"; // Apply the theme

#app {
    font-family: 'Avenir', Helvetica, Arial, sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    color: #2c3e50;
    height: 100%;
    overflow-y: hidden;
}
</style>
