import { createRouter, createWebHistory } from 'vue-router';
import LoginView from '../views/LoginView.vue';
import MainView from '../views/MainView.vue';
import RegisterView from '../views/RegisterView.vue'; // 新增

const routes = [
  {
    path: '/',
    name: 'Login',
    component: LoginView,
  },
  {
    path: '/register',
    name: 'Register',
    component: RegisterView, // 新增注册页面路由
  },
  {
    path: '/main',
    name: 'Main',
    component: MainView,
    meta: { requiresAuth: true },
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

// 路由守卫：检查是否登录
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token');
  if (to.meta.requiresAuth && !token) {
    next('/');
  } else {
    next();
  }
});

export default router;