"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { Navbar } from "@/components/navbar"
import { Footer } from "@/components/footer"
import { BookOpen, Code2, Link, FileText, ArrowRight } from "lucide-react"

const ease = [0.22, 1, 0.36, 1] as const

const RESOURCES = [
  {
    title: "Official SDK Integrations",
    desc: "Complete documentation and sample code for our Go, Node.js, and Python client SDKs featuring client-side fail-open logic.",
    link: "/docs/INTEGRATION_GUIDE.md",
    icon: Code2
  },
  {
    title: "Edge Alignment Blueprint",
    desc: "A comprehensive guide on deploying rate limits at the network edge via Cloudflare Workers and geodistributed Upstash Redis replica nodes.",
    link: "/docs/EDGE_ALIGNMENT_GUIDE.md",
    icon: Link
  },
  {
    title: "Advanced WAF & Failover Plans",
    desc: "Learn how to configure JWT origin authentication shields, block IP CIDR ranges, restrict traffic by country, and implement local memory cache fallbacks.",
    link: "/docs/ADVANCED_WAF_FAILOVER.md",
    icon: FileText
  }
]

const ENDPOINTS = [
  { method: "POST", path: "/auth/register", desc: "Create developer account" },
  { method: "POST", path: "/auth/login", desc: "Acquire API access and refresh tokens" },
  { method: "GET", path: "/projects", desc: "List associated tenant projects" },
  { method: "POST", path: "/projects/:projectId/keys", desc: "Provision new rate limit API keys" },
  { method: "GET", path: "/gateway/*path", desc: "Evaluate rate limit rules for key + route" }
]

export default function ResourcesPage() {
  return (
    <div className="min-h-screen dot-grid-bg">
      <Navbar />

      <main className="w-full px-6 py-12 lg:px-12">
        {/* Page Header */}
        <div className="mb-12">
          <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
            {"// DEVELOPER_RESOURCES / INTEGRATION_CENTRAL"}
          </span>
          <h1 className="mt-2 text-3xl lg:text-5xl font-mono font-bold uppercase tracking-tight">
            GUIDES, SDKS AND <span className="text-[#ea580c]">API</span> SPECS
          </h1>
          <p className="mt-4 max-w-2xl text-xs lg:text-sm font-mono text-muted-foreground leading-relaxed">
            Provision API keys, configure rate limiting rules, and integrate our clients in minutes. Access raw guides,
            quickstarts, and API specifications.
          </p>
        </div>

        {/* Guides List */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-12">
          {RESOURCES.map((res, i) => (
            <motion.div
              key={res.title}
              initial={{ opacity: 0, y: 12 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.05, duration: 0.5, ease }}
              className="border-2 border-foreground p-5 bg-background flex flex-col justify-between"
            >
              <div className="flex flex-col gap-3">
                <res.icon size={16} className="text-[#ea580c]" />
                <h3 className="text-sm font-mono font-bold uppercase tracking-wider">{res.title}</h3>
                <p className="text-xs font-mono text-muted-foreground leading-relaxed">{res.desc}</p>
              </div>

              <div className="mt-6 border-t border-foreground/10 pt-4">
                <a 
                  href={res.link}
                  className="inline-flex items-center gap-1 text-[11px] font-mono uppercase text-[#ea580c] hover:text-[#c2410c] transition-colors"
                >
                  Read doc <ArrowRight size={10} />
                </a>
              </div>
            </motion.div>
          ))}
        </div>

        {/* API Specification Directory */}
        <div className="border-2 border-foreground p-6 bg-background/50 backdrop-blur-sm">
          <div className="flex items-center justify-between border-b-2 border-foreground pb-4 mb-6">
            <span className="text-xs font-mono font-bold uppercase flex items-center gap-2">
              <BookOpen size={14} className="text-[#ea580c]" />
              02. Gateway Core API Reference
            </span>
            <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
              JSON_API_SPEC
            </span>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse font-mono text-xs">
              <thead>
                <tr className="border-b-2 border-foreground text-[10px] uppercase text-muted-foreground tracking-wider bg-foreground/5">
                  <th className="p-3">Method</th>
                  <th className="p-3">Endpoint Path</th>
                  <th className="p-3">Operation description</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-foreground/10">
                {ENDPOINTS.map((endpoint, i) => (
                  <tr key={i} className="hover:bg-muted/5">
                    <td className="p-3">
                      <span className={`px-2 py-0.5 font-bold border text-[9px] ${
                        endpoint.method === "POST" ? "text-green-500 border-green-500/20 bg-green-500/5" :
                        endpoint.method === "GET" ? "text-blue-500 border-blue-500/20 bg-blue-500/5" :
                        "text-foreground border-foreground/20"
                      }`}>
                        {endpoint.method}
                      </span>
                    </td>
                    <td className="p-3 font-semibold">{endpoint.path}</td>
                    <td className="p-3 text-muted-foreground">{endpoint.desc}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </main>

      <Footer />
    </div>
  )
}
