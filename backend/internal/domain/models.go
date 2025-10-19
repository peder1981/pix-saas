package domain

import (
	"time"

	"github.com/google/uuid"
)

// Merchant representa um cliente da plataforma (multi-tenant)
type Merchant struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string     `json:"name" gorm:"not null"`
	Document    string     `json:"document" gorm:"uniqueIndex;not null"` // CPF/CNPJ
	Email       string     `json:"email" gorm:"uniqueIndex;not null"`
	Phone       string     `json:"phone"`
	Active      bool       `json:"active" gorm:"default:true"`
	APIKey      string     `json:"-" gorm:"uniqueIndex;not null"` // Criptografado
	WebhookURL  string     `json:"webhook_url"`
	IPWhitelist []string   `json:"ip_whitelist" gorm:"type:text[]"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// User representa usuários do sistema (admin, merchant users)
type User struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MerchantID *uuid.UUID `json:"merchant_id,omitempty" gorm:"type:uuid;index"`
	Email      string     `json:"email" gorm:"uniqueIndex;not null"`
	Password   string     `json:"-" gorm:"not null"` // Hash bcrypt
	Name       string     `json:"name" gorm:"not null"`
	Role       UserRole   `json:"role" gorm:"not null"`
	Active     bool       `json:"active" gorm:"default:true"`
	LastLogin  *time.Time `json:"last_login,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleMerchant  UserRole = "merchant"
	RoleDeveloper UserRole = "developer"
)

// Provider representa uma instituição financeira
type Provider struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Code         string         `json:"code" gorm:"uniqueIndex;not null"` // bradesco, itau, bb, etc
	Name         string         `json:"name" gorm:"not null"`
	ISPB         string         `json:"ispb" gorm:"uniqueIndex;not null"` // Identificador SPB
	Type         ProviderType   `json:"type" gorm:"not null"`
	Active       bool           `json:"active" gorm:"default:true"`
	Config       ProviderConfig `json:"config" gorm:"type:jsonb"`
	Priority     int            `json:"priority" gorm:"default:0"` // Para fallback
	HealthStatus string         `json:"health_status" gorm:"default:'unknown'"`
	LastHealthAt *time.Time     `json:"last_health_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    *time.Time     `json:"deleted_at,omitempty" gorm:"index"`
}

type ProviderType string

const (
	ProviderTypeBank        ProviderType = "bank"
	ProviderTypeDigital     ProviderType = "digital_bank"
	ProviderTypeCooperative ProviderType = "cooperative"
	ProviderTypeFintech     ProviderType = "fintech"
	ProviderTypePSP         ProviderType = "psp"
)

// ProviderConfig armazena configurações específicas de cada provider
type ProviderConfig struct {
	BaseURL          string            `json:"base_url"`
	AuthURL          string            `json:"auth_url"`
	SandboxURL       string            `json:"sandbox_url,omitempty"`
	AuthType         string            `json:"auth_type"` // oauth2, mtls, api_key
	Timeout          int               `json:"timeout"`   // segundos
	MaxRetries       int               `json:"max_retries"`
	RequiresMTLS     bool              `json:"requires_mtls"`
	SupportedMethods []string          `json:"supported_methods"` // pix_key, account, qrcode
	CustomHeaders    map[string]string `json:"custom_headers,omitempty"`
}

// MerchantProvider representa a configuração de um merchant com um provider específico
type MerchantProvider struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MerchantID       uuid.UUID  `json:"merchant_id" gorm:"type:uuid;not null;index"`
	ProviderID       uuid.UUID  `json:"provider_id" gorm:"type:uuid;not null;index"`
	Active           bool       `json:"active" gorm:"default:true"`
	ClientID         string     `json:"-" gorm:"not null"` // Criptografado
	ClientSecret     string     `json:"-" gorm:"not null"` // Criptografado
	CertificateData  string     `json:"-"`                 // Certificado mTLS criptografado
	PrivateKeyData   string     `json:"-"`                 // Chave privada criptografada
	AccountAgency    string     `json:"account_agency"`
	AccountNumber    string     `json:"account_number"`
	AccountType      string     `json:"account_type"` // checking, savings
	PixKey           string     `json:"pix_key,omitempty"`
	PixKeyType       PixKeyType `json:"pix_key_type,omitempty"`
	LastTokenRefresh *time.Time `json:"last_token_refresh,omitempty"`
	TokenExpiresAt   *time.Time `json:"token_expires_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	// Relacionamentos
	Merchant Merchant `json:"merchant,omitempty" gorm:"foreignKey:MerchantID"`
	Provider Provider `json:"provider,omitempty" gorm:"foreignKey:ProviderID"`
}

