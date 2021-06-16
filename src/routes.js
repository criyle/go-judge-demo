import { createRouter, createWebHashHistory, createWebHistory } from 'vue-router';

import OnlineJudger from './components/OnlineJudger.vue';
import Submission from './components/Submission.vue';
import Terminal from './components/Terminal.vue';

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
        path: '/terminal',
        component: Terminal,
    },
    {
        path: '/:pathMatch(.*)',
        redirect: '/',
    },
];

const router = createRouter({
    history: createWebHistory(),
    routes,
});

export default router;