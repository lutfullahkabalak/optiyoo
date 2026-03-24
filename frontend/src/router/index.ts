import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/auth',
      name: 'auth',
      component: HomeView
    },
    {
      path: '/',
      name: 'dashboard',
      component: () => import('../views/DashboardView.vue')
    },
    {
      path: '/s/:id',
      name: 'survey',
      component: () => import('../views/SurveyView.vue')
    }
  ]
})

export default router
