package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMerchantValidation(t *testing.T) {
	merchant := &Merchant{
		ID:       uuid.New(),
		Name:     "Test Merchant",
		Document: "12345678000190",
		Email:    "test@example.com",
		Active:   true,
	}

	if merchant.Name == "" {
		t.Error("Merchant name should not be empty")
	}

	if merchant.Document == "" {
		t.Error("Merchant document should not be empty")
	}

	if merchant.Email == "" {
		t.Error("Merchant email should not be empty")
	}
}

func TestUserRoles(t *testing.T) {
	tests := []struct {
		name string
		role UserRole
		want string
	}{
		{
			name: "admin role",
			role: RoleAdmin,
			want: "admin",
		},
		{
			name: "merchant role",
			role: RoleMerchant,
			want: "merchant",
		},
		{
			name: "developer role",
			role: RoleDeveloper,
			want: "developer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.role) != tt.want {
				t.Errorf("Role = %v, want %v", tt.role, tt.want)
			}
		})
	}
}

func TestTransactionStatus(t *testing.T) {
	tests := []struct {
		name   string
		status TransactionStatus
		want   string
	}{
		{
			name:   "pending",
			status: TransactionStatusPending,
			want:   "pending",
		},
		{
			name:   "processing",
			status: TransactionStatusProcessing,
			want:   "processing",
		},
		{
			name:   "completed",
			status: TransactionStatusCompleted,
			want:   "completed",
		},
		{
			name:   "failed",
			status: TransactionStatusFailed,
			want:   "failed",
		},
		{
			name:   "cancelled",
			status: TransactionStatusCancelled,
			want:   "cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("Status = %v, want %v", tt.status, tt.want)
			}
		})
	}
}

func TestPixKeyTypes(t *testing.T) {
	tests := []struct {
		name    string
		keyType PixKeyType
		want    string
	}{
		{
			name:    "cpf",
			keyType: PixKeyTypeCPF,
			want:    "cpf",
		},
		{
			name:    "cnpj",
			keyType: PixKeyTypeCNPJ,
			want:    "cnpj",
		},
		{
			name:    "email",
			keyType: PixKeyTypeEmail,
			want:    "email",
		},
		{
			name:    "phone",
			keyType: PixKeyTypePhone,
			want:    "phone",
		},
		{
			name:    "random",
			keyType: PixKeyTypeRandom,
			want:    "random",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.keyType) != tt.want {
				t.Errorf("KeyType = %v, want %v", tt.keyType, tt.want)
			}
		})
	}
}

func TestTransactionCreation(t *testing.T) {
	merchantID := uuid.New()
	providerID := uuid.New()

	tx := &Transaction{
		ID:          uuid.New(),
		MerchantID:  merchantID,
		ProviderID:  providerID,
		ExternalID:  "ORDER-123",
		Amount:      10000,
		Description: "Test payment",
		Status:      TransactionStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if tx.MerchantID != merchantID {
		t.Errorf("MerchantID = %v, want %v", tx.MerchantID, merchantID)
	}

	if tx.Amount != 10000 {
		t.Errorf("Amount = %v, want 10000", tx.Amount)
	}

	if tx.Status != TransactionStatusPending {
		t.Errorf("Status = %v, want pending", tx.Status)
	}
}
