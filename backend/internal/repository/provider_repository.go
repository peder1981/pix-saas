package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pixsaas/backend/internal/domain"
	"gorm.io/gorm"
)

// ProviderRepository gerencia operações de providers
type ProviderRepository struct {
	db *gorm.DB
}

// NewProviderRepository cria um novo repositório de providers
func NewProviderRepository(db *gorm.DB) *ProviderRepository {
	return &ProviderRepository{db: db}
}

// Create cria um novo provider
func (r *ProviderRepository) Create(ctx context.Context, provider *domain.Provider) error {
	return r.db.WithContext(ctx).Create(provider).Error
}

// GetByID busca um provider por ID
func (r *ProviderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Provider, error) {
	var provider domain.Provider
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

// GetByCode busca um provider por código
func (r *ProviderRepository) GetByCode(ctx context.Context, code string) (*domain.Provider, error) {
	var provider domain.Provider
	err := r.db.WithContext(ctx).Where("code = ? AND deleted_at IS NULL", code).First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

// GetByISPB busca um provider por ISPB
func (r *ProviderRepository) GetByISPB(ctx context.Context, ispb string) (*domain.Provider, error) {
	var provider domain.Provider
	err := r.db.WithContext(ctx).Where("ispb = ? AND deleted_at IS NULL", ispb).First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

// Update atualiza um provider
func (r *ProviderRepository) Update(ctx context.Context, provider *domain.Provider) error {
	return r.db.WithContext(ctx).Save(provider).Error
}

// Delete deleta um provider (soft delete)
func (r *ProviderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Provider{}).Where("id = ?", id).Update("deleted_at", now).Error
}

// List lista todos os providers
func (r *ProviderRepository) List(ctx context.Context, activeOnly bool) ([]domain.Provider, error) {
	var providers []domain.Provider
	query := r.db.WithContext(ctx).Where("deleted_at IS NULL")

	if activeOnly {
		query = query.Where("active = true")
	}

	err := query.Order("priority DESC, name ASC").Find(&providers).Error
	return providers, err
}

// UpdateHealthStatus atualiza o status de saúde de um provider
func (r *ProviderRepository) UpdateHealthStatus(ctx context.Context, id uuid.UUID, status string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Provider{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"health_status":  status,
			"last_health_at": now,
			"updated_at":     now,
		}).Error
}

// GetHealthyProviders retorna providers saudáveis ordenados por prioridade
func (r *ProviderRepository) GetHealthyProviders(ctx context.Context) ([]domain.Provider, error) {
	var providers []domain.Provider
	err := r.db.WithContext(ctx).
		Where("active = true AND deleted_at IS NULL AND health_status = ?", "healthy").
		Order("priority DESC").
		Find(&providers).Error
	return providers, err
}

// MerchantProviderRepository gerencia configurações de merchant-provider
type MerchantProviderRepository struct {
	db *gorm.DB
}

// NewMerchantProviderRepository cria um novo repositório
func NewMerchantProviderRepository(db *gorm.DB) *MerchantProviderRepository {
	return &MerchantProviderRepository{db: db}
}

// Create cria uma nova configuração merchant-provider
func (r *MerchantProviderRepository) Create(ctx context.Context, mp *domain.MerchantProvider) error {
	return r.db.WithContext(ctx).Create(mp).Error
}

// GetByID busca uma configuração por ID
func (r *MerchantProviderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.MerchantProvider, error) {
	var mp domain.MerchantProvider
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("Provider").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&mp).Error
	if err != nil {
		return nil, err
	}
	return &mp, nil
}

// GetByMerchantAndProvider busca configuração específica
func (r *MerchantProviderRepository) GetByMerchantAndProvider(ctx context.Context, merchantID, providerID uuid.UUID) (*domain.MerchantProvider, error) {
	var mp domain.MerchantProvider
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("Provider").
		Where("merchant_id = ? AND provider_id = ? AND deleted_at IS NULL", merchantID, providerID).
		First(&mp).Error
	if err != nil {
		return nil, err
	}
	return &mp, nil
}

// ListByMerchant lista todas as configurações de um merchant
func (r *MerchantProviderRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID, activeOnly bool) ([]domain.MerchantProvider, error) {
	var mps []domain.MerchantProvider
	query := r.db.WithContext(ctx).
		Preload("Provider").
		Where("merchant_id = ? AND deleted_at IS NULL", merchantID)

	if activeOnly {
		query = query.Where("active = true")
	}

	err := query.Find(&mps).Error
	return mps, err
}

// Update atualiza uma configuração
func (r *MerchantProviderRepository) Update(ctx context.Context, mp *domain.MerchantProvider) error {
	return r.db.WithContext(ctx).Save(mp).Error
}

// Delete deleta uma configuração (soft delete)
func (r *MerchantProviderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.MerchantProvider{}).Where("id = ?", id).Update("deleted_at", now).Error
}

// UpdateTokenInfo atualiza informações de token
func (r *MerchantProviderRepository) UpdateTokenInfo(ctx context.Context, id uuid.UUID, expiresAt time.Time) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.MerchantProvider{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_token_refresh": now,
			"token_expires_at":   expiresAt,
			"updated_at":         now,
		}).Error
}
