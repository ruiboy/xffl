import './style.css'
import typescriptLogo from './typescript.svg'
import viteLogo from '/vite.svg'
import { setupCounter } from './counter.ts'
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { ApolloClient, createHttpLink, InMemoryCache } from '@apollo/client/core'
import { DefaultApolloClient } from '@vue/apollo-composable'
import App from './App.vue'
import router from './router'

document.querySelector<HTMLDivElement>('#app')!.innerHTML = `
  <div>
    <a href="https://vite.dev" target="_blank">
      <img src="${viteLogo}" class="logo" alt="Vite logo" />
    </a>
    <a href="https://www.typescriptlang.org/" target="_blank">
      <img src="${typescriptLogo}" class="logo vanilla" alt="TypeScript logo" />
    </a>
    <h1>Vite + TypeScript</h1>
    <div class="card">
      <button id="counter" type="button"></button>
    </div>
    <p class="read-the-docs">
      Click on the Vite and TypeScript logos to learn more
    </p>
  </div>
`

setupCounter(document.querySelector<HTMLButtonElement>('#counter')!)

// Create Apollo Client
const httpLink = createHttpLink({
  uri: 'http://localhost:8080/query',
  credentials: 'include',
  headers: {
    'Content-Type': 'application/json',
  },
})

const apolloClient = new ApolloClient({
  link: httpLink,
  cache: new InMemoryCache(),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: 'no-cache',
    },
  },
})

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.provide(DefaultApolloClient, apolloClient)

app.mount('#app')
