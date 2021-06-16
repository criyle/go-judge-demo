import { createApp } from 'vue';
import App from './App.vue';

import FomanticUI from 'vue-fomantic-ui'
import 'fomantic-ui-css/semantic.min.css'

import router from './routes';

const app = createApp(App);

app.use(router);
app.use(FomanticUI);

app.mount("#app");
