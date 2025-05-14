package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Envelope map[string]any

func WriteJSON(res http.ResponseWriter, status int, data Envelope) error {
	js, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	res.Write(js)
	return nil
}

func ReadIdParam(req *http.Request) (int64, error) {
	idParam := chi.URLParam(req, "id")

	if idParam == "" {
		return 0, errors.New("invalid id param")
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return 0, errors.New("invalid ID parameter type")
	}
	return id, nil
}
