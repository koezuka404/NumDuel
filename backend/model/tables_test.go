package model_test

import (
	"testing"

	"github.com/numduel/numduel/model"
)

func TestModelTableNames(t *testing.T) {
	tests := map[string]string{
		"users":              (model.User{}).TableName(),
		"games":              (model.Game{}).TableName(),
		"activity_logs":      (model.ActivityLog{}).TableName(),
		"login_logs":         (model.LoginLog{}).TableName(),
		"ws_connection_logs": (model.WSConnectionLog{}).TableName(),
		"rankings":           (model.Ranking{}).TableName(),
		"matching_queue":     (model.MatchingQueueEntry{}).TableName(),
	}
	for want, got := range tests {
		if got != want {
			t.Fatalf("%s TableName = %q", want, got)
		}
	}
}
