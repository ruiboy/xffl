import gql from 'graphql-tag'

export const GET_FFL_TEAM_BUILDER = gql`
  query GetFFLTeamBuilder($seasonId: ID!) {
    fflSeason(id: $seasonId) {
      id
      name
      ladder {
        id
        club { id name }
        players {
          nodes {
            id
            player { id name }
          }
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

export const GET_FFL_SEASON_CLUBS = gql`
  query GetFFLSeasonClubs($seasonId: ID!) {
    fflSeason(id: $seasonId) {
      id
      ladder {
        id
        club { id name }
      }
    }
  }
`

export const GET_FFL_CLUB_SEASON = gql`
  query GetFFLClubSeason($seasonId: ID!, $clubId: ID!) {
    fflClubSeason(seasonId: $seasonId, clubId: $clubId) {
      id
      club { id name }
      season { id name }
      players {
        nodes {
          id
          player { id name aflPlayerId }
        }
        totalCount
      }
    }
  }
`

export const SEARCH_AFL_PLAYERS = gql`
  query SearchAFLPlayers($query: String!) {
    aflPlayerSearch(query: $query) {
      id
      name
    }
  }
`


export const GET_FFL_ROUND_BY_AFL_ROUND = gql`
  query GetFFLRoundByAflRound($aflRoundId: ID!) {
    fflRoundByAflRound(aflRoundId: $aflRoundId) {
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
    }
  }
`

export const GET_AFL_LIVE_ROUND = gql`
  query GetAFLLiveRoundForFFL {
    aflLiveRound {
      round { id }
      startDate
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
