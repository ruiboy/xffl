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
