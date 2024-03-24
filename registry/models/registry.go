package models

type Registry struct {
	DbName   string `json:"db_name"`
	DocCount int    `json:"doc_count"`
}
