package domain

import (
	"fmt"
	"strings"
	"time"
)

// DocumentType represents the type of search document
type DocumentType string

const (
	DocumentTypePlayer     DocumentType = "player"
	DocumentTypeClub       DocumentType = "club"
	DocumentTypePlayerMatch DocumentType = "player_match"
	DocumentTypeClubSeason DocumentType = "club_season"
)

// SearchDocument represents a document in the search index
type SearchDocument struct {
	ID           string                 `json:"id"`
	Type         DocumentType           `json:"type"`
	Source       string                 `json:"source"` // "afl" or "ffl"
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata"`
	IndexedAt    time.Time              `json:"indexed_at"`
	LastModified time.Time              `json:"last_modified"`
}

// PlayerDocument represents player data for search indexing
type PlayerDocument struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	ClubID   uint   `json:"club_id"`
	ClubName string `json:"club_name"`
	Source   string `json:"source"` // "afl" or "ffl"
}

// ToSearchDocument converts PlayerDocument to SearchDocument
func (p PlayerDocument) ToSearchDocument() SearchDocument {
	return SearchDocument{
		ID:      p.generateID(),
		Type:    DocumentTypePlayer,
		Source:  p.Source,
		Title:   p.Name,
		Content: p.generateContent(),
		Tags:    p.generateTags(),
		Metadata: map[string]interface{}{
			"player_id": p.ID,
			"club_id":   p.ClubID,
			"club_name": p.ClubName,
		},
		IndexedAt:    time.Now(),
		LastModified: time.Now(),
	}
}

func (p PlayerDocument) generateID() string {
	return fmt.Sprintf("%s_player_%d", p.Source, p.ID)
}

func (p PlayerDocument) generateContent() string {
	return fmt.Sprintf("%s plays for %s in %s", p.Name, p.ClubName, strings.ToUpper(p.Source))
}

func (p PlayerDocument) generateTags() []string {
	return []string{
		p.Source,
		"player",
		strings.ToLower(p.ClubName),
		strings.ToLower(p.Name),
	}
}

// ClubDocument represents club data for search indexing
type ClubDocument struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Source string `json:"source"` // "afl" or "ffl"
}

// ToSearchDocument converts ClubDocument to SearchDocument
func (c ClubDocument) ToSearchDocument() SearchDocument {
	return SearchDocument{
		ID:      c.generateID(),
		Type:    DocumentTypeClub,
		Source:  c.Source,
		Title:   c.Name,
		Content: c.generateContent(),
		Tags:    c.generateTags(),
		Metadata: map[string]interface{}{
			"club_id": c.ID,
		},
		IndexedAt:    time.Now(),
		LastModified: time.Now(),
	}
}

func (c ClubDocument) generateID() string {
	return fmt.Sprintf("%s_club_%d", c.Source, c.ID)
}

func (c ClubDocument) generateContent() string {
	return fmt.Sprintf("%s club in %s", c.Name, strings.ToUpper(c.Source))
}

func (c ClubDocument) generateTags() []string {
	return []string{
		c.Source,
		"club",
		strings.ToLower(c.Name),
	}
}