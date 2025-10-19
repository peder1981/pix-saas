package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/pixsaas/backend/configs"
	"github.com/pixsaas/backend/internal/domain"
	"github.com/pixsaas/backend/internal/security"
)

var (
	db  *gorm.DB
	cfg *configs.Config
)

func main() {
	var err error

	// Carregar configuração
	cfg, err = configs.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("Erro ao carregar configuração: %v", err)
	}

	// Conectar ao banco
	db, err = gorm.Open(postgres.Open(cfg.Database.GetDSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco: %v", err)
	}

	// Inicializar encryption service
	encryptionKey, err := base64.StdEncoding.DecodeString(cfg.Encryption.Key)
	if err != nil || len(encryptionKey) != 32 {
		log.Fatalf("Chave de criptografia inválida")
	}

	_, err = security.NewEncryptionService(encryptionKey)
	if err != nil {
		log.Fatalf("Erro ao criar serviço de criptografia: %v", err)
	}

	// Executar CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "pixsaas-cli",
	Short: "PIX SaaS CLI - Ferramenta administrativa",
	Long: `CLI para gerenciamento do PIX SaaS.
	
Permite adicionar/remover providers, configurar merchants e gerenciar credenciais.`,
}

func init() {
	rootCmd.AddCommand(providerCmd)
	rootCmd.AddCommand(merchantCmd)
	rootCmd.AddCommand(keysCmd)
}

// Provider commands
var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Gerenciar providers",
}

var providerAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adicionar novo provider",
	Run: func(cmd *cobra.Command, args []string) {
		code, err := cmd.Flags().GetString("code")
		if err != nil {
			log.Fatalf("Erro ao obter flag code: %v", err)
		}
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatalf("Erro ao obter flag name: %v", err)
		}
		ispb, err := cmd.Flags().GetString("ispb")
		if err != nil {
			log.Fatalf("Erro ao obter flag ispb: %v", err)
		}
		providerType, err := cmd.Flags().GetString("type")
		if err != nil {
			log.Fatalf("Erro ao obter flag type: %v", err)
		}
		baseURL, err := cmd.Flags().GetString("base-url")
		if err != nil {
			log.Fatalf("Erro ao obter flag base-url: %v", err)
		}
		authURL, err := cmd.Flags().GetString("auth-url")
		if err != nil {
			log.Fatalf("Erro ao obter flag auth-url: %v", err)
		}

		provider := &domain.Provider{
			Code:   code,
			Name:   name,
			ISPB:   ispb,
			Type:   domain.ProviderType(providerType),
			Active: true,
			Config: domain.ProviderConfig{
				BaseURL:      baseURL,
				AuthURL:      authURL,
				Timeout:      30,
				MaxRetries:   3,
				RequiresMTLS: true,
			},
		}

		if err := db.Create(provider).Error; err != nil {
			log.Fatalf("Erro ao criar provider: %v", err)
		}

		fmt.Printf("✅ Provider '%s' criado com sucesso (ID: %s)\n", name, provider.ID)
	},
}

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "Listar providers",
	Run: func(cmd *cobra.Command, args []string) {
		var providers []domain.Provider
		if err := db.Find(&providers).Error; err != nil {
			log.Fatalf("Erro ao listar providers: %v", err)
		}

		fmt.Println("\n📋 Providers cadastrados:")
		fmt.Println("─────────────────────────────────────────────────────────────")
		for _, p := range providers {
			status := "🔴 Inativo"
			if p.Active {
				status = "🟢 Ativo"
			}
			fmt.Printf("%-20s | %-30s | %s | %s\n", p.Code, p.Name, p.ISPB, status)
		}
		fmt.Println("─────────────────────────────────────────────────────────────")
	},
}

var providerDeleteCmd = &cobra.Command{
	Use:   "delete [code]",
	Short: "Deletar provider",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		code := args[0]

		result := db.Where("code = ?", code).Delete(&domain.Provider{})
		if result.Error != nil {
			log.Fatalf("Erro ao deletar provider: %v", result.Error)
		}

		if result.RowsAffected == 0 {
			fmt.Printf("⚠️  Provider '%s' não encontrado\n", code)
			return
		}

		fmt.Printf("✅ Provider '%s' deletado com sucesso\n", code)
	},
}

func init() {
	providerCmd.AddCommand(providerAddCmd)
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerDeleteCmd)

	providerAddCmd.Flags().String("code", "", "Código do provider (ex: bradesco)")
	providerAddCmd.Flags().String("name", "", "Nome do provider")
	providerAddCmd.Flags().String("ispb", "", "ISPB do provider")
	providerAddCmd.Flags().String("type", "bank", "Tipo (bank, digital_bank, cooperative, fintech, psp)")
	providerAddCmd.Flags().String("base-url", "", "URL base da API")
	providerAddCmd.Flags().String("auth-url", "", "URL de autenticação")

	if err := providerAddCmd.MarkFlagRequired("code"); err != nil {
		log.Printf("Erro ao marcar flag como obrigatória: %v", err)
	}
	if err := providerAddCmd.MarkFlagRequired("name"); err != nil {
		log.Printf("Erro ao marcar flag como obrigatória: %v", err)
	}
	if err := providerAddCmd.MarkFlagRequired("ispb"); err != nil {
		log.Printf("Erro ao marcar flag como obrigatória: %v", err)
	}
	if err := providerAddCmd.MarkFlagRequired("base-url"); err != nil {
		log.Printf("Erro ao marcar flag como obrigatória: %v", err)
	}
	if err := providerAddCmd.MarkFlagRequired("auth-url"); err != nil {
		log.Printf("Erro ao marcar flag como obrigatória: %v", err)
	}
}

// Merchant commands
var merchantCmd = &cobra.Command{
	Use:   "merchant",
	Short: "Gerenciar merchants",
}

var merchantListCmd = &cobra.Command{
	Use:   "list",
	Short: "Listar merchants",
	Run: func(cmd *cobra.Command, args []string) {
		var merchants []domain.Merchant
		if err := db.Find(&merchants).Error; err != nil {
			log.Fatalf("Erro ao listar merchants: %v", err)
		}

		fmt.Println("\n📋 Merchants cadastrados:")
		fmt.Println("─────────────────────────────────────────────────────────────")
		for _, m := range merchants {
			status := "🔴 Inativo"
			if m.Active {
				status = "🟢 Ativo"
			}
			fmt.Printf("%-30s | %-20s | %s | %s\n", m.Name, m.Document, m.Email, status)
		}
		fmt.Println("─────────────────────────────────────────────────────────────")
	},
}

func init() {
	merchantCmd.AddCommand(merchantListCmd)
}

// Keys commands
var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Gerenciar chaves de criptografia",
}

var keysGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Gerar nova chave de criptografia",
	Run: func(cmd *cobra.Command, args []string) {
		key, err := security.GenerateKeyBase64()
		if err != nil {
			log.Fatalf("Erro ao gerar chave: %v", err)
		}

		fmt.Println("\n🔑 Nova chave de criptografia gerada:")
		fmt.Println("─────────────────────────────────────────────────────────────")
		fmt.Println(key)
		fmt.Println("─────────────────────────────────────────────────────────────")
		fmt.Println("\n⚠️  IMPORTANTE:")
		fmt.Println("1. Guarde esta chave em local seguro")
		fmt.Println("2. Adicione ao arquivo .env como ENCRYPTION_KEY")
		fmt.Println("3. Nunca compartilhe esta chave")
		fmt.Println("4. Se perder esta chave, não será possível descriptografar dados existentes")
	},
}

func init() {
	keysCmd.AddCommand(keysGenerateCmd)
}
