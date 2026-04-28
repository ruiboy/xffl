import gql from 'graphql-tag'

export const GET_AFL_ROUND_STATS = gql`
  query GetAFLRoundStats($roundId: ID!) {
    aflRound(id: $roundId) {
      id
      name
      season { id }
      matches {
        id
        statsImportStatus
        statsImportedAt
        homeClubMatch { id club { id name } score playerMatches { id } }
        awayClubMatch { id club { id name } score playerMatches { id } }
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
        matches {
          id
          homeClubMatch { id club { id name } }
          awayClubMatch { id club { id name } }
        }
      }
    }
  }
`
