package models

import (
	"testing"
	"time"
)

func TestPlayerStatus(t *testing.T) {
	player := Player{
		ID:   1,
		Name: "Test Player",
	}

	status := PlayerStatus{
		Player:    player,
		HasHosted: true,
		GameDate:  time.Now(),
	}

	if status.ID != 1 {
		t.Errorf("Expected ID 1, got %d", status.ID)
	}

	if status.Name != "Test Player" {
		t.Errorf("Expected name 'Test Player', got %s", status.Name)
	}

	if !status.HasHosted {
		t.Errorf("Expected HasHosted to be true")
	}
}

func TestApplyWinsRanks(t *testing.T) {
	entries := []LeaderboardEntry{
		{PlayerName: "Anna", Wins: 5},
		{PlayerName: "Ben", Wins: 3},
		{PlayerName: "Chris", Wins: 3},
		{PlayerName: "Dina", Wins: 1},
	}

	applyWinsRanks(entries)

	expectedRanks := []int{1, 2, 2, 4}
	for i, expected := range expectedRanks {
		if entries[i].Rank != expected {
			t.Fatalf("entry %d: expected rank %d, got %d", i, expected, entries[i].Rank)
		}
	}
}

func TestApplyPointsRanks(t *testing.T) {
	entries := []LeaderboardEntry{
		{PlayerName: "Anna", Points: 10, Wins: 3, Seconds: 1},
		{PlayerName: "Ben", Points: 10, Wins: 2, Seconds: 4},
		{PlayerName: "Chris", Points: 7, Wins: 2, Seconds: 1},
		{PlayerName: "Dina", Points: 7, Wins: 2, Seconds: 1},
	}

	applyPointsRanks(entries)

	expectedRanks := []int{1, 2, 3, 3}
	for i, expected := range expectedRanks {
		if entries[i].Rank != expected {
			t.Fatalf("entry %d: expected rank %d, got %d", i, expected, entries[i].Rank)
		}
	}
}
