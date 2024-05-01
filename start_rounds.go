package main

import (
	"sort"
	"time"

	"gorm.io/gorm"
)

func StartRounds(db *gorm.DB, startTime time.Time, currTime time.Time) ([]StartRoundsItem, error) {
	// Fetch all round configs for the clinic
	roundConfigs, err := getRoundConfigs(db)
	if err != nil {
		panic("Failed to get round configs")
	}

	// Fetch all rounds for the clinic udring time window
	rounds, err := getRounds(db, startTime, currTime)
	if err != nil {
		panic("Failed to get rounds")
	}

	// Put exisisting rounds in a map as round timestamp -> StartRoundItems
	// TODO: When we add building & program info, we could build a composite key here
	roundsMap := make(map[string]StartRoundsItem)
	for _, round := range rounds {
		roundsMap[round.RoundTimestamp] = StartRoundsItem{
			Status:         round.Status,
			RoundTimestamp: round.RoundTimestamp,
		}
	}

	// Create rounds for each round config
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

		// Walk through the time window and add new rounds to map as needed
		roundsMap = createRoundsForConfig(roundType, startTime, currTime, roundsMap)
	}

	// Convert the map to a slice
	var startRounds []StartRoundsItem
	for _, v := range roundsMap {
		startRounds = append(startRounds, v)
	}

	// Sort by round timestamp
	sort.Slice(startRounds, func(i, j int) bool {
		return startRounds[i].RoundTimestamp < startRounds[j].RoundTimestamp
	})

	// Mark old rounds as MISSED
	startRounds = formatMissedRounds(startRounds, currTime)

	// Add a next round if needed
	startRounds = appendFutureRoundIfNeeded(db, startRounds, currTime, roundConfigs)

	return startRounds, nil
}

// Create rounds for a given round type and add to the rounds map
func createRoundsForConfig(
	roundType RoundType, startTime time.Time, currTime time.Time, roundsMap map[string]StartRoundsItem) map[string]StartRoundsItem {
	tempTime := startTime

	// Walk forward in time, creating rounds as needed, until we reach the current time
	for !tempTime.After(currTime) {
		_, ok := roundsMap[tempTime.Format(time.RFC3339)]
		if !ok {
			existingRound := StartRoundsItem{
				Status:         "NOT_STARTED",
				RoundTimestamp: tempTime.Format(time.RFC3339),
			}
			roundsMap[tempTime.Format(time.RFC3339)] = existingRound
		}
		tempTime = tempTime.Add(time.Duration(roundType.DurationAmt) * time.Minute)
	}
	return roundsMap
}

// Mark old rounds as MISSED
func formatMissedRounds(roundItems []StartRoundsItem, currTime time.Time) []StartRoundsItem {
	// Mark all rounds that are NOT_STARTED as MISSED if they are 30 minutes old compared to currTime
	for i, round := range roundItems {
		roundTime, _ := time.Parse(time.RFC3339, round.RoundTimestamp)
		if round.Status == "NOT_STARTED" && currTime.Sub(roundTime) >= 30*time.Minute {
			roundItems[i].Status = "MISSED"
		}
	}
	return roundItems
}

// Add a future round if there is no NOT_STARTED round in the list
func appendFutureRoundIfNeeded(db *gorm.DB, roundItems []StartRoundsItem, currTime time.Time, configs []RoundConfig) []StartRoundsItem {
	// Look for a round that is NOT_STARTED anywhere in the list
	hasNotStarted := false
	for _, round := range roundItems {
		if round.Status == "NOT_STARTED" {
			hasNotStarted = true
			break
		}
	}

	if hasNotStarted {
		return roundItems
	}

	// Find the smallest round config duration. This will be the next round.
	// When we have building/program info for these configs, we might have a few configs we need to hold onto
	minDuration := 0
	for _, config := range configs {
		roundType, err := getRoundType(db, config.RoundTypeId)
		if err != nil {
			panic("Failed to get round type")
		}
		if minDuration == 0 || roundType.DurationAmt < minDuration {
			minDuration = roundType.DurationAmt
		}
	}

	newRound := StartRoundsItem{
		Status:         "NOT_STARTED",
		RoundTimestamp: currTime.Add(time.Duration(minDuration) * time.Minute).Format(time.RFC3339),
	}
	roundItems = append(roundItems, newRound)

	return roundItems
}
