package models

type Note struct {
	Name           string `json:"name"`
	Text           string `json:"text"`
	Status         string `json:"status"`
	DelegationUser string `json:"delegationUser"`
	CompletionTime string `json:"completionTime"`
}
