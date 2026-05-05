import gql from 'graphql-tag'

export const IMPORT_AFL_MATCH_STATS = gql`
  mutation ImportAFLMatchStats($matchId: ID!) {
    importAFLMatchStats(matchId: $matchId) {
      matchId
      homeClubName
      awayClubName
      homePlayerCount
      awayPlayerCount
      unmatchedPlayers {
        parsedName
        clubMatchId
        kicks handballs marks hitouts tackles goals behinds
      }
    }
  }
`

export const MARK_AFL_MATCH_STATS_COMPLETE = gql`
  mutation MarkAFLMatchStatsComplete($matchId: ID!, $complete: Boolean!) {
    markAFLMatchStatsComplete(matchId: $matchId, complete: $complete) {
      id
      dataStatus
    }
  }
`

export const RESOLVE_AFL_PLAYER_MATCH = gql`
  mutation ResolveAFLPlayerMatch($input: ResolveAFLPlayerMatchInput!) {
    resolveAFLPlayerMatch(input: $input) {
      id
    }
  }
`

export const ADD_AFL_PLAYER = gql`
  mutation AddAFLPlayer($input: AddAFLPlayerInput!) {
    addAFLPlayer(input: $input) {
      id
    }
  }
`

export const ADD_AFL_PLAYER_SEASON = gql`
  mutation AddAFLPlayerSeason($input: AddAFLPlayerSeasonInput!) {
    addAFLPlayerSeason(input: $input) {
      id
    }
  }
`

export const PARSE_TEAM_SUBMISSION = gql`
  mutation ParseFFLTeamSubmission($input: ParseFFLTeamSubmissionInput!) {
    parseFFLTeamSubmission(input: $input) {
      resolvedPlayers {
        parsedName
        clubHint
        resolvedName
        resolvedClub
        position
        backupPositions
        interchangePosition
        score
        notes
        playerSeasonId
        confidence
      }
      needsReview
    }
  }
`

export const CONFIRM_TEAM_SUBMISSION = gql`
  mutation ConfirmFFLTeamSubmission($input: ConfirmFFLTeamSubmissionInput!) {
    confirmFFLTeamSubmission(input: $input) {
      id
      playerSeasonId
      player { id aflPlayer { name } }
      position
      backupPositions
      interchangePosition
      score
    }
  }
`