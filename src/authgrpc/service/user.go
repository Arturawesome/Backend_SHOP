package service

import (
	"BE2_API_Shop/authgrpc/model"
	"BE2_API_Shop/authgrpc/repository"
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type CreateUserInput struct {
	Email    string
	Name     string
	Surname  string
	Phone    string
	Password string
}

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context,
	in CreateUserInput) (model.User, error) {
	if in.Email == "" ||
		in.Name == "" ||
		in.Surname == "" ||
		in.Phone == "" ||
		in.Password == "" {

		return model.User{}, ErrInvalidData
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}
	user := model.User{
		ID:           uuid.New(),
		Email:        in.Email,
		Name:         in.Name,
		Surname:      in.Surname,
		Phone:        in.Phone,
		PasswordHash: string(hashedPassword),
	}

	return user, s.repo.Create(ctx, user)
}

func (s *UserService) GetByEmail(ctx context.Context,
	email string) (model.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *UserService) GetByID(ctx context.Context,
	id string) (model.User, error) {

	return s.repo.GetByID(ctx, id)
}

func (s *UserService) ChangePassword(ctx context.Context,
	userID string, password string) error {
	if userID == "" || password == "" {
		return ErrInvalidData
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	err = s.repo.ChangePassword(ctx, userID, string(hashedPassword))
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) Authenticate(
	ctx context.Context,
	email string,
	password string,
) (model.User, error) {

	if email == "" || password == "" {
		return model.User{}, ErrInvalidCredentials
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.User{}, ErrInvalidCredentials
		}

		return model.User{}, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)

	if err != nil {
		return model.User{}, ErrInvalidCredentials
	}

	return user, nil
}
