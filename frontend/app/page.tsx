import Link from 'next/link'
import { ArrowRight, Shield, Zap, Globe } from 'lucide-react'

export default function Home() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-blue-50 to-white dark:from-gray-900 dark:to-gray-800">
      {/* Header */}
      <header className="container mx-auto px-4 py-6">
        <nav className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <div className="w-10 h-10 bg-blue-600 rounded-lg flex items-center justify-center">
              <span className="text-white font-bold text-xl">P</span>
            </div>
            <span className="text-2xl font-bold text-gray-900 dark:text-white">PIX SaaS</span>
          </div>
          <div className="flex items-center space-x-4">
            <Link 
              href="/login" 
              className="text-gray-600 hover:text-gray-900 dark:text-gray-300 dark:hover:text-white"
            >
              Login
            </Link>
            <Link 
              href="/dashboard" 
              className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition"
            >
              Dashboard
            </Link>
          </div>
        </nav>
      </header>

      {/* Hero Section */}
      <section className="container mx-auto px-4 py-20 text-center">
        <h1 className="text-5xl md:text-6xl font-bold text-gray-900 dark:text-white mb-6">
          Pagamentos PIX
          <br />
          <span className="text-blue-600">Simplificados</span>
        </h1>
        <p className="text-xl text-gray-600 dark:text-gray-300 mb-8 max-w-2xl mx-auto">
          Integre pagamentos PIX de múltiplos bancos brasileiros através de uma única API normalizada. 
          Seguro, rápido e escalável.
        </p>
        <div className="flex items-center justify-center space-x-4">
          <Link 
            href="/dashboard" 
            className="bg-blue-600 text-white px-8 py-4 rounded-lg hover:bg-blue-700 transition flex items-center space-x-2 text-lg font-semibold"
          >
            <span>Começar Agora</span>
            <ArrowRight className="w-5 h-5" />
          </Link>
          <Link 
            href="/docs" 
            className="border-2 border-gray-300 text-gray-700 dark:text-gray-300 dark:border-gray-600 px-8 py-4 rounded-lg hover:border-gray-400 transition text-lg font-semibold"
          >
            Documentação
          </Link>
        </div>
      </section>

      {/* Features */}
      <section className="container mx-auto px-4 py-20">
        <div className="grid md:grid-cols-3 gap-8">
          <div className="bg-white dark:bg-gray-800 p-8 rounded-xl shadow-lg">
            <div className="w-12 h-12 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center mb-4">
              <Shield className="w-6 h-6 text-blue-600 dark:text-blue-400" />
            </div>
            <h3 className="text-xl font-bold text-gray-900 dark:text-white mb-2">
              Segurança PCI DSS
            </h3>
            <p className="text-gray-600 dark:text-gray-300">
              Criptografia AES-256-GCM, JWT, auditoria completa e compliance com LGPD.
            </p>
          </div>

          <div className="bg-white dark:bg-gray-800 p-8 rounded-xl shadow-lg">
            <div className="w-12 h-12 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center mb-4">
              <Zap className="w-6 h-6 text-green-600 dark:text-green-400" />
            </div>
            <h3 className="text-xl font-bold text-gray-900 dark:text-white mb-2">
              Alta Performance
            </h3>
            <p className="text-gray-600 dark:text-gray-300">
              APIs otimizadas, rate limiting e pronto para escalar horizontalmente.
            </p>
          </div>

          <div className="bg-white dark:bg-gray-800 p-8 rounded-xl shadow-lg">
            <div className="w-12 h-12 bg-purple-100 dark:bg-purple-900 rounded-lg flex items-center justify-center mb-4">
              <Globe className="w-6 h-6 text-purple-600 dark:text-purple-400" />
            </div>
            <h3 className="text-xl font-bold text-gray-900 dark:text-white mb-2">
              Multi-banco
            </h3>
            <p className="text-gray-600 dark:text-gray-300">
              Suporte a 18+ instituições financeiras brasileiras com uma única integração.
            </p>
          </div>
        </div>
      </section>

      {/* Stats */}
      <section className="bg-blue-600 dark:bg-blue-800 py-16">
        <div className="container mx-auto px-4">
          <div className="grid md:grid-cols-4 gap-8 text-center">
            <div>
              <div className="text-4xl font-bold text-white mb-2">18+</div>
              <div className="text-blue-100">Bancos Suportados</div>
            </div>
            <div>
              <div className="text-4xl font-bold text-white mb-2">99.9%</div>
              <div className="text-blue-100">Uptime SLA</div>
            </div>
            <div>
              <div className="text-4xl font-bold text-white mb-2">&lt;2s</div>
              <div className="text-blue-100">Tempo de Resposta</div>
            </div>
            <div>
              <div className="text-4xl font-bold text-white mb-2">24/7</div>
              <div className="text-blue-100">Suporte</div>
            </div>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="container mx-auto px-4 py-20 text-center">
        <h2 className="text-4xl font-bold text-gray-900 dark:text-white mb-4">
          Pronto para começar?
        </h2>
        <p className="text-xl text-gray-600 dark:text-gray-300 mb-8">
          Crie sua conta e comece a processar pagamentos PIX em minutos.
        </p>
        <Link 
          href="/dashboard" 
          className="bg-blue-600 text-white px-8 py-4 rounded-lg hover:bg-blue-700 transition inline-flex items-center space-x-2 text-lg font-semibold"
        >
          <span>Acessar Dashboard</span>
          <ArrowRight className="w-5 h-5" />
        </Link>
      </section>

      {/* Footer */}
      <footer className="bg-gray-900 text-white py-12">
        <div className="container mx-auto px-4">
          <div className="grid md:grid-cols-4 gap-8">
            <div>
              <h4 className="font-bold mb-4">PIX SaaS</h4>
              <p className="text-gray-400 text-sm">
                Plataforma completa para pagamentos PIX no Brasil.
              </p>
            </div>
            <div>
              <h4 className="font-bold mb-4">Produto</h4>
              <ul className="space-y-2 text-sm text-gray-400">
                <li><Link href="/features">Features</Link></li>
                <li><Link href="/pricing">Preços</Link></li>
                <li><Link href="/docs">Documentação</Link></li>
              </ul>
            </div>
            <div>
              <h4 className="font-bold mb-4">Empresa</h4>
              <ul className="space-y-2 text-sm text-gray-400">
                <li><Link href="/about">Sobre</Link></li>
                <li><Link href="/blog">Blog</Link></li>
                <li><Link href="/contact">Contato</Link></li>
              </ul>
            </div>
            <div>
              <h4 className="font-bold mb-4">Legal</h4>
              <ul className="space-y-2 text-sm text-gray-400">
                <li><Link href="/privacy">Privacidade</Link></li>
                <li><Link href="/terms">Termos</Link></li>
                <li><Link href="/security">Segurança</Link></li>
              </ul>
            </div>
          </div>
          <div className="border-t border-gray-800 mt-8 pt-8 text-center text-sm text-gray-400">
            © 2024 PIX SaaS. Todos os direitos reservados.
          </div>
        </div>
      </footer>
    </div>
  )
}
