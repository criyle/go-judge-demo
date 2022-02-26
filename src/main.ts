import { createApp } from 'vue';

import App from './App.vue';
import router from './routes';

// General Font
import 'vfonts/Lato.css'
// Monospace Font
import 'vfonts/FiraCode.css'

const app = createApp(App);

app.use(router);

app.mount("#app");
