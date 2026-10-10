package service

import (
	"BE2_API_Shop/shop/internal/dto"
	"BE2_API_Shop/shop/internal/repository"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type ImageService struct {
	repo  repository.ImageRepository
	cashe repository.ImageCashe
}

func NewImageService(repo repository.ImageRepository, cashe repository.ImageCashe) *ImageService {
	return &ImageService{repo: repo,
		cashe: cashe}
}

func (s *ImageService) AddImage(ctx context.Context, reqImage dto.ImageRequest, productID uuid.UUID) error {
	if productID == uuid.Nil {
		return ErrInvalidID
	}
	if len(reqImage.Image) == 0 {
		return ErrIncorrectInput
	}

	image := fromImageDTO(reqImage)
	err := s.repo.AddImage(ctx, image, productID)
	return mapRepositoryError(err)
}

func (s *ImageService) ChangeImage(ctx context.Context, imageID uuid.UUID, newImage dto.ImageRequest) error {
	if imageID == uuid.Nil {
		return ErrInvalidID
	}
	if len(newImage.Image) == 0 {
		return ErrIncorrectInput
	}

	if err := s.repo.ChangeImage(ctx, imageID, newImage.Image); err != nil {
		return mapRepositoryError(err)
	}
	if err := s.cashe.Delete(ctx, "image:"+imageID.String()); err != nil {
		fmt.Println("CACHE SET ERROR:", err)
	}
	return nil

}

func (s *ImageService) RemoveByID(ctx context.Context, ID uuid.UUID) error {
	if ID == uuid.Nil {
		return ErrInvalidID
	}

	if err := s.repo.RemoveByID(ctx, ID); err != nil {
		return mapRepositoryError(err)
	}

	if err := s.cashe.Delete(ctx, "image:"+ID.String()); err != nil {
		log.Println(err)
	}
	return nil
}

func (s *ImageService) GetByProductID(ctx context.Context, productID uuid.UUID) (dto.ImageResponse, error) {
	if productID == uuid.Nil {
		return dto.ImageResponse{}, ErrInvalidID
	}
	image, err := s.repo.GetByProductID(ctx, productID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.ImageResponse{}, ErrNotFound
		}

		return dto.ImageResponse{}, mapRepositoryError(err)
	}
	imagDTO := toImageDTO(image)

	return imagDTO, nil
}

func (s *ImageService) GetByID(ctx context.Context, ID uuid.UUID) (dto.ImageResponse, error) {
	if ID == uuid.Nil {
		return dto.ImageResponse{}, ErrInvalidID
	}
	cashKey := "image:" + ID.String()
	data, err := s.cashe.GetByID(ctx, cashKey)

	if err == nil {
		fmt.Println("IMage from cashe")
		return dto.ImageResponse{Image: data}, nil
	}

	if errors.Is(err, redis.Nil) {
		fmt.Println("Cache miss")
	} else {
		fmt.Println("CACHE ERROR:", err)
	}

	fmt.Println("Image from postgres")
	image, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.ImageResponse{}, ErrNotFound
		}
		return dto.ImageResponse{}, mapRepositoryError(err)
	}

	imageDTO := toImageDTO(image)

	if err := s.cashe.Set(ctx, cashKey, image.Image, 0); err != nil {
		fmt.Println("CACHE SET ERROR:", err)
	}
	return imageDTO, nil

}
