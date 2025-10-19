# GitHub Actions - Automação de CI/CD

Este diretório contém todos os workflows do GitHub Actions para automação de testes, builds, segurança e releases.

## 📁 Estrutura

```
.github/
├── workflows/
│   ├── tests.yml          # Testes, lint e build
│   ├── docker.yml         # Build e publicação de imagens Docker
│   ├── frontend.yml       # Testes do frontend
│   ├── release.yml        # Releases automatizados
│   └── codeql.yml         # Análise de segurança CodeQL
├── ISSUE_TEMPLATE/
│   ├── bug_report.md      # Template para reportar bugs
│   └── feature_request.md # Template para solicitar funcionalidades
├── PULL_REQUEST_TEMPLATE.md
└── README.md              # Este arquivo
```

## 🚀 Workflows

### 1. Tests (`tests.yml`)
Executa em cada push e PR nas branches `main` e `develop`.

**O que faz**:
- ✅ Roda testes em Go 1.21 e 1.22
- ✅ Gera relatório de cobertura
- ✅ Executa linter (golangci-lint)
- ✅ Faz scan de segurança (Gosec)
- ✅ Compila binários

**Duração**: ~3-5 minutos

### 2. Docker Build (`docker.yml`)
Executa em pushes para `main` e tags `v*`.

**O que faz**:
- ✅ Constrói imagem Docker otimizada
- ✅ Publica no GitHub Container Registry
- ✅ Faz scan de vulnerabilidades (Trivy)
- ✅ Versiona automaticamente

**Duração**: ~5-8 minutos

### 3. Frontend Tests (`frontend.yml`)
Executa quando há mudanças na pasta `frontend/`.

**O que faz**:
- ✅ Testa em Node.js 18.x e 20.x
- ✅ Executa lint e type checking
- ✅ Faz build de produção
- ✅ Roda Lighthouse CI

**Duração**: ~2-4 minutos

### 4. Release (`release.yml`)
Executa quando uma tag `v*` é criada.

**O que faz**:
- ✅ Compila para múltiplas plataformas
- ✅ Gera checksums
- ✅ Cria release no GitHub
- ✅ Anexa binários automaticamente

**Duração**: ~8-12 minutos

### 5. CodeQL (`codeql.yml`)
Executa em pushes, PRs e semanalmente.

**O que faz**:
- ✅ Análise de segurança avançada
- ✅ Detecta vulnerabilidades
- ✅ Analisa Go e JavaScript

**Duração**: ~10-15 minutos

## 🔧 Como Usar

### Para Desenvolvedores

1. **Antes de fazer commit**:
   ```bash
   ./scripts/validate-ci.sh
   ```

2. **Criar branch e fazer mudanças**:
   ```bash
   git checkout -b feature/minha-feature
   # ... fazer mudanças ...
   git add .
   git commit -m "feat: adiciona nova funcionalidade"
   git push origin feature/minha-feature
   ```

3. **Abrir Pull Request**:
   - Use o template automático
   - Aguarde os checks passarem
   - Solicite revisão

### Para Releases

1. **Atualizar CHANGELOG.md**
2. **Criar tag**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
3. **Aguardar workflow completar**
4. **Verificar release criado**

## 📊 Status dos Workflows

Você pode ver o status de todos os workflows em:
- **Actions tab** no GitHub
- **Badges** no README principal
- **Notificações** por email

## 🛡️ Segurança

Múltiplas camadas de segurança:
- **Gosec**: Análise estática de Go
- **Trivy**: Scan de containers
- **CodeQL**: Análise semântica
- **Dependabot**: Atualizações de dependências (recomendado)

## 📝 Templates

### Pull Request
Use o template para garantir que todos os pontos importantes sejam cobertos:
- Descrição clara
- Tipo de mudança
- Como foi testado
- Checklist de qualidade
- Impacto

### Issues
Escolha o template apropriado:
- **Bug Report**: Para reportar problemas
- **Feature Request**: Para sugerir melhorias

## 🔗 Links Úteis

- [Documentação Completa](../GITHUB_ACTIONS.md)
- [Resumo CI/CD](../CI_CD_SUMMARY.md)
- [Resultados dos Testes](../TEST_RESULTS.md)
- [GitHub Actions Docs](https://docs.github.com/en/actions)

## 💡 Dicas

1. **Execute testes localmente** antes de push
2. **Use commits semânticos** (feat, fix, docs, etc)
3. **Mantenha PRs pequenos** e focados
4. **Revise os checks** antes de solicitar revisão
5. **Atualize documentação** junto com código

## 🆘 Problemas Comuns

### Testes falhando
```bash
# Execute localmente
cd backend
go test -v ./...
```

### Lint falhando
```bash
# Execute localmente
cd backend
golangci-lint run ./...
```

### Build Docker falhando
```bash
# Teste localmente
docker build -f docker/Dockerfile.api -t test .
```

## 📞 Suporte

Se você encontrar problemas com os workflows:
1. Verifique os logs no Actions tab
2. Execute validação local
3. Consulte a documentação
4. Abra uma issue se necessário

---

**Desenvolvido por**: Peder Munksgaard (JMPM Tecnologia)  
**Última atualização**: 2025-01-19
