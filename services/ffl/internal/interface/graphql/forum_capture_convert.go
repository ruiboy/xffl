package graphql

import "xffl/services/ffl/internal/application/dataops"

// toPreviewedPage maps an application PreviewedPage to the GraphQL model.
func toPreviewedPage(p dataops.PreviewedPage) *FFLPreviewedPage {
	posts := make([]*FFLPreviewedPost, 0, len(p.Posts))
	for _, pp := range p.Posts {
		players := make([]*FFLParsedPlayer, 0, len(pp.Players))
		for _, pl := range pp.Players {
			players = append(players, &FFLParsedPlayer{
				Name:                pl.Name,
				ClubHint:            pl.ClubHint,
				Position:            pl.Position,
				BackupPositions:     pl.BackupPositions,
				InterchangePosition: pl.InterchangePosition,
				Score:               pl.Score,
				Notes:               pl.Notes,
			})
		}
		posts = append(posts, &FFLPreviewedPost{
			PostID:           pp.PostID,
			Author:           pp.Author,
			Team:             pp.Team,
			IsTeamSubmission: pp.IsTeamSubmission(),
			ParseError:       pp.ParseError,
			Text:             pp.Text,
			Players:          players,
		})
	}
	return &FFLPreviewedPage{Season: p.Season, RoundTitle: p.RoundTitle, TopicID: p.TopicID, Posts: posts}
}
