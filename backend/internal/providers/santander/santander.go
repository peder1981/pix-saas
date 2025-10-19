package santander

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pixsaas/backend/internal/providers"
)

// SantanderProvider implementa o provider do Santander
type SantanderProvider struct {
	config     providers.ProviderConfig
	httpClient providers.HTTPClient
}

// NewSantanderProvider cria uma nova instância do provider Santander
func NewSantanderProvider() *SantanderProvider {
	return &SantanderProvider{}
}

// GetCode retorna o código do provider
func (p *SantanderProvider) GetCode() string {
	return "santander"
}

// GetName retorna o nome do provider
func (p *SantanderProvider) GetName() string {
	return "Santander"
}

// Initialize inicializa o provider
func (p *SantanderProvider) Initialize(config providers.ProviderConfig) error {
	p.config = config
	p.httpClient = providers.NewHTTPClient(config.Timeout, config.RequiresMTLS)
	return nil
}

// Authenticate realiza autenticação OAuth2
func (p *SantanderProvider) Authenticate(ctx context.Context, credentials providers.ProviderCredentials) (*providers.AuthToken, error) {
	authURL := p.config.AuthURL
	
	payload := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     credentials.ClientID,
		"client_secret": credentials.ClientSecret,
	}
	
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	
	resp, err := p.httpClient.PostForm(ctx, authURL, payload, headers)
	if err != nil {
		return nil, providers.NewProviderError("AUTH_FAILED", "Falha na autenticação", err)
	}
	
	var authResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	
	if err := json.Unmarshal(resp, &authResp); err != nil {
		return nil, providers.NewProviderError("PARSE_ERROR", "Erro ao processar resposta", err)
	}
	
	return &providers.AuthToken{
		AccessToken: authResp.AccessToken,
		TokenType:   authResp.TokenType,
		ExpiresIn:   authResp.ExpiresIn,
		ExpiresAt:   time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second),
	}, nil
}

// RefreshToken renova o token
func (p *SantanderProvider) RefreshToken(ctx context.Context, refreshToken string) (*providers.AuthToken, error) {
	return nil, providers.NewProviderError("NOT_SUPPORTED", "Refresh token não suportado", nil)
}

// CreateTransfer cria uma transferência PIX
func (p *SantanderProvider) CreateTransfer(ctx context.Context, req *providers.TransferRequest) (*providers.TransferResponse, error) {
	url := fmt.Sprintf("%s/pix/v1/payments", p.config.BaseURL)
	
	payload := map[string]interface{}{
		"amount": map[string]interface{}{
			"value":    req.Amount,
			"currency": "BRL",
		},
		"payee": map[string]interface{}{
			"name":     req.PayeeName,
			"document": req.PayeeDocument,
			"pixKey":   req.PayeePixKey,
		},
		"description":  req.Description,
		"externalId":   req.ExternalID,
	}
	
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", req.AuthToken),
		"X-Application-Key": req.ClientID,
	}
	
	resp, err := p.httpClient.Post(ctx, url, payload, headers)
	if err != nil {
		return nil, providers.NewProviderError("TRANSFER_FAILED", "Falha ao criar transferência", err)
	}
	
	var santanderResp struct {
		TransactionId string `json:"transactionId"`
		EndToEndId    string `json:"endToEndId"`
		Status        string `json:"status"`
	}
	
	if err := json.Unmarshal(resp, &santanderResp); err != nil {
		return nil, providers.NewProviderError("PARSE_ERROR", "Erro ao processar resposta", err)
	}
	
	return &providers.TransferResponse{
		ProviderTxID: santanderResp.TransactionId,
		E2EID:        santanderResp.EndToEndId,
		Status:       mapStatus(santanderResp.Status),
		ProcessedAt:  timePtr(time.Now()),
	}, nil
}

// GetTransfer consulta uma transferência
func (p *SantanderProvider) GetTransfer(ctx context.Context, req *providers.GetTransferRequest) (*providers.TransferResponse, error) {
	url := fmt.Sprintf("%s/pix/v1/payments/%s", p.config.BaseURL, req.ProviderTxID)
	
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", req.AuthToken),
		"X-Application-Key": req.ClientID,
	}
	
	resp, err := p.httpClient.Get(ctx, url, headers)
	if err != nil {
		return nil, providers.NewProviderError("GET_FAILED", "Falha ao consultar transferência", err)
	}
	
	var santanderResp struct {
		TransactionId string `json:"transactionId"`
		EndToEndId    string `json:"endToEndId"`
		Status        string `json:"status"`
	}
	
	if err := json.Unmarshal(resp, &santanderResp); err != nil {
		return nil, providers.NewProviderError("PARSE_ERROR", "Erro ao processar resposta", err)
	}
	
	return &providers.TransferResponse{
		ProviderTxID: santanderResp.TransactionId,
		E2EID:        santanderResp.EndToEndId,
		Status:       mapStatus(santanderResp.Status),
	}, nil
}

