<template>
  <div id="app">
    <OnlineJudger/>
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

<style>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
  height: 100%;
  overflow-y: hidden;
}
</style>
