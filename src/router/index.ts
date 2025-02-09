import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        component: () => import("../views/OnlineJudger.vue"),
    },
    {
        path: '/submissions',
        component: () => import('../views/Submission.vue'),
    },
    {
        path: '/terminal',
        component: () => import('../views/Terminal.vue'),
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