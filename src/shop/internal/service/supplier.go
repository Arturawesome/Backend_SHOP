package service

import (
	"BE2_API_Shop/shop/internal/dto"
	"BE2_API_Shop/shop/internal/repository"

	"context"
	"errors"

	"github.com/google/uuid"
)

type SupplierService struct {
	repo repository.SupplierRepository
}

func NewSupplierService(repo repository.SupplierRepository) *SupplierService {
	return &SupplierService{repo: repo}
}

func (s *SupplierService) AddSupplier(ctx context.Context, reqSupplier dto.SupplierRequest) error {
	if reqSupplier.Name == "" ||
		reqSupplier.PhoneNumber == "" {
		return ErrIncorrectInput
	}
	if reqSupplier.Address.Country == "" ||
		reqSupplier.Address.City == "" ||
		reqSupplier.Address.Street == "" {
		return ErrInvalidAddressData
	}
	address := fromAddressDTO(reqSupplier.Address)
	supplier := fromSupplierDTO(reqSupplier, address.ID)
	err := s.repo.AddSupplier(ctx, supplier, address)
	return mapRepositoryError(err)
}
func (s *SupplierService) PatchAddress(ctx context.Context, ID uuid.UUID, reqNewAddress dto.AddressRequest) error {
	if ID == uuid.Nil {
		return ErrInvalidID
	}
	if reqNewAddress.City == "" ||
		reqNewAddress.Country == "" ||
		reqNewAddress.Street == "" {
		return ErrInvalidAddressData
	}
	newAddress := fromAddressDTO(reqNewAddress)
	err := s.repo.PatchAddress(ctx, ID, newAddress)
	return mapRepositoryError(err)
}
func (s *SupplierService) RemoveByID(ctx context.Context, ID uuid.UUID) error {
	if ID == uuid.Nil {
		return ErrInvalidID
	}
	err := s.repo.RemoveByID(ctx, ID)
	return mapRepositoryError(err)
}
func (s *SupplierService) GetAllSuppliers(ctx context.Context) ([]dto.SupplierResponse, error) {
	suppliers, err := s.repo.GetAllSuppliers(ctx)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return toSupplierDTOs(suppliers), nil
}
func (s *SupplierService) GetByID(ctx context.Context, ID uuid.UUID) (dto.SupplierResponse, error) {
	if ID == uuid.Nil {
		return dto.SupplierResponse{}, ErrInvalidID
	}
	supplier, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.SupplierResponse{}, ErrNotFound
		}

		return dto.SupplierResponse{}, mapRepositoryError(err)
	}
	return toSupplierDTO(supplier), nil
}
