import { ApolloClient, createHttpLink, InMemoryCache, ApolloLink } from '@apollo/client/core'

const aflLink = createHttpLink({ uri: 'http://localhost:8090/afl/query' })
const fflLink = createHttpLink({ uri: 'http://localhost:8090/ffl/query' })

const FFL_OPERATIONS = new Set([
  'GetFFLTeamBuilder',
  'GetFFLSeasonClubs',
  'GetFFLClubSeason',
  'GetFFLLatestRound',
  'GetFFLSeason',
  'AddFFLPlayerToSeason',
  'RemoveFFLPlayerFromSeason',
  'AddFFLSquadPlayer',
  'SetFFLLineup',
])

const routingLink = new ApolloLink((operation, forward) => {
  const link = FFL_OPERATIONS.has(operation.operationName) ? fflLink : aflLink
  return link.request(operation, forward)
})

export const apolloClient = new ApolloClient({
  link: routingLink,
  cache: new InMemoryCache(),
})
