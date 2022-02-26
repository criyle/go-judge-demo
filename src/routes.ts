import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';

const OnlineJudger = () => import("./components/OnlineJudger.vue");
const Submission = () => import('./components/Submission.vue');
const Terminal = () => import('./components/Terminal.vue');

const routes: RouteRecordRaw[] = [
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