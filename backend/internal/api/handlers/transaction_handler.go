package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pixsaas/backend/internal/audit"
	"github.com/pixsaas/backend/internal/domain"
	"github.com/pixsaas/backend/internal/providers"
	"github.com/pixsaas/backend/internal/repository"
	"github.com/pixsaas/backend/internal/security"
	"gorm.io/gorm"
)

// TransactionHandler gerencia transações PIX
type TransactionHandler struct {
	db                   *gorm.DB
	txRepo               *repository.TransactionRepository
	merchantRepo         *repository.MerchantRepository
	providerRepo         *repository.ProviderRepository
	merchantProviderRepo *repository.MerchantProviderRepository
	auditService         *audit.AuditService
	encryptionService    *security.EncryptionService
	providerRegistry     *providers.ProviderRegistry
}

// NewTransactionHandler cria um novo handler de transações
func NewTransactionHandler(
	db *gorm.DB,
	auditService *audit.AuditService,
	encryptionService *security.EncryptionService,
	providerRegistry *providers.ProviderRegistry,
) *TransactionHandler {
	return &TransactionHandler{
		db:                   db,
		txRepo:               repository.NewTransactionRepository(db),
		merchantRepo:         repository.NewMerchantRepository(db),
		providerRepo:         repository.NewProviderRepository(db),
		merchantProviderRepo: repository.NewMerchantProviderRepository(db),
		auditService:         auditService,
		encryptionService:    encryptionService,
		providerRegistry:     providerRegistry,
	}
}

