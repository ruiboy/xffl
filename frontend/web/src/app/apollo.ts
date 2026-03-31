import { ApolloClient, createHttpLink, InMemoryCache, ApolloLink } from '@apollo/client/core'

const aflLink = createHttpLink({ uri: 'http://localhost:8090/afl/query' })
const fflLink = createHttpLink({ uri: 'http://localhost:8090/ffl/query' })

const routingLink = new ApolloLink((operation, forward) => {
  const isFFL = operation.query.definitions.some(
    def => def.kind === 'OperationDefinition' &&
      def.selectionSet.selections.some(
        sel => sel.kind === 'Field' && sel.name.value.startsWith('ffl')
      )
  )
  return isFFL ? fflLink.request(operation, forward) : aflLink.request(operation, forward)
})

export const apolloClient = new ApolloClient({
  link: routingLink,
  cache: new InMemoryCache(),
})
