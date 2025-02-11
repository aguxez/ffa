package filewatch

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/aguxez/ffa/models"
)

// ParseFoods reads and parses food data from CSV
func ParseFoods(path string) ([]models.Food, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening foods file: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}
	if len(header) != 1 || header[0] != "Food Name" {
		return nil, fmt.Errorf("invalid header format: expected ['Food Name'], got %v", header)
	}

	var foods []models.Food
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading record: %w", err)
		}

		if len(record) != 1 {
			return nil, fmt.Errorf("invalid record format at food: %v", record)
		}

		foods = append(foods, models.Food{Name: record[0]})
	}

	return foods, nil
}

// ParseMacroData reads and parses macro data from CSV
func ParseMacroData(path string) ([]models.MacroDay, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening macro data file: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}

	expectedHeader := []string{
		"Date", "Expenditure", "Trend Weight (kg)", "Weight (kg)",
		"Calories (kcal)", "Protein (g)", "Fat (g)", "Carbs (g)",
		"Target Calories (kcal)", "Target Protein (g)", "Target Fat (g)", "Target Carbs (g)",
	}
	if len(header) != len(expectedHeader) {
		return nil, fmt.Errorf("invalid header length: expected %d columns, got %d", len(expectedHeader), len(header))
	}
	for i, h := range header {
		if h != expectedHeader[i] {
			return nil, fmt.Errorf("invalid header: expected %s at position %d, got %s", expectedHeader[i], i, h)
		}
	}

	var days []models.MacroDay
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading record: %w", err)
		}

		if len(record) != 12 {
			return nil, fmt.Errorf("invalid record length: %v", record)
		}

		date, err := time.Parse("1/2/2006", record[0])
		if err != nil {
			return nil, fmt.Errorf("parsing date %s: %w", record[0], err)
		}

		expenditure, err := strconv.Atoi(record[1])
		if err != nil {
			return nil, fmt.Errorf("parsing expenditure %s: %w", record[1], err)
		}

		trendWeight, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return nil, fmt.Errorf("parsing trend weight %s: %w", record[2], err)
		}

		weight, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			return nil, fmt.Errorf("parsing weight %s: %w", record[3], err)
		}

		days = append(days, models.MacroDay{
			Date:        date,
			Expenditure: expenditure,
			TrendWeight: trendWeight,
			Weight:      weight,
			Actual: models.MacroInfo{
				Calories: parseInt(record[4]),
				Protein:  parseInt(record[5]),
				Fat:      parseInt(record[6]),
				Carbs:    parseInt(record[7]),
			},
			Target: models.MacroInfo{
				Calories: parseInt(record[8]),
				Protein:  parseInt(record[9]),
				Fat:      parseInt(record[10]),
				Carbs:    parseInt(record[11]),
			},
		})
	}

	return days, nil
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
