package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pixsaas/backend/internal/domain"
	"gorm.io/gorm"
)

// AuditService gerencia logs de auditoria
type AuditService struct {
	db *gorm.DB
}

// NewAuditService cria um novo serviço de auditoria
func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{
		db: db,
	}
}

// LogEntry representa uma entrada de log
type LogEntry struct {
	MerchantID    *uuid.UUID
	UserID        *uuid.UUID
	TransactionID *uuid.UUID
	Action        string
	Resource      string
	Method        string
	Path          string
	IPAddress     string
	UserAgent     string
	RequestBody   interface{}
	ResponseCode  int
	ResponseBody  interface{}
	ErrorMessage  string
	Duration      int64 // Milissegundos
	Metadata      map[string]interface{}
}

// Log registra uma entrada de auditoria
func (s *AuditService) Log(ctx context.Context, entry *LogEntry) error {
	auditLog := &domain.AuditLog{
		ID:            uuid.New(),
		MerchantID:    entry.MerchantID,
		UserID:        entry.UserID,
		TransactionID: entry.TransactionID,
		Action:        entry.Action,
		Resource:      entry.Resource,
		Method:        entry.Method,
		Path:          entry.Path,
		IPAddress:     entry.IPAddress,
		UserAgent:     entry.UserAgent,
		ResponseCode:  entry.ResponseCode,
		ErrorMessage:  entry.ErrorMessage,
		Duration:      entry.Duration,
		CreatedAt:     time.Now(),
	}

	// Converter RequestBody para map
	if entry.RequestBody != nil {
		if reqMap, ok := entry.RequestBody.(map[string]interface{}); ok {
			auditLog.RequestBody = reqMap
		} else {
			// Tentar serializar e deserializar
			jsonData, _ := json.Marshal(entry.RequestBody)
			var reqMap map[string]interface{}
			_ = json.Unmarshal(jsonData, &reqMap) //nolint:errcheck
			auditLog.RequestBody = reqMap
		}
	}

	// Converter ResponseBody para map
	if entry.ResponseBody != nil {
		if respMap, ok := entry.ResponseBody.(map[string]interface{}); ok {
			auditLog.ResponseBody = respMap
		} else {
			jsonData, _ := json.Marshal(entry.ResponseBody)
			var respMap map[string]interface{}
			_ = json.Unmarshal(jsonData, &respMap) //nolint:errcheck
			auditLog.ResponseBody = respMap
		}
	}

	if entry.Metadata != nil {
		auditLog.Metadata = entry.Metadata
	}

	return s.db.WithContext(ctx).Create(auditLog).Error
}

// LogTransaction registra uma operação de transação
func (s *AuditService) LogTransaction(ctx context.Context, merchantID, userID, transactionID uuid.UUID, action string, metadata map[string]interface{}) error {
	return s.Log(ctx, &LogEntry{
		MerchantID:    &merchantID,
		UserID:        &userID,
		TransactionID: &transactionID,
		Action:        action,
		Resource:      "transaction",
		Metadata:      metadata,
	})
}

// LogAuthentication registra tentativas de autenticação
func (s *AuditService) LogAuthentication(ctx context.Context, email, ipAddress string, success bool, errorMsg string) error {
	action := "auth_success"
	responseCode := 200

	if !success {
		action = "auth_failed"
		responseCode = 401
	}

	return s.Log(ctx, &LogEntry{
		Action:       action,
		Resource:     "authentication",
		IPAddress:    ipAddress,
		ResponseCode: responseCode,
		ErrorMessage: errorMsg,
		Metadata: map[string]interface{}{
			"email":   email,
			"success": success,
		},
	})
}

// LogAPIAccess registra acesso à API
func (s *AuditService) LogAPIAccess(ctx context.Context, merchantID uuid.UUID, method, path, ipAddress, userAgent string, statusCode int, duration int64) error {
	return s.Log(ctx, &LogEntry{
		MerchantID:   &merchantID,
		Action:       "api_access",
		Resource:     "api",
		Method:       method,
		Path:         path,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ResponseCode: statusCode,
		Duration:     duration,
	})
}

// LogProviderOperation registra operações com providers
func (s *AuditService) LogProviderOperation(ctx context.Context, merchantID, transactionID uuid.UUID, provider, operation string, success bool, errorMsg string, duration int64) error {
	action := "provider_" + operation
	if !success {
		action += "_failed"
	}

	return s.Log(ctx, &LogEntry{
		MerchantID:    &merchantID,
		TransactionID: &transactionID,
		Action:        action,
		Resource:      "provider",
		Duration:      duration,
		ErrorMessage:  errorMsg,
		Metadata: map[string]interface{}{
			"provider":  provider,
			"operation": operation,
			"success":   success,
		},
	})
}

