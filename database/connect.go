package database

import (
	"github.com/aguxez/ffa/config"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	Conn   *gorm.DB
	Config config.AppConfig
}

// NewDatabase creates a connection to the database using the DATABASE_URL environment
// variable. If the connection is successful, it returns a pointer to the GORM
// database connection. If there is an error, it logs an error message and
// returns a nil *gorm.DB and the error.
func NewDatabase(cfg config.AppConfig) (*Database, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		// Let us handle transactions manually so we don't have performance
		// pernalization by default.
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return &Database{}, err
	}

	log.Info().Msg("Connected to database. Checking if we need to run migrations...")
	err = db.AutoMigrate(
		&User{},
		&UserFood{},
		&UserPlanSchedule{},
		&UserDietaryRestriction{},
		&UserMealPlan{},
		&UserMealPlanFood{},
		&UserPayment{},
	)

	if err != nil {
		return &Database{}, err
	}

	return &Database{Conn: db, Config: cfg}, nil
}
