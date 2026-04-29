import { ApolloClient, createHttpLink, InMemoryCache } from '@apollo/client/core'

const gatewayUrl = import.meta.env.VITE_GATEWAY_URL ?? 'http://localhost:8090'

export const apolloClient = new ApolloClient({
  link: createHttpLink({ uri: `${gatewayUrl}/query` }),
  cache: new InMemoryCache(),
})
