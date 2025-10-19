package bradesco

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pixsaas/backend/internal/domain"
	"github.com/pixsaas/backend/internal/providers"
)

const (
	ProviderCode = "bradesco"
	ProviderName = "Bradesco"
	ISPB         = "60746948"
)

// BradescoProvider implementa a interface PixProvider para o Bradesco
type BradescoProvider struct {
	config     domain.ProviderConfig
	httpClient *http.Client
	baseURL    string
	authURL    string
}

// NewBradescoProvider cria uma nova instância do provider Bradesco
func NewBradescoProvider() *BradescoProvider {
	return &BradescoProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *BradescoProvider) GetCode() string {
	return ProviderCode
}

func (p *BradescoProvider) GetName() string {
	return ProviderName
}

func (p *BradescoProvider) Initialize(config domain.ProviderConfig) error {
	p.config = config
	p.baseURL = config.BaseURL
	p.authURL = config.AuthURL

	if config.Timeout > 0 {
		p.httpClient.Timeout = time.Duration(config.Timeout) * time.Second
	}

	return nil
}

func (p *BradescoProvider) Authenticate(ctx context.Context, credentials providers.ProviderCredentials) (*providers.AuthToken, error) {
	// Bradesco usa OAuth 2.0 com mTLS
	if p.config.RequiresMTLS {
		cert, err := tls.X509KeyPair(credentials.Certificate, credentials.PrivateKey)
		if err != nil {
			return nil, &providers.ProviderError{
				Code:    "CERT_ERROR",
				Message: "Erro ao carregar certificado mTLS",
				Details: map[string]interface{}{"error": err.Error()},
			}
		}

		p.httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
				MinVersion:   tls.VersionTLS12,
			},
		}
	}

	// Requisição OAuth2
	data := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     credentials.ClientID,
		"client_secret": credentials.ClientSecret,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, &providers.ProviderError{
			Code:    "MARSHAL_ERROR",
			Message: "Erro ao serializar dados",
			Details: map[string]interface{}{"error": err.Error()},
		}
	}
	req, err := http.NewRequestWithContext(ctx, "POST", p.authURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, &providers.ProviderError{
			Code:      "AUTH_ERROR",
			Message:   "Erro ao autenticar com Bradesco",
			Retryable: true,
			Details:   map[string]interface{}{"error": err.Error()},
		}
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("failed to read response body")
		}
		return nil, &providers.ProviderError{
			Code:       "AUTH_FAILED",
			Message:    "Autenticação falhou",
			StatusCode: resp.StatusCode,
			Details:    map[string]interface{}{"response": string(body)},
		}
	}

	var authResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, err
	}

	return &providers.AuthToken{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		TokenType:    authResp.TokenType,
		ExpiresIn:    authResp.ExpiresIn,
		ExpiresAt:    time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second),
	}, nil
}

func (p *BradescoProvider) RefreshToken(ctx context.Context, refreshToken string) (*providers.AuthToken, error) {
	// Implementar refresh token
	return nil, fmt.Errorf("refresh token não implementado")
}

func (p *BradescoProvider) CreateTransfer(ctx context.Context, req *providers.TransferRequest) (*providers.TransferResponse, error) {
	// Endpoint: POST /v1/spi/solicitar-transferencia
	endpoint := fmt.Sprintf("%s/v1/spi/solicitar-transferencia", p.baseURL)

	// Montar payload conforme documentação Bradesco
	payload := map[string]interface{}{
		"idTransacao": req.ExternalID,
		"valor":       float64(req.Amount) / 100.0, // Converter centavos para reais
		"descricao":   req.Description,
	}

	// Dados do pagador
	pagador := make(map[string]interface{})
	if req.PayerPixKey != "" {
		pagador["chavePix"] = req.PayerPixKey
	} else {
		pagador["banco"] = req.PayerBank
		pagador["agencia"] = req.PayerAccountAgency
		pagador["conta"] = req.PayerAccountNumber
		pagador["tipoConta"] = req.PayerAccountType
	}
	if req.PayerDocument != "" {
		pagador["cpfCnpj"] = req.PayerDocument
	}
	payload["pagador"] = pagador

	// Dados do recebedor
	recebedor := make(map[string]interface{})
	if req.PayeePixKey != "" {
		recebedor["chavePix"] = req.PayeePixKey
	} else {
		recebedor["banco"] = req.PayeeBank
		recebedor["agencia"] = req.PayeeAccountAgency
		recebedor["conta"] = req.PayeeAccountNumber
		recebedor["tipoConta"] = req.PayeeAccountType
	}
	if req.PayeeDocument != "" {
		recebedor["cpfCnpj"] = req.PayeeDocument
	}
	payload["recebedor"] = recebedor

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, &providers.ProviderError{
			Code:    "MARSHAL_ERROR",
			Message: "Erro ao serializar payload",
			Details: map[string]interface{}{"error": err.Error()},
		}
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	// Token deve ser passado via context ou armazenado

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, &providers.ProviderError{
			Code:      "TRANSFER_ERROR",
			Message:   "Erro ao criar transferência",
			Retryable: true,
			Details:   map[string]interface{}{"error": err.Error()},
		}
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		body = []byte("failed to read response body")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, &providers.ProviderError{
			Code:       "TRANSFER_FAILED",
			Message:    "Transferência falhou",
			StatusCode: resp.StatusCode,
			Details:    map[string]interface{}{"response": string(body)},
		}
	}

	var bradescoResp struct {
		IDTransacao string  `json:"idTransacao"`
		EndToEndID  string  `json:"endToEndId"`
		Status      string  `json:"status"`
		Valor       float64 `json:"valor"`
		DataHora    string  `json:"dataHora"`
		Motivo      string  `json:"motivo,omitempty"`
	}

	if err := json.Unmarshal(body, &bradescoResp); err != nil {
		return nil, err
	}

	// Mapear status do Bradesco para nosso status
	status := mapBradescoStatus(bradescoResp.Status)

	response := &providers.TransferResponse{
		ProviderTxID: bradescoResp.IDTransacao,
		E2EID:        bradescoResp.EndToEndID,
		Status:       status,
		Amount:       req.Amount,
		Description:  req.Description,
		RawResponse:  map[string]interface{}{"bradesco": bradescoResp},
	}

	if bradescoResp.Status == "REJEITADA" || bradescoResp.Status == "ERRO" {
		response.ErrorMessage = bradescoResp.Motivo
	}

	if bradescoResp.Status == "CONCLUIDA" {
		now := time.Now()
		response.CompletedAt = &now
	}

	return response, nil
}

