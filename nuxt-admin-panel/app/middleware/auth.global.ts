export default defineNuxtRouteMiddleware((to) => {
  const token = useCookie('tf_token')

  if (to.path !== '/login' && !token.value) {
    return navigateTo('/login')
  }
  if (to.path === '/login' && token.value) {
    return navigateTo('/')
  }
})
