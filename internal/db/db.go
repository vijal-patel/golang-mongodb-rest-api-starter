package db

import (
	"context"
	"golang-mongodb-rest-api-starter/internal/constants"
	"golang-mongodb-rest-api-starter/internal/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
)

func DoesEntityExist(coll *mongo.Collection, id, organizationId string, model interface{}) (bool, error) {
	if err := FindById(coll, id, model, bson.D{{Key: "_id", Value: 1}}); err != nil {
		return false, err
	}

	return true, nil
}

func FindOne(coll *mongo.Collection, filter interface{}, model interface{}, projection interface{}) error {
	var opts *options.FindOneOptions
	var err error
	if projection != nil {
		opts = options.FindOne().SetProjection(projection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(constants.DbDefaultContextTimeout)*time.Second)
	defer cancel()

	if err = coll.FindOne(ctx, filter, opts).Decode(model); err != nil && err == mongo.ErrNoDocuments {
		return nil
	}
	return err
}

func FindById(coll *mongo.Collection, id string, model interface{}, projection interface{}) error {
	var opts *options.FindOneOptions
	var err error

	if projection != nil {
		opts = options.FindOne().SetProjection(projection)
	}
	objectId, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objectId}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(constants.DbDefaultContextTimeout)*time.Second)
	defer cancel()

	if err = coll.FindOne(ctx, filter, opts).Decode(model); err != nil && err == mongo.ErrNoDocuments {
		return nil
	}

	return err
}

func FindMany(coll *mongo.Collection, filter interface{}, limit int64, offset int64, models interface{}, projection interface{}, sortOptions interface{}) (bool, int64, error) {
	var totalDocuments int64
	hasNext := false

	opts := options.Find()
	opts = opts.SetSkip(offset).SetLimit(limit)
	if projection != nil {
		opts = opts.SetProjection(projection)
	}

	if sortOptions != nil {
		opts = opts.SetSort(sortOptions)
		collation := &options.Collation{Locale: "en_US"}
		opts = opts.SetCollation(collation)
	} else {
		opts = opts.SetSort(bson.D{{Key: "createdAt", Value: -1}})
	}

	ctxCount, cancelCount := context.WithTimeout(context.Background(), time.Duration(constants.DbDefaultContextTimeout)*time.Second)
	defer cancelCount()

	ctxFind, cancelFind := context.WithTimeout(context.Background(), time.Duration(constants.DbDefaultContextTimeout)*time.Second)
	defer cancelFind()

	ctxAll, cancelAll := context.WithTimeout(context.Background(), time.Duration(constants.DbDefaultContextTimeout)*time.Second)
	defer cancelAll()

	g := new(errgroup.Group)
	g.Go(func() error {
		count, err := coll.CountDocuments(ctxCount, filter)
		if err != nil {
			return err
		}
		totalDocuments = count
		return nil
	})
	g.Go(func() error {
		cursor, err := coll.Find(ctxFind, filter, opts)
		if err != nil {
			return err
		}

		if err = cursor.All(ctxAll, models); err != nil {
			if err == mongo.ErrNoDocuments {
				return nil
			}
			return err
		}
		return nil
	})

	// Wait for all db ops to complete.
	if err := g.Wait(); err != nil {
		return false, 0, err
	}

	if totalDocuments > offset+limit {
		hasNext = true
	}
	return hasNext, totalDocuments, nil
}

func InsertOne(coll *mongo.Collection, model interface{}) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(constants.DbDefaultContextTimeout)*time.Second)
	defer cancel()

	return coll.InsertOne(ctx, model)
}

func InsertMany(coll *mongo.Collection, models []interface{}) (*mongo.InsertManyResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(constants.DbDefaultContextTimeout)*time.Second)
	defer cancel()

	return coll.InsertMany(ctx, models)
}

func UpdateOne(coll *mongo.Collection, id string, update interface{}) (*mongo.UpdateResult, error) {
	objectId, _ := primitive.ObjectIDFromHex(id)
	// // update := bson.D{{Key: "$set", Value: model}}
	// update := bson.M{
	// 	"$set": updatedFields,
	// 	"$currentDate": bson.M{
	// 		"updatedAt": true, // Update the "updatedAt" field with the current date.
	// 	},
	// }
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(constants.DbDefaultContextTimeout)*time.Second)
	defer cancel()

	return coll.UpdateByID(ctx, objectId, update)
}

func UpdateMany(coll *mongo.Collection, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(constants.DbUpdateManyContextTimeout)*time.Second)
	defer cancel()

	return coll.UpdateMany(ctx, filter, update)
}

func DeleteById(coll *mongo.Collection, id string) (*mongo.DeleteResult, error) {
	objectId, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objectId}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(constants.DbDefaultContextTimeout)*time.Second)
	defer cancel()

	return coll.DeleteOne(ctx, filter)
}

func DeleteByMany(coll *mongo.Collection, filter interface{}) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(constants.DbDefaultContextTimeout)*time.Second)
	defer cancel()

	return coll.DeleteMany(ctx, filter)
}

func ConvertToObjectIds(stringIds []string) ([]primitive.ObjectID, error) {
	var objectIds []primitive.ObjectID
	for _, id := range stringIds {
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return objectIds, err
		}
		objectIds = append(objectIds, objectId)
	}
	return objectIds, nil
}

func ConvertToObjectId(stringId string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(stringId)
}

func GetSortType(sortType string) int8 {
	return utils.IfTernary[int8](sortType == constants.SortAsc, 1, -1)
}

func GetFilterCount(coll *mongo.Collection, filter interface{}) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(constants.DbDefaultContextTimeout)*time.Second)
	defer cancel()
	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}
