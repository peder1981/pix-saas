package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pixsaas/backend/internal/domain"
	"gorm.io/gorm"
)

// MerchantRepository gerencia operações de merchants
type MerchantRepository struct {
	db *gorm.DB
}

// NewMerchantRepository cria um novo repositório de merchants
func NewMerchantRepository(db *gorm.DB) *MerchantRepository {
	return &MerchantRepository{db: db}
}

// Create cria um novo merchant
func (r *MerchantRepository) Create(ctx context.Context, merchant *domain.Merchant) error {
	return r.db.WithContext(ctx).Create(merchant).Error
}

// GetByID busca um merchant por ID
func (r *MerchantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Merchant, error) {
	var merchant domain.Merchant
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&merchant).Error
	if err != nil {
		return nil, err
	}
	return &merchant, nil
}

// GetByDocument busca um merchant por documento (CPF/CNPJ)
func (r *MerchantRepository) GetByDocument(ctx context.Context, document string) (*domain.Merchant, error) {
	var merchant domain.Merchant
	err := r.db.WithContext(ctx).Where("document = ? AND deleted_at IS NULL", document).First(&merchant).Error
	if err != nil {
		return nil, err
	}
	return &merchant, nil
}

// GetByEmail busca um merchant por email
func (r *MerchantRepository) GetByEmail(ctx context.Context, email string) (*domain.Merchant, error) {
	var merchant domain.Merchant
	err := r.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&merchant).Error
	if err != nil {
		return nil, err
	}
	return &merchant, nil
}

// GetByAPIKey busca um merchant por API key
func (r *MerchantRepository) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Merchant, error) {
	var merchant domain.Merchant
	err := r.db.WithContext(ctx).Where("api_key = ? AND active = true AND deleted_at IS NULL", apiKey).First(&merchant).Error
	if err != nil {
		return nil, err
	}
	return &merchant, nil
}

// Update atualiza um merchant
func (r *MerchantRepository) Update(ctx context.Context, merchant *domain.Merchant) error {
	return r.db.WithContext(ctx).Save(merchant).Error
}

// Delete deleta um merchant (soft delete)
func (r *MerchantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Merchant{}).Where("id = ?", id).Update("deleted_at", now).Error
}

// List lista merchants com paginação
func (r *MerchantRepository) List(ctx context.Context, limit, offset int) ([]domain.Merchant, int64, error) {
	var merchants []domain.Merchant
	var total int64
	
	err := r.db.WithContext(ctx).Model(&domain.Merchant{}).Where("deleted_at IS NULL").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	err = r.db.WithContext(ctx).Where("deleted_at IS NULL").Limit(limit).Offset(offset).Find(&merchants).Error
	return merchants, total, err
}

// SetActive ativa/desativa um merchant
func (r *MerchantRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	return r.db.WithContext(ctx).Model(&domain.Merchant{}).Where("id = ?", id).Update("active", active).Error
}
