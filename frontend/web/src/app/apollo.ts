import { ApolloClient, createHttpLink, InMemoryCache, ApolloLink } from '@apollo/client/core'

const gatewayUrl = import.meta.env.VITE_GATEWAY_URL ?? 'http://localhost:8090'

const aflLink = createHttpLink({ uri: `${gatewayUrl}/afl/query` })
const fflLink = createHttpLink({ uri: `${gatewayUrl}/ffl/query` })

const FFL_OPERATIONS = new Set([
  'GetFFLTeamBuilder',
  'GetFFLSeasonClubs',
  'GetFFLClubSeason',
  'GetFFLRoundByAflRound',
  'GetFFLRoundIdsByAflRound',
  'GetFFLSeason',
  'GetFFLRound',
  'GetFFLMatch',
  'GetFFLSeasonPositions',
  'AddFFLPlayerToSeason',
  'RemoveFFLPlayerFromSeason',
  'AddFFLSquadPlayer',
  'SetFFLTeam',
  'GetFFLDataOps',
  'ParseTeamSubmission',
  'ConfirmTeamSubmission',
])

const routingLink = new ApolloLink((operation, forward) => {
  const link = FFL_OPERATIONS.has(operation.operationName) ? fflLink : aflLink
  return link.request(operation, forward)
})

export const apolloClient = new ApolloClient({
  link: routingLink,
  cache: new InMemoryCache(),
})
