package dataops

import (
	"context"
	"sync"

	"xffl/services/ffl/internal/application"
)

// CapturedPost is one forum post as captured by the userscript.
type CapturedPost struct {
	PostID    string
	Author    string
	Timestamp string
	HTML      string
}

// CapturedPageParams is a captured forum page (one page of a round's thread).
type CapturedPageParams struct {
	Season     string
	RoundTitle string
	TopicID    string
	Posts      []CapturedPost
}

// PreviewedPost is a captured post after HTML→text, team identification, and parsing.
type PreviewedPost struct {
	PostID     string
	Author     string
	Team       string // parser format; "" if the author is unknown (escape hatch)
	Text       string // raw HTML converted to text, kept so an unparseable post (e.g. a squads thread) is still workable
	Players    []application.ParsedPlayerRow
	ParseError string
}

// IsTeamSubmission is a first-pass heuristic: a post that parsed into players.
func (p PreviewedPost) IsTeamSubmission() bool { return len(p.Players) > 0 }

// PreviewedPage is a captured page with each post previewed. No DB writes.
type PreviewedPage struct {
	Season     string
	RoundTitle string
	TopicID    string
	Posts      []PreviewedPost
}

// ForumCaptureBuffer is an in-session (ephemeral) store of previewed capture
// pages — slice 1 of the historical import. Not persisted; cleared on restart.
// Durable progress comes later from the committed data, not this buffer.
type ForumCaptureBuffer struct {
	forum application.ForumProcessor
	mu    sync.Mutex
	pages []PreviewedPage
}

func NewForumCaptureBuffer(forum application.ForumProcessor) *ForumCaptureBuffer {
	return &ForumCaptureBuffer{forum: forum}
}

// Ingest previews a captured page — for each post: identify the team from the
// author, convert the content HTML to text, and parse it — then stores and
// returns the preview. Unknown authors are flagged (Team == "") and not parsed.
func (b *ForumCaptureBuffer) Ingest(ctx context.Context, params CapturedPageParams) PreviewedPage {
	page := PreviewedPage{Season: params.Season, RoundTitle: params.RoundTitle, TopicID: params.TopicID}
	for _, post := range params.Posts {
		pp := PreviewedPost{PostID: post.PostID, Author: post.Author, Team: b.forum.TeamForAuthor(post.Author)}
		// Always keep the plain text: an unparseable post (unknown author, or a
		// non-team thread like a squads paste) shows nothing without it.
		pp.Text = b.forum.HTMLToText(post.HTML)
		// Unknown author → attribute from the post's own content, so a team posted
		// on someone else's behalf is still parsed and attributed.
		if pp.Team == "" {
			pp.Team = b.forum.DetectFormat(pp.Text)
		}
		if pp.Team != "" {
			rows, err := b.forum.Parse(ctx, pp.Team, pp.Text)
			if err != nil {
				pp.ParseError = err.Error()
			} else {
				pp.Players = rows
			}
		}
		page.Posts = append(page.Posts, pp)
	}
	b.mu.Lock()
	b.pages = append(b.pages, page)
	b.mu.Unlock()
	return page
}

// Pages returns all pages previewed this session (most recent last).
func (b *ForumCaptureBuffer) Pages() []PreviewedPage {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]PreviewedPage, len(b.pages))
	copy(out, b.pages)
	return out
}

// Clear drops all captured pages.
func (b *ForumCaptureBuffer) Clear() {
	b.mu.Lock()
	b.pages = nil
	b.mu.Unlock()
}
