package services

import (
	"context"
	"fmt"
	"strings"
	"transfers-api/internal/config"
	"transfers-api/internal/enums"
	"transfers-api/internal/known_errors"
	"transfers-api/internal/models"
)

//go:generate mockery --name TransfersRepository --structname TransfersRepositoryMock --filename transfers_repository_mock.go --output mocks --outpkg mocks

type TransfersRepository interface {
	Create(ctx context.Context, transfer models.Transfer) (string, error)
	GetByID(ctx context.Context, id string) (models.Transfer, error)
	Update(ctx context.Context, transfer models.Transfer) error
	Delete(ctx context.Context, id string) error
	GetByUserId(ctx context.Context, userId string) (models.Transfer, error)
}

type TransfersPublisher interface {
	Publish(operation string, transferID string) error
}

type TransfersService struct {
	businessCfg        config.BusinessConfig
	transfersRepo      TransfersRepository
	transfersCache     TransfersRepository
	transfersPublisher TransfersPublisher
}

func NewTransfersService(businessCfg config.BusinessConfig, transfersRepo TransfersRepository, transfersCache TransfersRepository, transfersPublisher TransfersPublisher) *TransfersService {
	return &TransfersService{
		businessCfg:        businessCfg,
		transfersRepo:      transfersRepo,
		transfersCache:     transfersCache,
		transfersPublisher: transfersPublisher,
	}
}

func (s *TransfersService) Create(ctx context.Context, transfer models.Transfer) (string, error) {
	if strings.TrimSpace(transfer.SenderID) == "" {
		return "", fmt.Errorf("sender_id is required: %w", known_errors.ErrBadRequest)
	}
	if strings.TrimSpace(transfer.ReceiverID) == "" {
		return "", fmt.Errorf("sender_id is required: %w", known_errors.ErrBadRequest)
	}
	if transfer.Currency == enums.CurrencyUnknown {
		return "", fmt.Errorf("invalid currency %s: %w", transfer.Currency.String(), known_errors.ErrBadRequest)
	}
	if transfer.Amount <= 0 {
		return "", fmt.Errorf("amount should be greater than 0: %w", known_errors.ErrBadRequest)
	}
	if strings.TrimSpace(transfer.State) == "" { // TODO: replace with enums.ParseState
		return "", fmt.Errorf("state is required: %w", known_errors.ErrBadRequest)
	}
	id, err := s.transfersRepo.Create(ctx, transfer)
	if err != nil {
		return "", fmt.Errorf("error creating transfer in repository: %w", err)
	}

	if _, err := s.transfersCache.Create(ctx, transfer); err != nil {
		return "", fmt.Errorf("error caching transfer with ID %s: %v", id, err)
	}

	return id, nil
}

func (s *TransfersService) GetByID(ctx context.Context, id string) (models.Transfer, error) {
	transfer, err := s.transfersRepo.GetByID(ctx, id)
	if err != nil {
		return models.Transfer{}, fmt.Errorf("error getting transfer %s from repository: %w", id, err)
	}
	return transfer, nil
}

func (s *TransfersService) Update(ctx context.Context, transfer models.Transfer) error {
	if strings.TrimSpace(transfer.ID) == "" {
		return fmt.Errorf("ID is required: %w", known_errors.ErrBadRequest)
	}
	if strings.TrimSpace(transfer.SenderID) == "" &&
		strings.TrimSpace(transfer.ReceiverID) == "" &&
		transfer.Currency == enums.CurrencyUnknown &&
		transfer.Amount <= 0 &&
		strings.TrimSpace(transfer.State) == "" {
		return fmt.Errorf("error updating transfer %s: no fields to update: %w", transfer.ID, known_errors.ErrBadRequest)
	}
	if err := s.transfersRepo.Update(ctx, transfer); err != nil {
		return fmt.Errorf("error updating transfer %s in repository: %w", transfer.ID, err)
	}
	return nil
}

func (s *TransfersService) Delete(ctx context.Context, id string) error {
	if err := s.transfersRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("error deleting transfer %s from repository: %w", id, err)
	}
	return nil
}

func (s *TransfersService) GetByUserId(ctx context.Context, userId string) (models.Transfer, error) {
	transfer, err := s.transfersRepo.GetByUserId(ctx, userId)
	if err != nil {
		return models.Transfer{}, fmt.Errorf("error getting transfer for user %s from repository: %w", userId, err)
	}
	return transfer, nil
}
