package models

import "time"

type SharedUser struct {
	Username string `json:"username"`
	Access   string `json:"access"`
}
type Note struct {
	Id             int64        `json:"id,omitempty"`
	Name           string       `json:"name"`
	Text           string       `json:"text"`
	Status         string       `json:"status"`
	DelegationUser string       `json:"delegationUser,omitempty"`
	CompletionTime string       `json:"completionTime,omitempty"`
	SharedUsers    []SharedUser `json:"sharedUsers,omitempty"`
	CreatedAt      time.Time    `json:"createdAt"`
}

type Response struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Status  string      `json:"status"`
	Data    interface{} `json:"data"`
}

func NewResponse(status, message string, error, data interface{}) Response {
	return Response{
		Message: message,
		Error:   error,
		Status:  status,
		Data:    data,
	}
}
