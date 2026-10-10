package web

import (
	"BE2_API_Shop/shop/internal/dto"
	"net/http"
)

// ClientAdd godoc
//
// @Summary Add client
// @Description Create a new client
// @Tags clients
// @Accept json
// @Security BearerAuth
// @Produce json
// @Param client body dto.ClientRequest true "Client data"
// @Success 201 "Created"
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /clients [post]
func (app *Application) ClientAdd(w http.ResponseWriter, r *http.Request) {
	var req dto.ClientRequest
	err := readJSON(r, &req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	err = app.services.Client.AddClient(r.Context(), req)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// ClientRemove godoc
//
// @Summary Remove client
// @Description Remove client by ID
// @Tags clients
// @Security BearerAuth
// @Param id path string true "Client ID"
// @Success 204 "No Content"
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 409 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /clients/{id} [delete]
func (app *Application) ClientRemove(w http.ResponseWriter, r *http.Request) {
	clientID, err := readID(r)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	err = app.services.Client.RemoveByID(r.Context(), clientID)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ClientSearch godoc
//
// @Summary Search clients
// @Description Search clients by name and surname
// @Tags clients
// @Security BearerAuth
// @Produce json
// @Param name query string true "Client name"
// @Param surname query string true "Client surname"
// @Success 200 {array} dto.ClientResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /clients/search [get]
func (app *Application) ClientSearch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	surname := r.URL.Query().Get("surname")

	clients, err := app.services.Client.GetByNameSurname(r.Context(), name, surname)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}
	_ = writeJSON(w, http.StatusOK, clients)

}

// ClientList godoc
//
// @Summary Get clients
// @Description Get all clients with optional pagination
// @Tags clients
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Maximum number of clients"
// @Param offset query int false "Number of clients to skip"
// @Success 200 {array} dto.ClientResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /clients [get]
func (app *Application) ClientList(w http.ResponseWriter, r *http.Request) {
	limit, err := readIntQuery(r, "limit")
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	offset, err := readIntQuery(r, "offset")
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	clients, err := app.services.Client.GetAllClients(r.Context(), limit, offset)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, clients)

}

// ClientPatchAddress godoc
//
// @Summary Change client address
// @Description Update address of an existing client
// @Tags clients
// @Security BearerAuth
// @Accept json
// @Param id path string true "Client ID"
// @Param address body dto.AddressRequest true "New address"
// @Success 204 "No Content"
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /clients/{id}/address [patch]
func (app *Application) ClientPatchAddress(w http.ResponseWriter, r *http.Request) {
	clientID, err := readID(r)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	var reqNewAddress dto.AddressRequest
	err = readJSON(r, &reqNewAddress)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	err = app.services.Client.PatchClientAddress(r.Context(), clientID, reqNewAddress)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
