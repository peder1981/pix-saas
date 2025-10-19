# Guia de Resolução de Problemas Git

**Autor**: Peder Munksgaard (JMPM Tecnologia)  
**Data**: 2025-01-19

---

## 🔧 Problema Resolvido

### Situação Encontrada
- Branch local e remoto divergiram
- Arquivos modificados não commitados
- Arquivos novos não rastreados

### Solução Aplicada
```bash
# 1. Salvar mudanças temporariamente
git stash

# 2. Sincronizar com remoto usando rebase
git pull --rebase origin main

# 3. Restaurar mudanças salvas
git stash pop

# 4. Adicionar todos os arquivos
git add -A

# 5. Fazer commit
git commit -m "mensagem descritiva"

# 6. Enviar para o GitHub
git push origin main
```

---

## 📋 Problemas Comuns e Soluções

### 1. Branch Divergiu (Your branch and 'origin/main' have diverged)

**Problema**: Commits diferentes no local e no remoto

**Solução A - Rebase (Recomendado)**:
```bash
git stash                      # Salvar mudanças
git pull --rebase origin main  # Sincronizar com rebase
git stash pop                  # Restaurar mudanças
```

**Solução B - Merge**:
```bash
git pull origin main           # Fazer merge
# Resolver conflitos se houver
git add .
git commit -m "merge: Resolver conflitos"
```

**Solução C - Forçar (CUIDADO!)**:
```bash
# Só use se tiver certeza que quer sobrescrever o remoto
git push --force origin main
```

---

### 2. Arquivos Não Rastreados (Untracked files)

**Problema**: Arquivos novos não estão no Git

**Solução**:
```bash
# Adicionar arquivo específico
git add nome-do-arquivo.md

# Adicionar todos os arquivos
git add -A

# Ou
git add .
```

---

### 3. Mudanças Não Commitadas (Changes not staged)

**Problema**: Arquivos modificados mas não commitados

**Solução**:
```bash
# Ver o que mudou
git diff

# Adicionar mudanças
git add arquivo.md

# Ou adicionar tudo
git add -A

# Fazer commit
git commit -m "descrição das mudanças"
```

---

### 4. Desfazer Mudanças Locais

**Problema**: Quer descartar mudanças não commitadas

**Solução**:
```bash
# Descartar mudanças em arquivo específico
git restore arquivo.md

# Descartar todas as mudanças
git restore .

# Remover arquivos não rastreados
git clean -fd
```

---

### 5. Desfazer Último Commit (Mantendo Mudanças)

**Problema**: Commit errado, mas quer manter as mudanças

**Solução**:
```bash
# Desfazer último commit, mantendo mudanças
git reset --soft HEAD~1

# Editar arquivos
# ...

# Fazer novo commit
git commit -m "mensagem corrigida"
```

---

### 6. Desfazer Último Commit (Descartando Mudanças)

**Problema**: Commit completamente errado

**Solução**:
```bash
# CUIDADO: Isso apaga as mudanças!
git reset --hard HEAD~1
```

---

### 7. Conflitos de Merge

**Problema**: Conflitos ao fazer pull/merge

**Solução**:
```bash
# 1. Fazer pull
git pull origin main

# 2. Git mostrará arquivos com conflito
# Editar arquivos e resolver conflitos manualmente
# Procurar por: <<<<<<< HEAD, =======, >>>>>>> 

# 3. Após resolver
git add arquivo-resolvido.md

# 4. Continuar merge
git commit -m "merge: Resolver conflitos"
```

---

### 8. Esqueci de Fazer Pull Antes de Commitar

**Problema**: Fez commit local mas esqueceu de puxar mudanças do remoto

**Solução**:
```bash
# Opção 1: Rebase (mantém histórico limpo)
git pull --rebase origin main

# Opção 2: Merge
git pull origin main
```

---

### 9. Arquivo Grande Demais

**Problema**: Git rejeita arquivo > 100MB

**Solução**:
```bash
# Remover arquivo do staging
git rm --cached arquivo-grande.zip

# Adicionar ao .gitignore
echo "arquivo-grande.zip" >> .gitignore

# Usar Git LFS para arquivos grandes
git lfs install
git lfs track "*.zip"
git add .gitattributes
```

---

### 10. Credenciais Inválidas

**Problema**: Git pede senha toda hora ou rejeita credenciais

