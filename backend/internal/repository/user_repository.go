package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pixsaas/backend/internal/domain"
	"gorm.io/gorm"
)

// UserRepository gerencia operações de usuários
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository cria um novo repositório de usuários
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create cria um novo usuário
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID busca um usuário por ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail busca um usuário por email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update atualiza um usuário
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete deleta um usuário (soft delete)
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", id).Update("deleted_at", now).Error
}

// ListByMerchant lista usuários de um merchant
func (r *UserRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]domain.User, error) {
	var users []domain.User
	err := r.db.WithContext(ctx).Where("merchant_id = ? AND deleted_at IS NULL", merchantID).Find(&users).Error
	return users, err
}

// SetActive ativa/desativa um usuário
func (r *UserRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", id).Update("active", active).Error
}
