# 🚀 Guia de Configuração do GitHub

Este guia explica como configurar o repositório no GitHub para ativar todos os workflows de CI/CD.

## 📋 Pré-requisitos

- Conta no GitHub
- Git instalado localmente
- Repositório criado no GitHub

## 🔧 Passo a Passo

### 1. Criar Repositório no GitHub

1. Acesse [GitHub](https://github.com)
2. Clique em **New repository**
3. Configure:
   - **Nome**: `pix-saas` (ou nome de sua escolha)
   - **Descrição**: "Plataforma SaaS para pagamentos PIX"
   - **Visibilidade**: Private ou Public
   - **NÃO** inicialize com README (já temos um)
4. Clique em **Create repository**

### 2. Configurar Git Local

```bash
# Navegar para o diretório do projeto
cd /home/peder/CascadeProjects/windsurf-project/pix-saas

# Inicializar git (se ainda não foi feito)
git init

# Adicionar remote
git remote add origin https://github.com/SEU_USUARIO/pix-saas.git

# Ou se preferir SSH
git remote add origin git@github.com:SEU_USUARIO/pix-saas.git
```

### 3. Atualizar Badges no README

Edite o arquivo `README.md` e substitua `YOUR_USERNAME` pelo seu usuário do GitHub:

```markdown
[![Tests](https://github.com/SEU_USUARIO/pix-saas/actions/workflows/tests.yml/badge.svg)]
[![Docker Build](https://github.com/SEU_USUARIO/pix-saas/actions/workflows/docker.yml/badge.svg)]
...
```

### 4. Fazer o Primeiro Commit

```bash
# Adicionar todos os arquivos
git add .

# Fazer commit
git commit -m "feat: implementação inicial da plataforma PIX SaaS

- Backend completo em Go com Clean Architecture
- Frontend em Next.js 14
- CLI administrativa
- 33 testes unitários
- 5 workflows de CI/CD
- Documentação completa"

# Push para o GitHub
git push -u origin main
```

### 5. Configurar Permissões dos Workflows

1. Acesse seu repositório no GitHub
2. Vá em **Settings** → **Actions** → **General**
3. Em **Workflow permissions**, selecione:
   - ✅ **Read and write permissions**
   - ✅ **Allow GitHub Actions to create and approve pull requests**
4. Clique em **Save**

### 6. Habilitar GitHub Container Registry

Para publicar imagens Docker:

1. Vá em **Settings** → **Packages**
2. Configure visibilidade dos packages
3. As imagens serão publicadas em `ghcr.io/SEU_USUARIO/pix-saas-api`

### 7. Configurar Secrets (Opcional)

Para funcionalidades avançadas:

1. Vá em **Settings** → **Secrets and variables** → **Actions**
2. Clique em **New repository secret**
3. Adicione (opcional):
   - `CODECOV_TOKEN` - Para upload de cobertura no Codecov
   - Outros secrets conforme necessário

### 8. Habilitar GitHub Security

1. Vá em **Settings** → **Code security and analysis**
2. Habilite:
   - ✅ **Dependency graph**
   - ✅ **Dependabot alerts**
   - ✅ **Dependabot security updates**
   - ✅ **Code scanning** (CodeQL já configurado)
   - ✅ **Secret scanning**

### 9. Configurar Branch Protection (Recomendado)

1. Vá em **Settings** → **Branches**
2. Clique em **Add rule**
3. Configure para a branch `main`:
   - ✅ **Require a pull request before merging**
   - ✅ **Require status checks to pass before merging**
   - Selecione os checks obrigatórios:
     - `test`
     - `lint`
     - `build`
     - `security`
   - ✅ **Require branches to be up to date before merging**
   - ✅ **Include administrators**
4. Clique em **Create**

### 10. Verificar Workflows

1. Vá na aba **Actions**
2. Você verá os workflows executando automaticamente
3. Clique em cada workflow para ver detalhes
4. Aguarde todos os checks passarem ✅

## 🎯 Verificação

Após configurar, você deve ver:

### Na aba Actions
- ✅ Tests workflow executando
- ✅ Docker Build workflow executando (se push na main)
- ✅ CodeQL workflow executando

### Na aba Security
- ✅ Code scanning alerts (CodeQL)
- ✅ Dependabot alerts
- ✅ Secret scanning

### No README
- ✅ Badges mostrando status dos workflows

## 🔄 Fluxo de Trabalho

### Para Desenvolvimento

1. **Criar branch**:
   ```bash
   git checkout -b feature/nova-funcionalidade
   ```

2. **Fazer mudanças e testar localmente**:
   ```bash
   ./scripts/validate-ci.sh
   ```

3. **Commit e push**:
   ```bash
   git add .
   git commit -m "feat: adiciona nova funcionalidade"
   git push origin feature/nova-funcionalidade
   ```

4. **Abrir Pull Request**:
   - No GitHub, clique em **Compare & pull request**
   - Preencha o template
   - Aguarde os checks passarem
   - Solicite revisão

5. **Merge**:
   - Após aprovação, faça merge
   - Delete a branch

### Para Releases

1. **Atualizar CHANGELOG.md**:
   ```bash
   # Adicionar entrada para nova versão
   vim CHANGELOG.md
   git add CHANGELOG.md
   git commit -m "chore: prepare release v1.0.0"
   git push
   ```

2. **Criar tag**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

3. **Aguardar workflow**:
   - O workflow `release.yml` será executado
   - Binários serão compilados
   - Release será criado automaticamente

4. **Verificar release**:
   - Vá na aba **Releases**
   - Verifique binários anexados
   - Publique release notes

## 🐳 Usar Imagens Docker

Após o build, as imagens estarão disponíveis:

```bash
# Pull da imagem
docker pull ghcr.io/SEU_USUARIO/pix-saas-api:main

# Executar
docker run -p 8080:8080 ghcr.io/SEU_USUARIO/pix-saas-api:main
```

## 📊 Codecov (Opcional)

Para visualizar cobertura de código:

1. Acesse [Codecov.io](https://codecov.io)
2. Faça login com GitHub
3. Adicione o repositório
4. Copie o token
5. Adicione como secret `CODECOV_TOKEN`
6. Badge será atualizado automaticamente

## 🔍 Monitoramento

### Status dos Workflows
- **GitHub Actions tab**: Status em tempo real
- **Badges no README**: Status visual
- **Email**: Notificações de falhas (configurável)

### Métricas
- **Actions tab**: Tempo de execução, taxa de sucesso
- **Insights**: Estatísticas do repositório
- **Security tab**: Vulnerabilidades detectadas

## 🆘 Troubleshooting

### Workflows não executam

**Problema**: Workflows não aparecem na aba Actions

**Solução**:
1. Verifique se os arquivos estão em `.github/workflows/`
2. Verifique permissões em Settings → Actions
3. Faça um novo push para trigger

### Permissões negadas

**Problema**: Workflow falha com erro de permissão

**Solução**:
1. Vá em Settings → Actions → General
2. Habilite "Read and write permissions"
3. Re-execute o workflow

### Docker push falha

**Problema**: Não consegue fazer push da imagem

**Solução**:
1. Verifique se GHCR está habilitado
2. Verifique permissões do workflow
3. Verifique se está na branch main ou tag

### Badges não aparecem

**Problema**: Badges mostram "unknown"

**Solução**:
1. Aguarde primeiro workflow completar
2. Verifique se o nome do workflow está correto
3. Atualize a página (cache)

## 📚 Recursos Adicionais

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Branch Protection Rules](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/defining-the-mergeability-of-pull-requests/about-protected-branches)
- [Codecov Documentation](https://docs.codecov.com/)

## ✅ Checklist de Configuração

- [ ] Repositório criado no GitHub
- [ ] Git remote configurado
- [ ] Badges atualizados no README
- [ ] Primeiro commit feito
- [ ] Permissões dos workflows configuradas
- [ ] GitHub Container Registry habilitado
- [ ] GitHub Security habilitado
- [ ] Branch protection configurado (opcional)
- [ ] Workflows executando com sucesso
- [ ] Documentação revisada

## 🎉 Pronto!

Seu repositório está configurado e pronto para desenvolvimento colaborativo com CI/CD automatizado!

---

**Desenvolvido por**: Peder Munksgaard (JMPM Tecnologia)  
**Última atualização**: 2025-01-19  
**Versão**: 1.0.0
