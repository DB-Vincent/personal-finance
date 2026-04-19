package response

import (
	"encoding/json"
	"net/http"
)

type ErrorDetail struct {
	Field  string `json:"field,omitempty"`
	Reason string `json:"reason"`
}

type ErrorBody struct {
	Code    int           `json:"code"`
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type errorResponse struct {
	Error ErrorBody `json:"error"`
}

var statusText = map[int]string{
	http.StatusBadRequest:          "INVALID_ARGUMENT",
	http.StatusUnauthorized:        "UNAUTHORIZED",
	http.StatusForbidden:           "FORBIDDEN",
	http.StatusNotFound:            "NOT_FOUND",
	http.StatusConflict:            "ALREADY_EXISTS",
	http.StatusInternalServerError: "INTERNAL",
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, status int, message string, details ...ErrorDetail) {
	s, ok := statusText[status]
	if !ok {
		s = "UNKNOWN"
	}

	resp := errorResponse{
		Error: ErrorBody{
			Code:    status,
			Status:  s,
			Message: message,
		},
	}
	if len(details) > 0 {
		resp.Error.Details = details
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

type ListResponse struct {
	Items         any    `json:"items"`
	NextPageToken string `json:"next_page_token,omitempty"`
	TotalSize     int64  `json:"total_size"`
}

func List(w http.ResponseWriter, items any, nextPageToken string, totalSize int64) {
	JSON(w, http.StatusOK, ListResponse{
		Items:         items,
		NextPageToken: nextPageToken,
		TotalSize:     totalSize,
	})
}
