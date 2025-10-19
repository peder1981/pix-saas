# GitHub Actions - Workflows de CI/CD

## 📋 Visão Geral

Este projeto utiliza GitHub Actions para garantir qualidade de código, segurança e automação de deploys. Todos os workflows são executados automaticamente em pushes e pull requests.

## 🔄 Workflows Implementados

### 1. Tests (`tests.yml`)

**Trigger**: Push e Pull Request nas branches `main` e `develop`

**Funcionalidades**:
- ✅ Executa testes unitários em múltiplas versões do Go (1.21, 1.22)
- ✅ Verifica race conditions com `-race` flag
- ✅ Gera relatório de cobertura de código
- ✅ Upload automático para Codecov
- ✅ Lint com golangci-lint
- ✅ Build dos binários (API e CLI)
- ✅ Scan de segurança com Gosec

**Jobs**:
1. **test**: Executa testes com cobertura
2. **lint**: Análise estática de código
3. **build**: Compila binários
4. **security**: Scan de vulnerabilidades

**Artefatos Gerados**:
- Relatório de cobertura HTML
- Binários compilados (API e CLI)
- Relatório SARIF de segurança

### 2. Docker Build (`docker.yml`)

**Trigger**: Push na branch `main` e tags `v*`

**Funcionalidades**:
- ✅ Build de imagens Docker otimizadas
- ✅ Push automático para GitHub Container Registry (ghcr.io)
- ✅ Versionamento automático com tags semânticas
- ✅ Cache de layers para builds rápidos
- ✅ Scan de vulnerabilidades com Trivy
- ✅ Upload de resultados para GitHub Security

**Tags Geradas**:
- `main` - Última versão da branch principal
- `v1.0.0` - Versão específica
- `v1.0` - Major.Minor
- `sha-abc123` - Commit específico

### 3. Frontend Tests (`frontend.yml`)

**Trigger**: Push e Pull Request que afetam a pasta `frontend/`

**Funcionalidades**:
- ✅ Testes em múltiplas versões do Node.js (18.x, 20.x)
- ✅ Lint com ESLint
- ✅ Type checking com TypeScript
- ✅ Build de produção
- ✅ Lighthouse CI para métricas de performance

**Jobs**:
1. **test**: Testes e build do frontend
2. **lighthouse**: Análise de performance e acessibilidade

### 4. Release (`release.yml`)

**Trigger**: Push de tags `v*` (ex: `v1.0.0`)

**Funcionalidades**:
- ✅ Executa testes antes do release
- ✅ Build cross-platform (Linux, macOS, Windows)
- ✅ Suporte para AMD64 e ARM64
- ✅ Geração de checksums SHA256
- ✅ Release notes automáticas
- ✅ Upload de binários para GitHub Releases

**Plataformas Suportadas**:
- Linux AMD64
- Linux ARM64
- macOS AMD64
- macOS ARM64 (Apple Silicon)
- Windows AMD64

### 5. CodeQL (`codeql.yml`)

**Trigger**: 
- Push e Pull Request nas branches `main` e `develop`
- Schedule: Toda segunda-feira à meia-noite

**Funcionalidades**:
- ✅ Análise de segurança avançada
- ✅ Detecção de vulnerabilidades
- ✅ Análise de código Go e JavaScript
- ✅ Integração com GitHub Security

## 🛡️ Segurança

### Ferramentas de Segurança Integradas

1. **Gosec**: Scanner de segurança para Go
   - Detecta problemas comuns de segurança
   - Gera relatórios SARIF
   - Integrado ao GitHub Security

2. **Trivy**: Scanner de vulnerabilidades em containers
   - Analisa imagens Docker
   - Detecta CVEs conhecidos
   - Verifica dependências

3. **CodeQL**: Análise semântica de código
   - Detecta padrões de vulnerabilidade
   - Análise profunda de fluxo de dados
   - Suporte para múltiplas linguagens

## 📊 Qualidade de Código

### golangci-lint

Configuração em `.golangci.yml` com 25+ linters habilitados:

