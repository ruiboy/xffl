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
  query GetFFLClubSeason($id: ID!) {
    fflClubSeason(id: $id) {
      id
      club { id name }
      season { id name }
      players {
        nodes {
          id
          player { id aflPlayerId aflPlayer { id name } }
          aflPlayerSeason {
            id
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


export const GET_FFL_ROUND_CLUB_MATCHES = gql`
  query GetFFLRoundClubMatches($id: ID!) {
    fflRound(id: $id) {
      matches {
        homeClubMatch { id clubSeasonId }
        awayClubMatch { id clubSeasonId }
      }
    }
  }
`

export const GET_FFL_CLUB_MATCH = gql`
  query GetFFLClubMatch($id: ID!) {
    fflClubMatch(id: $id) {
      id
      clubSeasonId
      roundId
      seasonId
      club { id }
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
      season {
        id
        name
        rounds {
          id
          name
          aflRoundId
          matches {
            homeClubMatch { id clubSeasonId }
            awayClubMatch { id clubSeasonId }
          }
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
          playerMatches {
            id
            playerSeasonId
            player { aflPlayer { name } }
            position
            status
            aflStatus
            backupPositions
            interchangePosition
            score
            playerSeason {
              aflPlayerSeason {
                clubSeason { club { name } }
              }
            }
            aflPlayerMatch {
              clubMatch { match { id } }
              goals kicks handballs marks tackles hitouts
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
            player { aflPlayer { name } }
            position
            status
            aflStatus
            backupPositions
            interchangePosition
            score
            playerSeason {
              aflPlayerSeason {
                clubSeason { club { name } }
              }
            }
            aflPlayerMatch {
              clubMatch { match { id } }
              goals kicks handballs marks tackles hitouts
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
        season { id name rounds { id name } }
      }
      homeClubMatch {
        id
        clubSeasonId
        club { id name }
        score
        playerMatches {
          id
          playerSeasonId
          player { aflPlayer { name } }
          position
          status
          aflStatus
          backupPositions
          interchangePosition
          score
          playerSeason {
            aflPlayerSeason {
              clubSeason { club { name } }
            }
          }
          aflPlayerMatch {
            goals kicks handballs marks tackles hitouts
          }
        }
      }
      awayClubMatch {
        id
        clubSeasonId
        club { id name }
        score
        playerMatches {
          id
          playerSeasonId
          player { aflPlayer { name } }
          position
          status
          aflStatus
          backupPositions
          interchangePosition
          score
          playerSeason {
            aflPlayerSeason {
              clubSeason { club { name } }
            }
          }
          aflPlayerMatch {
            goals kicks handballs marks tackles hitouts
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
              player { id aflPlayer { id name } }
              position
              status
              aflStatus
              backupPositions
              interchangePosition
              score
              aflPlayerMatch {
                clubMatch {
                  club { name }
                  match {
                    id
                    dataStatus
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
              player { id aflPlayer { id name } }
              position
              status
              aflStatus
              backupPositions
              interchangePosition
              score
              aflPlayerMatch {
                clubMatch {
                  club { name }
                  match {
                    id
                    dataStatus
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


export const GET_AFL_SEASON_CLUB_SEASONS = gql`
  query GetAFLSeasonClubSeasons($fflSeasonId: ID!) {
    fflSeason(id: $fflSeasonId) {
      aflSeason {
        ladder {
          id
          club { name }
        }
      }
    }
  }
`

export const GET_AFL_PLAYER_SEASON_STATS = gql`
  query GetAFLPlayerSeasonStats($id: ID!) {
    aflPlayerSeason(id: $id) {
      id
      player { id name }
      clubSeason {
        club { id name }
        season { id name }
      }
      matches {
        id
        status
        kicks
        handballs
        marks
        tackles
        hitouts
        goals
        behinds
        score
        clubMatch {
          club { id name }
          match {
            round { id name }
            homeClubMatch { club { id name } }
            awayClubMatch { club { id name } }
          }
        }
      }
    }
  }
`

export const GET_FFL_PLAYER_STINTS = gql`
  query GetFFLPlayerStints($aflPlayerSeasonId: ID!) {
    fflPlayerSeasonsByAflPlayerSeason(aflPlayerSeasonId: $aflPlayerSeasonId) {
      id
      club { id name }
      fromRoundId
      toRoundId
      playerMatches {
        id
        position
        status
        aflStatus
        score
        aflPlayerMatch {
          id
          clubMatch {
            match {
              round { id name }
              homeClubMatch { club { id name } }
              awayClubMatch { club { id name } }
            }
          }
        }
      }
    }
  }
`

export const SEARCH_AFL_PLAYERS = gql`
  query SearchAFLPlayersForFFL($query: String!) {
    aflPlayerSearch(query: $query) {
      id
      name
      latestPlayerSeason {
        id
        clubSeason {
          id
          club { name }
          season { name }
        }
      }
    }
  }
`
