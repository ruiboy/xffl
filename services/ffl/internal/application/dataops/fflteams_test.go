package dataops

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application"
)

// fakeForumProcessor identifies only "ruiboy" and parses any text into one player.
type fakeForumProcessor struct{}

func (fakeForumProcessor) HTMLToText(html string) string { return html }
func (fakeForumProcessor) TeamForAuthor(author string) string {
	if author == "ruiboy" {
		return "Ruiboys"
	}
	return ""
}
func (fakeForumProcessor) Parse(_ context.Context, _, _ string) ([]application.ParsedPlayerRow, error) {
	return []application.ParsedPlayerRow{{Name: "Jake Waterman", Position: "goals"}}, nil
}

func TestForumCaptureBuffer_Ingest(t *testing.T) {
	buf := NewForumCaptureBuffer(fakeForumProcessor{})

	page := buf.Ingest(context.Background(), CapturedPageParams{
		Season:     "2025",
		RoundTitle: "Rnd 1",
		Posts: []CapturedPost{
			{PostID: "1", Author: "ruiboy", HTML: "..."},
			{PostID: "2", Author: "randomguy", HTML: "banter"},
		},
	})

	require.Len(t, page.Posts, 2)

	known := page.Posts[0]
	assert.Equal(t, "Ruiboys", known.Team)
	assert.True(t, known.IsTeamSubmission())
	assert.Len(t, known.Players, 1)

	unknown := page.Posts[1]
	assert.Empty(t, unknown.Team, "unknown author is flagged, not parsed")
	assert.False(t, unknown.IsTeamSubmission())
	assert.Empty(t, unknown.Players)

	// Stored in the session buffer.
	assert.Len(t, buf.Pages(), 1)
	buf.Clear()
	assert.Empty(t, buf.Pages())
}
