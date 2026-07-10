package domain

import "testing"

func TestExperimentStatuses(t *testing.T) {
	if ExperimentStatusPending != "pending" {
		t.Errorf("expected ExperimentStatusPending to be 'pending', got %s", ExperimentStatusPending)
	}
	if ExperimentStatusRunning != "running" {
		t.Errorf("expected ExperimentStatusRunning to be 'running', got %s", ExperimentStatusRunning)
	}
	if ExperimentStatusCompleted != "completed" {
		t.Errorf("expected ExperimentStatusCompleted to be 'completed', got %s", ExperimentStatusCompleted)
	}
	if ExperimentStatusFailed != "failed" {
		t.Errorf("expected ExperimentStatusFailed to be 'failed', got %s", ExperimentStatusFailed)
	}
	if ExperimentStatusRecovered != "recovered" {
		t.Errorf("expected ExperimentStatusRecovered to be 'recovered', got %s", ExperimentStatusRecovered)
	}
}
