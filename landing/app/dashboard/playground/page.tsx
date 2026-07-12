"use client"

import * as React from "react"
import { TerminalSquare, Send, Zap } from "lucide-react"

import {
  Panel,
  PanelHeader,
  BrutalButton,
  Field,
  SelectField,
  InlineError,
  Label,
  StatusBadge,
} from "@/components/dashboard/kit"
import { API_BASE } from "@/lib/api"

interface Result {
  status: number
  ok: boolean
  durationMs: number
  headers: Record<string, string>
  body: string
}

const RATE_HEADERS = ["x-ratelimit-limit", "x-ratelimit-remaining", "x-ratelimit-reset", "retry-after"]

export default function PlaygroundPage() {
  const [apiKey, setApiKey] = React.useState("")
  const [method, setMethod] = React.useState("GET")
  const [path, setPath] = React.useState("/hello")
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [result, setResult] = React.useState<Result | null>(null)

  const fire = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setResult(null)
    if (!apiKey.trim()) return setError("Enter an API key")
    const clean = path.startsWith("/") ? path : `/${path}`
    setLoading(true)
    const start = performance.now()
    try {
      const res = await fetch(`${API_BASE}/gateway${clean}`, {
        method,
        headers: { "X-API-Key": apiKey.trim() },
      })
      const body = await res.text()
      const headers: Record<string, string> = {}
      res.headers.forEach((v, k) => (headers[k] = v))
      setResult({
        status: res.status,
        ok: res.ok,
        durationMs: Math.round(performance.now() - start),
        headers,
        body,
      })
    } catch {
      setError("Request failed — is the gateway reachable?")
    } finally {
      setLoading(false)
    }
  }

  const decision = result ? (result.status === 429 ? "blocked" : result.ok ? "allowed" : "error") : null

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-lg font-bold uppercase tracking-widest">Gateway Playground</h1>
        <p className="mt-1 text-[10px] uppercase tracking-wider text-muted-foreground">
          Fire test requests through the limiter and watch policies trigger
        </p>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Request builder */}
        <Panel>
          <PanelHeader title="Request" icon={TerminalSquare} />
          <form onSubmit={fire} className="flex flex-col gap-4 p-4">
            <InlineError message={error} />
            <Field
              label="API Key (X-API-Key)"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="lim_live_..."
              hint="Generate one under API Keys."
            />
            <div className="grid grid-cols-[110px_1fr] gap-3">
              <SelectField label="Method" value={method} onChange={(e) => setMethod(e.target.value)}>
                {["GET", "POST", "PUT", "DELETE", "PATCH"].map((m) => (
                  <option key={m}>{m}</option>
                ))}
              </SelectField>
              <Field
                label="Gateway Path"
                value={path}
                onChange={(e) => setPath(e.target.value)}
                placeholder="/users/profile"
              />
            </div>
            <div className="rounded-none border-2 border-dashed border-foreground/30 p-3">
              <Label>Resolved URL</Label>
              <code className="mt-1 block break-all text-[11px] text-[#ea580c]">
                {API_BASE}/gateway{path.startsWith("/") ? path : `/${path}`}
              </code>
            </div>
            <BrutalButton type="submit" variant="primary" icon={Send} loading={loading} className="justify-center">
              Send Request
            </BrutalButton>
          </form>
        </Panel>

        {/* Response */}
        <Panel accent={decision === "blocked"}>
          <PanelHeader
            title="Response"
            icon={Zap}
            action={
              result ? (
                <span className="flex items-center gap-2 text-xs font-bold tabular-nums">
                  <StatusBadge status={decision!} />
                  {result.status} · {result.durationMs}ms
                </span>
              ) : undefined
            }
          />
          {!result ? (
            <p className="p-4 text-[11px] uppercase text-muted-foreground">
              Send a request to inspect the status, rate-limit headers and body.
            </p>
          ) : (
            <div className="flex flex-col gap-4 p-4">
              <div>
                <Label>Rate Limit Headers</Label>
                <div className="mt-1 flex flex-col gap-1">
                  {RATE_HEADERS.filter((h) => result.headers[h]).length === 0 ? (
                    <span className="text-[11px] text-muted-foreground">None returned.</span>
                  ) : (
                    RATE_HEADERS.filter((h) => result.headers[h]).map((h) => (
                      <div key={h} className="flex justify-between border-b border-foreground/10 py-1 text-xs">
                        <span className="text-muted-foreground">{h}</span>
                        <span className="font-bold tabular-nums text-[#ea580c]">{result.headers[h]}</span>
                      </div>
                    ))
                  )}
                </div>
              </div>
              <div>
                <Label>Body</Label>
                <pre className="mt-1 max-h-[200px] overflow-auto border-2 border-foreground bg-muted/10 p-3 text-[11px]">
                  {result.body || "(empty)"}
                </pre>
              </div>
            </div>
          )}
        </Panel>
      </div>
    </div>
  )
}
