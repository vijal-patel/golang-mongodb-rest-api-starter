package models

import "time"

type User struct {
	Id               string     `bson:"_id,omitempty" json:"id"`
	Name             string     `bson:"name,omitempty" json:"name"`
	Email            string     `bson:"email,omitempty" json:"email"`
	Password         string     `bson:"password,omitempty" json:"-"`
	CreatedAt        time.Time  `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt        time.Time  `bson:"updatedAt,omitempty" json:"updatedAt"`
	CreatedById      string     `bson:"createdById,omitempty" json:"createdById"`
	UpdatedById      string     `bson:"updatedById,omitempty" json:"updatedById"`
	Roles            []string   `bson:"roles,omitempty" json:"roles"`
	Permissions      []string   `bson:"permissions,omitempty" json:"permissions"`
	Confirmed        bool       `bson:"confirmed,omitempty" json:"confirmed"`
	ConfirmOtp       string     `bson:"confirmOtp,omitempty" json:"-"`
	LastConfirmOtpAt *time.Time `bson:"confirmOtpAt,omitempty" json:"-"`
	LoginOtp         string     `bson:"loginOtp,omitempty" json:"-"`
	OrganizationId   string     `bson:"organizationId" json:"organizationId"`
}

type CreateUserRequest struct {
	Name           string   `json:"name" validate:"required,nourl,max=255"`
	Email          string   `json:"email" validate:"required,email,max=255"`
	Role           string   `json:"role" validate:"required,role"`
	Permissions    []string `bson:"permissions,omitempty" json:"permissions"`
	LoginOtp       string
	HashedLoginOtp string
}

type BulkCreateUsersRequest struct {
	Users []CreateUserRequest `json:"users" validate:"required,min=1,dive,min=1"`
}
type UpdateUserRequest struct {
	Name           string   `bson:"name,omitempty" json:"name,omitempty" validate:"omitempty,nourl,min=1,max=255"`
	Email          string   `bson:"email,omitempty" json:"email,omitempty" validate:"omitempty,email,max=255"`
	Roles          []string `bson:"roles,omitempty"`
	Permissions    []string `bson:"permissions" json:"permissions"`
	UpdatedById    string   `bson:"updatedById,omitempty"`
	ConfirmOtp     string   `bson:"confirmOtp,omitempty"`
	Confirmed      bool     `bson:"confirmed,omitempty"`
	OrganizationId string   `bson:"organizationId,omitempty"`
}

type UserResponse struct {
	Id            string   `json:"_id,omitempty"`
	Name          string   `json:"name,omitempty"`
	Email         string   `json:"email,omitempty"`
	Roles         []string `json:"roles,omitempty"`
	Organizations []string `json:"organizations,omitempty"`
	Confirmed     bool     `json:"confirmed,omitempty"`
	Permissions   []string `json:"permissions"`
}

type UpdateUserOTPRequest struct {
	Confirmed  bool   `bson:"confirmed,omitempty"`
	ConfirmOtp string `bson:"confirmOtp"`
}

type UpdatePasswordRequest struct {
	Password string `bson:"password"`
}
