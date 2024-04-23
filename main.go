package main

import (
	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDatabase() *gorm.DB {
	// Clear the database so we start fresh Delete test.db
	os.Remove("test.db")

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&RoundType{})
	db.AutoMigrate(&RoundConfig{})
	db.AutoMigrate(&RoundAssignment{})
	db.AutoMigrate(&Round{})
	db.AutoMigrate(&RoundRoundType{})
	db.AutoMigrate(&RoundMember{})

	return db
}

func main() {
	fmt.Println("Hello, World!")

	// // Set up database
	// db := setupDatabase()

}
