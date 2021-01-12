package controller

import (
	"github.com/go-chi/render"
	"github.com/onflow/cadence"
	"net/http"
)

type UtilsHandler struct{}

func NewUtilsHandler() *UtilsHandler {
	return &UtilsHandler{}
}

func (u *UtilsHandler) VersionHandler(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, struct {
		Version string `json:"version"`
	}{
		cadence.Version,
	})
}
