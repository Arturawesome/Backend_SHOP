package service

import (
	"BE2_API_Shop/shop/internal/dto"
	"BE2_API_Shop/shop/internal/repository"
	"context"

	"github.com/google/uuid"
)

type ClientService struct {
	repo repository.ClientRepository
}

func NewClientService(repo repository.ClientRepository) *ClientService {
	return &ClientService{repo: repo}
}

func (s *ClientService) AddClient(ctx context.Context, reqClient dto.ClientRequest) error {
	if reqClient.Name == "" || reqClient.Surname == "" ||
		reqClient.Gender == "" || reqClient.Birthday.IsZero() {
		return ErrInvalidClientData
	}
	if reqClient.Address.Country == "" ||
		reqClient.Address.City == "" ||
		reqClient.Address.Street == "" {
		return ErrInvalidAddressData
	}

	modelAddress := fromAddressDTO(reqClient.Address)
	modelClient := fromClientDTO(reqClient, modelAddress.ID)
	err := s.repo.AddClient(ctx, modelClient, modelAddress)
	return mapRepositoryError(err)
}
func (s *ClientService) RemoveByID(ctx context.Context, ID uuid.UUID) error {
	if ID == uuid.Nil {
		return ErrInvalidID
	}
	err := s.repo.RemoveByID(ctx, ID)
	return mapRepositoryError(err)
}

func (s *ClientService) GetByNameSurname(ctx context.Context, name string,
	surname string) ([]dto.ClientResponse, error) {
	if name == "" || surname == "" {
		return nil, ErrInvalidClientData
	}

	listModelCLient, err := s.repo.GetByNameSurname(ctx, name, surname)
	if err != nil {
		return nil, mapRepositoryError(err)
	}

	return toClientDTOs(listModelCLient), nil
}

func (s *ClientService) GetAllClients(ctx context.Context, limit *int,
	offset *int) ([]dto.ClientResponse, error) {

	if limit != nil && *limit < 0 {
		return nil, ErrIncorrectInput
	}
	if offset != nil && *offset < 0 {
		return nil, ErrIncorrectInput
	}

	listModelCLient, err := s.repo.GetAllClients(ctx, limit, offset)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return toClientDTOs(listModelCLient), nil
}

func (s *ClientService) PatchClientAddress(ctx context.Context, clientID uuid.UUID,
	reqNewAddress dto.AddressRequest) error {
	if clientID == uuid.Nil {
		return ErrInvalidID
	}
	if reqNewAddress.City == "" || reqNewAddress.Country == "" || reqNewAddress.Street == "" {
		return ErrIncorrectInput
	}

	newAddress := fromAddressDTO(reqNewAddress)

	err := s.repo.PatchClientAddress(ctx, clientID, newAddress)

	return mapRepositoryError(err)
}
