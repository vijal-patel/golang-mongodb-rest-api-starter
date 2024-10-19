package models

import (
	"time"
)

type Post struct {
	Id             string    `bson:"_id,omitempty" json:"id"`
	Name           string    `bson:"name,omitempty" json:"name"`
	OrganizationId string    `bson:"organizationId,omitempty" json:"organizationId"`
	CreatedAt      time.Time `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt      time.Time `bson:"updatedAt,omitempty" json:"updatedAt"`
	CreatedById    string    `bson:"createdById,omitempty" json:"createdById"`
	UpdatedById    string    `bson:"updatedById,omitempty" json:"updatedById"`
}

type CreatePostRequest struct {
	Name        string `bson:"name,omitempty" json:"name" validate:"required,max=255"`
	CreatedById string `bson:"createdById,omitempty" `
}

type UpdatePostRequest struct {
	Name        string `bson:"name,omitempty" json:"name" validate:"max=255,max=255"`
	UpdatedById string
}
