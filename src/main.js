import Vue from 'vue';
import App from './App.vue';
import VueMaterial from 'vue-material';
import 'vue-material/dist/vue-material.min.css';
import 'vue-material/dist/theme/default.css';

import router from './routes';

Vue.use(VueMaterial);

Vue.config.productionTip = false;

Vue.filter('cpu', function (value) {
  if (value) {
    return value + ' ms';
  } else {
    return '0 ms';
  }
});

Vue.filter('memory', function (value) {
  if (value) {
    return value + ' kB';
  } else {
    return '0 kB';
  }
});

new Vue({
  render: h => h(App),
  router,
}).$mount('#app');
