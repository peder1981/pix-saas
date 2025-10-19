package bb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pixsaas/backend/internal/providers"
)

// BBProvider implementa o provider do Banco do Brasil
type BBProvider struct {
	config     providers.ProviderConfig
	httpClient providers.HTTPClient
}

// NewBBProvider cria uma nova instância do provider Banco do Brasil
func NewBBProvider() *BBProvider {
	return &BBProvider{}
}

// GetCode retorna o código do provider
func (p *BBProvider) GetCode() string {
	return "banco_do_brasil"
}

// GetName retorna o nome do provider
func (p *BBProvider) GetName() string {
	return "Banco do Brasil"
}

// Initialize inicializa o provider com configurações
func (p *BBProvider) Initialize(config providers.ProviderConfig) error {
	p.config = config
	p.httpClient = providers.NewHTTPClient(config.Timeout, config.RequiresMTLS)
	return nil
}

// Authenticate realiza autenticação OAuth2 com o Banco do Brasil
func (p *BBProvider) Authenticate(ctx context.Context, credentials providers.ProviderCredentials) (*providers.AuthToken, error) {
	authURL := p.config.AuthURL
	
	payload := map[string]string{
		"grant_type": "client_credentials",
		"scope":      "cob.write cob.read pix.write pix.read",
	}
	
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	
	// Basic Auth com client_id e client_secret
	resp, err := p.httpClient.PostFormWithBasicAuth(
		ctx,
		authURL,
		payload,
		headers,
		credentials.ClientID,
		credentials.ClientSecret,
	)
	if err != nil {
		return nil, providers.NewProviderError("AUTH_FAILED", "Falha na autenticação", err)
	}
	
	var authResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	
	if err := json.Unmarshal(resp, &authResp); err != nil {
		return nil, providers.NewProviderError("PARSE_ERROR", "Erro ao processar resposta de autenticação", err)
	}
	
	return &providers.AuthToken{
		AccessToken: authResp.AccessToken,
		TokenType:   authResp.TokenType,
		ExpiresIn:   authResp.ExpiresIn,
		ExpiresAt:   time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second),
	}, nil
}

// RefreshToken renova o token de acesso
func (p *BBProvider) RefreshToken(ctx context.Context, refreshToken string) (*providers.AuthToken, error) {
	return nil, providers.NewProviderError("NOT_SUPPORTED", "Refresh token não suportado pelo BB", nil)
}

// CreateTransfer cria uma transferência PIX
func (p *BBProvider) CreateTransfer(ctx context.Context, req *providers.TransferRequest) (*providers.TransferResponse, error) {
	url := fmt.Sprintf("%s/pix/v1/pix", p.config.BaseURL)
	
	payload := map[string]interface{}{
		"valor": fmt.Sprintf("%.2f", float64(req.Amount)/100),
		"chave": req.PayeePixKey,
		"descricao": req.Description,
		"txid": req.ExternalID,
	}
	
	// Se não tiver chave PIX, usar dados bancários
	if req.PayeePixKey == "" {
		payload["favorecido"] = map[string]interface{}{
			"nome":      req.PayeeName,
			"cpfCnpj":   req.PayeeDocument,
			"banco":     req.PayeeISPB,
			"agencia":   req.PayeeAccountAgency,
			"conta":     req.PayeeAccountNumber,
			"tipoConta": mapAccountType(req.PayeeAccountType),
		}
	}
	
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", req.AuthToken),
	}
	
	resp, err := p.httpClient.Post(ctx, url, payload, headers)
	if err != nil {
		return nil, providers.NewProviderError("TRANSFER_FAILED", "Falha ao criar transferência", err)
	}
	
	var bbResp struct {
		EndToEndId string `json:"endToEndId"`
		TxId       string `json:"txid"`
		Status     string `json:"status"`
		Valor      string `json:"valor"`
	}
	
	if err := json.Unmarshal(resp, &bbResp); err != nil {
		return nil, providers.NewProviderError("PARSE_ERROR", "Erro ao processar resposta", err)
	}
	
	return &providers.TransferResponse{
		ProviderTxID: bbResp.TxId,
		E2EID:        bbResp.EndToEndId,
		Status:       mapStatus(bbResp.Status),
		ProcessedAt:  timePtr(time.Now()),
	}, nil
}

// GetTransfer consulta uma transferência
func (p *BBProvider) GetTransfer(ctx context.Context, req *providers.GetTransferRequest) (*providers.TransferResponse, error) {
	url := fmt.Sprintf("%s/pix/v1/pix/%s", p.config.BaseURL, req.ProviderTxID)
	
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", req.AuthToken),
	}
	
	resp, err := p.httpClient.Get(ctx, url, headers)
	if err != nil {
		return nil, providers.NewProviderError("GET_FAILED", "Falha ao consultar transferência", err)
	}
	
	var bbResp struct {
		EndToEndId string `json:"endToEndId"`
		TxId       string `json:"txid"`
		Status     string `json:"status"`
		Valor      string `json:"valor"`
	}
	
	if err := json.Unmarshal(resp, &bbResp); err != nil {
		return nil, providers.NewProviderError("PARSE_ERROR", "Erro ao processar resposta", err)
	}
	
	return &providers.TransferResponse{
		ProviderTxID: bbResp.TxId,
		E2EID:        bbResp.EndToEndId,
		Status:       mapStatus(bbResp.Status),
	}, nil
}

