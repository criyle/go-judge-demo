<template>
  <div id="app">
    <n-config-provider :theme="themeRef">
      <n-global-style />
      <n-loading-bar-provider>
        <n-layout position="absolute" class="root-layout">
          <n-layout-header bordered style="
              display: grid;
              grid-template-columns: 1fr auto;
              padding: 0 32px 0 0;
            ">
            <div>
              <n-menu :value="menuValue" :options="menuOptions" mode="horizontal"
                @update:value="handleMenuUpdateValue" />
            </div>

            <div style="display: flex; align-items: center">
              <n-button @click="handelThemeChange">{{ themeName }}</n-button>
            </div>
          </n-layout-header>

          <n-layout-content :native-scrollbar="false" position="absolute" style="top: 48px; bottom: 0px">
            <div class="container">
              <router-view v-slot="{ Component }">
                <keep-alive>
                  <component :is="Component" />
                </keep-alive>
              </router-view>
            </div>
          </n-layout-content>
        </n-layout>
      </n-loading-bar-provider>
    </n-config-provider>
  </div>
</template>

<script setup lang="ts">
import {
  darkTheme,
  NButton,
  NConfigProvider,
  NGlobalStyle,
  NLayout,
  NLayoutContent,
  NLayoutHeader,
  NLoadingBarProvider,
  NMenu,
  useOsTheme,
} from "naive-ui";
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

const route = useRoute();
const router = useRouter();
const menuOptions = [
  {
    label: "GO Judge",
    key: "home",
    path: "/",
  },
  {
    label: "Submission",
    key: "submission",
    path: "/submissions",
  },
  {
    label: "Terminal",
    key: "terminal",
    path: "/terminal",
  },
];

const menuValue = computed(() => {
  const option = menuOptions.filter((v) => v.path === route.path);
  if (option.length > 0) {
    return option[0].key;
  }
  return "home";
});

const osThemeRef = useOsTheme();
const themeRef = ref(osThemeRef.value === "dark" ? darkTheme : null);

watch(osThemeRef, (newVal) => {
  themeRef.value = newVal === "dark" ? darkTheme : null;
});

const isDarkTheme = () =>
  themeRef.value &&
  themeRef.value.common.baseColor === darkTheme.common.baseColor;

const themeName = computed(() => (isDarkTheme() ? "Dark" : "Light"));

const handelThemeChange = () => {
  themeRef.value = isDarkTheme() ? null : darkTheme;
};

const handleMenuUpdateValue = (_, option) => {
  router.push(option.path);
}
</script>

<style lang="css">
.container {
  width: 100%;
  max-width: 1100px;
  margin: auto;
  margin-top: 10px;
  margin-bottom: 10px;
}
</style>
