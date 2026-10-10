package web

import (
	"BE2_API_Shop/shop/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type errorResponse struct {
	Error string `json:"error" example:"incorrect input"`
}

func writeBinary(
	w http.ResponseWriter,
	status int,
	data []byte,
	filename string,
) error {
	w.Header().Set("Content-Type",
		"Application/octet-stream",
	)
	w.Header().Set(
		"Content-Disposition",
		fmt.Sprintf(
			`attachment; filename="%s"`,
			filename,
		),
	)
	w.WriteHeader(status)
	_, err := w.Write(data)
	return err
}
func grpcErrorToHTTP(err error) int {
	switch status.Code(err) {
	case codes.InvalidArgument:
		return http.StatusBadRequest

	case codes.Unauthenticated:
		return http.StatusUnauthorized

	case codes.PermissionDenied:
		return http.StatusForbidden

	case codes.NotFound:
		return http.StatusNotFound

	case codes.AlreadyExists:
		return http.StatusConflict

	default:
		return http.StatusInternalServerError
	}
}
func (app *Application) handleServiceError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, service.ErrInvalidClientData),
		errors.Is(err, service.ErrInvalidAddressData),
		errors.Is(err, service.ErrIncorrectInput),
		errors.Is(err, service.ErrInvalidID):
		errorJSON(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, service.ErrNotFound):
		errorJSON(w, http.StatusNotFound, err.Error())

	case errors.Is(err, service.ErrConflict):
		errorJSON(w, http.StatusConflict, err.Error())

	default:
		errorJSON(w, http.StatusInternalServerError, "internal server error")
	}
}

func readIntQuery(r *http.Request, name string) (*int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf(
			"%s must be an integer",
			name,
		)
	}
	return &n, nil
}

func readID(r *http.Request) (uuid.UUID, error) {

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {

		return uuid.Nil, errors.New("invalid id")
	}
	return id, nil
}

func errorJSON(w http.ResponseWriter, status int, messege string) {
	writeJSON(w, status, errorResponse{Error: messege})
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "Application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func readJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(dst)

	if err != nil {
		return mapJSONError(err)
	}

	err = decoder.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must contain only a single JSON object")
	}

	return nil
}

func mapJSONError(err error) error {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError

	switch {
	case errors.As(err, &syntaxError):
		return fmt.Errorf(
			"body contains badly-formed JSON at character %d",
			syntaxError.Offset,
		)

	case errors.Is(err, io.ErrUnexpectedEOF):
		return errors.New("body contains badly-formed JSON")

	case errors.As(err, &unmarshalTypeError):
		if unmarshalTypeError.Field != "" {
			return fmt.Errorf(
				"body contains incorrect JSON type for field %q",
				unmarshalTypeError.Field,
			)
		}

		return fmt.Errorf(
			"body contains incorrect JSON type at character %d",
			unmarshalTypeError.Offset,
		)

	case errors.Is(err, io.EOF):
		return errors.New("body must not be empty")

	default:
		return err
	}
}
