package main

import (
	"testing"
	"time"

	"gorm.io/gorm"
)

// Setup configs
func setupConfigs(db *gorm.DB) {
	// Create round types for 15, 30, and 60 minute rounds
	db.Create(&RoundType{
		Name:         "15 Minute Round",
		DurationAmt:  15,
		DurationUnit: "minutes",
	})
	db.Create(&RoundType{
		Name:         "30 Minute Round",
		DurationAmt:  30,
		DurationUnit: "minutes",
	})
	db.Create(&RoundType{
		Name:         "60 Minute Round",
		DurationAmt:  60,
		DurationUnit: "minutes",
	})

	// Create round configs for each round type
	var roundTypes []RoundType
	db.Find(&roundTypes)
	for _, roundType := range roundTypes {
		db.Create(&RoundConfig{
			RoundTypeId: roundType.ID,
			Enabled:     true,
		})
	}
}

func TestCreateRounds(t *testing.T) {
	tests := []struct {
		name                    string
		currTime                time.Time
		existingRounds          []Round
		existingRoundRoundTypes []RoundRoundType
		expectedRounds          []Round
		expectedRoundRoundTypes []RoundRoundType
	}{
		{
			name:           "No existing rounds - should create a new round and attach all three round types to it",
			currTime:       time.Date(2022, time.January, 10, 9, 30, 0, 0, time.UTC), // 9:30 AM, Jan 10, 2022
			existingRounds: []Round{},
			expectedRounds: []Round{
				{
					ID:             1,
					RoundTimestamp: "2022-01-10T09:30:00Z",
					Status:         "CREATED",
				},
			},
			existingRoundRoundTypes: []RoundRoundType{},
			expectedRoundRoundTypes: []RoundRoundType{
				{
					RoundTypeID: 1,
					RoundID:     1,
				},
				{
					RoundTypeID: 2,
					RoundID:     1,
				},
				{
					RoundTypeID: 3,
					RoundID:     1,
				},
			},
		},
		{
			name:     "Outdated existing round - create a new round and attach all types to it",
			currTime: time.Date(2022, time.January, 10, 9, 30, 0, 0, time.UTC), // 9:30 AM, Jan 10, 2022
			existingRounds: []Round{
				{
					ID:             1,
					RoundTimestamp: "2022-01-10T07:30:00Z",
					Status:         "CREATED",
				},
			},
			expectedRounds: []Round{
				{
					ID:             1,
					RoundTimestamp: "2022-01-10T07:30:00Z",
					Status:         "CREATED",
				},
				{
					ID:             2,
					RoundTimestamp: "2022-01-10T09:30:00Z",
					Status:         "CREATED",
				},
			},
			existingRoundRoundTypes: []RoundRoundType{
				{
					RoundTypeID: 1,
					RoundID:     1,
				},
				{
					RoundTypeID: 2,
					RoundID:     1,
				},
				{
					RoundTypeID: 3,
					RoundID:     1,
				},
			},
			expectedRoundRoundTypes: []RoundRoundType{
				{
					RoundTypeID: 1,
					RoundID:     1,
				},
				{
					RoundTypeID: 2,
					RoundID:     1,
				},
				{
					RoundTypeID: 3,
					RoundID:     1,
				},
				{
					RoundTypeID: 1,
					RoundID:     2,
				},
				{
					RoundTypeID: 2,
					RoundID:     2,
				},
				{
					RoundTypeID: 3,
					RoundID:     2,
				},
			},
		},
		{
			name:     "Only 15 minute round is due - create a new round and only attach the 15 minutes type to it",
			currTime: time.Date(2022, time.January, 10, 9, 30, 0, 0, time.UTC), // 9:30 AM, Jan 10, 2022
			existingRounds: []Round{
				{
					ID:             1,
					RoundTimestamp: "2022-01-10T09:15:00Z",
					Status:         "CREATED",
				},
			},
			expectedRounds: []Round{
				{
					ID:             1,
					RoundTimestamp: "2022-01-10T09:15:00Z",
					Status:         "CREATED",
				},
				{
					ID:             2,
					RoundTimestamp: "2022-01-10T09:30:00Z",
					Status:         "CREATED",
				},
			},
			existingRoundRoundTypes: []RoundRoundType{
				{
					RoundTypeID: 1,
					RoundID:     1,
				},
				{
					RoundTypeID: 2,
					RoundID:     1,
				},
				{
					RoundTypeID: 3,
					RoundID:     1,
				},
			},
			expectedRoundRoundTypes: []RoundRoundType{
				{
					RoundTypeID: 1,
					RoundID:     1,
				},
				{
					RoundTypeID: 2,
					RoundID:     1,
				},
				{
					RoundTypeID: 3,
					RoundID:     1,
				},
				{
					RoundTypeID: 1,
					RoundID:     2,
				},
			},
		},
	}

	for _, tt := range tests {
		db := setupDatabase()
		setupConfigs(db)

		// Create existing rounds
		for _, round := range tt.existingRounds {
			db.Create(&round)
		}
		// Create existing round round types
		for _, roundRoundType := range tt.existingRoundRoundTypes {
			db.Create(&roundRoundType)
		}

		// Call createRounds
		createRounds(db, tt.currTime)

		// Check that the rounds were created
		var rounds []Round
		db.Find(&rounds)
		if len(rounds) != len(tt.expectedRounds) {
			t.Errorf("Expected %d rounds, got %d", len(tt.expectedRounds), len(rounds))
		}
		for i := range tt.expectedRounds {
			if rounds[i].ID != tt.expectedRounds[i].ID ||
				rounds[i].RoundTimestamp != tt.expectedRounds[i].RoundTimestamp ||
				rounds[i].Status != tt.expectedRounds[i].Status {
				t.Errorf("Expected %v, got %v", tt.expectedRounds[i], rounds[i])
			}
		}

		// Check that the round round types were created
		var roundRoundTypes []RoundRoundType
		db.Find(&roundRoundTypes)
		if len(roundRoundTypes) != len(tt.expectedRoundRoundTypes) {
			t.Errorf("Expected %d round round types, got %d", len(tt.expectedRoundRoundTypes), len(roundRoundTypes))
		}
		for i := range roundRoundTypes {
			if roundRoundTypes[i].RoundTypeID != tt.expectedRoundRoundTypes[i].RoundTypeID ||
				roundRoundTypes[i].RoundID != tt.expectedRoundRoundTypes[i].RoundID {
				t.Errorf("Expected %v, got %v", tt.expectedRoundRoundTypes[i], roundRoundTypes[i])
			}
		}
	}
}
