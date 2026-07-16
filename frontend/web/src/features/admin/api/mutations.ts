import gql from 'graphql-tag'

// ---- Seasons + Fixtures builder ----

export const BUILD_FFL_SEASON = gql`
  mutation BuildFFLSeason($input: BuildFFLSeasonInput!) {
    buildFFLSeason(input: $input) {
      seasonId
      rulesId
      clubSeasons { clubName clubSeasonId }
    }
  }
`

export const ADD_FFL_ROUND = gql`
  mutation AddFFLRound($input: AddFFLRoundInput!) {
    addFFLRound(input: $input) { roundId name }
  }
`

export const ADD_FFL_FIXTURE = gql`
  mutation AddFFLFixture($input: AddFFLFixtureInput!) {
    addFFLFixture(input: $input) { matchId homeClubMatchId awayClubMatchId }
  }
`

export const GENERATE_FFL_HOME_AND_AWAY = gql`
  mutation GenerateFFLHomeAndAway($input: GenerateFFLHomeAndAwayInput!) {
    generateFFLHomeAndAway(input: $input) { roundId name }
  }
`

// ---- Calculate (ladder recompute) ----

export const RECALCULATE_AFL_LADDER = gql`
  mutation RecalculateAFLLadder($seasonId: ID!) {
    recalculateAFLLadder(seasonId: $seasonId)
  }
`

export const RECALCULATE_FFL_LADDER = gql`
  mutation RecalculateFFLLadder($seasonId: ID!) {
    recalculateFFLLadder(seasonId: $seasonId)
  }
`