**Categorias**:
- **Bugs**: errcheck, govet, staticcheck
- **Estilo**: gofmt, goimports, revive
- **Performance**: gocritic
- **Segurança**: gosec
- **Complexidade**: gocritic, revive

### Métricas de Cobertura

- **Target**: 80%+ de cobertura
- **Upload automático**: Codecov
- **Relatórios**: HTML gerado em cada build
- **Trend**: Monitoramento ao longo do tempo

## 🚀 Como Usar

### Executar Localmente

#### Testes
```bash
cd backend
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Lint
```bash
cd backend
golangci-lint run ./...
```

#### Build Docker
```bash
docker build -f docker/Dockerfile.api -t pix-saas-api .
```

### Criar um Release

1. Certifique-se de que todos os testes passam
2. Atualize o CHANGELOG.md
3. Crie e push uma tag:
```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

4. O workflow `release.yml` será executado automaticamente
5. Binários estarão disponíveis na página de Releases

## 📝 Templates

### Pull Request Template

Localizado em `.github/PULL_REQUEST_TEMPLATE.md`

**Seções**:
- Descrição das mudanças
- Tipo de mudança
- Como foi testado
- Checklist de qualidade
- Impacto (segurança, performance, compatibilidade)

### Issue Templates

#### Bug Report (`.github/ISSUE_TEMPLATE/bug_report.md`)
- Descrição do bug
- Passos para reproduzir
- Comportamento esperado vs atual
- Ambiente
- Logs

#### Feature Request (`.github/ISSUE_TEMPLATE/feature_request.md`)
- Descrição da funcionalidade
- Problema que resolve
- Casos de uso
- Impacto e prioridade

## 🔧 Configuração

### Secrets Necessários

Para funcionalidade completa, configure os seguintes secrets no GitHub:

1. **GITHUB_TOKEN**: Gerado automaticamente (já disponível)
2. **CODECOV_TOKEN**: Para upload de cobertura (opcional)

### Permissões

Os workflows requerem as seguintes permissões:

- `contents: write` - Para criar releases
- `packages: write` - Para push de imagens Docker
- `security-events: write` - Para upload de scans de segurança

## 📈 Monitoramento

### Status dos Workflows

Visualize o status em tempo real:
- GitHub Actions tab no repositório
- Badges no README.md
- Notificações por email (configurável)

### Métricas

- **Tempo médio de build**: ~3-5 minutos
- **Taxa de sucesso**: Target 95%+
- **Cobertura de código**: Target 80%+

## 🐛 Troubleshooting

### Testes Falhando

1. Execute localmente: `go test -v ./...`
2. Verifique logs do workflow
3. Valide dependências: `go mod tidy`

### Build Docker Falhando

1. Teste localmente: `docker build -f docker/Dockerfile.api .`
2. Verifique Dockerfile
3. Valide contexto de build

### Lint Falhando

1. Execute localmente: `golangci-lint run ./...`
2. Corrija warnings reportados
3. Atualize `.golangci.yml` se necessário

## 🎯 Melhores Práticas

1. **Sempre execute testes localmente antes de push**
2. **Mantenha PRs pequenos e focados**
3. **Escreva mensagens de commit descritivas**
4. **Atualize documentação junto com código**
5. **Revise relatórios de segurança regularmente**
6. **Monitore cobertura de testes**
7. **Use semantic versioning para releases**

## 📚 Recursos Adicionais

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint Linters](https://golangci-lint.run/usage/linters/)
- [Gosec Rules](https://github.com/securego/gosec#available-rules)
- [Trivy Documentation](https://aquasecurity.github.io/trivy/)
- [CodeQL Documentation](https://codeql.github.com/docs/)

## 🤝 Contribuindo

Para contribuir com melhorias nos workflows:

1. Teste mudanças em um fork primeiro
2. Documente novas funcionalidades
3. Atualize este documento
4. Abra um PR com descrição detalhada

---

**Desenvolvido por**: Peder Munksgaard (JMPM Tecnologia)  
**Última atualização**: 2025-01-19  
**Versão**: 1.0.0
