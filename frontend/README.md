# PIX SaaS Dashboard

Dashboard Next.js 14 para gerenciamento de transações PIX.

## 🚀 Tecnologias

- **Next.js 14** - App Router
- **React 18** - UI
- **TypeScript** - Type safety
- **TailwindCSS** - Styling
- **shadcn/ui** - Componentes
- **React Query** - Data fetching
- **Zustand** - State management
- **Recharts** - Gráficos
- **Lucide** - Ícones

## 📦 Instalação

```bash
cd frontend
npm install
```

## 🏃 Executar

```bash
# Desenvolvimento
npm run dev

# Build
npm run build

# Produção
npm start
```

## 🎨 Estrutura

```
frontend/
├── app/              # App Router (Next.js 14)
│   ├── (auth)/      # Rotas de autenticação
│   ├── (dashboard)/ # Rotas do dashboard
│   ├── layout.tsx   # Layout raiz
│   └── page.tsx     # Página inicial
├── components/       # Componentes React
│   ├── ui/          # Componentes base (shadcn/ui)
│   ├── charts/      # Gráficos
│   └── forms/       # Formulários
├── lib/             # Utilitários
│   ├── api.ts       # Cliente API
│   ├── auth.ts      # Autenticação
│   └── utils.ts     # Helpers
├── hooks/           # Custom hooks
├── store/           # Zustand stores
├── types/           # TypeScript types
└── public/          # Assets estáticos
```

## 🔐 Autenticação

O dashboard usa JWT tokens armazenados em cookies httpOnly.

## 📊 Features

- ✅ Dashboard com métricas
- ✅ Listagem de transações
- ✅ Filtros e busca
- ✅ Gráficos de volume
- ✅ Gerenciamento de API keys
- ✅ Configuração de webhooks
- ✅ Logs de auditoria
- ✅ Tema claro/escuro

## 🌐 Variáveis de Ambiente

Crie um arquivo `.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/v1
```

## 📝 Licença

Proprietary
