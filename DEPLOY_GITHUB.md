# Deploy para GitHub

## 📤 Push para o Repositório

### 1. Verificar Status

```bash
git status
git log --oneline -5
```

### 2. Push para GitHub

```bash
# Push inicial
git push -u origin main

# Pushes subsequentes
git push
```

### 3. Verificar no GitHub

Acesse: https://github.com/peder1981/pix-saas

## 🔒 Antes do Push - Checklist de Segurança

- [ ] Remover credenciais do código
- [ ] Verificar .gitignore
- [ ] Não commitar arquivos .env
- [ ] Não commitar certificados
- [ ] Não commitar chaves privadas

## 📋 Commits Realizados

1. **feat: Fase 1 e 2 - Fundação do SaaS PIX**
   - Estrutura Clean Architecture
   - Modelos de domínio
   - Providers (Bradesco, Itaú)
   - Segurança e auditoria

2. **feat: Fase 2, 3 e 5 - API completa, CLI e DevOps**
   - TransactionHandler
   - OpenAPI documentation
   - Docker Compose
   - CLI administrativa

3. **docs: Guia de instalação completo e estrutura frontend**
   - INSTALL.md
   - Frontend package.json
   - Instruções detalhadas

4. **docs: Resumo executivo completo do projeto**
   - SUMMARY.md
   - Visão geral
   - Métricas e roadmap

## 🌐 Configurar GitHub Pages (Opcional)

Para hospedar a documentação:

```bash
# Criar branch gh-pages
git checkout -b gh-pages

# Copiar documentação
mkdir docs-site
cp README.md docs-site/index.md
cp INSTALL.md docs-site/
cp SUMMARY.md docs-site/

# Commit e push
git add docs-site
git commit -m "docs: GitHub Pages"
git push origin gh-pages

# Voltar para main
git checkout main
```

Depois, nas configurações do repositório:
- Settings > Pages
- Source: Deploy from branch
- Branch: gh-pages / docs-site

## 🏷️ Criar Release

```bash
# Criar tag
git tag -a v1.0.0 -m "Release v1.0.0 - MVP Completo"

# Push tag
git push origin v1.0.0
```

No GitHub:
1. Ir em "Releases"
2. "Create a new release"
3. Escolher tag v1.0.0
4. Título: "v1.0.0 - MVP Completo"
5. Descrição: Copiar de SUMMARY.md

## 📝 Configurar README Badges

Adicionar ao README.md:

```markdown
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-Proprietary-red)
![Status](https://img.shields.io/badge/Status-MVP-green)
![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen)
```

## 🔐 Secrets do GitHub

Configurar secrets para CI/CD:
- Settings > Secrets and variables > Actions
- Adicionar:
  - `JWT_SECRET_KEY`
  - `ENCRYPTION_KEY`
  - `DATABASE_URL`

## 🤖 GitHub Actions (Futuro)

Criar `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: cd backend && go test ./...
```

## 📊 Configurar Issues e Projects

1. **Issues Templates**
   - Bug report
   - Feature request
   - Documentation

2. **Projects**
   - Criar board Kanban
   - Colunas: Backlog, In Progress, Review, Done

3. **Labels**
   - bug
   - enhancement
   - documentation
   - security
   - performance

## 🎯 Próximos Passos

Após o push:

1. [ ] Verificar repositório no GitHub
2. [ ] Criar primeira release (v1.0.0)
3. [ ] Configurar GitHub Actions
4. [ ] Adicionar badges ao README
5. [ ] Criar issues para próximas features
6. [ ] Configurar branch protection rules
7. [ ] Adicionar CONTRIBUTING.md
8. [ ] Configurar code owners

## 📞 Suporte

Se tiver problemas com o push:

```bash
# Verificar remote
git remote -v

# Reconfigurar se necessário
git remote set-url origin https://github.com/peder1981/pix-saas.git

# Forçar push (cuidado!)
git push -f origin main
```

## ✅ Checklist Final

- [x] Código commitado
- [x] .gitignore configurado
- [x] README.md completo
- [x] Documentação criada
- [ ] Push para GitHub
- [ ] Verificar no navegador
- [ ] Criar release
- [ ] Compartilhar com equipe