func (p *BradescoProvider) GetTransfer(ctx context.Context, txID string) (*providers.TransferResponse, error) {
	// Endpoint: GET /v1/spi/consultar-transferencia/{idTransacao}
	endpoint := fmt.Sprintf("%s/v1/spi/consultar-transferencia/%s", p.baseURL, txID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, &providers.ProviderError{
			Code:      "QUERY_ERROR",
			Message:   "Erro ao consultar transferência",
			Retryable: true,
			Details:   map[string]interface{}{"error": err.Error()},
		}
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		body = []byte("failed to read response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &providers.ProviderError{
			Code:       "QUERY_FAILED",
			Message:    "Consulta falhou",
			StatusCode: resp.StatusCode,
			Details:    map[string]interface{}{"response": string(body)},
		}
	}

	var bradescoResp struct {
		IDTransacao string  `json:"idTransacao"`
		EndToEndID  string  `json:"endToEndId"`
		Status      string  `json:"status"`
		Valor       float64 `json:"valor"`
		DataHora    string  `json:"dataHora"`
	}

	if err := json.Unmarshal(body, &bradescoResp); err != nil {
		return nil, err
	}

	return &providers.TransferResponse{
		ProviderTxID: bradescoResp.IDTransacao,
		E2EID:        bradescoResp.EndToEndID,
		Status:       mapBradescoStatus(bradescoResp.Status),
		Amount:       int64(bradescoResp.Valor * 100),
		RawResponse:  map[string]interface{}{"bradesco": bradescoResp},
	}, nil
}

func (p *BradescoProvider) CancelTransfer(ctx context.Context, txID string) error {
	return fmt.Errorf("cancelamento não suportado pelo Bradesco")
}

func (p *BradescoProvider) CreateQRCodeStatic(ctx context.Context, req *providers.QRCodeRequest) (*providers.QRCodeResponse, error) {
	// TODO: Implementar criação de QR Code estático
	return nil, fmt.Errorf("QR Code estático não implementado")
}

func (p *BradescoProvider) CreateQRCodeDynamic(ctx context.Context, req *providers.QRCodeRequest) (*providers.QRCodeResponse, error) {
	// TODO: Implementar criação de QR Code dinâmico
	return nil, fmt.Errorf("QR Code dinâmico não implementado")
}

func (p *BradescoProvider) GetQRCode(ctx context.Context, qrCodeID string) (*providers.QRCodeResponse, error) {
	return nil, fmt.Errorf("consulta de QR Code não implementada")
}

func (p *BradescoProvider) ValidatePixKey(ctx context.Context, pixKey string, pixKeyType domain.PixKeyType) (*providers.PixKeyInfo, error) {
	// TODO: Implementar validação de chave PIX
	return nil, fmt.Errorf("validação de chave PIX não implementada")
}

func (p *BradescoProvider) HealthCheck(ctx context.Context) error {
	// Fazer uma requisição simples para verificar conectividade
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL, http.NoBody)
	if err != nil {
		return err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck

	if resp.StatusCode >= 500 {
		return fmt.Errorf("provider unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

func (p *BradescoProvider) GetSupportedMethods() []string {
	return []string{"pix_key", "account", "transfer"}
}

// mapBradescoStatus mapeia status do Bradesco para nosso domínio
func mapBradescoStatus(bradescoStatus string) domain.TransactionStatus {
	switch bradescoStatus {
	case "EM_PROCESSAMENTO":
		return domain.TransactionStatusProcessing
	case "CONCLUIDA":
		return domain.TransactionStatusCompleted
	case "REJEITADA", "ERRO":
		return domain.TransactionStatusFailed
	case "CANCELADA":
		return domain.TransactionStatusCancelled
	default:
		return domain.TransactionStatusPending
	}
}