// CancelTransfer cancela uma transferência
func (p *SantanderProvider) CancelTransfer(ctx context.Context, req *providers.CancelTransferRequest) error {
	return providers.NewProviderError("NOT_SUPPORTED", "Cancelamento não suportado", nil)
}

// CreateQRCodeStatic cria um QR Code estático
func (p *SantanderProvider) CreateQRCodeStatic(ctx context.Context, req *providers.QRCodeRequest) (*providers.QRCodeResponse, error) {
	url := fmt.Sprintf("%s/pix/v1/qrcodes/static", p.config.BaseURL)
	
	payload := map[string]interface{}{
		"amount": req.Amount,
		"pixKey": req.PixKey,
		"description": req.Description,
	}
	
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", req.AuthToken),
	}
	
	resp, err := p.httpClient.Post(ctx, url, payload, headers)
	if err != nil {
		return nil, providers.NewProviderError("QRCODE_FAILED", "Falha ao gerar QR Code", err)
	}
	
	var qrResp struct {
		QRCodeId string `json:"qrcodeId"`
		QRCode   string `json:"qrcode"`
		Image    string `json:"image"`
	}
	
	if err := json.Unmarshal(resp, &qrResp); err != nil {
		return nil, providers.NewProviderError("PARSE_ERROR", "Erro ao processar resposta", err)
	}
	
	return &providers.QRCodeResponse{
		QRCodeID:    qrResp.QRCodeId,
		QRCode:      qrResp.QRCode,
		QRCodeImage: qrResp.Image,
	}, nil
}

// CreateQRCodeDynamic cria um QR Code dinâmico
func (p *SantanderProvider) CreateQRCodeDynamic(ctx context.Context, req *providers.QRCodeRequest) (*providers.QRCodeResponse, error) {
	return p.CreateQRCodeStatic(ctx, req)
}

// GetQRCode consulta um QR Code
func (p *SantanderProvider) GetQRCode(ctx context.Context, req *providers.GetQRCodeRequest) (*providers.QRCodeResponse, error) {
	url := fmt.Sprintf("%s/pix/v1/qrcodes/%s", p.config.BaseURL, req.QRCodeID)
	
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", req.AuthToken),
	}
	
	resp, err := p.httpClient.Get(ctx, url, headers)
	if err != nil {
		return nil, providers.NewProviderError("GET_FAILED", "Falha ao consultar QR Code", err)
	}
	
	var qrResp struct {
		QRCodeId string `json:"qrcodeId"`
		QRCode   string `json:"qrcode"`
		Image    string `json:"image"`
		Status   string `json:"status"`
	}
	
	if err := json.Unmarshal(resp, &qrResp); err != nil {
		return nil, providers.NewProviderError("PARSE_ERROR", "Erro ao processar resposta", err)
	}
	
	return &providers.QRCodeResponse{
		QRCodeID:    qrResp.QRCodeId,
		QRCode:      qrResp.QRCode,
		QRCodeImage: qrResp.Image,
		Status:      qrResp.Status,
	}, nil
}

// ValidatePixKey valida uma chave PIX
func (p *SantanderProvider) ValidatePixKey(ctx context.Context, req *providers.ValidatePixKeyRequest) (*providers.ValidatePixKeyResponse, error) {
	return nil, providers.NewProviderError("NOT_IMPLEMENTED", "Validação de chave não implementada", nil)
}

// HealthCheck verifica a saúde do provider
func (p *SantanderProvider) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", p.config.BaseURL)
	
	_, err := p.httpClient.Get(ctx, url, nil)
	if err != nil {
		return providers.NewProviderError("HEALTH_CHECK_FAILED", "Provider indisponível", err)
	}
	
	return nil
}

// GetSupportedMethods retorna os métodos suportados
func (p *SantanderProvider) GetSupportedMethods() []string {
	return []string{
		"transfer",
		"qrcode_static",
		"qrcode_dynamic",
		"get_transfer",
		"get_qrcode",
	}
}

// Funções auxiliares
func mapStatus(status string) providers.TransactionStatus {
	switch status {
	case "COMPLETED", "SETTLED":
		return providers.TransactionStatusCompleted
	case "PROCESSING", "PENDING":
		return providers.TransactionStatusProcessing
	case "CANCELLED", "REJECTED":
		return providers.TransactionStatusCancelled
	case "FAILED":
		return providers.TransactionStatusFailed
	default:
		return providers.TransactionStatusPending
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
