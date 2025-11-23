package handler

import (
	"github.com/sadsnake231/pr-reviewer-service/internal/domain"
)

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func errorResponse(code domain.ErrorCode, message string) ErrorResponse {
	resp := ErrorResponse{}
	resp.Error.Code = string(code)
	resp.Error.Message = message
	return resp
}
