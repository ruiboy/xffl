import gql from 'graphql-tag'

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
