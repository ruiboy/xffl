import gql from 'graphql-tag'

export const UPDATE_FFL_PLAYER_SEASON = gql`
  mutation UpdateFFLPlayerSeason($input: UpdateFFLPlayerSeasonInput!) {
    updateFFLPlayerSeason(input: $input) {
      id
      notes
      costCents
    }
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
  mutation RemoveFFLPlayerFromSeason($id: ID!, $toRoundId: ID!) {
    removeFFLPlayerFromSeason(id: $id, toRoundId: $toRoundId)
  }
`

export const ADD_FFL_SQUAD_PLAYER = gql`
  mutation AddFFLSquadPlayer($input: AddFFLSquadPlayerInput!) {
    addFFLSquadPlayer(input: $input) {
      id
      clubSeasonId
    }
  }
`

export const SET_FFL_TEAM = gql`
  mutation SetFFLTeam($input: SetFFLTeamInput!) {
    setFFLTeam(input: $input) {
      id
      playerSeasonId
      player { id name }
      position
      status
      backupPositions
      interchangePosition
      score
    }
  }
`
