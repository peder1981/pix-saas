package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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
	Initialize(config ProviderConfig) error
	
	// Authenticate realiza autenticação e obtém token de acesso
	Authenticate(ctx context.Context, credentials ProviderCredentials) (*AuthToken, error)
	
	// RefreshToken renova o token de acesso
	RefreshToken(ctx context.Context, refreshToken string) (*AuthToken, error)
	
	// CreateTransfer cria uma transferência PIX
	CreateTransfer(ctx context.Context, req *TransferRequest) (*TransferResponse, error)
	
	// GetTransfer consulta uma transferência PIX
	GetTransfer(ctx context.Context, req *GetTransferRequest) (*TransferResponse, error)
	
	// CancelTransfer cancela uma transferência PIX (se suportado)
	CancelTransfer(ctx context.Context, req *CancelTransferRequest) error
	
	// CreateQRCodeStatic cria um QR Code estático
	CreateQRCodeStatic(ctx context.Context, req *QRCodeRequest) (*QRCodeResponse, error)
	
	// CreateQRCodeDynamic cria um QR Code dinâmico
	CreateQRCodeDynamic(ctx context.Context, req *QRCodeRequest) (*QRCodeResponse, error)
	
	// GetQRCode consulta informações de um QR Code
	GetQRCode(ctx context.Context, req *GetQRCodeRequest) (*QRCodeResponse, error)
	
	// ValidatePixKey valida se uma chave PIX existe e retorna informações
	ValidatePixKey(ctx context.Context, req *ValidatePixKeyRequest) (*ValidatePixKeyResponse, error)
	
	// HealthCheck verifica se o provider está saudável
	HealthCheck(ctx context.Context) error
	
	// GetSupportedMethods retorna os métodos suportados pelo provider
	GetSupportedMethods() []string
}

// ProviderConfig representa a configuração de um provider
type ProviderConfig struct {
	BaseURL      string
	AuthURL      string
	SandboxURL   string
	Timeout      int
	MaxRetries   int
	RequiresMTLS bool
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
	
	// Auth token
	AuthToken string
	ClientID  string
}

// GetTransferRequest representa uma requisição de consulta de transferência
type GetTransferRequest struct {
	ProviderTxID string
	AuthToken    string
	ClientID     string
}

// CancelTransferRequest representa uma requisição de cancelamento
type CancelTransferRequest struct {
	ProviderTxID string
	AuthToken    string
	ClientID     string
	Reason       string
}

// GetQRCodeRequest representa uma requisição de consulta de QR Code
type GetQRCodeRequest struct {
	QRCodeID  string
	AuthToken string
	ClientID  string
}

// ValidatePixKeyRequest representa uma requisição de validação de chave PIX
type ValidatePixKeyRequest struct {
	PixKey     string
	PixKeyType domain.PixKeyType
	AuthToken  string
	ClientID   string
}

// ValidatePixKeyResponse representa a resposta de validação de chave PIX
type ValidatePixKeyResponse struct {
	Valid       bool
	PixKey      string
	PixKeyType  domain.PixKeyType
	Name        string
	Document    string
	Bank        string
	ISPB        string
	AccountType string
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
	PixKey           string
	PayeePixKey      string
	PayeePixKeyType  domain.PixKeyType
	ExpiresIn        int // Segundos (para QR Code dinâmico)
	AllowChange      bool // Permite alterar valor
	Metadata         map[string]interface{}
	AuthToken        string
	ClientID         string
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

// HTTPClient é um cliente HTTP para comunicação com providers
type HTTPClient struct {
	client      *http.Client
	timeout     time.Duration
	requireMTLS bool
}

// NewHTTPClient cria um novo cliente HTTP
func NewHTTPClient(timeout int, requireMTLS bool) HTTPClient {
	return HTTPClient{
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		timeout:     time.Duration(timeout) * time.Second,
		requireMTLS: requireMTLS,
	}
}

// Post faz uma requisição POST
func (c *HTTPClient) Post(ctx context.Context, url string, payload interface{}, headers map[string]string) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Get faz uma requisição GET
func (c *HTTPClient) Get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// PostForm faz uma requisição POST com form data
func (c *HTTPClient) PostForm(ctx context.Context, urlStr string, data map[string]string, headers map[string]string) ([]byte, error) {
	form := url.Values{}
	for key, value := range data {
		form.Add(key, value)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", urlStr, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// PostFormWithBasicAuth faz uma requisição POST com form data e Basic Auth
func (c *HTTPClient) PostFormWithBasicAuth(ctx context.Context, urlStr string, data map[string]string, headers map[string]string, username, password string) ([]byte, error) {
	form := url.Values{}
	for key, value := range data {
		form.Add(key, value)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", urlStr, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// NewProviderError cria um novo erro de provider
func NewProviderError(code, message string, err error) error {
	if err != nil {
		message = fmt.Sprintf("%s: %v", message, err)
	}
	return &ProviderError{
		Code:    code,
		Message: message,
	}
}

// TransactionStatus representa o status de uma transação
type TransactionStatus = domain.TransactionStatus

// Constantes de status
const (
	TransactionStatusPending    = domain.TransactionStatusPending
	TransactionStatusProcessing = domain.TransactionStatusProcessing
	TransactionStatusCompleted  = domain.TransactionStatusCompleted
	TransactionStatusFailed     = domain.TransactionStatusFailed
	TransactionStatusCancelled  = domain.TransactionStatusCancelled
)
