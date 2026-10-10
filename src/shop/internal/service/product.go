package service

import (
	"BE2_API_Shop/shop/internal/dto"
	"BE2_API_Shop/shop/internal/model"
	"BE2_API_Shop/shop/internal/repository"
	"context"
	"errors"

	"github.com/google/uuid"
)

type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) AddProduct(ctx context.Context, reqProduct dto.ProductRequest) error {
	if reqProduct.Name == "" ||
		reqProduct.Category == "" ||
		reqProduct.Price <= 0 ||
		reqProduct.AvailableStock < 0 ||
		reqProduct.SupplierID == uuid.Nil {
		return ErrIncorrectInput
	}
	product := fromProductDTO(reqProduct)
	err := s.repo.AddProduct(ctx, product)

	return mapRepositoryError(err)
}
func (s *ProductService) DecreaseProduct(ctx context.Context, ID uuid.UUID, num int) error {
	if ID == uuid.Nil || num <= 0 {
		return ErrIncorrectInput
	}
	err := s.repo.DecreaseProduct(ctx, ID, num)
	return mapRepositoryError(err)
}
func (s *ProductService) GetByID(ctx context.Context, ID uuid.UUID) (dto.ProductResponse, error) {
	if ID == uuid.Nil {
		return dto.ProductResponse{}, ErrInvalidID
	}
	product, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.ProductResponse{}, ErrNotFound
		}

		return dto.ProductResponse{}, mapRepositoryError(err)
	}

	return toProductDTO(product), nil
}
func (s *ProductService) GetAvailableProducts(ctx context.Context) ([]dto.ProductResponse, error) {
	products, err := s.repo.GetAvailableProducts(ctx)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	return toProductDTOs(products), nil
}
func (s *ProductService) RemoveByID(ctx context.Context, ID uuid.UUID) error {
	if ID == uuid.Nil {
		return ErrInvalidID
	}
	err := s.repo.RemoveByID(ctx, ID)
	return mapRepositoryError(err)
}

func (s *ProductService) UpdateProduct(ctx context.Context, event *model.ProductUpdateEvent) error {
	if event == nil ||
		event.ID == uuid.Nil ||
		event.Price < 0 ||
		event.AvailableStock < 0 {
		return ErrIncorrectInput
	}
	return s.repo.UpdateProduct(ctx, event)
}
