import {defineAsyncComponent } from 'vue';
import { createRouter, createWebHistory } from 'vue-router';

const OnlineJudger = () => import("./components/OnlineJudger.vue");
const Submission = () => import('./components/Submission.vue');
const Terminal = () => import('./components/Terminal.vue');

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