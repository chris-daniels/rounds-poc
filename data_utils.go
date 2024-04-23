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
