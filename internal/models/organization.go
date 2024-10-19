package models

import "time"

type Organization struct {
	Id          string    `bson:"_id,omitempty" json:"id"`
	CreatedAt   time.Time `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt,omitempty" json:"updatedAt"`
	CreatedById string    `bson:"createdById,omitempty" json:"createdById"`
	UpdatedById string    `bson:"updatedById,omitempty" json:"updatedById"`
	Name        string    `bson:"name,omitempty" json:"name"`
	Email       string    `bson:"email,omitempty" json:"email"`
}

type CreateOrganizationRequest struct {
	Name  string `json:"name,omitempty" validate:"required,max=255"`
	Email string `json:"email,omitempty" validate:"required,email,max=255"`
}

type UpdateOrganizationRequest struct {
	Name   string `bson:"name,omitempty" json:"name,omitempty" validate:"min=1,max=255"`
	Email  string `bson:"email,omitempty" json:"email,omitempty" validate:"email,max=255"`
}

func NewOrganizationResponse(organization Organization) *Organization {
	return &organization
}

func NewOrganizationsResponse(organizations []Organization) *[]Organization {
	return &organizations
}
