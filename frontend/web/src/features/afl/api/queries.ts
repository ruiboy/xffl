import gql from 'graphql-tag'

export const GET_AFL_LIVE_ROUND_IDS = gql`
  query GetAFLLiveRoundIds {
    aflLiveRound {
      round { id season { id } }
      startDate
    }
  }
`

export const GET_AFL_LIVE_ROUND = gql`
  query GetAFLLiveRound {
    aflLiveRound {
      round {
        id
        name
        season {
          id
          name
          ladder {
            id
            club { id name }
            played
            won
            lost
            drawn
            for
            against
            premiershipPoints
          }
          rounds {
            id
            name
          }
        }
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
      startDate
    }
  }
`

export const GET_ROUND = gql`
  query GetRound($seasonId: ID!) {
    aflSeason(id: $seasonId) {
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
            playerMatches {
              id
              playerSeasonId
              player { id name }
              status
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
            score
            playerMatches {
              id
              playerSeasonId
              player { id name }
              status
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
              status
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
              status
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
