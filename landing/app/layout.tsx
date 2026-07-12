import type { Metadata, Viewport } from 'next'
import { JetBrains_Mono } from 'next/font/google'
import { GeistPixelGrid } from 'geist/font/pixel'
import { ThemeProvider } from '@/components/theme-provider'
import { AuthProvider } from '@/lib/auth'

import './globals.css'

const jetbrainsMono = JetBrains_Mono({
  subsets: ['latin'],
  variable: '--font-mono',
})

export const metadata: Metadata = {
  title: 'Limiter.io - Distributed API Rate Limiting Platform',
  description:
    'Limiter.io is a high-performance, multi-tenant distributed rate limiting platform for APIs. Atomic Redis + Lua evaluation, five algorithms (token bucket, fixed & sliding window, sliding log, leaky bucket), a drop-in Go SDK with fail-open safety, Kafka-backed analytics, and a real-time WebSocket dashboard. Sub-millisecond evaluation latency.',
  keywords: [
    'API rate limiting',
    'distributed rate limiter',
    'token bucket',
    'sliding window rate limit',
    'leaky bucket algorithm',
    'Redis Lua rate limiting',
    'Go rate limiter SDK',
    'multi-tenant rate limiting',
    'API gateway throttling',
    'Kafka analytics pipeline',
    'real-time API dashboard',
    'developer infrastructure',
    'Gin middleware rate limit',
    'Next.js dashboard',
    'API quota management',
    'rate limiting as a service',
  ],
  authors: [{ name: 'Limiter.io' }],
  creator: 'Limiter.io',
  publisher: 'Limiter.io',
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  openGraph: {
    type: 'website',
    locale: 'en_US',
    title: 'Limiter.io — Distributed API Rate Limiting Platform',
    description:
      'Atomic Redis + Lua rate limiting with five algorithms, a fail-open Go SDK, Kafka-backed analytics, and a real-time WebSocket dashboard. Sub-millisecond evaluation latency.',
    siteName: 'Limiter.io',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Limiter.io — Distributed API Rate Limiting',
    description:
      'Protect your API with atomic Redis rate limiting, five algorithms, a drop-in Go SDK, and a live analytics dashboard.',
    creator: '@limiterio',
  },
  category: 'technology',
}

export const viewport: Viewport = {
  themeColor: '#F2F1EA',
  width: 'device-width',
  initialScale: 1,
  maximumScale: 5,
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en" className={`${jetbrainsMono.variable} ${GeistPixelGrid.variable}`} suppressHydrationWarning>
      <body className="font-mono antialiased">
        <ThemeProvider attribute="class" defaultTheme="light" enableSystem={false} disableTransitionOnChange>
          <AuthProvider>{children}</AuthProvider>
        </ThemeProvider>
      </body>
    </html>
  )
}
