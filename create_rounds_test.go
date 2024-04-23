package main

import (
	"testing"
	"time"

	"gorm.io/gorm"
)

type RoundWithTypes struct {
	ID             uint
	RoundTimestamp string
	RoundTypes     string
}

func setupRoundConfigs(db *gorm.DB) {
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
		expectedRoundsCount     int
		expectedRounds          []RoundWithTypes
	}{
		{
			name:                    "No existing rounds - should create rounds for the last 12 hours. Count only to keep test succinct",
			currTime:                time.Date(2022, time.January, 10, 9, 30, 0, 0, time.UTC), // 9:30 AM, Jan 10, 2022
			existingRounds:          []Round{},
			existingRoundRoundTypes: []RoundRoundType{},
			expectedRoundsCount:     49,
		},
		{
			name:     "Outdated existing round from 60 minutes ago - create rounds for the last hour",
			currTime: time.Date(2022, time.January, 10, 9, 30, 0, 0, time.UTC), // 9:30 AM, Jan 10, 2022
			existingRounds: []Round{
				{
					ID:             1,
					RoundTimestamp: "2022-01-10T08:30:00Z",
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
			expectedRoundsCount: 5,
			expectedRounds: []RoundWithTypes{
				{
					ID:             1,
					RoundTimestamp: "2022-01-10T08:30:00Z",
					RoundTypes:     "15 Minute Round,30 Minute Round,60 Minute Round",
				},
				{
					ID:             2,
					RoundTimestamp: "2022-01-10T08:45:00Z",
					RoundTypes:     "15 Minute Round",
				},
				{
					ID:             3,
					RoundTimestamp: "2022-01-10T09:00:00Z",
					RoundTypes:     "15 Minute Round,30 Minute Round",
				},
				{
					ID:             4,
					RoundTimestamp: "2022-01-10T09:15:00Z",
					RoundTypes:     "15 Minute Round",
				},
				{
					ID:             5,
					RoundTimestamp: "2022-01-10T09:30:00Z",
					RoundTypes:     "15 Minute Round,30 Minute Round,60 Minute Round",
				},
			},
		},
		{
			name:     "Rounds are up-to-date - no new rounds should be created",
			currTime: time.Date(2022, time.January, 10, 9, 30, 0, 0, time.UTC), // 9:30 AM, Jan 10, 2022
			existingRounds: []Round{
				{
					ID:             1,
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
			expectedRoundsCount: 1,
			expectedRounds: []RoundWithTypes{
				{
					ID:             1,
					RoundTimestamp: "2022-01-10T09:30:00Z",
					RoundTypes:     "15 Minute Round,30 Minute Round,60 Minute Round",
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
		// Create existing round-round-types
		for _, roundRoundType := range tt.existingRoundRoundTypes {
			db.Create(&roundRoundType)
		}

		// Call createRounds
		CreateRounds(db, tt.currTime)

		// Get round with joins by grouping by round id and timestamp
		var roundsWithTypes []*RoundWithTypes
		db.Table("rounds").
			Select("rounds.id, round_timestamp, group_concat(name) as round_types").
			Joins("JOIN round_round_types ON rounds.id = round_round_types.round_id").
			Joins("JOIN round_types ON round_round_types.round_type_id = round_types.id").
			Group("rounds.id, round_timestamp").
			Scan(&roundsWithTypes)

		// Compare counts
		if len(roundsWithTypes) != tt.expectedRoundsCount {
			t.Errorf("Expected %d rounds, got %d", len(tt.expectedRounds), len(roundsWithTypes))
		}
		for i := range tt.expectedRounds {
			if roundsWithTypes[i].ID != tt.expectedRounds[i].ID ||
				roundsWithTypes[i].RoundTimestamp != tt.expectedRounds[i].RoundTimestamp {
				t.Errorf("Expected %v, got %v", tt.expectedRounds[i], roundsWithTypes[i])
			}
		}
	}
}
