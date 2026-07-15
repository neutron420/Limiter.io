"use client"

import { useState, Suspense } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { motion } from "framer-motion"
import { Cpu, Home } from "lucide-react"
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@/components/ui/card"
import { Field, SubmitButton, InlineError } from "@/components/dashboard/kit"
import { useAuth } from "@/lib/auth"
import { ApiError } from "@/lib/api"

export default function LoginPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-background text-foreground flex flex-col items-center justify-center p-6 font-mono">
          <Cpu className="h-8 w-8 animate-spin text-[#ea580c]" />
        </div>
      }
    >
      <LoginContent />
    </Suspense>
  )
}

function LoginContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { login, loginWithGoogle } = useAuth()
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [loading, setLoading] = useState(false)
  const [googleLoading, setGoogleLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const nextUrl = searchParams.get("next") || "/dashboard"

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setLoading(true)
    try {
      await login(email, password)
      router.push(nextUrl)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Unable to reach the auth server")
      setLoading(false)
    }
  }

  const handleGoogleLogin = async () => {
    setError(null)
    setGoogleLoading(true)
    try {
      await loginWithGoogle()
      router.push(nextUrl)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Google authentication failed")
      setGoogleLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col items-center justify-center p-6 relative overflow-hidden font-mono selection:bg-[#ea580c] selection:text-background">
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_80%_80%_at_50%_-20%,rgba(120,119,198,0.15),rgba(255,255,255,0))]" />

      <a
        href="/"
        className="absolute top-6 left-6 flex items-center gap-2 border-2 border-foreground px-3 py-1.5 text-xs uppercase hover:translate-x-[-2px] hover:translate-y-[-2px] hover:shadow-[4px_4px_0px_0px_rgba(234,88,12,1)] bg-background transition-all duration-200"
      >
        <Home size={12} />
        <span>Return Home</span>
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
                SECURE_AUTH // LOG_IN
              </CardTitle>
            </div>
            <CardDescription className="text-xs text-muted-foreground uppercase mt-2">
              Access the API rate limiting controller dashboard.
            </CardDescription>
          </CardHeader>

          <CardContent className="pt-6">
            <div className="flex flex-col gap-5">
              <InlineError message={error} />

              <form onSubmit={handleSubmit} className="flex flex-col gap-5">
                <Field
                  label="User Email Address"
                  type="email"
                  required
                  autoComplete="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="developer@limiter.io"
                />

                <Field
                  label="Account Password"
                  type="password"
                  required
                  autoComplete="current-password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••••••"
                />

                <div className="flex justify-end -mt-2">
                  <a
                    href="/forgot-password"
                    className="text-[10px] uppercase tracking-wider text-muted-foreground hover:text-[#ea580c]"
                  >
                    Forgot password?
                  </a>
                </div>

                <SubmitButton loading={loading}>SUBMIT CREDENTIALS</SubmitButton>
              </form>

              <div className="relative flex items-center justify-center my-1">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t-2 border-foreground/10"></div>
                </div>
                <span className="relative px-3 text-[10px] uppercase bg-background text-muted-foreground font-bold">OR CONTINUE WITH</span>
              </div>

              <button
                type="button"
                disabled={loading || googleLoading}
                onClick={handleGoogleLogin}
                className="flex items-center justify-center gap-3 border-2 border-foreground bg-background py-2.5 text-xs font-bold uppercase tracking-wider transition-all duration-150 disabled:opacity-50 hover:translate-x-[-2px] hover:translate-y-[-2px] hover:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] active:translate-x-[0px] active:translate-y-[0px] active:shadow-none cursor-pointer"
              >
                <svg className="h-4 w-4 shrink-0" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
                  <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.56-2.77c-.98.66-2.23 1.06-3.72 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
                  <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.63z" fill="#FBBC05"/>
                  <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z" fill="#EA4335"/>
                </svg>
                <span>{googleLoading ? "Connecting..." : "Google"}</span>
              </button>
            </div>
          </CardContent>

          <CardFooter className="border-t-2 border-foreground pt-6 text-[10px] text-muted-foreground uppercase flex justify-between">
            <span>New operator?</span>
            <a
              href="/register"
              className="text-foreground hover:text-[#ea580c] font-bold underline decoration-[#ea580c] decoration-2 underline-offset-4"
            >
              Register here
            </a>
          </CardFooter>
        </Card>
      </motion.div>
    </div>
  )
}
