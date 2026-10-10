package web

import (
	"BE2_API_Shop/shop/internal/dto"
	"net/http"
)

// ProductAdd godoc
//
// @Summary Add product
// @Description Create a new product
// @Tags products
// @Accept json
// @Security BearerAuth
// @Param product body dto.ProductRequest true "Product data"
// @Success 201 "Created"
// @Failure 400 {object} errorResponse
// @Failure 409 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /products [post]
func (app *Application) ProductAdd(w http.ResponseWriter, r *http.Request) {
	var req dto.ProductRequest

	err := readJSON(r, &req)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	err = app.services.Product.AddProduct(r.Context(), req)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// ProductDecrease godoc
//
// @Summary Decrease product stock
// @Description Decrease the available stock of a product by the specified amount
// @Tags products
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param num query int true "Amount to decrease"
// @Success 204 "No Content"
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /products/{id}/stock [patch]
func (app *Application) ProductDecrease(w http.ResponseWriter, r *http.Request) {
	productID, err := readID(r)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	num, err := readIntQuery(r, "num")
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if num == nil {
		errorJSON(w, http.StatusBadRequest, "num is required")
		return
	}

	err = app.services.Product.DecreaseProduct(
		r.Context(),
		productID,
		*num,
	)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ProductGetByID godoc
//
// @Summary Get product
// @Description Get a product by ID
// @Tags products
// @Security BearerAuth
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} dto.ProductResponse
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /products/{id} [get]
func (app *Application) ProductGetByID(w http.ResponseWriter, r *http.Request) {
	productID, err := readID(r)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	product, err := app.services.Product.GetByID(r.Context(), productID)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, product)
}

// ProductGetAvailable godoc
//
// @Summary Get available products
// @Description Get all products with available stock
// @Tags products
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.ProductResponse
// @Failure 500 {object} errorResponse
// @Router /products [get]
func (app *Application) ProductGetAvailable(w http.ResponseWriter, r *http.Request) {
	products, err := app.services.Product.GetAvailableProducts(r.Context())
	if err != nil {
		app.handleServiceError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, products)
}

// ProductRemoveByID godoc
//
// @Summary Remove product
// @Description Delete a product by ID
// @Tags products
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 204 "No Content"
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /products/{id} [delete]
func (app *Application) ProductRemoveByID(w http.ResponseWriter, r *http.Request) {
	productID, err := readID(r)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	err = app.services.Product.RemoveByID(r.Context(), productID)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
