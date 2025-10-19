package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pixsaas/backend/internal/domain"
	"gorm.io/gorm"
)

// TransactionRepository gerencia operações de transações
type TransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository cria um novo repositório de transações
func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create cria uma nova transação
func (r *TransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

// GetByID busca uma transação por ID
func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	var tx domain.Transaction
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("Provider").
		Where("id = ?", id).
		First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// GetByExternalID busca uma transação por ID externo (do merchant)
func (r *TransactionRepository) GetByExternalID(ctx context.Context, merchantID uuid.UUID, externalID string) (*domain.Transaction, error) {
	var tx domain.Transaction
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("Provider").
		Where("merchant_id = ? AND external_id = ?", merchantID, externalID).
		First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// GetByE2EID busca uma transação por End-to-End ID
func (r *TransactionRepository) GetByE2EID(ctx context.Context, e2eID string) (*domain.Transaction, error) {
	var tx domain.Transaction
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("Provider").
		Where("e2e_id = ?", e2eID).
		First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// GetByProviderTxID busca uma transação por ID do provider
func (r *TransactionRepository) GetByProviderTxID(ctx context.Context, providerID uuid.UUID, providerTxID string) (*domain.Transaction, error) {
	var tx domain.Transaction
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("Provider").
		Where("provider_id = ? AND provider_tx_id = ?", providerID, providerTxID).
		First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// Update atualiza uma transação
func (r *TransactionRepository) Update(ctx context.Context, tx *domain.Transaction) error {
	return r.db.WithContext(ctx).Save(tx).Error
}

// UpdateStatus atualiza apenas o status de uma transação
func (r *TransactionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TransactionStatus) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == domain.TransactionStatusProcessing {
		updates["processed_at"] = time.Now()
	} else if status == domain.TransactionStatusCompleted {
		updates["completed_at"] = time.Now()
	} else if status == domain.TransactionStatusCancelled {
		updates["cancelled_at"] = time.Now()
	}

	return r.db.WithContext(ctx).Model(&domain.Transaction{}).Where("id = ?", id).Updates(updates).Error
}

// ListByMerchant lista transações de um merchant com filtros
func (r *TransactionRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID, filters map[string]interface{}, limit, offset int) ([]domain.Transaction, int64, error) {
	var transactions []domain.Transaction
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Transaction{}).Where("merchant_id = ?", merchantID)

	// Aplicar filtros
	if status, ok := filters["status"].(domain.TransactionStatus); ok {
		query = query.Where("status = ?", status)
	}

	if txType, ok := filters["type"].(domain.TransactionType); ok {
		query = query.Where("type = ?", txType)
	}

	if startDate, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("created_at >= ?", startDate)
	}

	if endDate, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("created_at <= ?", endDate)
	}

	if minAmount, ok := filters["min_amount"].(int64); ok {
		query = query.Where("amount >= ?", minAmount)
	}

	if maxAmount, ok := filters["max_amount"].(int64); ok {
		query = query.Where("amount <= ?", maxAmount)
	}

	// Contar total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Buscar transações
	err := query.
		Preload("Provider").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	return transactions, total, err
}

// GetStatistics retorna estatísticas de transações
func (r *TransactionRepository) GetStatistics(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	baseQuery := r.db.WithContext(ctx).Model(&domain.Transaction{}).
		Where("merchant_id = ? AND created_at BETWEEN ? AND ?", merchantID, startDate, endDate)

	// Total de transações
	var total int64
	baseQuery.Count(&total)
	stats["total"] = total

	// Total por status
	var statusStats []struct {
		Status domain.TransactionStatus
		Count  int64
		Amount int64
	}
	baseQuery.Select("status, COUNT(*) as count, SUM(amount) as amount").
		Group("status").
		Scan(&statusStats)
	stats["by_status"] = statusStats

	// Total por tipo
	var typeStats []struct {
		Type   domain.TransactionType
		Count  int64
		Amount int64
	}
	baseQuery.Select("type, COUNT(*) as count, SUM(amount) as amount").
		Group("type").
		Scan(&typeStats)
	stats["by_type"] = typeStats

	// Volume total
	var totalAmount int64
	baseQuery.Where("status = ?", domain.TransactionStatusCompleted).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalAmount)
	stats["total_amount"] = totalAmount

	// Taxa de sucesso
	var successCount int64
	baseQuery.Where("status = ?", domain.TransactionStatusCompleted).Count(&successCount)
	successRate := float64(0)
	if total > 0 {
		successRate = float64(successCount) / float64(total) * 100
	}
	stats["success_rate"] = successRate

	return stats, nil
}

// GetPendingTransactions busca transações pendentes ou em processamento
func (r *TransactionRepository) GetPendingTransactions(ctx context.Context, limit int) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := r.db.WithContext(ctx).
		Preload("Merchant").
		Preload("Provider").
		Where("status IN ?", []domain.TransactionStatus{
			domain.TransactionStatusPending,
			domain.TransactionStatusProcessing,
		}).
		Order("created_at ASC").
		Limit(limit).
		Find(&transactions).Error

	return transactions, err
}

// GetExpiredQRCodes busca QR Codes expirados
func (r *TransactionRepository) GetExpiredQRCodes(ctx context.Context) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	now := time.Now()

	err := r.db.WithContext(ctx).
		Where("type IN ? AND qr_code_expires_at < ? AND status = ?",
			[]domain.TransactionType{
				domain.TransactionTypeQRCodeStatic,
				domain.TransactionTypeQRCodeDynamic,
			},
			now,
			domain.TransactionStatusPending,
		).
		Find(&transactions).Error

	return transactions, err
}
