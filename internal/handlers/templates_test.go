package handlers

import (
	"bytes"
	"html/template"
	"testing"
	"time"

	"github.com/klausbreyer/pokerhans/internal/models"
)

func TestTemplatesParse(t *testing.T) {
	tmpl, err := template.ParseGlob("../../web/templates/*.html")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}

	for _, name := range []string{"layout", "home_content", "season_content", "leaderboard_content"} {
		if tmpl.Lookup(name) == nil {
			t.Fatalf("expected template %q to be defined", name)
		}
	}
}

func TestLayoutExecutesPages(t *testing.T) {
	tmpl, err := template.ParseGlob("../../web/templates/*.html")
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}

	testCases := []struct {
		name string
		data any
	}{
		{
			name: "home",
			data: struct {
				Page        string
				Seasons     []models.Season
				CurrentYear int
			}{
				Page:        "home",
				Seasons:     []models.Season{},
				CurrentYear: 2026,
			},
		},
		{
			name: "season",
			data: struct {
				Page           string
				Seasons        []models.Season
				CurrentSeason  models.Season
				VisitedPlayers []models.PlayerStatus
				ToVisitPlayers []models.PlayerStatus
				Games          []models.Game
				AllPlayers     []models.Player
				CurrentDate    string
				CurrentYear    int
				IsLatestSeason bool
			}{
				Page:          "season",
				CurrentSeason: models.Season{ID: 1, Name: "Season 1"},
				CurrentDate:   time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
				CurrentYear:   2026,
			},
		},
		{
			name: "leaderboard",
			data: struct {
				Page                   string
				Seasons                []models.Season
				OverallWins            []models.LeaderboardEntry
				PointsLeaderboard      []models.LeaderboardEntry
				LeaderboardStartSeason int
				LatestSeasonID         int
				CurrentYear            int
			}{
				Page:                   "leaderboard",
				LeaderboardStartSeason: 4,
				LatestSeasonID:         5,
				CurrentYear:            2026,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := tmpl.ExecuteTemplate(&buf, "layout", tc.data); err != nil {
				t.Fatalf("failed to execute layout: %v", err)
			}
			if buf.Len() == 0 {
				t.Fatal("expected rendered HTML")
			}
		})
	}
}
