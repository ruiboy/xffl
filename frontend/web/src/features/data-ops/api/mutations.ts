import gql from 'graphql-tag'

export const CLEAR_FFL_FORUM_CAPTURES = gql`
  mutation ClearFFLForumCaptures {
    clearFFLForumCaptures
  }
`

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
      match {
        id
        dataStatus
        homeClubMatch { id score playerMatches { id } }
        awayClubMatch { id score playerMatches { id } }
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

export const ADD_FFL_PLAYER_TO_SEASON = gql`
  mutation AddFFLPlayerToSeasonForDataOps($input: AddFFLPlayerToSeasonInput!) {
    addFFLPlayerToSeason(input: $input) {
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

export const MARK_FFL_TEAM_FINAL = gql`
  mutation MarkFFLTeamFinal($input: MarkFFLTeamFinalInput!) {
    markFFLTeamFinal(input: $input)
  }
`

export const MARK_FFL_TEAM_SUBMITTED = gql`
  mutation MarkFFLTeamSubmitted($input: MarkFFLTeamFinalInput!) {
    markFFLTeamSubmitted(input: $input)
  }
`

export const RECALCULATE_FFL_CLUB_MATCH_SCORE = gql`
  mutation RecalculateFFLClubMatchScore($clubMatchId: ID!) {
    recalculateFFLClubMatchScore(clubMatchId: $clubMatchId)
  }
`
