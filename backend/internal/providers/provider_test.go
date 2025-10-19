package providers

import (
	"context"
	"testing"
	"time"
)

func TestNewProviderRegistry(t *testing.T) {
	registry := NewProviderRegistry()
	if registry == nil {
		t.Fatal("NewProviderRegistry() returned nil")
	}

	if registry.providers == nil {
		t.Fatal("NewProviderRegistry() providers map is nil")
	}
}

func TestProviderRegistryRegisterAndGet(t *testing.T) {
	registry := NewProviderRegistry()

	// Create mock provider
	mockProvider := &MockProvider{
		code: "test",
		name: "Test Provider",
	}

	// Register
	registry.Register(mockProvider)

	// Get
	provider, exists := registry.Get("test")
	if !exists {
		t.Fatal("Get() provider not found")
	}

	if provider.GetCode() != "test" {
		t.Errorf("Get() code = %v, want test", provider.GetCode())
	}
}

func TestProviderRegistryGetNonExistent(t *testing.T) {
	registry := NewProviderRegistry()

	_, exists := registry.Get("nonexistent")
	if exists {
		t.Error("Get() should return false for non-existent provider")
	}
}

func TestProviderRegistryGetAll(t *testing.T) {
	registry := NewProviderRegistry()

	mock1 := &MockProvider{code: "provider1", name: "Provider 1"}
	mock2 := &MockProvider{code: "provider2", name: "Provider 2"}

	registry.Register(mock1)
	registry.Register(mock2)

	all := registry.GetAll()
	if len(all) != 2 {
		t.Errorf("GetAll() length = %v, want 2", len(all))
	}
}

func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClient(30, false)

	if client.timeout != 30*time.Second {
		t.Errorf("NewHTTPClient() timeout = %v, want 30s", client.timeout)
	}

	if client.requireMTLS != false {
		t.Error("NewHTTPClient() requireMTLS should be false")
	}
}

func TestNewProviderError(t *testing.T) {
	err := NewProviderError("TEST_CODE", "Test message", nil)

	if err == nil {
		t.Fatal("NewProviderError() returned nil")
	}

	provErr, ok := err.(*ProviderError)
	if !ok {
		t.Fatal("NewProviderError() did not return *ProviderError")
	}

	if provErr.Code != "TEST_CODE" {
		t.Errorf("Code = %v, want TEST_CODE", provErr.Code)
	}

	if provErr.Message != "Test message" {
		t.Errorf("Message = %v, want 'Test message'", provErr.Message)
	}
}

func TestProviderErrorWithWrappedError(t *testing.T) {
	originalErr := context.DeadlineExceeded
	err := NewProviderError("TIMEOUT", "Request timeout", originalErr)

	provErr, ok := err.(*ProviderError)
	if !ok {
		t.Fatal("NewProviderError() did not return *ProviderError")
	}

	// Should contain both messages
	if provErr.Message == "Request timeout" {
		t.Error("Message should include wrapped error")
	}
}

// MockProvider for testing
type MockProvider struct {
	code string
	name string
}

func (m *MockProvider) GetCode() string {
	return m.code
}

func (m *MockProvider) GetName() string {
	return m.name
}

func (m *MockProvider) Initialize(config ProviderConfig) error {
	return nil
}

func (m *MockProvider) Authenticate(ctx context.Context, credentials ProviderCredentials) (*AuthToken, error) {
	return &AuthToken{
		AccessToken: "mock-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}, nil
}

func (m *MockProvider) RefreshToken(ctx context.Context, refreshToken string) (*AuthToken, error) {
	return nil, NewProviderError("NOT_SUPPORTED", "Refresh not supported", nil)
}

func (m *MockProvider) CreateTransfer(ctx context.Context, req *TransferRequest) (*TransferResponse, error) {
	return &TransferResponse{
		ProviderTxID: "mock-tx-123",
		E2EID:        "E12345678901234567890123456789012",
		Status:       TransactionStatusCompleted,
	}, nil
}

func (m *MockProvider) GetTransfer(ctx context.Context, req *GetTransferRequest) (*TransferResponse, error) {
	return &TransferResponse{
		ProviderTxID: req.ProviderTxID,
		Status:       TransactionStatusCompleted,
	}, nil
}

func (m *MockProvider) CancelTransfer(ctx context.Context, req *CancelTransferRequest) error {
	return nil
}

func (m *MockProvider) CreateQRCodeStatic(ctx context.Context, req *QRCodeRequest) (*QRCodeResponse, error) {
	return &QRCodeResponse{
		QRCodeID: "qr-123",
		QRCode:   "00020126580014br.gov.bcb.pix",
	}, nil
}

func (m *MockProvider) CreateQRCodeDynamic(ctx context.Context, req *QRCodeRequest) (*QRCodeResponse, error) {
	return m.CreateQRCodeStatic(ctx, req)
}

func (m *MockProvider) GetQRCode(ctx context.Context, req *GetQRCodeRequest) (*QRCodeResponse, error) {
	return &QRCodeResponse{
		QRCodeID: req.QRCodeID,
		QRCode:   "00020126580014br.gov.bcb.pix",
	}, nil
}

func (m *MockProvider) ValidatePixKey(ctx context.Context, req *ValidatePixKeyRequest) (*ValidatePixKeyResponse, error) {
	return &ValidatePixKeyResponse{
		Valid:  true,
		PixKey: req.PixKey,
		Name:   "Test User",
	}, nil
}

func (m *MockProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *MockProvider) GetSupportedMethods() []string {
	return []string{"transfer", "qrcode"}
}
