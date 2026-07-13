"use client"

import * as React from "react"
import { TerminalSquare, Send, Zap } from "lucide-react"
import { useProject } from "@/lib/project-context"
import { Rule } from "@/lib/types"
import { api } from "@/lib/api"

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
  const { current } = useProject()
  
  // Rule simulation states
  const [rules, setRules] = React.useState<Rule[]>([])
  const [selectedRuleId, setSelectedRuleId] = React.useState("")
  const [simRps, setSimRps] = React.useState("10")
  const [simCount, setSimCount] = React.useState("50")
  const [simSteps, setSimSteps] = React.useState<any[]>([])
  const [simLoading, setSimLoading] = React.useState(false)
  const [simError, setSimError] = React.useState<string | null>(null)

  // Standard request states
  const [apiKey, setApiKey] = React.useState("")
  const [method, setMethod] = React.useState("GET")
  const [path, setPath] = React.useState("/hello")
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [result, setResult] = React.useState<Result | null>(null)

  const loadRules = React.useCallback(async () => {
    if (!current) return
    try {
      const data = await api.get<Rule[]>(`/projects/${current.id}/rules`)
      setRules(data ?? [])
      if (data && data.length > 0) {
        setSelectedRuleId(data[0].id)
      }
    } catch {
      // ignore
    }
  }, [current])

  React.useEffect(() => {
    loadRules()
  }, [loadRules])

  const runSimulation = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!current || !selectedRuleId) return
    setSimError(null)
    setSimLoading(true)
    try {
      const steps = await api.post<any[]>(
        `/projects/${current.id}/rules/${selectedRuleId}/simulate`,
        {
          requests_per_second: Number(simRps),
          num_requests: Number(simCount),
        }
      )
      setSimSteps(steps ?? [])
    } catch (err) {
      setSimError(err instanceof Error ? err.message : "Simulation failed")
    } finally {
      setSimLoading(false)
    }
  }

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

      {/* Simulation Section */}
      {current && (
        <Panel>
          <PanelHeader title="Algorithm Dry-Run Simulator" icon={TerminalSquare} />
          <div className="grid grid-cols-1 gap-6 p-4 lg:grid-cols-[300px_1fr]">
            <form onSubmit={runSimulation} className="flex flex-col gap-4">
              <InlineError message={simError} />
              <SelectField
                label="Select Policy Rule"
                value={selectedRuleId}
                onChange={(e) => setSelectedRuleId(e.target.value)}
                hint="The selected rule's parameters (limit, period, burst, algorithm) will be simulated."
              >
                {rules.map((r) => (
                  <option key={r.id} value={r.id}>
                    {r.name} ({r.algorithm.replace("_", " ")})
                  </option>
                ))}
              </SelectField>

              <div className="grid grid-cols-2 gap-3">
                <Field
                  label="Simulated RPS"
                  type="number"
                  min={1}
                  max={100}
                  value={simRps}
                  onChange={(e) => setSimRps(e.target.value)}
                />
                <Field
                  label="Total Hits"
                  type="number"
                  min={1}
                  max={200}
                  value={simCount}
                  onChange={(e) => setSimCount(e.target.value)}
                />
              </div>

              <BrutalButton type="submit" variant="primary" loading={simLoading} className="justify-center cursor-pointer">
                RUN SIMULATION
              </BrutalButton>
            </form>

            <div className="flex flex-col border-2 border-foreground bg-muted/5 min-h-[250px]">
              <div className="border-b-2 border-foreground bg-muted/10 p-3 flex justify-between text-[10px] font-bold uppercase tracking-wider">
                <span>Timeline Result Log</span>
                {simSteps.length > 0 && (
                  <span>
                    Blocked: {simSteps.filter((s) => !s.allowed).length} / {simSteps.length}
                  </span>
                )}
              </div>
              {simSteps.length === 0 ? (
                <div className="flex flex-1 items-center justify-center text-[10px] text-muted-foreground uppercase p-6 text-center">
                  Configure simulation parameters and click run to view timeline.
                </div>
              ) : (
                <div className="flex-1 max-h-[350px] overflow-y-auto p-2 flex flex-col gap-1 font-mono text-[10px]">
                  {simSteps.map((s) => (
                    <div
                      key={s.request_number}
                      className={`flex justify-between items-center p-1.5 border-2 border-foreground ${
                        s.allowed ? "bg-green-500/10 border-green-600/30" : "bg-red-500/10 border-red-600/30"
                      }`}
                    >
                      <span className="font-bold">
                        HIT #{String(s.request_number).padStart(3, "0")} · {new Date(s.timestamp).toLocaleTimeString()}
                      </span>
                      <div className="flex items-center gap-3">
                        <span className="text-muted-foreground uppercase">
                          Remaining: {s.remaining} / {s.limit}
                        </span>
                        <span
                          className={`font-bold px-2 py-0.5 border-2 border-foreground ${
                            s.allowed
                              ? "bg-green-500 text-white shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]"
                              : "bg-red-500 text-white shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]"
                          }`}
                        >
                          {s.allowed ? "ALLOWED" : "RATE_LIMITED"}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </Panel>
      )}
    </div>
  )
}
