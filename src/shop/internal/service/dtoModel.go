package service

import (
	"BE2_API_Shop/shop/internal/dto"
	"BE2_API_Shop/shop/internal/model"
	"time"

	"github.com/google/uuid"
)

func fromClientDTO(req dto.ClientRequest, addressID uuid.UUID) model.Client {
	return model.Client{
		ID:               uuid.New(),
		Name:             req.Name,
		Surname:          req.Surname,
		Birthday:         req.Birthday,
		Gender:           req.Gender,
		RegistrationDate: time.Now(),
		AddressID:        addressID,
	}
}

func toClientDTO(client model.Client) dto.ClientResponse {
	return dto.ClientResponse{
		ID:               client.ID,
		Name:             client.Name,
		Surname:          client.Surname,
		Birthday:         client.Birthday,
		Gender:           client.Gender,
		RegistrationDate: client.RegistrationDate,
	}

}

func fromAddressDTO(req dto.AddressRequest) model.Address {
	return model.Address{
		ID:      uuid.New(),
		Country: req.Country,
		City:    req.City,
		Street:  req.Street,
	}
}
func toClientDTOs(clients []model.Client) []dto.ClientResponse {
	result := make([]dto.ClientResponse, 0, len(clients))

	for _, client := range clients {
		result = append(result, toClientDTO(client))
	}

	return result
}

func fromImageDTO(req dto.ImageRequest) model.Image {
	return model.Image{ID: uuid.New(),
		Image: req.Image}
}
func toImageDTO(image model.Image) dto.ImageResponse {
	return dto.ImageResponse{
		Image: image.Image}
}

func fromProductDTO(req dto.ProductRequest) model.Product {
	return model.Product{
		ID:             uuid.New(),
		Name:           req.Name,
		Category:       req.Category,
		Price:          req.Price,
		AvailableStock: req.AvailableStock,
		LastUpdateDate: time.Now(),
		SupplierID:     req.SupplierID,
		ImageID:        nil,
	}
}

func toProductDTO(product model.Product) dto.ProductResponse {
	return dto.ProductResponse{
		ID:             product.ID,
		Name:           product.Name,
		Category:       product.Category,
		Price:          product.Price,
		AvailableStock: product.AvailableStock,
		LastUpdateDate: product.LastUpdateDate,
		SupplierID:     product.SupplierID,
		ImageID:        product.ImageID,
	}
}
func toProductDTOs(products []model.Product) []dto.ProductResponse {
	result := make([]dto.ProductResponse, 0, len(products))

	for _, p := range products {
		result = append(result, toProductDTO(p))
	}

	return result
}

func toSupplierDTO(supplier model.Supplier) dto.SupplierResponse {
	return dto.SupplierResponse{
		ID:          supplier.ID,
		Name:        supplier.Name,
		AddressID:   supplier.AddressID,
		PhoneNumber: supplier.PhoneNumber,
	}
}
func toSupplierDTOs(suppliers []model.Supplier) []dto.SupplierResponse {
	result := make([]dto.SupplierResponse, 0, len(suppliers))

	for _, s := range suppliers {
		result = append(result, toSupplierDTO(s))
	}
	return result

}
func fromSupplierDTO(req dto.SupplierRequest, addressID uuid.UUID) model.Supplier {
	return model.Supplier{
		ID:          uuid.New(),
		Name:        req.Name,
		PhoneNumber: req.PhoneNumber,
		AddressID:   addressID,
	}
}
