package models

type PatchLogLevelRequest struct {
	Level string `json:"level" validate:"required"`
}
