import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('../views/DashboardView.vue')
    },
    {
      path: '/survey/:id',
      name: 'survey',
      component: () => import('../views/SurveyView.vue')
    },
    {
      path: '/create-survey',
      name: 'create-survey',
      component: () => import('../views/CreateSurveyView.vue')
    }
  ]
})

export default router