type PixKeyType string

const (
	PixKeyTypeCPF     PixKeyType = "cpf"
	PixKeyTypeCNPJ    PixKeyType = "cnpj"
	PixKeyTypeEmail   PixKeyType = "email"
	PixKeyTypePhone   PixKeyType = "phone"
	PixKeyTypeRandom  PixKeyType = "random"
	PixKeyTypeAccount PixKeyType = "account"
)

// Transaction representa uma transação PIX
type Transaction struct {
	ID           uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MerchantID   uuid.UUID         `json:"merchant_id" gorm:"type:uuid;not null;index"`
	ProviderID   uuid.UUID         `json:"provider_id" gorm:"type:uuid;not null;index"`
	ExternalID   string            `json:"external_id" gorm:"uniqueIndex"` // ID do merchant
	ProviderTxID string            `json:"provider_tx_id" gorm:"index"`    // ID do banco
	E2EID        string            `json:"e2e_id" gorm:"uniqueIndex"`      // End-to-end ID do PIX
	Type         TransactionType   `json:"type" gorm:"not null"`
	Status       TransactionStatus `json:"status" gorm:"not null;index"`
	Amount       int64             `json:"amount" gorm:"not null"` // Centavos
	Currency     string            `json:"currency" gorm:"default:'BRL'"`
	Description  string            `json:"description"`

	// Pagador
	PayerName          string     `json:"payer_name"`
	PayerDocument      string     `json:"payer_document"`
	PayerPixKey        string     `json:"payer_pix_key,omitempty"`
	PayerPixKeyType    PixKeyType `json:"payer_pix_key_type,omitempty"`
	PayerAccountAgency string     `json:"payer_account_agency,omitempty"`
	PayerAccountNumber string     `json:"payer_account_number,omitempty"`
	PayerBank          string     `json:"payer_bank,omitempty"`

	// Recebedor
	PayeeName          string     `json:"payee_name"`
	PayeeDocument      string     `json:"payee_document"`
	PayeePixKey        string     `json:"payee_pix_key,omitempty"`
	PayeePixKeyType    PixKeyType `json:"payee_pix_key_type,omitempty"`
	PayeeAccountAgency string     `json:"payee_account_agency,omitempty"`
	PayeeAccountNumber string     `json:"payee_account_number,omitempty"`
	PayeeBank          string     `json:"payee_bank,omitempty"`

	// QR Code (se aplicável)
	QRCode          string     `json:"qr_code,omitempty"`
	QRCodeImage     string     `json:"qr_code_image,omitempty"` // Base64
	QRCodeExpiresAt *time.Time `json:"qr_code_expires_at,omitempty"`

	// Metadata
	Metadata     map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	ErrorCode    string                 `json:"error_code,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`

	// Timestamps
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relacionamentos
	Merchant Merchant `json:"merchant,omitempty" gorm:"foreignKey:MerchantID"`
	Provider Provider `json:"provider,omitempty" gorm:"foreignKey:ProviderID"`
}

type TransactionType string

const (
	TransactionTypeTransfer      TransactionType = "transfer"
	TransactionTypeQRCodeStatic  TransactionType = "qrcode_static"
	TransactionTypeQRCodeDynamic TransactionType = "qrcode_dynamic"
	TransactionTypePixCopyPaste  TransactionType = "pix_copy_paste"
)

type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "pending"
	TransactionStatusProcessing TransactionStatus = "processing"
	TransactionStatusCompleted  TransactionStatus = "completed"
	TransactionStatusFailed     TransactionStatus = "failed"
	TransactionStatusCancelled  TransactionStatus = "cancelled"
	TransactionStatusRefunded   TransactionStatus = "refunded"
)

