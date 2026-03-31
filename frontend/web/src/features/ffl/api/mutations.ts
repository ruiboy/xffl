import gql from 'graphql-tag'

export const ADD_FFL_PLAYER_TO_SEASON = gql`
  mutation AddFFLPlayerToSeason($input: AddFFLPlayerToSeasonInput!) {
    addFFLPlayerToSeason(input: $input) {
      id
      playerId
      clubSeasonId
    }
  }
`

export const REMOVE_FFL_PLAYER_FROM_SEASON = gql`
  mutation RemoveFFLPlayerFromSeason($id: ID!) {
    removeFFLPlayerFromSeason(id: $id)
  }
`
