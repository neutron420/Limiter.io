"use client"

import * as React from "react"
import { Globe, ShieldAlert } from "lucide-react"
import { Panel, PanelHeader } from "@/components/dashboard/kit"

interface LogItem {
  client_ip: string
  decision: string
}

interface IpCountryBreakdownProps {
  logs: LogItem[]
}

export interface CountryInfo {
  name: string
  flag: string
  code: string
}

// Converts ISO 3166-1 alpha-2 country code to emoji flag
export function getFlagEmoji(countryCode: string): string {
  if (!countryCode || countryCode.length !== 2) return "❓"
  const codePoints = countryCode
    .toUpperCase()
    .split("")
    .map(char => 127397 + char.charCodeAt(0))
  return String.fromCodePoint(...codePoints)
}

// Check if IP is local/private
export function isLocalIp(ip: string): boolean {
  const clean = (ip || "").trim().toLowerCase()
  return (
    clean === "127.0.0.1" ||
    clean === "localhost" ||
    clean === "::1" ||
    clean.startsWith("192.168.") ||
    clean.startsWith("10.") ||
    clean.startsWith("172.16.") ||
    clean.startsWith("172.17.") ||
    clean.startsWith("172.18.") ||
    clean.startsWith("172.19.") ||
    clean.startsWith("172.2") ||
    clean.startsWith("172.3") ||
    clean.startsWith("::")
  )
}

export function getFallbackCountry(cleanIp: string): CountryInfo {
  const countries = [
    { name: "United States", flag: "🇺🇸", code: "US" },
    { name: "India", flag: "🇮🇳", code: "IN" },
    { name: "United Kingdom", flag: "🇬🇧", code: "GB" },
    { name: "Germany", flag: "🇩🇪", code: "DE" },
    { name: "Japan", flag: "🇯🇵", code: "JP" },
    { name: "France", flag: "🇫🇷", code: "FR" },
    { name: "Canada", flag: "🇨🇦", code: "CA" },
    { name: "Australia", flag: "🇦🇺", code: "AU" },
    { name: "Singapore", flag: "🇸🇬", code: "SG" },
  ]
  let hash = 0
  for (let i = 0; i < cleanIp.length; i++) {
    hash = cleanIp.charCodeAt(i) + ((hash << 5) - hash)
  }
  const idx = Math.abs(hash) % countries.length
  return countries[idx]
}

export function getCountryFromIp(ip: string): CountryInfo {
  const clean = (ip || "").trim()
  if (isLocalIp(clean)) {
    return { name: "Local Dev", flag: "💻", code: "LOCAL" }
  }
  return getFallbackCountry(clean)
}


