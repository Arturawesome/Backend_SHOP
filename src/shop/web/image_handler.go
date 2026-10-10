package web

import (
	"BE2_API_Shop/shop/internal/dto"
	"net/http"
)

// ImageAdd godoc
//
// @Summary Add product image
// @Description Add an image to a product by product ID
// @Tags images
// @Security BearerAuth
// @Accept json
// @Param id path string true "Product ID"
// @Param image body dto.ImageRequest true "Image data"
// @Success 201 "Created"
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /products/{id}/image [post]
func (app *Application) ImageAdd(w http.ResponseWriter, r *http.Request) {
	productID, err := readID(r)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	var reqImage dto.ImageRequest
	err = readJSON(r, &reqImage)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	err = app.services.Image.AddImage(r.Context(), reqImage, productID)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)

}

// ImageChange godoc
//
// @Summary Change image
// @Description Replace an existing image by image ID
// @Tags images
// @Security BearerAuth
// @Accept json
// @Param id path string true "Image ID"
// @Param image body dto.ImageRequest true "New image data"
// @Success 204 "No Content"
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /images/{id} [put]
func (app *Application) ImageChange(w http.ResponseWriter, r *http.Request) {
	imageID, err := readID(r)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	var reqImage dto.ImageRequest
	err = readJSON(r, &reqImage)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	err = app.services.Image.ChangeImage(r.Context(), imageID, reqImage)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ImageRemove godoc
//
// @Summary Remove image
// @Description Delete an image by image ID
// @Tags images
// @Security BearerAuth
// @Param id path string true "Image ID"
// @Success 204 "No Content"
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /images/{id} [delete]
func (app *Application) ImageRemove(w http.ResponseWriter, r *http.Request) {
	imageID, err := readID(r)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	err = app.services.Image.RemoveByID(r.Context(), imageID)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}

// ImageGetByProductID godoc
//
// @Summary Get product image
// @Description Get the image associated with a product by product ID
// @Tags images
// @Security BearerAuth
// @Produce Application/octet-stream
// @Param id path string true "Product ID"
// @Success 200 {file} binary
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /products/{id}/image [get]
func (app *Application) ImageGetByProductID(w http.ResponseWriter, r *http.Request) {
	productID, err := readID(r)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	image, err := app.services.Image.GetByProductID(r.Context(), productID)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}
	//writeJSON(w, http.StatusOK, image)
	err = writeBinary(w, http.StatusOK, image.Image, "image.bin")
	if err != nil {
		return
	}
}

// ImageGet godoc
//
// @Summary Get image
// @Description Get an image by image ID
// @Tags images
// @Security BearerAuth
// @Produce Application/octet-stream
// @Param id path string true "Image ID"
// @Success 200 {file} binary
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /images/{id} [get]

func (app *Application) ImageGet(w http.ResponseWriter, r *http.Request) {
	imageID, err := readID(r)
	if err != nil {
		errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	image, err := app.services.Image.GetByID(r.Context(), imageID)
	if err != nil {
		app.handleServiceError(w, err)
		return
	}
	//writeJSON(w, http.StatusOK, image)
	err = writeBinary(w, http.StatusOK, image.Image, "image.bin")
	if err != nil {
		return
	}
}
