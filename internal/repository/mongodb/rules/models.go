package mongodb

import (
	"log"

	"github.com/shopspring/decimal"
	"github.com/sirawong/point-accumulate-interview/internal/domain/entity"
	"github.com/sirawong/point-accumulate-interview/internal/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Rule struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Name       string             `bson:"name"`
	RuleType   entity.RuleType    `bson:"rule_type"`
	Conditions Conditions         `bson:"conditions"`
	Reward     Reward             `bson:"reward"`
	Status     string             `bson:"status"`
}

type Reward struct {
	Value     int64    `bson:"value"`
	RatioUnit *float64 `bson:"ratio_unit,omitempty"`
}

type Conditions struct {
	MinAmount   primitive.Decimal128 `bson:"min_amount"`
	BranchID    string               `bson:"branch_id"`
	CategoryIDs []string             `bson:"category_ids"`
}

func (r Rule) ToDomain() (*entity.Rule, error) {
	value, err := decimal.NewFromString(r.Conditions.MinAmount.String())
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}

	return &entity.Rule{
		ID:       r.ID.Hex(),
		Name:     r.Name,
		RuleType: r.RuleType,
		Status:   r.Status,
		Reward: entity.Reward{
			Value:     r.Reward.Value,
			RatioUnit: r.Reward.RatioUnit,
		},
		Conditions: entity.Conditions{
			MinAmount:  value,
			BranchID:   r.Conditions.BranchID,
			CategoryID: r.Conditions.CategoryIDs,
		},
	}, nil
}

type Rules []*Rule

func (r Rules) ToDomain() []entity.Rule {
	rules := make([]entity.Rule, 0, len(r))
	for _, rule := range r {
		value, err := rule.ToDomain()
		if err != nil {
			log.Println(errors.ErrInternal, err)
			continue
		}
		rules = append(rules, *value)
	}
	return rules
}
