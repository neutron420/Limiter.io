"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { Navbar } from "@/components/navbar"
import { Footer } from "@/components/footer"
import { ShieldAlert, Globe, Server, CheckCircle2, Star, Mail } from "lucide-react"

const ease = [0.22, 1, 0.36, 1] as const

const ENTERPRISE_FEATURES = [
  {
    title: "Dedicated Edge Deployments",
    desc: "Deploy isolated rate limit nodes co-located with your servers. Reached globally via Anycast routing for absolute minimum network latency.",
    icon: Server
  },
  {
    title: "Dynamic SLA Commitments",
    desc: "Legally backed SLA uptime guarantees of up to 99.999% with dedicated site reliability engineer coverage on-call 24/7.",
    icon: CheckCircle2
  },
  {
    title: "VPC Peering & Cryptographic Shields",
    desc: "Route all configuration and execution traffic securely through AWS VPC Peering, GCP Interconnect, or private Cloudflare Tunnels.",
    icon: Globe
  },
  {
    title: "Compliance & Audit Vault",
    desc: "Access logs mapped directly for SOC2 Type II, GDPR, and PCI-DSS compliance audits. Secure, write-once storage with immediate export.",
    icon: ShieldAlert
  }
]

export default function EnterprisePage() {
  return (
    <div className="min-h-screen dot-grid-bg">
      <Navbar />

      <main className="w-full px-6 py-12 lg:px-12">
        {/* Page Header */}
        <div className="mb-12">
          <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
            {"// ENTERPRISE / COMPLIANCE_GUARANTEE"}
          </span>
          <h1 className="mt-2 text-3xl lg:text-5xl font-mono font-bold uppercase tracking-tight">
            INFRASTRUCTURE AT <span className="text-[#ea580c]">GLOBAL</span> SCALE
          </h1>
          <p className="mt-4 max-w-2xl text-xs lg:text-sm font-mono text-muted-foreground leading-relaxed">
            Enterprise demands isolation, speed, and reliability. Limiter.io provides single-tenant setups,
            unlimited execution streams, custom SLA rules, and multi-region synchronization.
          </p>
        </div>

        {/* Feature Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-12">
          {ENTERPRISE_FEATURES.map((feature, i) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, x: i % 2 === 0 ? -20 : 20 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.05, duration: 0.5, ease }}
              className="border-2 border-foreground p-6 bg-background/50 backdrop-blur-sm flex flex-col justify-between"
            >
              <div className="flex items-start gap-4">
                <div className="border border-foreground/20 p-2 bg-foreground/5 text-[#ea580c]">
                  <feature.icon size={16} />
                </div>
                <div className="flex-1">
                  <h3 className="text-sm font-mono font-bold uppercase tracking-wider">{feature.title}</h3>
                  <p className="mt-2 text-xs font-mono text-muted-foreground leading-relaxed">{feature.desc}</p>
                </div>
              </div>
            </motion.div>
          ))}
        </div>

        {/* Custom Rules SLA Block */}
        <div className="border-2 border-foreground p-6 mb-12 bg-background">
          <div className="flex items-center justify-between border-b-2 border-foreground pb-4 mb-6">
            <span className="text-xs font-mono font-bold uppercase flex items-center gap-2">
              <Star size={14} className="text-[#ea580c]" />
              Enterprise custom rules engine
            </span>
            <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
              VPC_LUA_COMPILATION
            </span>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <div className="flex flex-col gap-4">
              <p className="text-xs font-mono text-muted-foreground leading-relaxed">
                Standard rate limiters bind you to rigid algorithm constraints. Our Enterprise layer features a **Custom Sandboxed Lua Engine**.
                You can write and deploy custom rate logic directly from the console, syncing to edge nodes in under 2 seconds.
              </p>
              <p className="text-xs font-mono text-muted-foreground leading-relaxed">
                Perfect for complex pricing models (e.g. rate limit based on payload size, query parameters, user tier status, or API token balance).
              </p>
            </div>
            
            <div className="bg-foreground text-background p-4 font-mono text-[10px] overflow-x-auto border border-foreground/10 flex flex-col justify-between">
              <span className="text-background/40 uppercase tracking-widest text-[8px] mb-2">CUSTOM_RECONCILE.lua</span>
              <pre><code>{`local balance = redis.call("GET", KEYS[1])
local weight = tonumber(ARGV[1])
if balance and tonumber(balance) >= weight then
  redis.call("DEBY", KEYS[1], weight)
  return 1
end
return 0`}</code></pre>
            </div>
          </div>
        </div>

        {/* Contact form / CTA block */}
        <div className="border-2 border-foreground bg-foreground p-6 text-background text-center lg:py-12">
          <h2 className="text-xl lg:text-3xl font-mono font-bold uppercase tracking-widest mb-4">
            DISCUSS TECHNICAL REQUIREMENTS
          </h2>
          <p className="max-w-xl mx-auto text-xs font-mono text-background/60 leading-relaxed mb-8">
            Ready to design a low-latency rate limiting architecture for your services? Connect with an SRE infrastructure engineer.
          </p>

          <a 
            href="mailto:sales@limiter.io"
            className="inline-flex items-center gap-2 bg-[#ea580c] hover:bg-[#c2410c] text-white font-mono uppercase tracking-widest px-6 py-3 text-xs font-bold transition-colors cursor-pointer"
          >
            <Mail size={14} />
            Contact sales@limiter.io
          </a>
        </div>
      </main>

      <Footer />
    </div>
  )
}
