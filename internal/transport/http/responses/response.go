package responses

type Response struct {
	Code string `json:"code"`
	Message string  `json:"message"`
}

type ErrorResponse struct {
	Error Response `json:"error"`
}

func Error(code, msg string) ErrorResponse {
	return ErrorResponse{
		Error: Response{Code: code, Message: msg},
	}
}