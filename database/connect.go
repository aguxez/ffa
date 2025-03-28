package database

import (
	"os"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	Conn *gorm.DB
}

// Connect creates a connection to the database using the DATABASE_URL environment
// variable. If the connection is successful, it returns a pointer to the GORM
// database connection. If there is an error, it logs an error message and
// returns a nil *gorm.DB and the error.
func Connect() (*Database, error) {
	dbUrl := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		return &Database{}, err
	}

	log.Info().Msg("Connected to database. Checking if we need to run migrations...")
	db.AutoMigrate(
		&User{},
		&UserFood{},
		&UserPlanSchedule{},
		&UserDietaryRestriction{},
		&UserMealPlan{},
		&UserMealPlanFood{},
	)

	return &Database{Conn: db}, nil
}
