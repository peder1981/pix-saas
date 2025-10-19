package providers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pixsaas/backend/internal/domain"
)

// PixProvider é a interface que todos os providers bancários devem implementar
type PixProvider interface {
	// GetCode retorna o código único do provider (bradesco, itau, bb, etc)
	GetCode() string
	
	// GetName retorna o nome do provider
	GetName() string
	
	// Initialize inicializa o provider com as configurações
	Initialize(config domain.ProviderConfig) error
	
	// Authenticate realiza autenticação e obtém token de acesso
	Authenticate(ctx context.Context, credentials ProviderCredentials) (*AuthToken, error)
	
	// RefreshToken renova o token de acesso
	RefreshToken(ctx context.Context, refreshToken string) (*AuthToken, error)
	
	// CreateTransfer cria uma transferência PIX
	CreateTransfer(ctx context.Context, req *TransferRequest) (*TransferResponse, error)
	
	// GetTransfer consulta uma transferência PIX
	GetTransfer(ctx context.Context, txID string) (*TransferResponse, error)
	
	// CancelTransfer cancela uma transferência PIX (se suportado)
	CancelTransfer(ctx context.Context, txID string) error
	
	// CreateQRCodeStatic cria um QR Code estático
	CreateQRCodeStatic(ctx context.Context, req *QRCodeRequest) (*QRCodeResponse, error)
	
	// CreateQRCodeDynamic cria um QR Code dinâmico
	CreateQRCodeDynamic(ctx context.Context, req *QRCodeRequest) (*QRCodeResponse, error)
	
	// GetQRCode consulta informações de um QR Code
	GetQRCode(ctx context.Context, qrCodeID string) (*QRCodeResponse, error)
	
	// ValidatePixKey valida se uma chave PIX existe e retorna informações
	ValidatePixKey(ctx context.Context, pixKey string, pixKeyType domain.PixKeyType) (*PixKeyInfo, error)
	
	// HealthCheck verifica se o provider está saudável
	HealthCheck(ctx context.Context) error
	
	// GetSupportedMethods retorna os métodos suportados pelo provider
	GetSupportedMethods() []string
}

// ProviderCredentials representa as credenciais de autenticação
type ProviderCredentials struct {
	ClientID       string
	ClientSecret   string
	Certificate    []byte // Certificado mTLS
	PrivateKey     []byte // Chave privada mTLS
	AccountAgency  string
	AccountNumber  string
	AccountType    string
	PixKey         string
	PixKeyType     domain.PixKeyType
}

// AuthToken representa um token de autenticação
type AuthToken struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int       // Segundos
	ExpiresAt    time.Time
	Scope        string
}

// TransferRequest representa uma requisição de transferência PIX
type TransferRequest struct {
	ExternalID  string // ID do merchant
	Amount      int64  // Centavos
	Description string
	
	// Pagador
	PayerName          string
	PayerDocument      string
	PayerPixKey        string
	PayerPixKeyType    domain.PixKeyType
	PayerAccountAgency string
	PayerAccountNumber string
	PayerAccountType   string
	PayerBank          string
	
	// Recebedor
	PayeeName          string
	PayeeDocument      string
	PayeePixKey        string
	PayeePixKeyType    domain.PixKeyType
	PayeeAccountAgency string
	PayeeAccountNumber string
	PayeeAccountType   string
	PayeeBank          string
	PayeeISPB          string
	
	// Metadata adicional
	Metadata map[string]interface{}
}

// TransferResponse representa a resposta de uma transferência PIX
type TransferResponse struct {
	ProviderTxID string
	E2EID        string // End-to-end ID
	Status       domain.TransactionStatus
	Amount       int64
	Description  string
	
	// Informações do pagador
	PayerName          string
	PayerDocument      string
	PayerPixKey        string
	PayerPixKeyType    domain.PixKeyType
	PayerAccountAgency string
	PayerAccountNumber string
	PayerBank          string
	
	// Informações do recebedor
	PayeeName          string
	PayeeDocument      string
	PayeePixKey        string
	PayeePixKeyType    domain.PixKeyType
	PayeeAccountAgency string
	PayeeAccountNumber string
	PayeeBank          string
	
	// Timestamps
	ProcessedAt *time.Time
	CompletedAt *time.Time
	
	// Erro (se houver)
	ErrorCode    string
	ErrorMessage string
	
	// Dados brutos do provider
	RawResponse map[string]interface{}
}

// QRCodeRequest representa uma requisição de criação de QR Code
type QRCodeRequest struct {
	ExternalID       string
	Amount           int64 // Centavos (0 para QR Code estático sem valor)
	Description      string
	PayeeName        string
	PayeeDocument    string
	PayeePixKey      string
	PayeePixKeyType  domain.PixKeyType
	ExpiresIn        int // Segundos (para QR Code dinâmico)
	AllowChange      bool // Permite alterar valor
	Metadata         map[string]interface{}
}

// QRCodeResponse representa a resposta de criação/consulta de QR Code
type QRCodeResponse struct {
	QRCodeID     string
	QRCode       string // Código PIX copia e cola
	QRCodeImage  string // Base64 da imagem
	Amount       int64
	Description  string
	Status       string
	ExpiresAt    *time.Time
	CreatedAt    time.Time
	RawResponse  map[string]interface{}
}

// PixKeyInfo representa informações de uma chave PIX
type PixKeyInfo struct {
	PixKey      string
	PixKeyType  domain.PixKeyType
	Name        string
	Document    string
	Bank        string
	ISPB        string
	AccountType string
	CreatedAt   time.Time
}

// ProviderError representa um erro específico do provider
type ProviderError struct {
	Code       string
	Message    string
	StatusCode int
	Retryable  bool
	Details    map[string]interface{}
}

func (e *ProviderError) Error() string {
	return e.Message
}

// ProviderRegistry gerencia todos os providers registrados
type ProviderRegistry struct {
	providers map[string]PixProvider
}

// NewProviderRegistry cria um novo registro de providers
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]PixProvider),
	}
}

// Register registra um novo provider
func (r *ProviderRegistry) Register(provider PixProvider) {
	r.providers[provider.GetCode()] = provider
}

// Get retorna um provider pelo código
func (r *ProviderRegistry) Get(code string) (PixProvider, bool) {
	provider, exists := r.providers[code]
	return provider, exists
}

// GetAll retorna todos os providers registrados
func (r *ProviderRegistry) GetAll() map[string]PixProvider {
	return r.providers
}

// ProviderManager gerencia a interação com providers
type ProviderManager struct {
	registry *ProviderRegistry
}

// NewProviderManager cria um novo gerenciador de providers
func NewProviderManager(registry *ProviderRegistry) *ProviderManager {
	return &ProviderManager{
		registry: registry,
	}
}

// ExecuteWithFallback executa uma operação com fallback para outros providers
func (r *ProviderManager) ExecuteWithFallback(
	ctx context.Context,
	merchantID uuid.UUID,
	operation func(provider PixProvider) error,
) error {
	// TODO: Implementar lógica de fallback baseada em prioridade e saúde dos providers
	return nil
}

// GetHealthyProvider retorna um provider saudável para o merchant
func (r *ProviderManager) GetHealthyProvider(
	ctx context.Context,
	merchantID uuid.UUID,
	preferredProvider string,
) (PixProvider, error) {
	// TODO: Implementar lógica de seleção de provider baseada em saúde
	return nil, nil
}
