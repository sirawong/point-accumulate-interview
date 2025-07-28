package mongodb

import (
	"context"

	"github.com/sirawong/point-accumulate-interview/internal/domain/entity"
	"github.com/sirawong/point-accumulate-interview/internal/domain/repository"
	"github.com/sirawong/point-accumulate-interview/internal/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	RuleCollection = "rules"
)

type ruleRepository struct {
	collection *mongo.Collection
}

func NewRuleRepository(db *mongo.Database) repository.RuleRepository {
	return &ruleRepository{
		collection: db.Collection(RuleCollection),
	}
}

func (r ruleRepository) GetActiveRules(ctx context.Context, branchIDWithCategoryIDs map[string][]string) ([]entity.Rule, error) {
	filter, err := filterActiveBranchIDWithCategoryIDs(branchIDWithCategoryIDs)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}
	defer cursor.Close(ctx)

	var rules Rules

	err = cursor.All(ctx, &rules)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}

	return rules.ToDomain(), nil
}
