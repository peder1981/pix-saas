#!/bin/bash

# Script para fazer push do projeto PIX SaaS para GitHub
# Autor: Peder Munksgaard
# Data: 19/10/2025

echo "🚀 PIX SaaS - Push para GitHub"
echo "================================"
echo ""

# Verificar se estamos no diretório correto
if [ ! -f "README.md" ]; then
    echo "❌ Erro: Execute este script na raiz do projeto"
    exit 1
fi

# Verificar status do git
echo "📊 Status do repositório:"
git status --short
echo ""

# Mostrar commits
echo "📝 Commits a serem enviados:"
git log --oneline -7
echo ""

# Confirmar push
read -p "Deseja fazer push para GitHub? (s/N) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Ss]$ ]]; then
    echo "❌ Push cancelado"
    exit 0
fi

# Fazer push
echo ""
echo "📤 Fazendo push para GitHub..."
git push -u origin main

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Push realizado com sucesso!"
    echo ""
    echo "🎉 Projeto disponível em:"
    echo "   https://github.com/peder1981/pix-saas"
    echo ""
    echo "📚 Próximos passos:"
    echo "   1. Verificar repositório no GitHub"
    echo "   2. Criar primeira release (v1.0.0)"
    echo "   3. Configurar GitHub Actions"
    echo "   4. Adicionar badges ao README"
    echo "   5. Executar: cd backend && go mod download"
    echo "   6. Testar API localmente"
    echo ""
else
    echo ""
    echo "❌ Erro ao fazer push"
    echo "   Verifique suas credenciais e conexão"
    exit 1
fi
