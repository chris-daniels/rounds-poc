package main

import (
	"testing"
	"time"
)

func TestStartRounds(t *testing.T) {
	tests := []struct {
		name                string
		startTime           time.Time
		currTime            time.Time
		existingRounds      []Round
		expectedRoundsCount int
		expectedRounds      []StartRoundsItem
	}{
		{
			name:                "Should return 49 rounds for the last 12 hours. Count only to keep test succinct.",
			startTime:           time.Date(2022, time.January, 9, 21, 30, 0, 0, time.UTC), // 9:30 PM, Jan 9, 2022
			currTime:            time.Date(2022, time.January, 10, 9, 30, 0, 0, time.UTC), // 9:30 AM, Jan 10, 2022
			existingRounds:      []Round{},
			expectedRoundsCount: 49, // 4 per hour, include the current hour: 48 + 1 = 49
		},
		{
			name:      "Should display STARTED rounds if they are already created. Time window of one hour",
			startTime: time.Date(2022, time.January, 10, 8, 30, 0, 0, time.UTC), // 8:30 AM, Jan 10, 2022
			currTime:  time.Date(2022, time.January, 10, 9, 30, 0, 0, time.UTC), // 9:30 AM, Jan 10, 2022
			existingRounds: []Round{
				{
					ID:             1,
					RoundTimestamp: "2022-01-10T09:00:00Z",
					Status:         "STARTED",
				},
			},
			expectedRoundsCount: 5,
			expectedRounds: []StartRoundsItem{
				{
					Status:         "MISSED",
					RoundTimestamp: "2022-01-10T08:30:00Z",
				},
				{
					Status:         "MISSED",
					RoundTimestamp: "2022-01-10T08:45:00Z",
				},
				{
					Status:         "STARTED",
					RoundTimestamp: "2022-01-10T09:00:00Z",
				},
				{
					Status:         "NOT_STARTED",
					RoundTimestamp: "2022-01-10T09:15:00Z",
				},
				{
					Status:         "NOT_STARTED",
					RoundTimestamp: "2022-01-10T09:30:00Z",
				},
			},
		},
		{
			name:      "Should display COMPLETE rounds if they are already created. Time window of one hour",
			startTime: time.Date(2022, time.January, 10, 8, 30, 0, 0, time.UTC), // 8:30 AM, Jan 10, 2022
			currTime:  time.Date(2022, time.January, 10, 9, 30, 0, 0, time.UTC), // 9:30 AM, Jan 10, 2022
			existingRounds: []Round{
				{
					ID:             1,
					RoundTimestamp: "2022-01-10T08:45:00Z",
					Status:         "COMPLETE",
				},
			},
			expectedRoundsCount: 5,
			expectedRounds: []StartRoundsItem{
				{
					Status:         "MISSED",
					RoundTimestamp: "2022-01-10T08:30:00Z",
				},
				{
					Status:         "COMPLETE",
					RoundTimestamp: "2022-01-10T08:45:00Z",
				},
				{
					Status:         "MISSED",
					RoundTimestamp: "2022-01-10T09:00:00Z",
				},
				{
					Status:         "NOT_STARTED",
					RoundTimestamp: "2022-01-10T09:15:00Z",
				},
				{
					Status:         "NOT_STARTED",
					RoundTimestamp: "2022-01-10T09:30:00Z",
				},
			},
		},
		{
			name:      "If all rounds have been started, should create a new round at the next time interval. Time window of one hour",
			startTime: time.Date(2022, time.January, 10, 8, 30, 0, 0, time.UTC), // 8:30 AM, Jan 10, 2022
			currTime:  time.Date(2022, time.January, 10, 9, 30, 0, 0, time.UTC), // 9:30 AM, Jan 10, 2022
			existingRounds: []Round{
				{
					ID:             1,
					RoundTimestamp: "2022-01-10T08:30:00Z",
					Status:         "COMPLETE",
				},
				{
					ID:             2,
					RoundTimestamp: "2022-01-10T08:45:00Z",
					Status:         "COMPLETE",
				},
				{
					ID:             3,
					RoundTimestamp: "2022-01-10T09:00:00Z",
					Status:         "COMPLETE",
				},
				{
					ID:             4,
					RoundTimestamp: "2022-01-10T09:15:00Z",
					Status:         "STARTED",
				},
				{
					ID:             5,
					RoundTimestamp: "2022-01-10T09:30:00Z",
					Status:         "STARTED",
				},
			},
			expectedRoundsCount: 6,
			expectedRounds: []StartRoundsItem{
				{
					RoundTimestamp: "2022-01-10T08:30:00Z",
					Status:         "COMPLETE",
				},
				{
					RoundTimestamp: "2022-01-10T08:45:00Z",
					Status:         "COMPLETE",
				},
				{
					RoundTimestamp: "2022-01-10T09:00:00Z",
					Status:         "COMPLETE",
				},
				{
					RoundTimestamp: "2022-01-10T09:15:00Z",
					Status:         "STARTED",
				},
				{
					RoundTimestamp: "2022-01-10T09:30:00Z",
					Status:         "STARTED",
				},
				{
					RoundTimestamp: "2022-01-10T09:45:00Z",
					Status:         "NOT_STARTED",
				},
			},
		},
	}

	for _, tt := range tests {
		db := setupDatabase()
		setupRoundConfigs(db)

		// Create existing rounds
		for _, round := range tt.existingRounds {
			db.Create(&round)
		}

		startRoundsItems, err := StartRounds(db, tt.startTime, tt.currTime)
		if err != nil {
			t.Fatalf("StartRounds failed: %v", err)
		}

		// Check the number of rounds created
		if len(startRoundsItems) != tt.expectedRoundsCount {
			t.Fatalf("Expected %v rounds, got %v", tt.expectedRoundsCount, len(startRoundsItems))
		}

		// Check contents of the rounds created
		for i, expectedRound := range tt.expectedRounds {
			if startRoundsItems[i].Status != expectedRound.Status {
				t.Fatalf("Expected round status %v, got %v", expectedRound.Status, startRoundsItems[i].Status)
			}
			if startRoundsItems[i].RoundTimestamp != expectedRound.RoundTimestamp {
				t.Fatalf("Expected round timestamp %v, got %v", expectedRound.RoundTimestamp, startRoundsItems[i].RoundTimestamp)
			}
		}
	}
}
