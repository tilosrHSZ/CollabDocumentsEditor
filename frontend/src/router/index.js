import { createRouter, createWebHistory } from 'vue-router'
import Login from '../views/Login.vue'
import Home from '../views/Home.vue'
import Editor from '../views/Editor.vue'
import Register from '../views/Register.vue'

const routes = [
  { path: '/login', component: Login },
  { path: '/register', component: Register },
  { path: '/', component: Home },
  { path: '/edit/:id', component: Editor },
  { path: '/profile', component: () => import('../views/Profile.vue') }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  // 统一用 sessionStorage
  const user = sessionStorage.getItem('user') || localStorage.getItem('user')
  
  const whiteList = ['/login', '/register']
  
  if (!user && !whiteList.includes(to.path)) {
    // 没登录且不在白名单 -> 去登录页
    next('/login')
  } else if (user && whiteList.includes(to.path)) {
    // 已登录还想去登录页 -> 回主页
    next('/')
  } else {
    // 正常放行
    next()
  }
})

export default router