import gql from 'graphql-tag'

export const CREATE_FFL_PLAYER = gql`
  mutation CreateFFLPlayer($input: CreateFFLPlayerInput!) {
    createFFLPlayer(input: $input) {
      id
      name
    }
  }
`

export const UPDATE_FFL_PLAYER = gql`
  mutation UpdateFFLPlayer($input: UpdateFFLPlayerInput!) {
    updateFFLPlayer(input: $input) {
      id
      name
    }
  }
`

export const DELETE_FFL_PLAYER = gql`
  mutation DeleteFFLPlayer($id: ID!) {
    deleteFFLPlayer(id: $id)
  }
`

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
