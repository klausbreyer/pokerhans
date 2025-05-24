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
