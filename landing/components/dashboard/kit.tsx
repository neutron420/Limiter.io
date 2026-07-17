"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { ArrowRight, Loader2, type LucideIcon } from "lucide-react"

import { cn } from "@/lib/utils"

/* ------------------------------------------------------------------ *
 * Shared brutalist primitives for the dashboard.
 * Matches the app aesthetic: 2px foreground borders, hard offset
 * shadows, zero radius, mono uppercase labels, #ea580c accent.
 * ------------------------------------------------------------------ */

const ORANGE = "#ea580c"

/** A bordered card with a hard black offset shadow. */
export function Panel({
  className,
  accent = false,
  children,
  ...props
}: React.HTMLAttributes<HTMLDivElement> & { accent?: boolean }) {
  return (
    <div
      className={cn(
        "border-2 border-foreground bg-background rounded-none",
        accent
          ? "shadow-[6px_6px_0px_0px_rgba(234,88,12,1)]"
          : "shadow-[6px_6px_0px_0px_rgba(0,0,0,1)]",
        className,
      )}
      {...props}
    >
      {children}
    </div>
  )
}

export function PanelHeader({
  title,
  subtitle,
  icon: Icon,
  action,
}: {
  title: string
  subtitle?: string
  icon?: LucideIcon
  action?: React.ReactNode
}) {
  return (
    <div className="flex flex-row items-center justify-between gap-4 border-b-2 border-foreground p-4">
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          {Icon && <Icon size={15} className="text-[#ea580c] shrink-0" />}
          <h2 className="text-sm font-bold uppercase tracking-widest truncate">{title}</h2>
        </div>
        {subtitle && (
          <p className="mt-1 text-[10px] uppercase tracking-wider text-muted-foreground">{subtitle}</p>
        )}
      </div>
      {action}
    </div>
  )
}

/** Small uppercase mono caption used above values/fields. */
export function Label({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <span className={cn("text-[10px] uppercase tracking-wider text-muted-foreground", className)}>
      {children}
    </span>
  )
}

/** Metric card: caption + big number + hint, with a corner icon. */
export function StatCard({
  label,
  value,
  hint,
  icon: Icon,
  iconClassName = "text-[#ea580c]",
}: {
  label: string
  value: React.ReactNode
  hint?: string
  icon?: LucideIcon
  iconClassName?: string
}) {
  return (
    <Panel accent>
      <div className="flex flex-row items-center justify-between border-b-2 border-foreground p-3">
        <Label className="font-bold tracking-widest">{label}</Label>
        {Icon && <Icon className={cn("h-4 w-4", iconClassName)} />}
      </div>
      <div className="p-4">
        <div className="text-2xl font-bold tracking-tight tabular-nums">{value}</div>
        {hint && <p className="mt-1 text-[10px] uppercase text-muted-foreground">{hint}</p>}
      </div>
    </Panel>
  )
}

type BtnVariant = "primary" | "outline" | "danger" | "ghost"

const btnBase =
  "inline-flex items-center justify-center gap-2 font-mono text-xs font-bold uppercase tracking-wider border-2 border-foreground rounded-none px-3 py-2 transition-all duration-150 disabled:opacity-50 disabled:pointer-events-none cursor-pointer select-none"

const btnVariants: Record<BtnVariant, string> = {
  primary:
    "bg-foreground text-background hover:bg-[#ea580c] hover:text-background shadow-[3px_3px_0px_0px_rgba(0,0,0,1)] hover:shadow-[3px_3px_0px_0px_rgba(234,88,12,1)] hover:-translate-x-[1px] hover:-translate-y-[1px]",
  outline:
    "bg-background text-foreground hover:bg-muted/10 shadow-[3px_3px_0px_0px_rgba(0,0,0,1)] hover:shadow-[3px_3px_0px_0px_rgba(234,88,12,1)]",
  danger:
    "bg-background text-red-500 border-red-500 hover:bg-red-500 hover:text-background shadow-[3px_3px_0px_0px_rgba(239,68,68,0.4)]",
  ghost: "border-transparent shadow-none hover:text-[#ea580c] px-2",
}

export function BrutalButton({
  variant = "outline",
  loading = false,
  icon: Icon,
  className,
  children,
  disabled,
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: BtnVariant
  loading?: boolean
  icon?: LucideIcon
}) {
  return (
    <button
      className={cn(btnBase, btnVariants[variant], className)}
      disabled={disabled || loading}
      {...props}
    >
      {loading ? <Loader2 size={13} className="animate-spin" /> : Icon && <Icon size={13} />}
      {children}
    </button>
  )
}

