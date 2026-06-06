package models

type ExtractionResponse map[string]map[string][]struct {
	Text                string
	TextCoordinates     Coordinates
	SelectionCoordinate Coordinates
}

type Coordinates struct {
	X1 float64 `json:"x1" example:"43.122"`
	Y1 float64 `json:"y1" example:"52.125"`
	X2 float64 `json:"x2" example:"13"`
	Y2 float64 `json:"y2" example:"27.853"`
}
