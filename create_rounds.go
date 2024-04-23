package main

import (
	"time"

	"gorm.io/gorm"
)

func getRoundConfigs(db *gorm.DB) ([]RoundConfig, error) {
	var roundConfigs []RoundConfig
	db.Find(&roundConfigs)
	return roundConfigs, nil
}

func getRoundType(db *gorm.DB, roundTypeId uint) (RoundType, error) {
	var roundType RoundType
	db.Where("id = ?", roundTypeId).First(&roundType)
	return roundType, nil
}

func getLastRoundForType(db *gorm.DB, roundTypeId uint) (Round, error) {
	// Get the most recent round of this type
	// Need to join the round_round_type table
	var round Round
	db.Joins("JOIN round_round_types ON rounds.id = round_round_types.round_id").
		Where("round_round_types.round_type_id = ?", roundTypeId).
		Order("rounds.created_at desc").
		First(&round)
	return round, nil
}

func getRoundForTime(db *gorm.DB, t time.Time) (Round, error) {
	var round Round
	db.Where("round_timestamp = ?", t.Format(time.RFC3339)).First(&round)
	return round, nil
}

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
func createRounds(db *gorm.DB, t time.Time) {
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
