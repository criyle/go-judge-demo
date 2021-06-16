<template>
  <div id="app">
    <sui-menu class="toolbar">
      <router-link to="/" v-slot="{ isExactActive }">
        <sui-menu-item header :active="isExactActive">
          GO Judger
        </sui-menu-item>
      </router-link>

      <router-link to="/submissions" v-slot="{ isExactActive }">
        <sui-menu-item :active="isExactActive">
          Submissions
        </sui-menu-item>
      </router-link>

      <router-link to="/terminal" v-slot="{ isExactActive }">
        <sui-menu-item :active="isExactActive"> Terminal </sui-menu-item>
      </router-link>

      <!-- <sui-menu-item md-label="Submissions" to="/submissions"
        >Submissions</sui-menu-item
      >
      <sui-menu-item md-label="Terminal" to="/terminal">Terminal</sui-menu-item> -->
    </sui-menu>
    <div class="container">
      <router-view v-slot="{ Component }">
        <keep-alive>
          <component :is="Component" />
        </keep-alive>
      </router-view>
    </div>
  </div>
</template>

<style lang="css">
.container {
  width: 100%;
  max-width: 1100px;
  margin: auto;
  margin-top: 10px;
  margin-bottom: 10px;
}
</style>
