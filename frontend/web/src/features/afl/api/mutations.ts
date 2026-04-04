import gql from 'graphql-tag'

export const UPDATE_PLAYER_MATCH = gql`
  mutation UpdatePlayerMatch($input: UpdateAFLPlayerMatchInput!) {
    updateAFLPlayerMatch(input: $input) {
      id
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
`
