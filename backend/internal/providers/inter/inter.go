package inter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pixsaas/backend/internal/providers"
)

// InterProvider implementa o provider do Banco Inter
type InterProvider struct {
	config     providers.ProviderConfig
	httpClient providers.HTTPClient
}

// NewInterProvider cria uma nova instância do provider Inter
func NewInterProvider() *InterProvider {
	return &InterProvider{}
}

// GetCode retorna o código do provider
func (p *InterProvider) GetCode() string {
	return "inter"
}

// GetName retorna o nome do provider
func (p *InterProvider) GetName() string {
	return "Banco Inter"
}

// Initialize inicializa o provider
func (p *InterProvider) Initialize(config providers.ProviderConfig) error {
	p.config = config
	p.httpClient = providers.NewHTTPClient(config.Timeout, config.RequiresMTLS)
	return nil
}

// Authenticate realiza autenticação OAuth2
func (p *InterProvider) Authenticate(ctx context.Context, credentials providers.ProviderCredentials) (*providers.AuthToken, error) {
	authURL := p.config.AuthURL

	payload := map[string]string{
		"client_id":     credentials.ClientID,
		"client_secret": credentials.ClientSecret,
		"scope":         "extrato.read boleto-cobranca.read boleto-cobranca.write pagamento-pix.write pagamento-pix.read",
		"grant_type":    "client_credentials",
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
func (p *InterProvider) RefreshToken(ctx context.Context, refreshToken string) (*providers.AuthToken, error) {
	return nil, providers.NewProviderError("NOT_SUPPORTED", "Refresh token não suportado", nil)
}

// CreateTransfer cria uma transferência PIX
func (p *InterProvider) CreateTransfer(ctx context.Context, req *providers.TransferRequest) (*providers.TransferResponse, error) {
	url := fmt.Sprintf("%s/banking/v2/pix", p.config.BaseURL)

	payload := map[string]interface{}{
		"valor": float64(req.Amount) / 100,
		"destinatario": map[string]interface{}{
			"nome":    req.PayeeName,
			"cpfCnpj": req.PayeeDocument,
			"chave":   req.PayeePixKey,
		},
		"descricao": req.Description,
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", req.AuthToken),
	}

	resp, err := p.httpClient.Post(ctx, url, payload, headers)
	if err != nil {
		return nil, providers.NewProviderError("TRANSFER_FAILED", "Falha ao criar transferência", err)
	}

	var interResp struct {
		CodigoSolicitacao string `json:"codigoSolicitacao"`
		EndToEndId        string `json:"endToEndId"`
		Status            string `json:"status"`
	}

	if err := json.Unmarshal(resp, &interResp); err != nil {
		return nil, providers.NewProviderError("PARSE_ERROR", "Erro ao processar resposta", err)
	}

	return &providers.TransferResponse{
		ProviderTxID: interResp.CodigoSolicitacao,
		E2EID:        interResp.EndToEndId,
		Status:       mapStatus(interResp.Status),
		ProcessedAt:  timePtr(time.Now()),
	}, nil
}

// GetTransfer consulta uma transferência
func (p *InterProvider) GetTransfer(ctx context.Context, req *providers.GetTransferRequest) (*providers.TransferResponse, error) {
	url := fmt.Sprintf("%s/banking/v2/pix/%s", p.config.BaseURL, req.ProviderTxID)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", req.AuthToken),
	}

	resp, err := p.httpClient.Get(ctx, url, headers)
	if err != nil {
		return nil, providers.NewProviderError("GET_FAILED", "Falha ao consultar transferência", err)
	}

	var interResp struct {
		CodigoSolicitacao string `json:"codigoSolicitacao"`
		EndToEndId        string `json:"endToEndId"`
		Status            string `json:"status"`
	}

	if err := json.Unmarshal(resp, &interResp); err != nil {
		return nil, providers.NewProviderError("PARSE_ERROR", "Erro ao processar resposta", err)
	}

	return &providers.TransferResponse{
		ProviderTxID: interResp.CodigoSolicitacao,
		E2EID:        interResp.EndToEndId,
		Status:       mapStatus(interResp.Status),
	}, nil
}

// CancelTransfer cancela uma transferência
func (p *InterProvider) CancelTransfer(ctx context.Context, req *providers.CancelTransferRequest) error {
	return providers.NewProviderError("NOT_SUPPORTED", "Cancelamento não suportado", nil)
}

// CreateQRCodeStatic cria um QR Code estático
func (p *InterProvider) CreateQRCodeStatic(ctx context.Context, req *providers.QRCodeRequest) (*providers.QRCodeResponse, error) {
	url := fmt.Sprintf("%s/banking/v2/pix/qrcode-estatico", p.config.BaseURL)

	payload := map[string]interface{}{
		"valor":              float64(req.Amount) / 100,
		"chave":              req.PixKey,
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

	var qrResp struct {
		TxId          string `json:"txid"`
		PixCopiaECola string `json:"pixCopiaECola"`
	}

	if err := json.Unmarshal(resp, &qrResp); err != nil {
		return nil, providers.NewProviderError("PARSE_ERROR", "Erro ao processar resposta", err)
	}

	return &providers.QRCodeResponse{
		QRCodeID: qrResp.TxId,
		QRCode:   qrResp.PixCopiaECola,
	}, nil
}

// CreateQRCodeDynamic cria um QR Code dinâmico
func (p *InterProvider) CreateQRCodeDynamic(ctx context.Context, req *providers.QRCodeRequest) (*providers.QRCodeResponse, error) {
	url := fmt.Sprintf("%s/banking/v2/pix/qrcode-dinamico", p.config.BaseURL)

	payload := map[string]interface{}{
		"valor": map[string]interface{}{
			"original": float64(req.Amount) / 100,
		},
		"chave":              req.PixKey,
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

	var qrResp struct {
		TxId          string `json:"txid"`
		PixCopiaECola string `json:"pixCopiaECola"`
	}

	if err := json.Unmarshal(resp, &qrResp); err != nil {
		return nil, providers.NewProviderError("PARSE_ERROR", "Erro ao processar resposta", err)
	}

	return &providers.QRCodeResponse{
		QRCodeID: qrResp.TxId,
		QRCode:   qrResp.PixCopiaECola,
	}, nil
}

// GetQRCode consulta um QR Code
func (p *InterProvider) GetQRCode(ctx context.Context, req *providers.GetQRCodeRequest) (*providers.QRCodeResponse, error) {
	url := fmt.Sprintf("%s/banking/v2/pix/qrcode/%s", p.config.BaseURL, req.QRCodeID)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", req.AuthToken),
	}

	resp, err := p.httpClient.Get(ctx, url, headers)
	if err != nil {
		return nil, providers.NewProviderError("GET_FAILED", "Falha ao consultar QR Code", err)
	}

	var qrResp struct {
		TxId          string `json:"txid"`
		PixCopiaECola string `json:"pixCopiaECola"`
		Status        string `json:"status"`
	}

	if err := json.Unmarshal(resp, &qrResp); err != nil {
		return nil, providers.NewProviderError("PARSE_ERROR", "Erro ao processar resposta", err)
	}

	return &providers.QRCodeResponse{
		QRCodeID: qrResp.TxId,
		QRCode:   qrResp.PixCopiaECola,
		Status:   qrResp.Status,
	}, nil
}

// ValidatePixKey valida uma chave PIX
func (p *InterProvider) ValidatePixKey(ctx context.Context, req *providers.ValidatePixKeyRequest) (*providers.ValidatePixKeyResponse, error) {
	return nil, providers.NewProviderError("NOT_IMPLEMENTED", "Validação de chave não implementada", nil)
}

// HealthCheck verifica a saúde do provider
func (p *InterProvider) HealthCheck(ctx context.Context) error {
	// Inter não tem endpoint de health check público
	return nil
}

// GetSupportedMethods retorna os métodos suportados
func (p *InterProvider) GetSupportedMethods() []string {
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
	case "REALIZADO", "CONCLUIDO":
		return providers.TransactionStatusCompleted
	case "EM_PROCESSAMENTO", "PENDENTE":
		return providers.TransactionStatusProcessing
	case "CANCELADO", "DEVOLVIDO":
		return providers.TransactionStatusCancelled
	case "ERRO", "REJEITADO":
		return providers.TransactionStatusFailed
	default:
		return providers.TransactionStatusPending
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
