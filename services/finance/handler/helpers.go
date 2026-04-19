package handler

import (
	"errors"
	"net/http"

	"github.com/DB-Vincent/personal-finance/pkg/response"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func userIDFromHeader(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(r.Header.Get("X-User-ID"))
}

func validationErrors(err error) []response.ErrorDetail {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil
	}
	details := make([]response.ErrorDetail, len(ve))
	for i, fe := range ve {
		details[i] = response.ErrorDetail{
			Field:  fe.Field(),
			Reason: fe.Tag(),
		}
	}
	return details
}