/** The split orange-arrow submit button used on auth + primary CTAs. */
export function SubmitButton({
  loading,
  children,
  className,
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement> & { loading?: boolean }) {
  return (
    <motion.button
      whileHover={{ scale: 1.01 }}
      whileTap={{ scale: 0.98 }}
      type="submit"
      disabled={loading || props.disabled}
      className={cn(
        "group flex w-full items-center justify-center gap-0 border-2 border-foreground font-mono text-xs font-bold uppercase tracking-wider transition-all duration-200 disabled:opacity-60",
        "bg-foreground text-background shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:bg-[#ea580c] hover:shadow-[4px_4px_0px_0px_rgba(255,255,255,1)] cursor-pointer",
        className,
      )}
      {...(props as any)}
    >
      <span className="flex h-9 w-9 items-center justify-center border-r-2 border-background bg-[#ea580c] transition-colors group-hover:border-foreground">
        <ArrowRight size={14} strokeWidth={2} className="text-background" />
      </span>
      <span className="flex-1 py-2.5">{loading ? "PROCESSING..." : children}</span>
    </motion.button>
  )
}

/** Labeled text input matching the auth pages. */
export const Field = React.forwardRef<
  HTMLInputElement,
  React.InputHTMLAttributes<HTMLInputElement> & { label?: string; hint?: string }
>(function Field({ label, hint, className, children, ...props }, ref) {
  return (
    <label className="flex flex-col gap-2">
      {label && <Label className="tracking-wider">{label}</Label>}
      {children ? children : (
        <input
          ref={ref}
          className={cn(
            "w-full border-2 border-foreground bg-background px-3 py-2 font-mono text-sm transition-all focus:border-[#ea580c] focus:shadow-[4px_4px_0px_0px_rgba(234,88,12,1)] focus:outline-none",
            className,
          )}
          {...props}
        />
      )}
      {hint && <span className="text-[10px] text-muted-foreground">{hint}</span>}
    </label>
  )
})

/** Native select styled to match Field. */
export const SelectField = React.forwardRef<
  HTMLSelectElement,
  React.SelectHTMLAttributes<HTMLSelectElement> & { label?: string; hint?: string }
>(function SelectField({ label, hint, className, children, ...props }, ref) {
  return (
    <label className="flex flex-col gap-2">
      {label && <Label className="tracking-wider">{label}</Label>}
      <select
        ref={ref}
        className={cn(
          "w-full border-2 border-foreground bg-background px-3 py-2 font-mono text-sm transition-all focus:border-[#ea580c] focus:shadow-[4px_4px_0px_0px_rgba(234,88,12,1)] focus:outline-none",
          className,
        )}
        {...props}
      >
        {children}
      </select>
      {hint && <span className="text-[10px] text-muted-foreground">{hint}</span>}
    </label>
  )
})

/** Allowed / blocked style status pill. */
export function StatusBadge({ status }: { status: string }) {
  const s = status.toLowerCase()
  const ok = s === "allowed" || s === "active"
  const bad = s === "blocked" || s === "revoked" || s === "expired"
  return (
    <span
      className={cn(
        "inline-block border px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider",
        ok && "border-green-500/30 bg-green-500/10 text-green-500",
        bad && "border-red-500/30 bg-red-500/10 text-red-500",
        !ok && !bad && "border-foreground/30 bg-muted/10 text-muted-foreground",
      )}
    >
      {status}
    </span>
  )
}

export function InlineError({ message }: { message?: string | null }) {
  if (!message) return null
  return (
    <div className="border-2 border-red-500 bg-red-500/10 px-3 py-2 text-xs font-bold uppercase tracking-wider text-red-500">
      {"⚠ "}
      {message}
    </div>
  )
}

export function Spinner({ label = "LOADING" }: { label?: string }) {
  return (
    <div className="flex items-center justify-center gap-2 py-12 text-xs uppercase tracking-widest text-muted-foreground">
      <Loader2 size={14} className="animate-spin text-[#ea580c]" />
      {label}
    </div>
  )
}

export function EmptyState({
  icon: Icon,
  title,
  hint,
  action,
}: {
  icon?: LucideIcon
  title: string
  hint?: string
  action?: React.ReactNode
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 border-2 border-dashed border-foreground/40 p-12 text-center">
      {Icon && <Icon size={28} className="text-muted-foreground" />}
      <div className="text-sm font-bold uppercase tracking-widest">{title}</div>
      {hint && <p className="max-w-sm text-[11px] uppercase text-muted-foreground">{hint}</p>}
      {action}
    </div>
  )
}

/** Brutalist modal overlay. */
export function Modal({
  open,
  onClose,
  title,
  children,
}: {
  open: boolean
  onClose: () => void
  title: string
  children: React.ReactNode
}) {
  React.useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && onClose()
    window.addEventListener("keydown", onKey)
    return () => window.removeEventListener("keydown", onKey)
  }, [open, onClose])

  if (!open) return null
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-background/70 backdrop-blur-sm" onClick={onClose} />
      <motion.div
        initial={{ opacity: 0, y: 12, scale: 0.98 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        transition={{ duration: 0.2 }}
        className="relative z-10 w-full max-w-md border-2 border-foreground bg-background shadow-[8px_8px_0px_0px_rgba(234,88,12,1)]"
      >
        <div className="border-b-2 border-foreground p-4">
          <h3 className="text-sm font-bold uppercase tracking-widest">{title}</h3>
        </div>
        <div className="p-4">{children}</div>
      </motion.div>
    </div>
  )
}

export { ORANGE }
