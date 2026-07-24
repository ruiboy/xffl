import gql from 'graphql-tag'
import { LAST_N } from '../utils/playerStats'


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
              id
              club { id name }
            }
            statsAll: stats(method: MEAN) {
              goals kicks handballs marks tackles hitouts
            }
            statsLastN: stats(lastN: ${LAST_N}, method: MEAN) {
              goals kicks handballs marks tackles hitouts
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
      aflRoundId
      seasonId
      club { id }
    }
  }
`

export const GET_FFL_CLUB_MATCH_TEAM = gql`
  query GetFFLClubMatchTeam($id: ID!) {
    fflClubMatch(id: $id) {
      id
      playerMatches {
        playerSeasonId
        position
        backupPositions
        interchangePosition
        player { aflPlayer { name } }
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

export const GET_FFL_ROUND_ID_BY_AFL_ROUND = gql`
  query GetFFLRoundIdByAflRound($aflRoundId: ID!) {
    fflRoundByAflRound(aflRoundId: $aflRoundId) {
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
          premiershipPoints
          extraPoints
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
  query GetFFLRound($id: ID!, $aflRoundId: ID) {
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
            clubMatches { id clubSeasonId }
          }
        }
      }
      matches {
        id
        venue
        startTime
        result
        matchStyle
        clubMatches {
          id
          side
          clubSeasonId
          club { id name }
          score
          dataStatus
          suggestedSubstitutions { kind replacedPmId replacingPmId }
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
                id
                clubSeason { club { name } }
                stats(upToRoundId: $aflRoundId) { goals kicks handballs marks tackles hitouts games }
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
      matchStyle
      clubMatches {
        id
        side
        clubSeasonId
        club { id name }
        score
        dataStatus
        suggestedSubstitutions { kind replacedPmId replacingPmId }
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
              id
              clubSeason { id club { name } }
              stats { goals kicks handballs marks tackles hitouts games }
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
        premiershipPoints
        extraPoints
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
        id
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
      player {
        id
        name
        playerSeasons {
          id
          clubSeason {
            id
            club { id name }
            season { id name }
          }
        }
      }
      clubSeason {
        id
        club { id name }
        season { id name }
      }
      statsAll: stats(method: MEAN) {
        goals kicks handballs marks tackles hitouts games
      }
      statsLastN: stats(lastN: ${LAST_N}, method: MEAN) {
        goals kicks handballs marks tackles hitouts
      }
      statsMedian: stats(method: MEDIAN) {
        goals kicks handballs marks tackles hitouts
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
            id
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
      clubSeasonId
      fromRoundId
      toRoundId
      playerMatches {
        id
        matchId
        position
        backupPositions
        interchangePosition
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

export const GET_FFL_SEASON_ROUND_MAPPING = gql`
  query GetFflSeasonRoundMapping($seasonId: ID!) {
    fflSeason(id: $seasonId) {
      rounds { id aflRoundId }
    }
  }
`

export const GET_AFL_CLUB_SEASON = gql`
  query GetAFLClubSeason($id: ID!) {
    aflClubSeason(id: $id) {
      id
      club { id name }
      season { id name }
      playerSeasons {
        id
        player { id name }
        stats {
          goals kicks handballs marks tackles hitouts
        }
        fflPlayerSeasons {
          id
          clubSeasonId
          club { id name }
          toRoundId
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
          season { id name }
        }
      }
    }
  }
`

export const GET_FREE_AGENTS = gql`
  query GetFreeAgents($fflSeasonId: ID!) {
    fflSeason(id: $fflSeasonId) {
      aflSeason {
        playerSeasons {
          nodes {
            id
            player { id name }
            clubSeason { id club { id name } }
            statsAll: stats(method: MEAN) {
              goals kicks handballs marks tackles hitouts games
            }
            statsLastN: stats(lastN: ${LAST_N}, method: MEAN) {
              goals kicks handballs marks tackles hitouts
            }
            fflPlayerSeasons { toRoundId }
          }
        }
      }
    }
  }
`

export const GET_PLAYER_STATS_CARD = gql`
  query GetPlayerStatsCard($id: ID!, $aflRoundId: ID) {
    aflPlayerSeason(id: $id) {
      id
      seasonAvg: stats(upToRoundId: $aflRoundId) { goals kicks handballs marks tackles hitouts games }
      lastN: stats(upToRoundId: $aflRoundId, lastN: ${LAST_N}) { goals kicks handballs marks tackles hitouts games }
      matches {
        id
        status
        goals kicks handballs marks tackles hitouts
        clubMatch { match { id round { id name } } }
      }
    }
  }
`

export const GET_FFL_SEASONS = gql`
  query GetFFLSeasons {
    fflSeasons {
      id
      name
    }
  }
`

export const GET_FFL_SEASON_LADDER = gql`
  query GetFFLSeasonLadder($id: ID!) {
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
        premiershipPoints
        extraPoints
      }
      rounds {
        id
        name
      }
    }
  }
`
