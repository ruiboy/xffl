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
