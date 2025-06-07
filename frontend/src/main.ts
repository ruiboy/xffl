import './style.css'
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { ApolloClient, createHttpLink, InMemoryCache } from '@apollo/client/core'
import { DefaultApolloClient } from '@vue/apollo-composable'
import PrimeVue from 'primevue/config'
import ConfirmationService from 'primevue/confirmationservice'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Dropdown from 'primevue/dropdown'
import ConfirmDialog from 'primevue/confirmdialog'
import App from './App.vue'
import router from './router'

// Import PrimeVue CSS theme
import 'primevue/resources/themes/aura-dark-lime/theme.css'
import 'primevue/resources/primevue.min.css'
import 'primeicons/primeicons.css'

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
app.use(PrimeVue, { 
  ripple: true
})
app.use(ConfirmationService)
app.provide(DefaultApolloClient, apolloClient)

// Register PrimeVue components globally
app.component('Button', Button)
app.component('DataTable', DataTable)
app.component('Column', Column)
app.component('Dialog', Dialog)
app.component('InputText', InputText)
app.component('Dropdown', Dropdown)
app.component('ConfirmDialog', ConfirmDialog)

app.mount('#app')
