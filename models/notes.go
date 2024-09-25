package models

type Note struct {
	Id             int64  `json:"id,omitempty"`
	Name           string `json:"name"`
	Text           string `json:"text"`
	Status         string `json:"status"`
	DelegationUser string `json:"delegationUser"`
	CompletionTime string `json:"completionTime"`
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
