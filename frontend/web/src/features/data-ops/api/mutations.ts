import gql from 'graphql-tag'

export const PARSE_TEAM_SUBMISSION = gql`
  mutation ParseTeamSubmission($input: ParseTeamSubmissionInput!) {
    parseTeamSubmission(input: $input) {
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
  mutation ConfirmTeamSubmission($input: ConfirmTeamSubmissionInput!) {
    confirmTeamSubmission(input: $input) {
      id
      playerSeasonId
      player { id name }
      position
      backupPositions
      interchangePosition
      score
    }
  }
`
