package service

import (
	"BE2_API_Shop/shop/internal/repository"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type Services struct {
	Client   *ClientService
	Product  *ProductService
	Supplier *SupplierService
	Image    *ImageService
}

func NewServices(
	clientRepo repository.ClientRepository,
	productRepo repository.ProductRepository,
	supplierRepo repository.SupplierRepository,
	imageRepo repository.ImageRepository,
	imageCashe repository.ImageCashe,
) *Services {
	return &Services{
		Client:   NewClientService(clientRepo),
		Product:  NewProductService(productRepo),
		Supplier: NewSupplierService(supplierRepo),
		Image:    NewImageService(imageRepo, imageCashe),
	}
}

var (
	ErrInvalidClientData  = errors.New("invalid client data")
	ErrInvalidID          = errors.New("invalid id")
	ErrNotFound           = errors.New("not found")
	ErrIncorrectInput     = errors.New("incorrect input")
	ErrInvalidAddressData = errors.New("invalid address data")
	ErrConflict           = errors.New("conflict")
)

func mapRepositoryError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			return ErrIncorrectInput

		case "23505":
			return ErrConflict

		case "23502":
			return ErrIncorrectInput

		case "23514":
			return ErrIncorrectInput

		case "22001":
			return ErrIncorrectInput
		}
	}

	return err
}
