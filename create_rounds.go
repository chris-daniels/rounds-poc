package main

import (
	"time"

	"gorm.io/gorm"
)

// Given a last round and a round type, "walk forward" until the current time, creating rounds as needed
func fillTimeWithRounds(db *gorm.DB, lastRound Round, roundType RoundType, currTime time.Time) {
	// Declare start time as 12 hours before the current time
	startTime := currTime.Add(-12 * time.Hour)

	// If last round is valid, set start time to last round's timestamp
	if lastRound.ID != 0 {
		startTime, _ = time.Parse(time.RFC3339, lastRound.RoundTimestamp)
		startTime = startTime.Add(time.Duration(roundType.DurationAmt) * time.Minute)
	}

	tempTime := startTime

	// Walk forward in time, creating rounds as needed
	for !tempTime.After(currTime) {
		// Look to see if a round already exists for this time
		round, err := getRoundForTime(db, tempTime)
		if err != nil {
			panic("Failed to get round for time")
		}

		// If a round doesn't exist, add the round type to it
		if round.ID == 0 {
			// Create a new round
			round = Round{
				RoundTimestamp: tempTime.Format(time.RFC3339),
				Status:         "CREATED",
			}
			db.Create(&round)
		}

		// Add the round type to the round round type table
		db.Create(&RoundRoundType{
			RoundID:     round.ID,
			RoundTypeID: roundType.ID,
		})

		// Move to the next time slice
		tempTime = tempTime.Add(time.Duration(roundType.DurationAmt) * time.Minute)
	}

}

// Create rounds for a given time
func CreateRounds(db *gorm.DB, t time.Time) {
	// Fetch round configs
	roundConfigs, err := getRoundConfigs(db)
	if err != nil {
		panic("Failed to get round configs")
	}

	// Iterate through round configs
	for _, roundConfig := range roundConfigs {
		// If round is not enabled, skip
		if !roundConfig.Enabled {
			continue
		}

		// Get the round type for this config
		roundType, err := getRoundType(db, roundConfig.RoundTypeId)
		if err != nil {
			panic("Failed to get round types")
		}

		// If round type duration is not minutes, let's panic for now
		if roundType.DurationUnit != "minutes" {
			panic("Only minutes round duration is supported for now")
		}

		// Get the most recent round of this type
		lastRound, err := getLastRoundForType(db, roundType.ID)
		if err != nil {
			panic("Failed to get last round")
		}

		// If there is a last round, and it is not too old, fill time with rounds
		fillTimeWithRounds(db, lastRound, roundType, t)
	}
}
