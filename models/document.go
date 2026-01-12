package models

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	DocumentUUID  uuid.UUID  `json:"documentUUID" example:"ba3ca973-5052-4030-a528-39b49736d8ad"`
	DocumentTitle *string    `json:"documentTitle,omitempty"`
	TimeCreated   *time.Time `json:"timeCreated,omitempty"`
	OwnerUUID     *uuid.UUID `json:"ownerUUID,omitempty"`
	OwnerType     *int8      `json:"ownerType,omitempty"`
	PdfBase64     *string    `json:"pdfBase64,omitempty"`
}

type Meta struct {
	DocumentUUID  uuid.UUID          `json:"documentUUID" example:"ba3ca973-5052-4030-a528-39b49736d8ad"`
	NumberOfPages *int               `json:"numberOfPages,omitempty" example:"31"`
	Width         *float32           `json:"width,omitempty" example:"1920"`
	Height        *float32           `json:"height,omitempty" example:"1080"`
	Images        *map[string]string `json:"images,omitempty"`
	OwnerUUID     *uuid.UUID         `json:"ownerUUID,omitempty" example:"34906041-2d68-45a2-9671-9f0ba89f31a9"`
	OwnerType     *int               `json:"ownerType,omitempty" example:"1"`
}