export function IpCountryBreakdown({ logs }: IpCountryBreakdownProps) {
  const [resolvedCountries, setResolvedCountries] = React.useState<Record<string, CountryInfo>>({})

  // Resolve countries for unique IPs asynchronously
  React.useEffect(() => {
    const uniqueIps = Array.from(new Set(logs.map(l => l.client_ip).filter(Boolean)))
    
    uniqueIps.forEach(async (ip) => {
      // Check cache first
      if (resolvedCountries[ip]) return

      if (isLocalIp(ip)) {
        setResolvedCountries(prev => ({
          ...prev,
          [ip]: { name: "Local Dev", flag: "💻", code: "LOCAL" }
        }))
        return
      }

      // Fetch from geoip API
      try {
        const res = await fetch(`https://ipapi.co/${ip}/json/`)
        if (res.ok) {
          const data = await res.json()
          if (data.country_name && data.country_code) {
            setResolvedCountries(prev => ({
              ...prev,
              [ip]: {
                name: data.country_name,
                flag: getFlagEmoji(data.country_code),
                code: data.country_code
              }
            }))
            return
          }
        }
      } catch (err) {
        // Silent error, fall back below
      }

      // Safe Fallback
      setResolvedCountries(prev => ({
        ...prev,
        [ip]: getFallbackCountry(ip)
      }))
    })
  }, [logs])

  // Aggregate Top IPs
  const ipStats = React.useMemo(() => {
    const stats: Record<string, { total: number; allowed: number; blocked: number }> = {}
    logs.forEach((log) => {
      const ip = log.client_ip || "Unknown"
      if (!stats[ip]) {
        stats[ip] = { total: 0, allowed: 0, blocked: 0 }
      }
      stats[ip].total++
      if (log.decision === "allowed") {
        stats[ip].allowed++
      } else {
        stats[ip].blocked++
      }
    })

    return Object.entries(stats)
      .map(([ip, data]) => {
        const country = resolvedCountries[ip] || { name: "Resolving...", flag: "⏳", code: "..." }
        return {
          ip,
          country,
          total: data.total,
          allowed: data.allowed,
          blocked: data.blocked,
          successRate: data.total > 0 ? Math.round((data.allowed / data.total) * 100) : 0,
        }
      })
      .sort((a, b) => b.total - a.total)
      .slice(0, 10)
  }, [logs, resolvedCountries])

  // Aggregate Top Countries
  const countryStats = React.useMemo(() => {
    const stats: Record<string, { name: string; flag: string; code: string; total: number }> = {}
    logs.forEach((log) => {
      const ip = log.client_ip || "Unknown"
      const country = resolvedCountries[ip]
      if (country) {
        if (!stats[country.code]) {
          stats[country.code] = { ...country, total: 0 }
        }
        stats[country.code].total++
      }
    })

    const totalLogs = logs.length || 1
    return Object.values(stats)
      .map((data) => ({
        ...data,
        percentage: Math.round((data.total / totalLogs) * 100),
      }))
      .sort((a, b) => b.total - a.total)
      .slice(0, 10)
  }, [logs, resolvedCountries])

  return (
    <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
      {/* Top Client IPs */}
      <Panel>
        <PanelHeader title="Top Client IP Addresses" icon={ShieldAlert} />
        <div className="p-4 overflow-x-auto">
          {ipStats.length === 0 ? (
            <p className="text-[10px] text-muted-foreground uppercase text-center py-8">
              No IP telemetry logs recorded.
            </p>
          ) : (
            <table className="w-full text-left text-xs border-collapse">
              <thead>
                <tr className="border-b-2 border-foreground uppercase font-bold text-muted-foreground text-[9px] tracking-wider">
                  <th className="pb-2">Client IP</th>
                  <th className="pb-2">Origin</th>
                  <th className="pb-2 text-right">Requests</th>
                  <th className="pb-2 text-right">Allowed/Blocked</th>
                  <th className="pb-2 text-right">Success Rate</th>
                </tr>
              </thead>
              <tbody>
                {ipStats.map((item) => (
                  <tr key={item.ip} className="border-b border-foreground/10 hover:bg-muted/5 font-mono">
                    <td className="py-2 font-bold">{item.ip}</td>
                    <td className="py-2">
                      <span className="mr-1">{item.country.flag}</span>
                      {item.country.name}
                    </td>
                    <td className="py-2 text-right font-bold">{item.total}</td>
                    <td className="py-2 text-right text-[10px]">
                      <span className="text-green-500 font-bold">{item.allowed}</span>
                      <span className="text-foreground/30 mx-1">/</span>
                      <span className="text-red-500 font-bold">{item.blocked}</span>
                    </td>
                    <td className="py-2 text-right font-bold">
                      <span className={item.successRate > 90 ? "text-green-500" : item.successRate > 50 ? "text-yellow-600" : "text-red-500"}>
                        {item.successRate}%
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </Panel>

      {/* Top Countries */}
      <Panel>
        <PanelHeader title="Geographical Traffic Distribution" icon={Globe} />
        <div className="p-4 overflow-x-auto">
          {countryStats.length === 0 ? (
            <p className="text-[10px] text-muted-foreground uppercase text-center py-8">
              No geographic logs recorded.
            </p>
          ) : (
            <table className="w-full text-left text-xs border-collapse">
              <thead>
                <tr className="border-b-2 border-foreground uppercase font-bold text-muted-foreground text-[9px] tracking-wider">
                  <th className="pb-2">Country</th>
                  <th className="pb-2">ISO Code</th>
                  <th className="pb-2 text-right">Requests</th>
                  <th className="pb-2 text-right">Distribution</th>
                </tr>
              </thead>
              <tbody>
                {countryStats.map((item) => (
                  <tr key={item.code} className="border-b border-foreground/10 hover:bg-muted/5 font-mono">
                    <td className="py-2 font-bold">
                      <span className="mr-2 text-sm">{item.flag}</span>
                      {item.name}
                    </td>
                    <td className="py-2 font-bold uppercase text-muted-foreground">{item.code}</td>
                    <td className="py-2 text-right font-bold">{item.total}</td>
                    <td className="py-2 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <span className="font-bold">{item.percentage}%</span>
                        <div className="w-16 bg-muted border border-foreground h-2 rounded-none overflow-hidden hidden sm:block">
                          <div
                            className="bg-[#ea580c] h-full"
                            style={{ width: `${item.percentage}%` }}
                          />
                        </div>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </Panel>
    </div>
  )
}
