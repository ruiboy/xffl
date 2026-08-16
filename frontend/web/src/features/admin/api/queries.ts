import gql from 'graphql-tag'

// Reference data for the season builder: the club registry (pick existing clubs
// by id) and the selectable scoring eras.
export const GET_FFL_BUILDER_REFS = gql`
  query FFLBuilderRefs {
    fflClubs {
      id
      name
    }
    fflRulesEras {
      id
      label
    }
  }
`

// The list of FFL seasons for the fixture-builder season picker.
export const GET_FFL_SEASON_LIST = gql`
  query FFLSeasonList {
    fflSeasons {
      id
      name
    }
  }
`

// All FFL seasons with their rules era and clubs, for the Seasons admin list.
export const GET_FFL_SEASONS_ADMIN = gql`
  query FFLSeasonsAdmin {
    fflSeasons {
      id
      name
      rulesId
      ladder {
        club {
          name
        }
      }
    }
  }
`

// A season's context for the fixture builder: its club_seasons (id + name) and
// the AFL season's rounds (for the per-round AFL round selector).
export const GET_FFL_BUILDER_SEASON = gql`
  query FFLBuilderSeason($id: ID!) {
    fflSeason(id: $id) {
      id
      name
      ladder {
        id
        club {
          id
          name
        }
      }
      aflSeason {
        id
        name
        rounds {
          id
          name
        }
      }
    }
  }
`

// A season's saved rounds with their fixtures, byes and lock state.
export const GET_FFL_SEASON_FIXTURES = gql`
  query FFLSeasonFixtures($seasonId: ID!) {
    fflSeasonFixtures(seasonId: $seasonId) {
      roundId
      name
      aflRoundId
      roundType
      locked
      matches {
        style
        clubSeasonIds
      }
    }
  }
`
