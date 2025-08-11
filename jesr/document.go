package jesr

import "github.com/google/uuid"

type Document struct {
	Uuid          uuid.UUID    `json:"documentUUID" example:"ba3ca973-5052-4030-a528-39b49736d8ad"`
	OwnerUUID     *uuid.UUID   `json:"ownerUUID,omitempty"`
	OwnerType     *string      `json:"ownerType,omitempty"`
	PdfBase64     *string      `json:"pdfBase64,omitempty"`
	SelectionData *[]Selection `json:"selectionData,omitempty"`
}

type Selection struct {
	Uuid            uuid.UUID                  `json:"selectionUUID"`
	DocumentUUID    *uuid.UUID                 `json:"documentUUID,omitempty"`
	IsComplete      bool                       `json:"isComplete,omitempty"`
	Settings        *string                    `json:"settings,omitempty"`
	SelectionBounds *map[int][]SelectionBounds `json:"selectionBounds,omitempty"`
}

type SelectionBounds struct {
	SelectionMethod *string `json:"extract_method" example:"None"`
	X1              float64 `json:"x1" example:"43.122"`
	X2              float64 `json:"x2" example:"13"`
	Y1              float64 `json:"y1" example:"52.125"`
	Y2              float64 `json:"y2" example:"27.853"`
}
