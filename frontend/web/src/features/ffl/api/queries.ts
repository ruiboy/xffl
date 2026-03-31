import gql from 'graphql-tag'

export const GET_FFL_TEAM_BUILDER = gql`
  query GetFFLTeamBuilder($seasonId: ID!) {
    fflSeason(id: $seasonId) {
      id
      name
      ladder {
        id
        club { id name }
        roster {
          playerSeasonId
          player { id name aflPlayerId }
        }
      }
      rounds {
        id
        name
        matches {
          id
          homeClubMatch {
            id
            club { id name }
            playerMatches {
              id
              playerSeasonId
              player { id name }
              position
              status
              backupPositions
              interchangePosition
              score
            }
          }
          awayClubMatch {
            id
            club { id name }
            playerMatches {
              id
              playerSeasonId
              player { id name }
              position
              status
              backupPositions
              interchangePosition
              score
            }
          }
        }
      }
    }
  }
`

export const GET_FFL_LATEST_ROUND = gql`
  query GetFFLLatestRound {
    fflLatestRound {
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
          percentage
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
  }
`

export const GET_FFL_SEASON = gql`
  query GetFFLSeason($id: ID!) {
    fflSeason(id: $id) {
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
        percentage
      }
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
              position
              status
              backupPositions
              interchangePosition
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
              position
              status
              backupPositions
              interchangePosition
              score
            }
          }
        }
      }
    }
  }
`
