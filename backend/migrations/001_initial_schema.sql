-- Habilitar extensões necessárias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Tabela de Merchants (Multi-tenant)
CREATE TABLE merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    document VARCHAR(20) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(20),
    active BOOLEAN DEFAULT true,
    api_key TEXT NOT NULL UNIQUE,
    webhook_url TEXT,
    ip_whitelist TEXT[],
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_merchants_document ON merchants(document);
CREATE INDEX idx_merchants_email ON merchants(email);
CREATE INDEX idx_merchants_active ON merchants(active);
CREATE INDEX idx_merchants_deleted_at ON merchants(deleted_at);

-- Tabela de Usuários
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID REFERENCES merchants(id),
    email VARCHAR(255) NOT NULL UNIQUE,
    password TEXT NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_merchant_id ON users(merchant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Tabela de Providers (Instituições Financeiras)
CREATE TABLE providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    ispb VARCHAR(8) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL,
    active BOOLEAN DEFAULT true,
    config JSONB NOT NULL,
    priority INTEGER DEFAULT 0,
    health_status VARCHAR(50) DEFAULT 'unknown',
    last_health_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_providers_code ON providers(code);
CREATE INDEX idx_providers_ispb ON providers(ispb);
CREATE INDEX idx_providers_active ON providers(active);
CREATE INDEX idx_providers_health_status ON providers(health_status);
CREATE INDEX idx_providers_deleted_at ON providers(deleted_at);

-- Tabela de Configurações Merchant-Provider
CREATE TABLE merchant_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    provider_id UUID NOT NULL REFERENCES providers(id),
    active BOOLEAN DEFAULT true,
    client_id TEXT NOT NULL,
    client_secret TEXT NOT NULL,
    certificate_data TEXT,
    private_key_data TEXT,
    account_agency VARCHAR(10),
    account_number VARCHAR(20),
    account_type VARCHAR(20),
    pix_key VARCHAR(255),
    pix_key_type VARCHAR(20),
    last_token_refresh TIMESTAMP,
    token_expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    UNIQUE(merchant_id, provider_id)
);

CREATE INDEX idx_merchant_providers_merchant_id ON merchant_providers(merchant_id);
CREATE INDEX idx_merchant_providers_provider_id ON merchant_providers(provider_id);
CREATE INDEX idx_merchant_providers_active ON merchant_providers(active);
CREATE INDEX idx_merchant_providers_deleted_at ON merchant_providers(deleted_at);

-- Tabela de Transações
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    provider_id UUID NOT NULL REFERENCES providers(id),
    external_id VARCHAR(255) UNIQUE,
    provider_tx_id VARCHAR(255),
    e2e_id VARCHAR(255) UNIQUE,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    amount BIGINT NOT NULL,
    currency VARCHAR(3) DEFAULT 'BRL',
    description TEXT,
    
    -- Pagador
    payer_name VARCHAR(255),
    payer_document VARCHAR(20),
    payer_pix_key VARCHAR(255),
    payer_pix_key_type VARCHAR(20),
    payer_account_agency VARCHAR(10),
    payer_account_number VARCHAR(20),
    payer_bank VARCHAR(255),
    
    -- Recebedor
    payee_name VARCHAR(255),
    payee_document VARCHAR(20),
    payee_pix_key VARCHAR(255),
    payee_pix_key_type VARCHAR(20),
    payee_account_agency VARCHAR(10),
    payee_account_number VARCHAR(20),
    payee_bank VARCHAR(255),
    
    -- QR Code
    qr_code TEXT,
    qr_code_image TEXT,
    qr_code_expires_at TIMESTAMP,
    
    -- Metadata
    metadata JSONB,
    error_code VARCHAR(50),
    error_message TEXT,
    
    -- Timestamps
    processed_at TIMESTAMP,
    completed_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_merchant_id ON transactions(merchant_id);
