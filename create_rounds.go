package main

import (
	"time"

	"gorm.io/gorm"
)

// Create rounds for a given current time
// In production, whatever async task runner we use would call this function and pass in the appropriate time
func CreateRounds(db *gorm.DB, currTime time.Time) {
	// Fetch all round configs for clinic
	roundConfigs, err := getRoundConfigs(db)
	if err != nil {
		panic("Failed to get round configs")
	}

	for _, roundConfig := range roundConfigs {
		// If round config is not enabled, skip
		if !roundConfig.Enabled {
			continue
		}

		// TODO: the "enabled time window" logic would go here

		// Get the round type for this config
		roundType, err := getRoundType(db, roundConfig.RoundTypeId)
		if err != nil {
			panic("Failed to get round types")
		}

		// If round type duration is not minutes, let's panic for now
		// A more complete implementation would have a flexible duration type
		if roundType.DurationUnit != "minutes" {
			panic("Only minutes round duration is supported for now")
		}

		// Get the most recent round of this type
		lastRound, err := getLastRoundForType(db, roundType.ID)
		if err != nil {
			panic("Failed to get last round")
		}

		// Given a last round (if any), and a round type, fill the time window with rounds
		fillTimeWithRounds(db, lastRound, roundType, currTime)
	}
}

// Given a last round (if any), and a round type, fill the time window with rounds
// Eventually, we would need to pass along building/program info here to get the right patients to add to the round
func fillTimeWithRounds(db *gorm.DB, lastRound Round, roundType RoundType, currTime time.Time) {
	// Declare start time as 12 hours before the current time
	startTime := currTime.Add(-12 * time.Hour)

	// If last round is valid, set start time to last round's timestamp, plus duration of round type
	if lastRound.ID != 0 {
		startTime, _ = time.Parse(time.RFC3339, lastRound.RoundTimestamp)
		startTime = startTime.Add(time.Duration(roundType.DurationAmt) * time.Minute)
	}

	// We will update tempTime as we walk through the time window
	tempTime := startTime

	// Walk forward in time, creating rounds as needed, until we reach the current time
	for !tempTime.After(currTime) {
		// Look to see if a round already exists for this time
		round, err := getRoundForTime(db, tempTime)
		if err != nil {
			panic("Failed to get round for time")
		}

		// If a round doesn't exist, create a new round at this time
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

		// Add members to the round
		// Eventually, we would need to pass along building/program info here to get the right patients to add to the round
		addMembersToRound(db, round.ID, roundType.ID)

		// Move to the next time slice
		tempTime = tempTime.Add(time.Duration(roundType.DurationAmt) * time.Minute)
	}

}

// Add members to the round
// Eventually, we would need to pass along building/program info here to get the right patients to add to the round
func addMembersToRound(db *gorm.DB, roundId uint, roundTypeId uint) {
	// Get existing round members for this roundId
	roundMembers, err := getRoundMembersForRound(db, roundId)
	if err != nil {
		panic("Failed to get round members")
	}

	// Put existing patient ids in a set
	patientIds := make(map[string]bool)
	for _, roundMember := range roundMembers {
		patientIds[roundMember.PatientId] = true
	}

	// Get round assignments for this roundTypeId
	roundAssignments, err := getRoundAssignmentsForRoundType(db, roundTypeId)
	if err != nil {
		panic("Failed to get round assignments")
	}

	// Iterate through round assignments, adding any new patients to the round
	for _, roundAssignment := range roundAssignments {
		// If patient id is not in the set, add it
		if _, ok := patientIds[roundAssignment.PatientId]; !ok {
			db.Create(&RoundMember{
				RoundId:   roundId,
				PatientId: roundAssignment.PatientId,
			})
		}
	}
}
