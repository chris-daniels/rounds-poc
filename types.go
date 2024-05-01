package main

import (
	"gorm.io/gorm"
)

type RoundType struct {
	gorm.Model
	ID           uint   `json:"id" gorm:"primaryKey"`
	Name         string `json:"name"`
	DurationAmt  int    `json:"durationAmt"`
	DurationUnit string `json:"durationUnit"`
}

type RoundConfig struct {
	gorm.Model
	ID          uint `json:"id" gorm:"primaryKey"`
	RoundTypeId uint `json:"roundType"`
	Enabled     bool `json:"enabled"`
}

type RoundAssignment struct {
	gorm.Model
	ID          uint   `json:"id" gorm:"primaryKey"`
	RoundTypeId uint   `json:"roundType"`
	PatientId   string `json:"patientId"`
}

type Round struct {
	gorm.Model
	ID             uint   `json:"id" gorm:"primaryKey"`
	RoundTimestamp string `json:"roundTimestamp"`
	Status         string `json:"status"`
}

type RoundRoundType struct {
	gorm.Model
	RoundID     uint `json:"round"`
	RoundTypeID uint `json:"roundType"`
}

type RoundMember struct {
	gorm.Model
	ID        uint   `json:"id" gorm:"primaryKey"`
	RoundId   uint   `json:"round"`
	Status    string `json:"status"`
	PatientId string `json:"patientId"`
}

type StartRoundsItem struct {
	RoundTimestamp string `json:"roundTimestamp"`
	Status         string `json:"status"`
}