CREATE INDEX idx_transactions_provider_id ON transactions(provider_id);
CREATE INDEX idx_transactions_external_id ON transactions(external_id);
CREATE INDEX idx_transactions_provider_tx_id ON transactions(provider_tx_id);
CREATE INDEX idx_transactions_e2e_id ON transactions(e2e_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_transactions_payee_document ON transactions(payee_document);
CREATE INDEX idx_transactions_payer_document ON transactions(payer_document);

-- Tabela de Logs de Auditoria (Retenção 5 anos)
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID REFERENCES merchants(id),
    user_id UUID REFERENCES users(id),
    transaction_id UUID REFERENCES transactions(id),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    method VARCHAR(10),
    path TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    request_body JSONB,
    response_code INTEGER,
    response_body JSONB,
    error_message TEXT,
    duration BIGINT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_merchant_id ON audit_logs(merchant_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_transaction_id ON audit_logs(transaction_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX idx_audit_logs_ip_address ON audit_logs(ip_address);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- Tabela de Webhooks
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    url TEXT NOT NULL,
    events TEXT[] NOT NULL,
    secret TEXT NOT NULL,
    active BOOLEAN DEFAULT true,
    max_retries INTEGER DEFAULT 3,
    timeout INTEGER DEFAULT 30,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_webhooks_merchant_id ON webhooks(merchant_id);
CREATE INDEX idx_webhooks_active ON webhooks(active);
CREATE INDEX idx_webhooks_deleted_at ON webhooks(deleted_at);

-- Tabela de Entregas de Webhook
CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    event VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    attempt INTEGER DEFAULT 1,
    status VARCHAR(50) NOT NULL,
    response_code INTEGER,
    response_body TEXT,
    error_message TEXT,
    next_retry_at TIMESTAMP,
    delivered_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_deliveries_webhook_id ON webhook_deliveries(webhook_id);
CREATE INDEX idx_webhook_deliveries_transaction_id ON webhook_deliveries(transaction_id);
CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status);
CREATE INDEX idx_webhook_deliveries_created_at ON webhook_deliveries(created_at);
CREATE INDEX idx_webhook_deliveries_next_retry_at ON webhook_deliveries(next_retry_at);

-- Tabela de API Keys
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    name VARCHAR(255) NOT NULL,
    key TEXT NOT NULL UNIQUE,
    prefix VARCHAR(20) NOT NULL,
    permissions TEXT[],
    active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_api_keys_merchant_id ON api_keys(merchant_id);
CREATE INDEX idx_api_keys_key ON api_keys(key);
CREATE INDEX idx_api_keys_active ON api_keys(active);
CREATE INDEX idx_api_keys_deleted_at ON api_keys(deleted_at);

-- Tabela de Refresh Tokens
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT false,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_revoked ON refresh_tokens(revoked);

-- Função para atualizar updated_at automaticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers para atualizar updated_at
CREATE TRIGGER update_merchants_updated_at BEFORE UPDATE ON merchants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_providers_updated_at BEFORE UPDATE ON providers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_merchant_providers_updated_at BEFORE UPDATE ON merchant_providers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_webhooks_updated_at BEFORE UPDATE ON webhooks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_api_keys_updated_at BEFORE UPDATE ON api_keys
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comentários nas tabelas
COMMENT ON TABLE merchants IS 'Clientes da plataforma (multi-tenant)';
COMMENT ON TABLE users IS 'Usuários do sistema (admin, merchant users)';
COMMENT ON TABLE providers IS 'Instituições financeiras (bancos, fintechs, PSPs)';
COMMENT ON TABLE merchant_providers IS 'Configurações de integração merchant-provider';
COMMENT ON TABLE transactions IS 'Transações PIX';
COMMENT ON TABLE audit_logs IS 'Logs de auditoria com retenção de 5 anos';
COMMENT ON TABLE webhooks IS 'Configurações de webhooks dos merchants';
COMMENT ON TABLE webhook_deliveries IS 'Histórico de entregas de webhooks';
COMMENT ON TABLE api_keys IS 'Chaves de API para autenticação';
COMMENT ON TABLE refresh_tokens IS 'Tokens de refresh JWT';
