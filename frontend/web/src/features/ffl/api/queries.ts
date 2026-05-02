import gql from 'graphql-tag'


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
          aflPlayerSeason {
            clubSeason {
              club { id name }
            }
          }
          fromRoundId
          toRoundId
          notes
          costCents
        }
        pageInfo { totalCount }
      }
    }
  }
`

export const GET_AFL_PLAYER_SEASONS = gql`
  query GetAFLPlayerSeasonsBySeason($seasonId: ID!, $query: String) {
    fflSeason(id: $seasonId) {
      aflSeason {
        playerSeasons(filter: { query: $query }) {
          nodes {
            id
            player { id name }
            clubSeason { club { name } }
          }
          pageInfo { totalCount }
        }
      }
    }
  }
`


export const GET_FFL_ROUND_IDS_BY_AFL_ROUND = gql`
  query GetFFLRoundIdsByAflRound($aflRoundId: ID!) {
    fflRoundByAflRound(aflRoundId: $aflRoundId) {
      id
      season { id }
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

export const GET_FFL_ROUND = gql`
  query GetFFLRound($id: ID!) {
    fflRound(id: $id) {
      id
      name
      aflRoundId
      aflRound {
        id
        season { id }
        matches {
          id
          statsImportStatus
          homeClubMatch { club { id } }
          awayClubMatch { club { id } }
        }
      }
      season {
        id
        name
        rounds { id name aflRoundId }
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
          playerMatches {
            id
            playerSeasonId
            player { name }
            position
            status
            backupPositions
            interchangePosition
            score
            playerSeason {
              aflPlayerSeason {
                clubSeason { club { id name } }
              }
            }
          }
        }
        awayClubMatch {
          id
          club { id name }
          score
          playerMatches {
            id
            playerSeasonId
            player { name }
            position
            status
            backupPositions
            interchangePosition
            score
            playerSeason {
              aflPlayerSeason {
                clubSeason { club { id name } }
              }
            }
          }
        }
      }
    }
  }
`

export const GET_FFL_MATCH = gql`
  query GetFFLMatch($id: ID!) {
    fflMatch(id: $id) {
      id
      venue
      result
      round {
        id
        name
        aflRoundId
        aflRound {
          id
          season { id }
          matches {
            id
            statsImportStatus
            homeClubMatch { club { id } }
            awayClubMatch { club { id } }
          }
        }
        season { id name rounds { id name } }
      }
      homeClubMatch {
        id
        club { id name }
        score
        playerMatches {
          id
          playerSeasonId
          player { name }
          position
          status
          backupPositions
          interchangePosition
          score
          playerSeason {
            aflPlayerSeason {
              clubSeason { club { id name } }
            }
          }
        }
      }
      awayClubMatch {
        id
        club { id name }
        score
        playerMatches {
          id
          playerSeasonId
          player { name }
          position
          status
          backupPositions
          interchangePosition
          score
          playerSeason {
            aflPlayerSeason {
              clubSeason { club { id name } }
            }
          }
        }
      }
    }
  }
`

export const GET_FFL_SEASON_POSITIONS = gql`
  query GetFFLSeasonPositions($id: ID!) {
    fflSeason(id: $id) {
      id
      rounds {
        id
        name
        matches {
          id
          homeClubMatch {
            id
            playerMatches {
              id
              playerSeasonId
              position
              backupPositions
              interchangePosition
            }
          }
          awayClubMatch {
            id
            playerMatches {
              id
              playerSeasonId
              position
              backupPositions
              interchangePosition
            }
          }
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
        aflRoundId
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
              aflPlayerMatch {
                clubMatch {
                  club { name }
                  match {
                    id
                    statsImportStatus
                    round { season { id } }
                  }
                }
              }
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
              aflPlayerMatch {
                clubMatch {
                  club { name }
                  match {
                    id
                    statsImportStatus
                    round { season { id } }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
`
