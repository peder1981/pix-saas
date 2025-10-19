package itau

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
	ProviderCode = "itau"
	ProviderName = "Itaú Unibanco"
	ISPB         = "60701190"
)

// ItauProvider implementa a interface PixProvider para o Itaú
type ItauProvider struct {
	config     domain.ProviderConfig
	httpClient *http.Client
	baseURL    string
	authURL    string
}

// NewItauProvider cria uma nova instância do provider Itaú
func NewItauProvider() *ItauProvider {
	return &ItauProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *ItauProvider) GetCode() string {
	return ProviderCode
}

func (p *ItauProvider) GetName() string {
	return ProviderName
}

func (p *ItauProvider) Initialize(config domain.ProviderConfig) error {
	p.config = config
	p.baseURL = config.BaseURL
	p.authURL = config.AuthURL

	if config.Timeout > 0 {
		p.httpClient.Timeout = time.Duration(config.Timeout) * time.Second
	}

	return nil
}

func (p *ItauProvider) Authenticate(ctx context.Context, credentials providers.ProviderCredentials) (*providers.AuthToken, error) {
	// Itaú usa OAuth 2.0 com Private Key JWT e mTLS
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

	// Requisição OAuth2 com client_credentials
	data := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     credentials.ClientID,
		"client_secret": credentials.ClientSecret,
		"scope":         "sispag",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, &providers.ProviderError{
			Code:    "MARSHAL_ERROR",
			Message: "Erro ao serializar dados de autenticação",
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
			Message:   "Erro ao autenticar com Itaú",
			Retryable: true,
			Details:   map[string]interface{}{"error": err.Error()},
		}
	}
	defer resp.Body.Close()

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
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, err
	}

	return &providers.AuthToken{
		AccessToken: authResp.AccessToken,
		TokenType:   authResp.TokenType,
		ExpiresIn:   authResp.ExpiresIn,
		ExpiresAt:   time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second),
		Scope:       authResp.Scope,
	}, nil
}

func (p *ItauProvider) RefreshToken(ctx context.Context, refreshToken string) (*providers.AuthToken, error) {
	return nil, fmt.Errorf("refresh token não implementado")
}

