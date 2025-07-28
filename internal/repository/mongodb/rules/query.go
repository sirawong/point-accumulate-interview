package mongodb

import (
	"github.com/sirawong/point-accumulate-interview/internal/errors"
	"go.mongodb.org/mongo-driver/bson"
)

func filterActiveBranchIDWithCategoryIDs(branchIDWithCategoryIDs map[string][]string) (bson.M, error) {
	if len(branchIDWithCategoryIDs) == 0 {
		return nil, errors.ErrInvalidArgument.WithMessage("branchIDWithCategoryIDs is empty")
	}

	filter := bson.A{}
	for branchID, CategoryIDs := range branchIDWithCategoryIDs {
		value := bson.M{"conditions.branch_id": branchID}
		if len(CategoryIDs) > 0 {
			value["conditions.category_ids"] = bson.M{"$in": CategoryIDs}
		}

		filter = append(filter, value)
	}

	return bson.M{"status": "ACTIVE", "$or": filter}, nil
}
