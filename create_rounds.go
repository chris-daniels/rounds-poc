package main

import (
	"fmt"
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

		// If there is a last round, and it is not too old, skip
		lastRoundTime, err := time.Parse(time.RFC3339, lastRound.RoundTimestamp)
		if err != nil {
			panic("Failed to parse last round time")
		}
		if lastRound.ID != 0 &&
			lastRoundTime.Add(time.Duration(roundType.DurationAmt)*time.Minute).After(t) {
			continue
		}

		// Look to see if a round already exists for this time
		round, err := getRoundForTime(db, t)
		if err != nil {
			panic("Failed to get round for time")
		}

		// If a round already exists, add the round type to it
		if round.ID != 0 {
			db.Create(&RoundRoundType{
				RoundID:     round.ID,
				RoundTypeID: roundType.ID,
			})
			continue
		}

		// Create a new round
		round = Round{
			RoundTimestamp: t.Format(time.RFC3339),
			Status:         "CREATED",
		}
		db.Create(&round)

		// Add the round type to the round round type table
		db.Create(&RoundRoundType{
			RoundID:     round.ID,
			RoundTypeID: roundType.ID,
		})
	}

	fmt.Println("Hello, World!")
}
