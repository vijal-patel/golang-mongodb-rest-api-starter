package builders

import (
	"go.mongodb.org/mongo-driver/bson"
)

type UpdateDocBuilder struct {
	fieldsToSet  interface{}
	fieldsToPush interface{}
	fieldsToPull interface{}
}

func NewUpdateDocBuilder() *UpdateDocBuilder {
	return &UpdateDocBuilder{}
}

func (updateDocBuilder *UpdateDocBuilder) WithSetFields(fieldsToSet interface{}) (u *UpdateDocBuilder) {
	updateDocBuilder.fieldsToSet = fieldsToSet
	return updateDocBuilder
}

func (updateDocBuilder *UpdateDocBuilder) WithPush(fieldsToPush interface{}) (u *UpdateDocBuilder) {
	updateDocBuilder.fieldsToPush = fieldsToPush
	return updateDocBuilder
}

func (updateDocBuilder *UpdateDocBuilder) WithPull(fieldsToPull interface{}) (u *UpdateDocBuilder) {
	updateDocBuilder.fieldsToPull = fieldsToPull
	return updateDocBuilder
}

func (updateDocBuilder *UpdateDocBuilder) Build() bson.M {
	update := bson.M{
		"$currentDate": bson.M{
			"updatedAt": true, 
		},
	}

	if updateDocBuilder.fieldsToPush != nil {
		update["$push"] = updateDocBuilder.fieldsToPush
	}

	if updateDocBuilder.fieldsToSet != nil {
		update["$set"] = updateDocBuilder.fieldsToSet
	}

	if updateDocBuilder.fieldsToPull != nil {
		update["$pull"] = updateDocBuilder.fieldsToPull
	}

	return update
}