**Solução**:
```bash
# Configurar cache de credenciais
git config --global credential.helper cache

# Ou armazenar permanentemente (Linux)
git config --global credential.helper store

# Ou usar SSH ao invés de HTTPS
# 1. Gerar chave SSH
ssh-keygen -t ed25519 -C "seu-email@example.com"

# 2. Adicionar ao ssh-agent
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519

# 3. Copiar chave pública
cat ~/.ssh/id_ed25519.pub
# Adicionar no GitHub: Settings > SSH and GPG keys

# 4. Mudar remote para SSH
git remote set-url origin git@github.com:usuario/repo.git
```

---

## 🔍 Comandos Úteis para Diagnóstico

### Ver Status
```bash
git status                    # Status atual
git log --oneline -10         # Últimos 10 commits
git log --graph --oneline     # Histórico visual
git remote -v                 # Ver remotes configurados
```

### Ver Diferenças
```bash
git diff                      # Mudanças não staged
git diff --staged             # Mudanças staged
git diff HEAD                 # Todas as mudanças
git diff origin/main          # Diferença com remoto
```

### Ver Histórico
```bash
git log                       # Histórico completo
git log --oneline             # Histórico resumido
git log --author="Nome"       # Commits de um autor
git log --since="2 days ago"  # Commits recentes
```

### Ver Branches
```bash
git branch                    # Branches locais
git branch -r                 # Branches remotas
git branch -a                 # Todas as branches
```

---

## 🚀 Workflow Recomendado

### Antes de Começar a Trabalhar
```bash
# 1. Atualizar branch main
git checkout main
git pull origin main

# 2. Criar branch para feature
git checkout -b feature/nova-funcionalidade
```

### Durante o Desenvolvimento
```bash
# 1. Fazer mudanças
# ...

# 2. Ver o que mudou
git status
git diff

# 3. Adicionar mudanças
git add arquivo1.md arquivo2.go

# 4. Commit
git commit -m "feat: descrição da mudança"

# 5. Push da branch
git push origin feature/nova-funcionalidade
```

### Ao Finalizar
```bash
# 1. Atualizar main
git checkout main
git pull origin main

# 2. Voltar para feature branch
git checkout feature/nova-funcionalidade

# 3. Rebase com main
git rebase main

# 4. Resolver conflitos se houver

# 5. Push (pode precisar de --force após rebase)
git push origin feature/nova-funcionalidade --force-with-lease

# 6. Abrir Pull Request no GitHub
```

---

## 📝 Boas Práticas

### Mensagens de Commit
```bash
# Formato: tipo: descrição curta

# Tipos comuns:
feat:     # Nova funcionalidade
fix:      # Correção de bug
docs:     # Documentação
style:    # Formatação
refactor: # Refatoração
test:     # Testes
chore:    # Manutenção

# Exemplos:
git commit -m "feat: adiciona autorun de testes"
git commit -m "fix: corrige erro de compilação"
git commit -m "docs: atualiza README com instruções"
```

### Commits Atômicos
- Um commit = uma mudança lógica
- Commits pequenos e frequentes
- Facilita revisão e rollback

### Branches
- `main` - produção, sempre estável
- `develop` - desenvolvimento
- `feature/nome` - novas funcionalidades
- `fix/nome` - correções
- `hotfix/nome` - correções urgentes

---

## 🆘 Em Caso de Emergência

### Salvou Tudo Antes de Fazer Besteira?
```bash
# Criar backup da branch atual
git branch backup-$(date +%Y%m%d-%H%M%S)
```

### Fez Besteira e Quer Voltar?
```bash
# Ver histórico de comandos
git reflog

# Voltar para um ponto específico
git reset --hard HEAD@{5}
```

### Perdeu Commits?
```bash
# Ver todos os commits, incluindo "perdidos"
git reflog

# Recuperar commit
git cherry-pick <commit-hash>
```

---

## 🔗 Links Úteis

- [Git Documentation](https://git-scm.com/doc)
- [GitHub Guides](https://guides.github.com/)
- [Git Cheat Sheet](https://education.github.com/git-cheat-sheet-education.pdf)
- [Oh Shit, Git!?!](https://ohshitgit.com/)

---

## ✅ Checklist Antes de Push

- [ ] `git status` - Verificar o que vai ser commitado
- [ ] `git diff` - Revisar mudanças
- [ ] Testes passando (`./scripts/autorun-tests.sh`)
- [ ] Mensagem de commit descritiva
- [ ] Pull antes de push (se trabalhando em main)
- [ ] Resolver conflitos se houver

---

**Desenvolvido por**: Peder Munksgaard (JMPM Tecnologia)  
**Última atualização**: 2025-01-19
