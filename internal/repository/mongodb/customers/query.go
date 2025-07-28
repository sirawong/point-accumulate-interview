package mongodb

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirawong/point-accumulate-interview/internal/domain/entity"
	apperr "github.com/sirawong/point-accumulate-interview/internal/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func filterCustomerIDs(customerIDs []string) bson.M {
	if len(customerIDs) == 0 {
		return bson.M{}
	}
	return bson.M{"customer_id": bson.M{"$in": customerIDs}}
}

func operationUpdateCustomerPoint(updateCustomer []entity.UpdateCustomer) ([]mongo.WriteModel, error) {
	if len(updateCustomer) == 0 {
		return nil, apperr.ErrInvalidArgument.Wrap(errors.New("updateCustomer is empty"))
	}

	operations := make([]mongo.WriteModel, 0, len(updateCustomer))
	for _, customer := range updateCustomer {
		records, err := fromRecords(customer.Records)
		if err != nil {
			return nil, err
		}
		incPayload := bson.M{"points": customer.PointsToAdd}
		for date, incValue := range customer.PointsByDate {
			fieldPath := fmt.Sprintf("points_by_date.%s", date)
			incPayload[fieldPath] = incValue
		}

		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{
				"customer_id": customer.CustomerID,
			}).
			SetUpdate(bson.M{
				"$inc": incPayload,
				"$set": bson.M{
					"last_purchase_date": customer.LastPurchaseDate,
					"updated_at":         time.Now(),
				},
				"$push": bson.M{
					"records": bson.M{"$each": records},
				},
				"$setOnInsert": bson.M{
					"customer_id": customer.CustomerID,
					"created_at":  time.Now(),
				},
			}).
			SetUpsert(true)

		operations = append(operations, model)
	}

	return operations, nil
}
