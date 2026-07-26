import gql from 'graphql-tag'

export const GET_AFL_ROUND_STATS = gql`
  query GetAFLRoundStats($roundId: ID!) {
    aflRound(id: $roundId) {
      id
      name
      matches {
        id
        dataStatus
        homeClubMatch { id clubSeasonId club { id name } score playerMatches { id } }
        awayClubMatch { id clubSeasonId club { id name } score playerMatches { id } }
      }
      byes {
        id
        club { id name }
      }
    }
  }
`

export const GET_FFL_DATA_OPS = gql`
  query GetFFLDataOps($seasonId: ID!) {
    fflSeason(id: $seasonId) {
      id
      name
      ladder {
        id
        club { id name }
      }
      rounds {
        id
        name
        aflRoundId
        matches {
          id
          matchStyle
          clubMatches { id clubSeasonId club { id name } dataStatus score }
        }
      }
    }
  }
`

export const GET_AFL_SEASON_CLUB_SEASONS = gql`
  query GetAFLSeasonClubSeasonsForDataOps($fflSeasonId: ID!) {
    fflSeason(id: $fflSeasonId) {
      aflSeason {
        id
        ladder {
          id
          club { name }
        }
      }
    }
  }
`

export const GET_FFL_CAPTURED_PAGES = gql`
  query GetFFLCapturedPages {
    fflCapturedPages {
      season
      roundTitle
      topicId
      posts {
        postId
        author
        team
        isTeamSubmission
        parseError
        text
        players {
          name
          clubHint
          position
          backupPositions
          interchangePosition
          score
        }
      }
    }
  }
`

// A season's context for the forum round importer: club_seasons (for attribution
// + the club dropdown) and rounds→club_matches (to map a club to its club_match,
// and to read back the derived score + reference notes after import).
export const GET_FFL_CAPTURE_SEASON = gql`
  query GetFFLCaptureSeason($id: ID!) {
    fflSeason(id: $id) {
      id
      name
      ladder { id club { id name } }
      rounds {
        id
        name
        matches {
          id
          matchStyle
          clubMatches { id clubSeasonId club { id name } dataStatus score notes }
        }
      }
    }
  }
`

// A single club_match's derived score + reference notes, read back after an
// import to show derived vs forum vs fixtures.
export const GET_FFL_CLUB_MATCH_SCORE = gql`
  query GetFFLClubMatchScore($id: ID!) {
    fflClubMatch(id: $id) {
      id
      score
      notes
    }
  }
`

// A season's context for the fixture importer: its club_seasons (for resolving
// the sheet's club names) and the AFL season's rounds (to map each parsed round).
export const GET_FFL_FIXTURE_IMPORT_SEASON = gql`
  query GetFFLFixtureImportSeason($id: ID!) {
    fflSeason(id: $id) {
      id
      name
      ladder {
        id
        club { id name }
      }
      aflSeason {
        id
        name
        rounds { id name }
      }
    }
  }
`

export const SEARCH_AFL_PLAYERS = gql`
  query SearchAFLPlayers($query: String!) {
    aflPlayerSearch(query: $query) {
      id
      name
      latestPlayerSeason {
        id
        clubSeason {
          id
          club { name }
          season { id name }
        }
      }
    }
  }
`
