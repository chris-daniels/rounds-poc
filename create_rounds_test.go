package main

import (
	"testing"
	"time"

	"gorm.io/gorm"
)

type RoundWithTypesAndMembers struct {
	ID             uint
	RoundTimestamp string
	RoundTypes     string
	RoundMembers   string
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

	// Create round assignments for each round type
	// Patient 1 has all three round types enabled
	db.Create(&RoundAssignment{
		RoundTypeId: 1,
		PatientId:   "patient1",
	})
	db.Create(&RoundAssignment{
		RoundTypeId: 2,
		PatientId:   "patient1",
	})
	db.Create(&RoundAssignment{
		RoundTypeId: 3,
		PatientId:   "patient1",
	})
	// Patient 2 has only 30 minute round type enabled
	db.Create(&RoundAssignment{
		RoundTypeId: 2,
		PatientId:   "patient2",
	})
	// Patient 3 has only 60 minute round type enabled
	db.Create(&RoundAssignment{
		RoundTypeId: 3,
		PatientId:   "patient3",
	})

}

func TestCreateRounds(t *testing.T) {
	tests := []struct {
		name                    string
		currTime                time.Time
		existingRounds          []Round
		existingRoundRoundTypes []RoundRoundType
		existingRoundMembers    []RoundMember
		expectedRoundsCount     int
		expectedRounds          []RoundWithTypesAndMembers
	}{
		{
			name:                    "No existing rounds - should create rounds for the last 12 hours. Count only to keep test succinct.",
			currTime:                time.Date(2022, time.January, 10, 9, 30, 0, 0, time.UTC), // 9:30 AM, Jan 10, 2022
			existingRounds:          []Round{},
			existingRoundRoundTypes: []RoundRoundType{},
			expectedRoundsCount:     49, // 4 per hour, include the current hour: 48 + 1 = 49
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
			existingRoundMembers: []RoundMember{
				{
					RoundId:   1,
					PatientId: "patient1",
				},
				{
					RoundId:   1,
					PatientId: "patient2",
				},
				{
					RoundId:   1,
					PatientId: "patient3",
				},
			},
			expectedRoundsCount: 5,
			expectedRounds: []RoundWithTypesAndMembers{
				{
					ID:             1,
					RoundTimestamp: "2022-01-10T08:30:00Z",
					RoundTypes:     "15 Minute Round,30 Minute Round,60 Minute Round",
					RoundMembers:   "patient1,patient2,patient3",
				},
				{
					ID:             2,
					RoundTimestamp: "2022-01-10T08:45:00Z",
					RoundTypes:     "15 Minute Round",
					RoundMembers:   "patient1",
				},
				{
					ID:             3,
					RoundTimestamp: "2022-01-10T09:00:00Z",
					RoundTypes:     "15 Minute Round,30 Minute Round",
					RoundMembers:   "patient1,patient2",
				},
				{
					ID:             4,
					RoundTimestamp: "2022-01-10T09:15:00Z",
					RoundTypes:     "15 Minute Round",
					RoundMembers:   "patient1",
				},
				{
					ID:             5,
					RoundTimestamp: "2022-01-10T09:30:00Z",
					RoundTypes:     "15 Minute Round,30 Minute Round,60 Minute Round",
					RoundMembers:   "patient1,patient2,patient3",
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
			existingRoundMembers: []RoundMember{
				{
					RoundId:   1,
					PatientId: "patient1",
				},
				{
					RoundId:   1,
					PatientId: "patient2",
				},
				{
					RoundId:   1,
					PatientId: "patient3",
				},
			},
			expectedRoundsCount: 1,
			expectedRounds: []RoundWithTypesAndMembers{
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
		// Create existing round members
		for _, roundMember := range tt.existingRoundMembers {
			db.Create(&roundMember)
		}

		// Call CreateRounds
		CreateRounds(db, tt.currTime)

		// Get rounds with types by joining round_round_types and grouping by round id and timestamp
		var roundsWithTypes []*RoundWithTypesAndMembers
		db.Table("rounds").
			Select("rounds.id, round_timestamp, group_concat(name) as round_types").
			Joins("JOIN round_round_types ON rounds.id = round_round_types.round_id").
			Joins("JOIN round_types ON round_round_types.round_type_id = round_types.id").
			Group("rounds.id, round_timestamp").
			Scan(&roundsWithTypes)

		//Compare with expected
		if len(roundsWithTypes) != tt.expectedRoundsCount {
			t.Errorf("Expected %d rounds, got %d", len(tt.expectedRounds), len(roundsWithTypes))
		}
		for i := range tt.expectedRounds {
			if roundsWithTypes[i].ID != tt.expectedRounds[i].ID ||
				roundsWithTypes[i].RoundTimestamp != tt.expectedRounds[i].RoundTimestamp {
				t.Errorf("Expected %v, got %v", tt.expectedRounds[i], roundsWithTypes[i])
			}
		}

		// Get rounds with members by joining round_members and grouping by round id and timestamp
		var roundsWithMembers []*RoundWithTypesAndMembers
		db.Table("rounds").
			Select("rounds.id, round_timestamp, group_concat(patient_id) as round_members").
			Joins("JOIN round_members ON rounds.id = round_members.round_id").
			Group("rounds.id, round_timestamp").
			Scan(&roundsWithMembers)

		//Compare with expected
		for i := range tt.expectedRounds {
			if roundsWithMembers[i].ID != tt.expectedRounds[i].ID ||
				roundsWithMembers[i].RoundTimestamp != tt.expectedRounds[i].RoundTimestamp {
				t.Errorf("Expected %v, got %v", tt.expectedRounds[i], roundsWithMembers[i])
			}
		}
	}
}
