# Changelog

Todas as mudanças notáveis neste projeto serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/),
e este projeto adere ao [Semantic Versioning](https://semver.org/lang/pt-BR/).

## [Unreleased]

### Adicionado
- GitHub Actions workflows para CI/CD
- Testes automatizados em múltiplas versões do Go
- Scan de segurança com Gosec e Trivy
- Análise de código com CodeQL
- Build automático de imagens Docker
- Release automático com binários cross-platform
- Templates para Pull Requests e Issues
- Script de validação local de CI
- Badges de status no README
- Documentação completa dos workflows

## [1.0.0] - 2025-01-19

### Adicionado
- Implementação completa do backend em Go
- API REST com Fiber framework
- Sistema de autenticação JWT
- Criptografia AES-256-GCM para dados sensíveis
- Suporte para múltiplos providers PIX
- Implementações para Bradesco, Itaú, BB, Santander e Inter
- Sistema de auditoria com retenção de 5 anos
- Dashboard frontend em Next.js 14
- CLI administrativa com Cobra
- Docker Compose para desenvolvimento
- Documentação completa da API (OpenAPI)
- Testes unitários para componentes críticos
- 33 testes passando com cobertura adequada

### Segurança
- Implementação de PCI DSS compliance
- TLS 1.3 obrigatório
- Rate limiting
- CORS configurável
- Proteção contra SQL Injection
- Proteção contra XSS
- Logs de auditoria completos

### Infraestrutura
- PostgreSQL com suporte a replicação
- Redis para cache (preparado)
- Prometheus para métricas
- Grafana para dashboards
- Vault para secrets (preparado)

## [0.1.0] - 2025-01-15

### Adicionado
- Estrutura inicial do projeto
- Modelos de domínio
- Arquitetura Clean Architecture
- Configuração básica do Go
- Estrutura de diretórios

---

## Tipos de Mudanças

- `Adicionado` para novas funcionalidades
- `Modificado` para mudanças em funcionalidades existentes
- `Descontinuado` para funcionalidades que serão removidas
- `Removido` para funcionalidades removidas
- `Corrigido` para correções de bugs
- `Segurança` para vulnerabilidades corrigidas

## Links

- [Unreleased]: https://github.com/YOUR_USERNAME/pix-saas/compare/v1.0.0...HEAD
- [1.0.0]: https://github.com/YOUR_USERNAME/pix-saas/releases/tag/v1.0.0
- [0.1.0]: https://github.com/YOUR_USERNAME/pix-saas/releases/tag/v0.1.0
