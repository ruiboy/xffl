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

export const SET_FFL_LINEUP = gql`
  mutation SetFFLLineup($input: SetFFLLineupInput!) {
    setFFLLineup(input: $input) {
      id
      playerSeasonId
      player { id name }
      position
      status
      score
    }
  }
`
