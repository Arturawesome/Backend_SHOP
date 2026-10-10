package web

import (
	"BE2_API_Shop/shop/internal/dto"
	"net/http"
)

// SupplierAdd godoc
//
// @Summary Add supplier
// @Description Create a new supplier
// @Tags suppliers
// @Accept json
// @Security BearerAuth
// @Param supplier body dto.SupplierRequest true "Supplier data"
// @Success 201 "Created"
// @Failure 400 {object} errorResponse
// @Failure 409 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /suppliers [post]
func (app *Application) SupplierAdd(w http.ResponseWriter, r *http.Request) {
	var req dto.SupplierRequest

	err := readJSON(r, &req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	err = app.services.Supplier.AddSupplier(r.Context(), req)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// SupplierPatchAddress godoc
//
// @Summary Change supplier address
// @Description Update address of an existing supplier
// @Tags suppliers
// @Accept json
// @Security BearerAuth
// @Param id path string true "Supplier ID"
// @Param address body dto.AddressRequest true "New address"
// @Success 204 "No Content"
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /suppliers/{id}/address [patch]
func (app *Application) SupplierPatchAddress(w http.ResponseWriter, r *http.Request) {
	supplierID, err := readID(r)
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

	err = app.services.Supplier.PatchAddress(
		r.Context(),
		supplierID,
		reqNewAddress,
	)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SupplierRemoveByID godoc
//
// @Summary Remove supplier
// @Description Delete a supplier by ID
// @Tags suppliers
// @Security BearerAuth
// @Param id path string true "Supplier ID"
// @Success 204 "No Content"
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 409 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /suppliers/{id} [delete]
func (app *Application) SupplierRemoveByID(w http.ResponseWriter, r *http.Request) {
	supplierID, err := readID(r)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	err = app.services.Supplier.RemoveByID(r.Context(), supplierID)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SupplierList godoc
//
// @Summary Get suppliers
// @Description Get list of all suppliers
// @Tags suppliers
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.SupplierResponse
// @Failure 500 {object} errorResponse
// @Router /suppliers [get]
func (app *Application) SupplierList(w http.ResponseWriter, r *http.Request) {
	suppliers, err := app.services.Supplier.GetAllSuppliers(r.Context())
	if err != nil {
		app.handleServiceError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, suppliers)
}

// SupplierGetByID godoc
//
// @Summary Get supplier
// @Description Get a supplier by ID
// @Tags suppliers
// @Security BearerAuth
// @Produce json
// @Param id path string true "Supplier ID"
// @Success 200 {object} dto.SupplierResponse
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /suppliers/{id} [get]
func (app *Application) SupplierGetByID(w http.ResponseWriter, r *http.Request) {
	supplierID, err := readID(r)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	supplier, err := app.services.Supplier.GetByID(r.Context(), supplierID)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, supplier)
}
