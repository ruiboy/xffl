package graphql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application/dataops"
	"xffl/services/ffl/internal/infrastructure/forum"
)

// Real THC post body (br-separated), trimmed, from the captured 2025 Round 1 page.
const thcSampleHTML = `THC - Round 1 <br>
Strap in<br>
<br>
Goals<br>
B King GCS<br>
J Cameron GEE<br>
<br>
Star<br>
N Daicos COL<i data-tag="post_body_end" class="hide"></i>`

// End-to-end (no DB) proof of slice 1: a captured page flows through the ingest
// resolver → forum HTML/parse → preview, and the query reflects it.
func TestIngestFFLForumPage_ParsesRealPost(t *testing.T) {
	r := &Resolver{Captures: dataops.NewForumCaptureBuffer(forum.NewParser())}
	ctx := context.Background()

	page, err := r.Mutation().IngestFFLForumPage(ctx, IngestFFLForumPageInput{
		Season:     "2025",
		RoundTitle: "Rnd 1- SheepDog Trials",
		TopicID:    "16265070",
		Posts: []*CapturedFFLPostInput{
			{PostID: "13140", Author: "thc", HTML: thcSampleHTML},
			{PostID: "99", Author: "randomguy", HTML: "just some banter, no team here"},
		},
	})
	require.NoError(t, err)
	require.Len(t, page.Posts, 2)

	thc := page.Posts[0]
	assert.Equal(t, "THC", thc.Team)
	assert.True(t, thc.IsTeamSubmission)
	require.NotEmpty(t, thc.Players)
	assert.Equal(t, "goals", thc.Players[0].Position)

	unknown := page.Posts[1]
	assert.Empty(t, unknown.Team, "unknown author is flagged, not parsed")
	assert.False(t, unknown.IsTeamSubmission)

	// The query reflects the in-session buffer.
	pages, err := r.Query().FflCapturedPages(ctx)
	require.NoError(t, err)
	require.Len(t, pages, 1)
	assert.Equal(t, "16265070", pages[0].TopicID)
}
