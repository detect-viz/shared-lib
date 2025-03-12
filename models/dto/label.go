package dto

type LabelDTO struct {
	ID     int64    `json:"id"`
	Key    string   `json:"key"`
	Values []string `json:"values"`
}
