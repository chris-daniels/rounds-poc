package main

import (
	"time"

	"gorm.io/gorm"
)

// Get all round configs for clinic
func getRoundConfigs(db *gorm.DB) ([]RoundConfig, error) {
	var roundConfigs []RoundConfig
	db.Find(&roundConfigs)
	return roundConfigs, nil
}

// Get round type by ID
func getRoundType(db *gorm.DB, roundTypeId uint) (RoundType, error) {
	var roundType RoundType
	db.Where("id = ?", roundTypeId).First(&roundType)
	return roundType, nil
}

// Get the most recent round of a given type
func getLastRoundForType(db *gorm.DB, roundTypeId uint) (Round, error) {
	var round Round
	db.Joins("JOIN round_round_types ON rounds.id = round_round_types.round_id").
		Where("round_round_types.round_type_id = ?", roundTypeId).
		Order("rounds.created_at desc").
		First(&round)
	return round, nil
}

// Get the round for a given time
func getRoundForTime(db *gorm.DB, t time.Time) (Round, error) {
	var round Round
	db.Where("round_timestamp = ?", t.Format(time.RFC3339)).First(&round)
	return round, nil
}

// Get the round members for a given round id
func getRoundMembersForRound(db *gorm.DB, roundId uint) ([]RoundMember, error) {
	var roundMembers []RoundMember
	db.Where("round_id = ?", roundId).Find(&roundMembers)
	return roundMembers, nil
}

// Get round assignments for a given round type id
func getRoundAssignmentsForRoundType(db *gorm.DB, roundTypeId uint) ([]RoundAssignment, error) {
	var roundAssignments []RoundAssignment
	db.Where("round_type_id = ?", roundTypeId).Find(&roundAssignments)
	return roundAssignments, nil
}
