"use client"

import * as React from "react"
import { Search, Filter, ShieldAlert, FileText, ChevronRight, X, Clock, HelpCircle } from "lucide-react"
import { Panel, PanelHeader } from "@/components/dashboard/kit"
import { getCountryFromIp } from "./IpCountryBreakdown"

interface LogItem {
  id: string
  project_id: string
  api_key_id: string
  request_id: string
  client_ip: string
  route: string
  status_code: number
  latency_ms: number
  decision: string
  blocked_reason?: string
  timestamp: string
}

interface AuditLogsExplorerProps {
  logs: LogItem[]
}

export function AuditLogsExplorer({ logs }: AuditLogsExplorerProps) {
  const [search, setSearch] = React.useState("")
  const [decisionFilter, setDecisionFilter] = React.useState("all")
  const [selectedLog, setSelectedLog] = React.useState<LogItem | null>(null)

  // Filter logs based on search query and decision filter
  const filteredLogs = React.useMemo(() => {
    return logs.filter((log) => {
      const route = (log.route || "").toLowerCase()
      const ip = (log.client_ip || "").toLowerCase()
      const reqId = (log.request_id || "").toLowerCase()
      const q = search.toLowerCase()
      
      const matchesSearch = route.includes(q) || ip.includes(q) || reqId.includes(q)
      const matchesDecision =
        decisionFilter === "all" ||
        (decisionFilter === "allowed" && log.decision === "allowed") ||
        (decisionFilter === "blocked" && log.decision !== "allowed")

      return matchesSearch && matchesDecision
    })
  }, [logs, search, decisionFilter])

  return (
    <div className="flex flex-col gap-4">
      {/* Search & Filter Bar */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
        {/* Search Input */}
        <div className="relative flex-1">
          <span className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-muted-foreground">
            <Search size={14} />
          </span>
          <input
            type="text"
            placeholder="Search logs by API route, client IP, or request ID..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-background border-2 border-foreground p-2 pl-9 text-xs font-bold uppercase placeholder-muted-foreground outline-none focus:ring-2 focus:ring-[#ea580c]"
          />
        </div>

        {/* Filter Switcher */}
        <div className="flex items-center gap-2 border-2 border-foreground bg-background p-1 select-none w-fit">
          <span className="text-[9px] font-extrabold uppercase px-2 text-muted-foreground flex items-center gap-1">
            <Filter size={10} /> Filter
          </span>
          <button
            onClick={() => setDecisionFilter("all")}
            className={`px-3 py-1 text-[9px] font-bold uppercase transition-all cursor-pointer ${
              decisionFilter === "all"
                ? "bg-foreground text-background"
                : "hover:bg-muted/10 text-foreground"
            }`}
          >
            All
          </button>
          <button
            onClick={() => setDecisionFilter("allowed")}
            className={`px-3 py-1 text-[9px] font-bold uppercase transition-all cursor-pointer ${
              decisionFilter === "allowed"
                ? "bg-green-500 text-white"
                : "hover:bg-muted/10 text-green-500"
            }`}
          >
            Allowed
          </button>
          <button
            onClick={() => setDecisionFilter("blocked")}
            className={`px-3 py-1 text-[9px] font-bold uppercase transition-all cursor-pointer ${
              decisionFilter === "blocked"
                ? "bg-red-500 text-white"
                : "hover:bg-muted/10 text-red-500"
            }`}
          >
            Blocked
          </button>
        </div>
      </div>

      {/* Logs Table */}
      <Panel>
        <PanelHeader title={`Audit Trail Log Stream (${filteredLogs.length} events)`} icon={FileText} />
        <div className="p-0 overflow-x-auto">
          {filteredLogs.length === 0 ? (
            <p className="text-[10px] text-muted-foreground uppercase text-center py-16">
              No matching log records found.
            </p>
          ) : (
            <table className="w-full text-left text-xs border-collapse">
              <thead>
                <tr className="border-b-2 border-foreground uppercase font-bold text-muted-foreground text-[9px] tracking-wider bg-muted/5">
                  <th className="p-3">Decision</th>
                  <th className="p-3">Route Pattern</th>
                  <th className="p-3">Client IP</th>
                  <th className="p-3 text-right">Latency</th>
                  <th className="p-3 text-right">HTTP Status</th>
                  <th className="p-3">Time</th>
                  <th className="p-3 text-center">Details</th>
                </tr>
              </thead>
              <tbody>
                {filteredLogs.map((log) => {
                  const country = getCountryFromIp(log.client_ip)
                  return (
                    <tr
                      key={log.id}
                      onClick={() => setSelectedLog(log)}
                      className="border-b border-foreground/10 hover:bg-muted/5 font-mono cursor-pointer transition-colors"
                    >
                      <td className="p-3">
                        <span
                          className={`inline-block px-1.5 py-0.5 text-[9px] font-bold uppercase border border-foreground ${
                            log.decision === "allowed"
                              ? "bg-green-500/10 text-green-600"
                              : "bg-red-500/10 text-red-600"
                          }`}
                        >
                          {log.decision}
                        </span>
                      </td>
                      <td className="p-3 font-bold truncate max-w-[200px]">{log.route}</td>
                      <td className="p-3 text-muted-foreground">
                        <span className="mr-1">{country.flag}</span>
                        {log.client_ip}
                      </td>
                      <td className="p-3 text-right font-bold">{log.latency_ms}ms</td>
                      <td className="p-3 text-right">
                        <span className={log.status_code >= 400 ? "text-red-500 font-bold" : "text-green-500 font-bold"}>
                          {log.status_code}
                        </span>
                      </td>
                      <td className="p-3 text-muted-foreground text-[10px]">
                        {new Date(log.timestamp).toLocaleTimeString()}
                      </td>
                      <td className="p-3 text-center">
                        <button className="text-foreground hover:text-[#ea580c] transition-colors p-1">
                          <ChevronRight size={14} />
                        </button>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          )}
        </div>
      </Panel>

      {/* Detail JSON Modal */}
      {selectedLog && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm">
          <div className="relative w-full max-w-lg bg-background border-4 border-foreground p-6 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] animate-in fade-in zoom-in-95 duration-150">
            {/* Header */}
            <div className="flex justify-between items-start border-b-2 border-foreground pb-4">
              <div>
                <span
                  className={`inline-block px-2 py-0.5 text-xs font-bold uppercase border-2 border-foreground mb-2 ${
                    selectedLog.decision === "allowed"
                      ? "bg-green-500 text-white"
                      : "bg-red-500 text-white"
                  }`}
                >
                  {selectedLog.decision}
                </span>
                <h3 className="text-sm font-bold uppercase font-mono truncate max-w-[320px]">
                  {selectedLog.route}
                </h3>
              </div>
              <button
                onClick={() => setSelectedLog(null)}
                className="border-2 border-foreground p-1 hover:bg-[#ea580c] hover:text-white transition-colors cursor-pointer"
              >
                <X size={16} />
              </button>
            </div>

            {/* Metadata Grid */}
            <div className="grid grid-cols-2 gap-4 py-4 border-b-2 border-foreground font-mono text-[11px] uppercase">
              <div>
                <span className="text-muted-foreground block text-[9px] font-bold">Client Host IP</span>
                <span className="font-bold flex items-center gap-1 mt-0.5">
                  <span>{getCountryFromIp(selectedLog.client_ip).flag}</span>
                  {selectedLog.client_ip}
                </span>
              </div>
              <div>
                <span className="text-muted-foreground block text-[9px] font-bold">Origin Country</span>
                <span className="font-bold mt-0.5">{getCountryFromIp(selectedLog.client_ip).name}</span>
              </div>
              <div>
                <span className="text-muted-foreground block text-[9px] font-bold">Execution Latency</span>
                <span className="font-bold text-foreground mt-0.5">{selectedLog.latency_ms} ms</span>
              </div>
              <div>
                <span className="text-muted-foreground block text-[9px] font-bold">HTTP Status Code</span>
                <span className="font-bold text-foreground mt-0.5">{selectedLog.status_code}</span>
              </div>
              <div className="col-span-2">
                <span className="text-muted-foreground block text-[9px] font-bold">Request Identifier</span>
                <code className="block bg-muted/30 border border-foreground/20 p-1 mt-0.5 text-[9px] lowercase break-all">
                  {selectedLog.request_id}
                </code>
              </div>
              {selectedLog.decision !== "allowed" && (
                <div className="col-span-2 bg-red-500/10 border border-red-500 p-2 text-red-600">
                  <span className="block text-[9px] font-extrabold flex items-center gap-1">
                    <ShieldAlert size={10} /> Restriction Reason
                  </span>
                  <span className="font-bold mt-0.5">{selectedLog.blocked_reason || "rate_limit_exceeded"}</span>
                </div>
              )}
            </div>

            {/* Raw JSON Code Block */}
            <div className="mt-4">
              <span className="text-muted-foreground block text-[9px] uppercase font-bold font-mono mb-1">Raw Telemetry Context</span>
              <pre className="bg-muted p-3 border-2 border-foreground text-[10px] font-mono overflow-auto max-h-[160px] whitespace-pre-wrap select-all">
                {JSON.stringify(selectedLog, null, 2)}
              </pre>
            </div>

            {/* Footer */}
            <div className="mt-6 flex justify-end">
              <button
                onClick={() => setSelectedLog(null)}
                className="bg-foreground text-background border-2 border-foreground hover:bg-[#ea580c] hover:text-white px-4 py-1.5 text-xs font-bold uppercase transition-colors cursor-pointer shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] hover:shadow-none translate-y-0.5"
              >
                Close Details
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
