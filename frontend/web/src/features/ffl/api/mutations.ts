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
      clubSeasonId
      aflPlayerSeasonId
      fromRoundId
    }
  }
`

export const REMOVE_FFL_PLAYER_FROM_SEASON = gql`
  mutation RemoveFFLPlayerFromSeason($input: RemoveFFLPlayerFromSeasonInput!) {
    removeFFLPlayerFromSeason(input: $input)
  }
`

export const SET_FFL_TEAM = gql`
  mutation SetFFLTeam($input: SetFFLTeamInput!) {
    setFFLTeam(input: $input) {
      id
      playerSeasonId
      player { id aflPlayer { id name } }
      position
      status
      backupPositions
      interchangePosition
      score
    }
  }
`

export const ADD_AFL_PLAYER = gql`
  mutation AddAFLPlayerForFFL($input: AddAFLPlayerInput!) {
    addAFLPlayer(input: $input) {
      id
    }
  }
`
