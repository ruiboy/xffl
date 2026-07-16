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

export const SAVE_FFL_FIXTURES = gql`
  mutation SaveFFLFixtures($input: SaveFFLFixturesInput!) {
    saveFFLFixtures(input: $input)
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