// CancelTransfer cancela uma transferência
func (p *BBProvider) CancelTransfer(ctx context.Context, req *providers.CancelTransferRequest) error {
	return providers.NewProviderError("NOT_SUPPORTED", "Cancelamento não suportado pelo BB", nil)
}

// CreateQRCodeStatic cria um QR Code estático
func (p *BBProvider) CreateQRCodeStatic(ctx context.Context, req *providers.QRCodeRequest) (*providers.QRCodeResponse, error) {
	url := fmt.Sprintf("%s/pix/v1/cobqrcode", p.config.BaseURL)
	
	payload := map[string]interface{}{
		"valor": map[string]interface{}{
			"original": fmt.Sprintf("%.2f", float64(req.Amount)/100),
		},
		"chave": req.PixKey,
		"solicitacaoPagador": req.Description,
	}
	
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", req.AuthToken),
	}
	
	resp, err := p.httpClient.Post(ctx, url, payload, headers)
	if err != nil {
		return nil, providers.NewProviderError("QRCODE_FAILED", "Falha ao gerar QR Code", err)
	}
	
	var bbResp struct {
		TxId      string `json:"txid"`
		QRCode    string `json:"qrcode"`
		ImagemQRCode string `json:"imagemQrcode"`
	}
	
	if err := json.Unmarshal(resp, &bbResp); err != nil {
		return nil, providers.NewProviderError("PARSE_ERROR", "Erro ao processar resposta", err)
	}
	
	return &providers.QRCodeResponse{
		QRCodeID:    bbResp.TxId,
		QRCode:      bbResp.QRCode,
		QRCodeImage: bbResp.ImagemQRCode,
	}, nil
}

// CreateQRCodeDynamic cria um QR Code dinâmico
func (p *BBProvider) CreateQRCodeDynamic(ctx context.Context, req *providers.QRCodeRequest) (*providers.QRCodeResponse, error) {
	return p.CreateQRCodeStatic(ctx, req)
}

// GetQRCode consulta um QR Code
func (p *BBProvider) GetQRCode(ctx context.Context, req *providers.GetQRCodeRequest) (*providers.QRCodeResponse, error) {
	url := fmt.Sprintf("%s/pix/v1/cobqrcode/%s", p.config.BaseURL, req.QRCodeID)
	
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", req.AuthToken),
	}
	
	resp, err := p.httpClient.Get(ctx, url, headers)
	if err != nil {
		return nil, providers.NewProviderError("GET_FAILED", "Falha ao consultar QR Code", err)
	}
	
	var bbResp struct {
		TxId      string `json:"txid"`
		QRCode    string `json:"qrcode"`
		ImagemQRCode string `json:"imagemQrcode"`
		Status    string `json:"status"`
	}
	
	if err := json.Unmarshal(resp, &bbResp); err != nil {
		return nil, providers.NewProviderError("PARSE_ERROR", "Erro ao processar resposta", err)
	}
	
	return &providers.QRCodeResponse{
		QRCodeID:    bbResp.TxId,
		QRCode:      bbResp.QRCode,
		QRCodeImage: bbResp.ImagemQRCode,
		Status:      bbResp.Status,
	}, nil
}

// ValidatePixKey valida uma chave PIX
func (p *BBProvider) ValidatePixKey(ctx context.Context, req *providers.ValidatePixKeyRequest) (*providers.ValidatePixKeyResponse, error) {
	return nil, providers.NewProviderError("NOT_IMPLEMENTED", "Validação de chave não implementada", nil)
}

// HealthCheck verifica a saúde do provider
func (p *BBProvider) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/pix/v1/health", p.config.BaseURL)
	
	_, err := p.httpClient.Get(ctx, url, nil)
	if err != nil {
		return providers.NewProviderError("HEALTH_CHECK_FAILED", "Provider indisponível", err)
	}
	
	return nil
}

// GetSupportedMethods retorna os métodos suportados
func (p *BBProvider) GetSupportedMethods() []string {
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
	case "ATIVA", "CONCLUIDA":
		return providers.TransactionStatusCompleted
	case "PENDENTE", "EM_PROCESSAMENTO":
		return providers.TransactionStatusProcessing
	case "REMOVIDA_PELO_USUARIO_RECEBEDOR", "REMOVIDA_PELO_PSP":
		return providers.TransactionStatusCancelled
	default:
		return providers.TransactionStatusPending
	}
}

func mapAccountType(accountType string) string {
	switch accountType {
	case "checking":
		return "CORRENTE"
	case "savings":
		return "POUPANCA"
	default:
		return "CORRENTE"
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