// LogWebhookDelivery registra tentativas de entrega de webhook
func (s *AuditService) LogWebhookDelivery(ctx context.Context, merchantID, webhookID, transactionID uuid.UUID, event string, attempt int, success bool, statusCode int, errorMsg string) error {
	action := "webhook_delivery"
	if !success {
		action = "webhook_delivery_failed"
	}

	return s.Log(ctx, &LogEntry{
		MerchantID:    &merchantID,
		TransactionID: &transactionID,
		Action:        action,
		Resource:      "webhook",
		ResponseCode:  statusCode,
		ErrorMessage:  errorMsg,
		Metadata: map[string]interface{}{
			"webhook_id": webhookID.String(),
			"event":      event,
			"attempt":    attempt,
			"success":    success,
		},
	})
}

// LogSecurityEvent registra eventos de segurança
func (s *AuditService) LogSecurityEvent(ctx context.Context, eventType, description, ipAddress string, severity string, metadata map[string]interface{}) error {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["severity"] = severity
	metadata["event_type"] = eventType

	return s.Log(ctx, &LogEntry{
		Action:    "security_event",
		Resource:  "security",
		IPAddress: ipAddress,
		Metadata:  metadata,
	})
}

// LogDataAccess registra acesso a dados sensíveis
func (s *AuditService) LogDataAccess(ctx context.Context, userID uuid.UUID, resource, action, ipAddress string, metadata map[string]interface{}) error {
	return s.Log(ctx, &LogEntry{
		UserID:    &userID,
		Action:    action,
		Resource:  resource,
		IPAddress: ipAddress,
		Metadata:  metadata,
	})
}

// QueryLogs busca logs de auditoria com filtros
func (s *AuditService) QueryLogs(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]domain.AuditLog, int64, error) {
	var logs []domain.AuditLog
	var total int64

	query := s.db.WithContext(ctx).Model(&domain.AuditLog{})

	// Aplicar filtros
	if merchantID, ok := filters["merchant_id"].(uuid.UUID); ok {
		query = query.Where("merchant_id = ?", merchantID)
	}

	if userID, ok := filters["user_id"].(uuid.UUID); ok {
		query = query.Where("user_id = ?", userID)
	}

	if transactionID, ok := filters["transaction_id"].(uuid.UUID); ok {
		query = query.Where("transaction_id = ?", transactionID)
	}

	if action, ok := filters["action"].(string); ok {
		query = query.Where("action = ?", action)
	}

	if resource, ok := filters["resource"].(string); ok {
		query = query.Where("resource = ?", resource)
	}

	if ipAddress, ok := filters["ip_address"].(string); ok {
		query = query.Where("ip_address = ?", ipAddress)
	}

	if startDate, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("created_at >= ?", startDate)
	}

	if endDate, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("created_at <= ?", endDate)
	}

	// Contar total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Buscar logs
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	return logs, total, err
}

// CleanupOldLogs remove logs mais antigos que o período de retenção
// Para compliance brasileiro: manter 5 anos
func (s *AuditService) CleanupOldLogs(ctx context.Context, retentionYears int) error {
	cutoffDate := time.Now().AddDate(-retentionYears, 0, 0)

	result := s.db.WithContext(ctx).
		Where("created_at < ?", cutoffDate).
		Delete(&domain.AuditLog{})

	return result.Error
}

// GetLogStatistics retorna estatísticas de logs
func (s *AuditService) GetLogStatistics(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total de logs
	var total int64
	s.db.WithContext(ctx).
		Model(&domain.AuditLog{}).
		Where("merchant_id = ? AND created_at BETWEEN ? AND ?", merchantID, startDate, endDate).
		Count(&total)
	stats["total"] = total

	// Logs por ação
	var actionStats []struct {
		Action string
		Count  int64
	}
	s.db.WithContext(ctx).
		Model(&domain.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("merchant_id = ? AND created_at BETWEEN ? AND ?", merchantID, startDate, endDate).
		Group("action").
		Scan(&actionStats)
	stats["by_action"] = actionStats

	// Logs por recurso
	var resourceStats []struct {
		Resource string
		Count    int64
	}
	s.db.WithContext(ctx).
		Model(&domain.AuditLog{}).
		Select("resource, COUNT(*) as count").
		Where("merchant_id = ? AND created_at BETWEEN ? AND ?", merchantID, startDate, endDate).
		Group("resource").
		Scan(&resourceStats)
	stats["by_resource"] = resourceStats

	// Erros
	var errorCount int64
	s.db.WithContext(ctx).
		Model(&domain.AuditLog{}).
		Where("merchant_id = ? AND created_at BETWEEN ? AND ? AND error_message != ''", merchantID, startDate, endDate).
		Count(&errorCount)
	stats["errors"] = errorCount

	return stats, nil
}
