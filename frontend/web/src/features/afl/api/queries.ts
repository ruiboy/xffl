import gql from 'graphql-tag'

export const GET_SEASONS = gql`
  query GetSeasons {
    aflSeasons {
      id
      name
    }
  }
`

export const GET_SEASON = gql`
  query GetSeason($id: ID!) {
    aflSeason(id: $id) {
      id
      name
      rounds {
        id
        name
        matches {
          id
          venue
          startTime
          result
          homeClubMatch {
            id
            club { id name }
            score
          }
          awayClubMatch {
            id
            club { id name }
            score
          }
        }
      }
    }
  }
`

export const GET_MATCH = gql`
  query GetMatch($seasonId: ID!) {
    aflSeason(id: $seasonId) {
      id
      rounds {
        id
        name
        matches {
          id
          venue
          startTime
          result
          homeClubMatch {
            id
            club { id name }
            rushedBehinds
            score
            playerMatches {
              id
              playerSeasonId
              player { id name }
              kicks
              handballs
              marks
              hitouts
              tackles
              goals
              behinds
              disposals
              score
            }
          }
          awayClubMatch {
            id
            club { id name }
            rushedBehinds
            score
            playerMatches {
              id
              playerSeasonId
              player { id name }
              kicks
              handballs
              marks
              hitouts
              tackles
              goals
              behinds
              disposals
              score
            }
          }
        }
      }
    }
  }
`