// AuditLog representa logs de auditoria (retenção 5 anos)
type AuditLog struct {
	ID            uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MerchantID    *uuid.UUID             `json:"merchant_id,omitempty" gorm:"type:uuid;index"`
	UserID        *uuid.UUID             `json:"user_id,omitempty" gorm:"type:uuid;index"`
	TransactionID *uuid.UUID             `json:"transaction_id,omitempty" gorm:"type:uuid;index"`
	Action        string                 `json:"action" gorm:"not null;index"`   // create, update, delete, auth, etc
	Resource      string                 `json:"resource" gorm:"not null;index"` // transaction, merchant, user, etc
	Method        string                 `json:"method"`                         // GET, POST, PUT, DELETE
	Path          string                 `json:"path"`
	IPAddress     string                 `json:"ip_address" gorm:"index"`
	UserAgent     string                 `json:"user_agent"`
	RequestBody   map[string]interface{} `json:"request_body,omitempty" gorm:"type:jsonb"`
	ResponseCode  int                    `json:"response_code"`
	ResponseBody  map[string]interface{} `json:"response_body,omitempty" gorm:"type:jsonb"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	Duration      int64                  `json:"duration"` // Milissegundos
	Metadata      map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt     time.Time              `json:"created_at" gorm:"index"`
}

// Webhook representa configurações de webhook
type Webhook struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MerchantID uuid.UUID  `json:"merchant_id" gorm:"type:uuid;not null;index"`
	URL        string     `json:"url" gorm:"not null"`
	Events     []string   `json:"events" gorm:"type:text[]"` // transaction.completed, transaction.failed, etc
	Secret     string     `json:"-" gorm:"not null"`         // Para HMAC signature
	Active     bool       `json:"active" gorm:"default:true"`
	MaxRetries int        `json:"max_retries" gorm:"default:3"`
	Timeout    int        `json:"timeout" gorm:"default:30"` // segundos
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	// Relacionamento
	Merchant Merchant `json:"merchant,omitempty" gorm:"foreignKey:MerchantID"`
}

// WebhookDelivery representa tentativas de entrega de webhook
type WebhookDelivery struct {
	ID            uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WebhookID     uuid.UUID              `json:"webhook_id" gorm:"type:uuid;not null;index"`
	TransactionID uuid.UUID              `json:"transaction_id" gorm:"type:uuid;not null;index"`
	Event         string                 `json:"event" gorm:"not null"`
	Payload       map[string]interface{} `json:"payload" gorm:"type:jsonb"`
	Attempt       int                    `json:"attempt" gorm:"default:1"`
	Status        string                 `json:"status" gorm:"not null"` // pending, success, failed
	ResponseCode  int                    `json:"response_code"`
	ResponseBody  string                 `json:"response_body,omitempty"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	NextRetryAt   *time.Time             `json:"next_retry_at,omitempty"`
	DeliveredAt   *time.Time             `json:"delivered_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at" gorm:"index"`

	// Relacionamentos
	Webhook     Webhook     `json:"webhook,omitempty" gorm:"foreignKey:WebhookID"`
	Transaction Transaction `json:"transaction,omitempty" gorm:"foreignKey:TransactionID"`
}

// APIKey representa chaves de API para autenticação
type APIKey struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MerchantID  uuid.UUID  `json:"merchant_id" gorm:"type:uuid;not null;index"`
	Name        string     `json:"name" gorm:"not null"`
	Key         string     `json:"-" gorm:"uniqueIndex;not null"` // Hash da chave
	Prefix      string     `json:"prefix" gorm:"not null"`        // Primeiros 8 chars para identificação
	Permissions []string   `json:"permissions" gorm:"type:text[]"`
	Active      bool       `json:"active" gorm:"default:true"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	// Relacionamento
	Merchant Merchant `json:"merchant,omitempty" gorm:"foreignKey:MerchantID"`
}

// RefreshToken representa tokens de refresh JWT
type RefreshToken struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Token     string     `json:"-" gorm:"uniqueIndex;not null"` // Hash do token
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null;index"`
	Revoked   bool       `json:"revoked" gorm:"default:false"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`

	// Relacionamento
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
