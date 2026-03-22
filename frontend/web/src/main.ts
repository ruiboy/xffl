import './app/main.css'
import { createApp } from 'vue'
import { DefaultApolloClient } from '@vue/apollo-composable'
import { apolloClient } from './app/apollo'
import router from './app/router'
import App from './App.vue'

const app = createApp(App)
app.use(router)
app.provide(DefaultApolloClient, apolloClient)
app.mount('#app')
