import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import Groups from '../views/Groups.vue'
import Users from '../views/Users.vue'
import Stats from '../views/Stats.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'dashboard', component: Dashboard },
    { path: '/groups', name: 'groups', component: Groups },
    { path: '/users', name: 'users', component: Users },
    { path: '/stats', name: 'stats', component: Stats },
  ],
})
