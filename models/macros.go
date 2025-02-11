package models

import "time"

type MacroDay struct {
	Date        time.Time
	Expenditure int
	TrendWeight float64
	Weight      float64
	Actual      MacroInfo
	Target      MacroInfo
}

type MacroInfo struct {
	Calories int
	Protein  int
	Fat      int
	Carbs    int
}
