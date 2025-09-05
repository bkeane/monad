import { createApp } from 'vue'
import VueTufte from 'vue-tufte'
import App from './App.vue'

const app = createApp(App)
app.use(VueTufte)
app.mount('#app')
