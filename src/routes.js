import Vue from 'vue';
import VueRouter from 'vue-router';

import OnlineJudger from './components/OnlineJudger.vue';
import Submission from './components/Submission.vue';

Vue.use(VueRouter);

const routes = [
    {
        path: '/',
        component: OnlineJudger,
    },
    {
        path: '/submissions',
        component: Submission,
    },
    {
        path: '*',
        redirect: '/',
    },
];

const router = new VueRouter({
    mode: 'history',
    routes,
});

export default router;