func (p *ItauProvider) CreateTransfer(ctx context.Context, req *providers.TransferRequest) (*providers.TransferResponse, error) {
	// Endpoint: POST /sispag/v1/pagamentos/pix
	endpoint := fmt.Sprintf("%s/sispag/v1/pagamentos/pix", p.baseURL)

	// Montar payload conforme documentação Itaú
	payload := map[string]interface{}{
		"id_requisicao": req.ExternalID,
		"valor":         float64(req.Amount) / 100.0,
		"descricao":     req.Description,
	}

	// Dados do pagador
	if req.PayerPixKey != "" {
		payload["chave_pagador"] = req.PayerPixKey
	} else {
		payload["conta_pagador"] = map[string]interface{}{
			"agencia": req.PayerAccountAgency,
			"conta":   req.PayerAccountNumber,
			"tipo":    req.PayerAccountType,
		}
	}

	// Dados do recebedor
	if req.PayeePixKey != "" {
		payload["chave_recebedor"] = req.PayeePixKey
		payload["tipo_chave"] = mapPixKeyTypeToItau(req.PayeePixKeyType)
	} else {
		payload["conta_recebedor"] = map[string]interface{}{
			"ispb":    req.PayeeISPB,
			"agencia": req.PayeeAccountAgency,
			"conta":   req.PayeeAccountNumber,
			"tipo":    req.PayeeAccountType,
		}
	}

	if req.PayeeDocument != "" {
		payload["cpf_cnpj_recebedor"] = req.PayeeDocument
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, &providers.ProviderError{
			Code:    "MARSHAL_ERROR",
			Message: "Erro ao serializar payload de transferência",
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
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		body = []byte("failed to read response body")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResp struct {
			Codigo   string `json:"codigo"`
			Mensagem string `json:"mensagem"`
			Detalhes []struct {
				Campo    string `json:"campo"`
				Mensagem string `json:"mensagem"`
			} `json:"detalhes"`
		}
		json.Unmarshal(body, &errorResp)

		return nil, &providers.ProviderError{
			Code:       errorResp.Codigo,
			Message:    errorResp.Mensagem,
			StatusCode: resp.StatusCode,
			Details:    map[string]interface{}{"response": string(body)},
		}
	}

	var itauResp struct {
		IDRequisicao  string  `json:"id_requisicao"`
		EndToEndID    string  `json:"end_to_end_id"`
		Status        string  `json:"status"`
		Valor         float64 `json:"valor"`
		DataPagamento string  `json:"data_pagamento"`
		Recebedor     struct {
			Nome     string `json:"nome"`
			CpfCnpj  string `json:"cpf_cnpj"`
			ChavePix string `json:"chave_pix,omitempty"`
		} `json:"recebedor"`
	}

	if err := json.Unmarshal(body, &itauResp); err != nil {
		return nil, err
	}

	status := mapItauStatus(itauResp.Status)

	response := &providers.TransferResponse{
		ProviderTxID:  itauResp.IDRequisicao,
		E2EID:         itauResp.EndToEndID,
		Status:        status,
		Amount:        req.Amount,
		Description:   req.Description,
		PayeeName:     itauResp.Recebedor.Nome,
		PayeeDocument: itauResp.Recebedor.CpfCnpj,
		PayeePixKey:   itauResp.Recebedor.ChavePix,
		RawResponse:   map[string]interface{}{"itau": itauResp},
	}

	if status == domain.TransactionStatusCompleted {
		now := time.Now()
		response.CompletedAt = &now
	}

	return response, nil
}

func (p *ItauProvider) GetTransfer(ctx context.Context, txID string) (*providers.TransferResponse, error) {
	// Endpoint: GET /sispag/v1/pagamentos/pix/{id_requisicao}
	endpoint := fmt.Sprintf("%s/sispag/v1/pagamentos/pix/%s", p.baseURL, txID)

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
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		body = []byte("failed to read response body")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, &providers.ProviderError{
			Code:       "NOT_FOUND",
			Message:    "Transferência não encontrada",
			StatusCode: resp.StatusCode,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &providers.ProviderError{
			Code:       "QUERY_FAILED",
			Message:    "Consulta falhou",
			StatusCode: resp.StatusCode,
			Details:    map[string]interface{}{"response": string(body)},
		}
	}

	var itauResp struct {
		IDRequisicao  string  `json:"id_requisicao"`
		EndToEndID    string  `json:"end_to_end_id"`
		Status        string  `json:"status"`
		Valor         float64 `json:"valor"`
		DataPagamento string  `json:"data_pagamento"`
	}

	if err := json.Unmarshal(body, &itauResp); err != nil {
		return nil, err
	}

	return &providers.TransferResponse{
		ProviderTxID: itauResp.IDRequisicao,
		E2EID:        itauResp.EndToEndID,
		Status:       mapItauStatus(itauResp.Status),
		Amount:       int64(itauResp.Valor * 100),
		RawResponse:  map[string]interface{}{"itau": itauResp},
	}, nil
}

func (p *ItauProvider) CancelTransfer(ctx context.Context, txID string) error {
	return fmt.Errorf("cancelamento não suportado pelo Itaú")
}

func (p *ItauProvider) CreateQRCodeStatic(ctx context.Context, req *providers.QRCodeRequest) (*providers.QRCodeResponse, error) {
	// Endpoint: POST /sispag/v1/qrcodes/estatico
	endpoint := fmt.Sprintf("%s/sispag/v1/qrcodes/estatico", p.baseURL)

	payload := map[string]interface{}{
		"chave_pix": req.PayeePixKey,
		"valor":     float64(req.Amount) / 100.0,
		"descricao": req.Description,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, &providers.ProviderError{
			Code:    "MARSHAL_ERROR",
			Message: "Erro ao serializar payload de QR Code",
			Details: map[string]interface{}{"error": err.Error()},
		}
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, &providers.ProviderError{
			Code:      "QRCODE_ERROR",
			Message:   "Erro ao criar QR Code",
			Retryable: true,
			Details:   map[string]interface{}{"error": err.Error()},
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		body = []byte("failed to read response body")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, &providers.ProviderError{
			Code:       "QRCODE_FAILED",
			Message:    "Criação de QR Code falhou",
			StatusCode: resp.StatusCode,
			Details:    map[string]interface{}{"response": string(body)},
		}
	}

	var qrResp struct {
		IDQRCode    string `json:"id_qrcode"`
		QRCode      string `json:"qrcode"`
		QRCodeImage string `json:"qrcode_imagem"`
	}

	if err := json.Unmarshal(body, &qrResp); err != nil {
		return nil, err
	}

	return &providers.QRCodeResponse{
		QRCodeID:    qrResp.IDQRCode,
		QRCode:      qrResp.QRCode,
		QRCodeImage: qrResp.QRCodeImage,
		Amount:      req.Amount,
		Description: req.Description,
		Status:      "active",
		CreatedAt:   time.Now(),
		RawResponse: map[string]interface{}{"itau": qrResp},
	}, nil
}

func (p *ItauProvider) CreateQRCodeDynamic(ctx context.Context, req *providers.QRCodeRequest) (*providers.QRCodeResponse, error) {
	// Endpoint: POST /sispag/v1/qrcodes/dinamico
	endpoint := fmt.Sprintf("%s/sispag/v1/qrcodes/dinamico", p.baseURL)

	payload := map[string]interface{}{
		"chave_pix":       req.PayeePixKey,
		"valor":           float64(req.Amount) / 100.0,
		"descricao":       req.Description,
		"expiracao":       req.ExpiresIn,
		"permite_alterar": req.AllowChange,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, &providers.ProviderError{
			Code:    "MARSHAL_ERROR",
			Message: "Erro ao serializar payload de QR Code dinâmico",
			Details: map[string]interface{}{"error": err.Error()},
		}
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, &providers.ProviderError{
			Code:      "QRCODE_ERROR",
			Message:   "Erro ao criar QR Code dinâmico",
			Retryable: true,
			Details:   map[string]interface{}{"error": err.Error()},
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		body = []byte("failed to read response body")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, &providers.ProviderError{
			Code:       "QRCODE_FAILED",
			Message:    "Criação de QR Code dinâmico falhou",
			StatusCode: resp.StatusCode,
			Details:    map[string]interface{}{"response": string(body)},
		}
	}

	var qrResp struct {
		IDQRCode    string `json:"id_qrcode"`
		QRCode      string `json:"qrcode"`
		QRCodeImage string `json:"qrcode_imagem"`
		Expiracao   string `json:"expiracao"`
	}

	if err := json.Unmarshal(body, &qrResp); err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)

	return &providers.QRCodeResponse{
		QRCodeID:    qrResp.IDQRCode,
		QRCode:      qrResp.QRCode,
		QRCodeImage: qrResp.QRCodeImage,
		Amount:      req.Amount,
		Description: req.Description,
		Status:      "active",
		ExpiresAt:   &expiresAt,
		CreatedAt:   time.Now(),
		RawResponse: map[string]interface{}{"itau": qrResp},
	}, nil
}

func (p *ItauProvider) GetQRCode(ctx context.Context, qrCodeID string) (*providers.QRCodeResponse, error) {
	endpoint := fmt.Sprintf("%s/sispag/v1/qrcodes/%s", p.baseURL, qrCodeID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, &providers.ProviderError{
			Code:      "QUERY_ERROR",
			Message:   "Erro ao consultar QR Code",
			Retryable: true,
			Details:   map[string]interface{}{"error": err.Error()},
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		body = []byte("failed to read response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &providers.ProviderError{
			Code:       "QUERY_FAILED",
			Message:    "Consulta de QR Code falhou",
			StatusCode: resp.StatusCode,
			Details:    map[string]interface{}{"response": string(body)},
		}
	}

	var qrResp struct {
		IDQRCode  string  `json:"id_qrcode"`
		QRCode    string  `json:"qrcode"`
		Status    string  `json:"status"`
		Valor     float64 `json:"valor"`
		Descricao string  `json:"descricao"`
	}

	if err := json.Unmarshal(body, &qrResp); err != nil {
		return nil, err
	}

	return &providers.QRCodeResponse{
		QRCodeID:    qrResp.IDQRCode,
		QRCode:      qrResp.QRCode,
		Amount:      int64(qrResp.Valor * 100),
		Description: qrResp.Descricao,
		Status:      qrResp.Status,
		RawResponse: map[string]interface{}{"itau": qrResp},
	}, nil
}

func (p *ItauProvider) ValidatePixKey(ctx context.Context, pixKey string, pixKeyType domain.PixKeyType) (*providers.PixKeyInfo, error) {
	// TODO: Implementar validação de chave PIX via DICT
	return nil, fmt.Errorf("validação de chave PIX não implementada")
}

func (p *ItauProvider) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL, http.NoBody)
	if err != nil {
		return err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("provider unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

func (p *ItauProvider) GetSupportedMethods() []string {
	return []string{"pix_key", "account", "transfer", "qrcode_static", "qrcode_dynamic"}
}

// mapItauStatus mapeia status do Itaú para nosso domínio
func mapItauStatus(itauStatus string) domain.TransactionStatus {
	switch itauStatus {
	case "PROCESSANDO", "AGENDADO":
		return domain.TransactionStatusProcessing
	case "LIQUIDADO", "CONCLUIDO":
		return domain.TransactionStatusCompleted
	case "REJEITADO", "ERRO":
		return domain.TransactionStatusFailed
	case "CANCELADO":
		return domain.TransactionStatusCancelled
	default:
		return domain.TransactionStatusPending
	}
}

// mapPixKeyTypeToItau mapeia tipo de chave PIX para formato Itaú
func mapPixKeyTypeToItau(keyType domain.PixKeyType) string {
	switch keyType {
	case domain.PixKeyTypeCPF:
		return "CPF"
	case domain.PixKeyTypeCNPJ:
		return "CNPJ"
	case domain.PixKeyTypeEmail:
		return "EMAIL"
	case domain.PixKeyTypePhone:
		return "TELEFONE"
	case domain.PixKeyTypeRandom:
		return "CHAVE_ALEATORIA"
	case domain.PixKeyTypeAccount:
		return "AGENCIA_CONTA"
	default:
		return "CPF"
	}
}
