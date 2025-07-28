package entity

import (
	"github.com/shopspring/decimal"
)

type RuleType string

const (
	PercentageRule RuleType = "PERCENTAGE"
	FixedPointRule RuleType = "FIXED_POINT"
	RatioRule      RuleType = "RATIO"
)

type Rule struct {
	ID         string
	Name       string
	RuleType   RuleType
	Conditions Conditions
	Reward     Reward
	Status     string
}

type Reward struct {
	Value     int64
	RatioUnit *float64
}

type Conditions struct {
	MinAmount  decimal.Decimal
	BranchID   string
	CategoryID []string
}
