"use client"

import { useState } from "react"
import { motion } from "framer-motion"
import { Cpu, Home, MailCheck } from "lucide-react"
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@/components/ui/card"
import { Field, SubmitButton, InlineError } from "@/components/dashboard/kit"
import { api, ApiError } from "@/lib/api"

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("")
  const [loading, setLoading] = useState(false)
  const [sent, setSent] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setLoading(true)
    try {
      await api.post("/auth/forgot-password", { email }, { auth: false })
    } catch (err) {
      // Reveal nothing about account existence — but surface transport errors.
      if (err instanceof ApiError && err.status >= 500) {
        setError("Server error. Try again later.")
        setLoading(false)
        return
      }
    }
    setLoading(false)
    setSent(true)
  }

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col items-center justify-center p-6 relative overflow-hidden font-mono selection:bg-[#ea580c] selection:text-background">
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_80%_80%_at_50%_-20%,rgba(120,119,198,0.15),rgba(255,255,255,0))]" />

      <a
        href="/login"
        className="absolute top-6 left-6 flex items-center gap-2 border-2 border-foreground px-3 py-1.5 text-xs uppercase hover:translate-x-[-2px] hover:translate-y-[-2px] hover:shadow-[4px_4px_0px_0px_rgba(234,88,12,1)] bg-background transition-all duration-200"
      >
        <Home size={12} />
        <span>Back to Login</span>
      </a>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="w-full max-w-md relative z-10"
      >
        <Card className="border-2 border-foreground shadow-[8px_8px_0px_0px_rgba(234,88,12,1)] bg-background">
          <CardHeader className="border-b-2 border-foreground pb-6">
            <div className="flex items-center gap-2">
              <Cpu size={18} className="text-[#ea580c]" />
              <CardTitle className="text-lg tracking-widest uppercase font-bold text-foreground">
                SECURE_AUTH // RECOVER
              </CardTitle>
            </div>
            <CardDescription className="text-xs text-muted-foreground uppercase mt-2">
              Request a password reset link for your operator account.
            </CardDescription>
          </CardHeader>

          <CardContent className="pt-6">
            {sent ? (
              <div className="flex flex-col items-center gap-3 py-6 text-center">
                <MailCheck size={28} className="text-[#ea580c]" />
                <p className="text-sm font-bold uppercase tracking-wider">Request received</p>
                <p className="text-[11px] uppercase text-muted-foreground">
                  If an account exists for that email, a reset link is on its way.
                </p>
              </div>
            ) : (
              <form onSubmit={handleSubmit} className="flex flex-col gap-5">
                <InlineError message={error} />
                <Field
                  label="User Email Address"
                  type="email"
                  required
                  autoComplete="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="developer@limiter.io"
                />
                <SubmitButton loading={loading}>SEND RESET LINK</SubmitButton>
              </form>
            )}
          </CardContent>

          <CardFooter className="border-t-2 border-foreground pt-6 text-[10px] text-muted-foreground uppercase flex justify-between">
            <span>Remembered it?</span>
            <a
              href="/login"
              className="text-foreground hover:text-[#ea580c] font-bold underline decoration-[#ea580c] decoration-2 underline-offset-4"
            >
              Login here
            </a>
          </CardFooter>
        </Card>
      </motion.div>
    </div>
  )
}
