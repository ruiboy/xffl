import gql from 'graphql-tag'

export const GET_AFL_ROUND_STATS = gql`
  query GetAFLRoundStats($roundId: ID!) {
    aflRound(id: $roundId) {
      id
      name
      season { id }
      matches {
        id
        dataStatus
        homeClubMatch { id clubSeasonId club { id name } score playerMatches { id } }
        awayClubMatch { id clubSeasonId club { id name } score playerMatches { id } }
      }
    }
  }
`

export const GET_FFL_DATA_OPS = gql`
  query GetFFLDataOps($seasonId: ID!) {
    fflSeason(id: $seasonId) {
      id
      name
      ladder {
        id
        club { id name }
      }
      rounds {
        id
        name
        aflRoundId
        matches {
          id
          homeClubMatch { id club { id name } dataStatus score }
          awayClubMatch { id club { id name } dataStatus score }
        }
      }
    }
  }
`

export const SEARCH_AFL_PLAYERS = gql`
  query SearchAFLPlayers($query: String!) {
    aflPlayerSearch(query: $query) {
      id
      name
      latestPlayerSeason {
        id
        clubSeason {
          id
          club { name }
          season { name }
        }
      }
    }
  }
`