// CreateTransferRequest representa uma requisição de transferência
type CreateTransferRequest struct {
	ExternalID   string `json:"external_id" validate:"required"`
	Amount       int64  `json:"amount" validate:"required,min=1"`
	Description  string `json:"description"`
	ProviderCode string `json:"provider_code,omitempty"`

	// Recebedor (obrigatório)
	PayeeName       string            `json:"payee_name" validate:"required"`
	PayeeDocument   string            `json:"payee_document" validate:"required"`
	PayeePixKey     string            `json:"payee_pix_key,omitempty"`
	PayeePixKeyType domain.PixKeyType `json:"payee_pix_key_type,omitempty"`
	PayeeAccount    *AccountInfo      `json:"payee_account,omitempty"`

	// Metadata opcional
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AccountInfo representa informações de conta bancária
type AccountInfo struct {
	Bank   string `json:"bank"`
	ISPB   string `json:"ispb"`
	Agency string `json:"agency"`
	Number string `json:"number"`
	Type   string `json:"type"` // checking, savings
}

// TransactionResponse representa a resposta de uma transação
type TransactionResponse struct {
	ID          uuid.UUID                `json:"id"`
	ExternalID  string                   `json:"external_id"`
	E2EID       string                   `json:"e2e_id,omitempty"`
	Status      domain.TransactionStatus `json:"status"`
	Amount      int64                    `json:"amount"`
	Description string                   `json:"description"`
	Provider    string                   `json:"provider"`
	CreatedAt   string                   `json:"created_at"`
	UpdatedAt   string                   `json:"updated_at"`
}

// CreateTransfer cria uma nova transferência PIX
func (h *TransactionHandler) CreateTransfer(c *fiber.Ctx) error {
	merchantID, ok := c.Locals("merchant_id").(*uuid.UUID)
	if !ok || merchantID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "merchant not found in context",
		})
	}

	var req CreateTransferRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Verificar se external_id já existe
	existing, _ := h.txRepo.GetByExternalID(c.Context(), *merchantID, req.ExternalID)
	if existing != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "external_id already exists",
		})
	}

	// Selecionar provider
	var selectedProvider *domain.Provider
	var merchantProvider *domain.MerchantProvider

	if req.ProviderCode != "" {
		// Provider específico solicitado
		provider, err := h.providerRepo.GetByCode(c.Context(), req.ProviderCode)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "provider not found",
			})
		}
		selectedProvider = provider

		// Buscar configuração do merchant para este provider
		mp, err := h.merchantProviderRepo.GetByMerchantAndProvider(c.Context(), *merchantID, provider.ID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "merchant not configured for this provider",
			})
		}
		merchantProvider = mp
	} else {
		// Selecionar provider automaticamente (primeiro ativo)
		mps, err := h.merchantProviderRepo.ListByMerchant(c.Context(), *merchantID, true)
		if err != nil || len(mps) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "no active providers configured",
			})
		}
		merchantProvider = &mps[0]
		selectedProvider = &merchantProvider.Provider
	}

	// Obter implementação do provider
	providerImpl, exists := h.providerRegistry.Get(selectedProvider.Code)
	if !exists {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "provider implementation not found",
		})
	}

	// Inicializar provider com configuração convertida
	providerConfig := providers.ProviderConfig{
		BaseURL:      selectedProvider.Config.BaseURL,
		AuthURL:      selectedProvider.Config.AuthURL,
		SandboxURL:   selectedProvider.Config.SandboxURL,
		Timeout:      selectedProvider.Config.Timeout,
		MaxRetries:   selectedProvider.Config.MaxRetries,
		RequiresMTLS: selectedProvider.Config.RequiresMTLS,
	}

	if err := providerImpl.Initialize(providerConfig); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to initialize provider",
		})
	}

	// Descriptografar credenciais
	clientID, _ := h.encryptionService.Decrypt(merchantProvider.ClientID)
	clientSecret, _ := h.encryptionService.Decrypt(merchantProvider.ClientSecret)

	// Autenticar com provider
	credentials := providers.ProviderCredentials{
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		AccountAgency: merchantProvider.AccountAgency,
		AccountNumber: merchantProvider.AccountNumber,
		AccountType:   merchantProvider.AccountType,
		PixKey:        merchantProvider.PixKey,
		PixKeyType:    merchantProvider.PixKeyType,
	}

	// TODO: Implementar cache de tokens
	_, authErr := providerImpl.Authenticate(c.Context(), credentials)
	if authErr != nil {
		h.auditService.LogProviderOperation(c.Context(), *merchantID, uuid.Nil, selectedProvider.Code, "authenticate", false, authErr.Error(), 0)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to authenticate with provider",
		})
	}

	// Criar requisição de transferência
	transferReq := &providers.TransferRequest{
		ExternalID:  req.ExternalID,
		Amount:      req.Amount,
		Description: req.Description,

		// Pagador (merchant)
		PayerAccountAgency: merchantProvider.AccountAgency,
		PayerAccountNumber: merchantProvider.AccountNumber,
		PayerAccountType:   merchantProvider.AccountType,
		PayerPixKey:        merchantProvider.PixKey,
		PayerPixKeyType:    merchantProvider.PixKeyType,

		// Recebedor
		PayeeName:       req.PayeeName,
		PayeeDocument:   req.PayeeDocument,
		PayeePixKey:     req.PayeePixKey,
		PayeePixKeyType: req.PayeePixKeyType,

		Metadata: req.Metadata,
	}

	if req.PayeeAccount != nil {
		transferReq.PayeeBank = req.PayeeAccount.Bank
		transferReq.PayeeISPB = req.PayeeAccount.ISPB
		transferReq.PayeeAccountAgency = req.PayeeAccount.Agency
		transferReq.PayeeAccountNumber = req.PayeeAccount.Number
		transferReq.PayeeAccountType = req.PayeeAccount.Type
	}

	// Criar transação no banco
	tx := &domain.Transaction{
		ID:              uuid.New(),
		MerchantID:      *merchantID,
		ProviderID:      selectedProvider.ID,
		ExternalID:      req.ExternalID,
		Type:            domain.TransactionTypeTransfer,
		Status:          domain.TransactionStatusPending,
		Amount:          req.Amount,
		Description:     req.Description,
		PayeeName:       req.PayeeName,
		PayeeDocument:   req.PayeeDocument,
		PayeePixKey:     req.PayeePixKey,
		PayeePixKeyType: req.PayeePixKeyType,
		Metadata:        req.Metadata,
	}

	if req.PayeeAccount != nil {
		tx.PayeeBank = req.PayeeAccount.Bank
		tx.PayeeAccountAgency = req.PayeeAccount.Agency
		tx.PayeeAccountNumber = req.PayeeAccount.Number
	}

	if err := h.txRepo.Create(c.Context(), tx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create transaction",
		})
	}

	// Executar transferência com provider
	transferResp, err := providerImpl.CreateTransfer(c.Context(), transferReq)
	if err != nil {
		// Atualizar transação como falha
		tx.Status = domain.TransactionStatusFailed
		if providerErr, ok := err.(*providers.ProviderError); ok {
			tx.ErrorCode = providerErr.Code
			tx.ErrorMessage = providerErr.Message
		} else {
			tx.ErrorMessage = err.Error()
		}
		h.txRepo.Update(c.Context(), tx)

		h.auditService.LogProviderOperation(c.Context(), *merchantID, tx.ID, selectedProvider.Code, "create_transfer", false, err.Error(), 0)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "transfer failed",
			"details": tx.ErrorMessage,
		})
	}

	// Atualizar transação com resposta do provider
	tx.ProviderTxID = transferResp.ProviderTxID
	tx.E2EID = transferResp.E2EID
	tx.Status = transferResp.Status
	tx.ProcessedAt = transferResp.ProcessedAt
	tx.CompletedAt = transferResp.CompletedAt

	if err := h.txRepo.Update(c.Context(), tx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update transaction",
		})
	}

	// Log de auditoria
	h.auditService.LogTransaction(c.Context(), *merchantID, uuid.Nil, tx.ID, "create_transfer", map[string]interface{}{
		"provider": selectedProvider.Code,
		"amount":   req.Amount,
		"status":   tx.Status,
	})

	// TODO: Enviar webhook se configurado

	return c.Status(fiber.StatusCreated).JSON(TransactionResponse{
		ID:          tx.ID,
		ExternalID:  tx.ExternalID,
		E2EID:       tx.E2EID,
		Status:      tx.Status,
		Amount:      tx.Amount,
		Description: tx.Description,
		Provider:    selectedProvider.Code,
		CreatedAt:   tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   tx.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetTransaction busca uma transação por ID
func (h *TransactionHandler) GetTransaction(c *fiber.Ctx) error {
	merchantID, ok := c.Locals("merchant_id").(*uuid.UUID)
	if !ok || merchantID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "merchant not found in context",
		})
	}

	txID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid transaction id",
		})
	}

	tx, err := h.txRepo.GetByID(c.Context(), txID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "transaction not found",
		})
	}

	// Verificar se pertence ao merchant
	if tx.MerchantID != *merchantID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	return c.JSON(TransactionResponse{
		ID:          tx.ID,
		ExternalID:  tx.ExternalID,
		E2EID:       tx.E2EID,
		Status:      tx.Status,
		Amount:      tx.Amount,
		Description: tx.Description,
		Provider:    tx.Provider.Code,
		CreatedAt:   tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   tx.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ListTransactions lista transações do merchant
func (h *TransactionHandler) ListTransactions(c *fiber.Ctx) error {
	merchantID, ok := c.Locals("merchant_id").(*uuid.UUID)
	if !ok || merchantID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "merchant not found in context",
		})
	}

	// Parâmetros de paginação
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	// Filtros
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = domain.TransactionStatus(status)
	}

	transactions, total, err := h.txRepo.ListByMerchant(c.Context(), *merchantID, filters, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list transactions",
		})
	}

	var response []TransactionResponse
	for _, tx := range transactions {
		response = append(response, TransactionResponse{
			ID:          tx.ID,
			ExternalID:  tx.ExternalID,
			E2EID:       tx.E2EID,
			Status:      tx.Status,
			Amount:      tx.Amount,
			Description: tx.Description,
			Provider:    tx.Provider.Code,
			CreatedAt:   tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   tx.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return c.JSON(fiber.Map{
		"data":   response,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}
