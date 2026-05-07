import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './styles.css'

const app = createApp(App)

app.config.errorHandler = (err, _instance, info) => {
  const payload = err instanceof Error
    ? { name: err.name, message: err.message, stack: err.stack }
    : { value: String(err) }
  // eslint-disable-next-line no-console
  console.error('[vue:errorHandler]', info, payload)
}

window.addEventListener('error', event => {
  // eslint-disable-next-line no-console
  console.error('[window.error]', event.message, {
    filename: event.filename,
    line: event.lineno,
    column: event.colno,
    stack: event.error?.stack,
  })
})

window.addEventListener('unhandledrejection', event => {
  const reason = event.reason
  const payload = reason instanceof Error
    ? { name: reason.name, message: reason.message, stack: reason.stack }
    : { value: String(reason) }
  // eslint-disable-next-line no-console
  console.error('[unhandledrejection]', payload)
})

app.use(createPinia()).use(router).mount('#app')
