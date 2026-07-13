"use client"

import { useState, Suspense } from "react"
import { useSearchParams, useRouter } from "next/navigation"
import { motion } from "framer-motion"
import { Cpu, Home, ShieldCheck } from "lucide-react"
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

function ResetPasswordForm() {
	const searchParams = useSearchParams()
	const router = useRouter()
	const token = searchParams.get("token") || ""

	const [password, setPassword] = useState("")
	const [confirmPassword, setConfirmPassword] = useState("")
	const [loading, setLoading] = useState(false)
	const [success, setSuccess] = useState(false)
	const [error, setError] = useState<string | null>(null)

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault()
		setError(null)

		if (!token) {
			setError("Missing or invalid password reset token.")
			return
		}

		if (password.length < 8) {
			setError("Password must be at least 8 characters.")
			return
		}

		if (password !== confirmPassword) {
			setError("Passwords do not match.")
			return
		}

		setLoading(true)
		try {
			await api.post("/auth/reset-password", { token, password }, { auth: false })
			setSuccess(true)
		} catch (err) {
			setError(err instanceof ApiError ? err.message : "Failed to reset password.")
		} finally {
			setLoading(false)
		}
	}

	return (
		<div className="w-full max-w-md relative z-10">
			<Card className="border-2 border-foreground shadow-[8px_8px_0px_0px_rgba(234,88,12,1)] bg-background">
				<CardHeader className="border-b-2 border-foreground pb-6">
					<div className="flex items-center gap-2">
						<Cpu size={18} className="text-[#ea580c]" />
						<CardTitle className="text-lg tracking-widest uppercase font-bold text-foreground">
							SECURE_AUTH // RESET
						</CardTitle>
					</div>
					<CardDescription className="text-xs text-muted-foreground uppercase mt-2">
						Set your new account operator password.
					</CardDescription>
				</CardHeader>

				<CardContent className="pt-6">
					{success ? (
						<div className="flex flex-col items-center gap-3 py-6 text-center">
							<ShieldCheck size={28} className="text-green-500" />
							<p className="text-sm font-bold uppercase tracking-wider">Password Updated</p>
							<p className="text-[11px] uppercase text-muted-foreground mb-4">
								Your password has been successfully reset.
							</p>
							<BrutalButton variant="primary" onClick={() => router.push("/login")}>
								PROCEED TO LOGIN
							</BrutalButton>
						</div>
					) : (
						<form onSubmit={handleSubmit} className="flex flex-col gap-5">
							<InlineError message={error} />
							{!token && (
								<div className="rounded-none border-2 border-danger bg-danger/10 p-3 text-[11px] text-danger uppercase font-bold">
									No reset token detected in query string. Reset will fail.
								</div>
							)}
							<Field
								label="New Password"
								type="password"
								required
								value={password}
								onChange={(e) => setPassword(e.target.value)}
								placeholder="••••••••••••"
							/>
							<Field
								label="Confirm Password"
								type="password"
								required
								value={confirmPassword}
								onChange={(e) => setConfirmPassword(e.target.value)}
								placeholder="••••••••••••"
							/>
							<SubmitButton loading={loading} disabled={!token}>RESET PASSWORD</SubmitButton>
						</form>
					)}
				</CardContent>

				<CardFooter className="border-t-2 border-foreground pt-6 text-[10px] text-muted-foreground uppercase flex justify-between">
					<span>Abort operation?</span>
					<a
						href="/login"
						className="text-foreground hover:text-[#ea580c] font-bold underline decoration-[#ea580c] decoration-2 underline-offset-4"
					>
						Back to Login
					</a>
				</CardFooter>
			</Card>
		</div>
	)
}

function BrutalButton({ children, variant, onClick }: { children: React.ReactNode; variant: string; onClick?: () => void }) {
	return (
		<button
			onClick={onClick}
			type="button"
			className="border-2 border-foreground px-4 py-2 text-xs font-bold uppercase tracking-wider bg-[#ea580c] text-white hover:translate-x-[-2px] hover:translate-y-[-2px] hover:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] active:translate-x-[0px] active:translate-y-[0px] active:shadow-none transition-all duration-150 cursor-pointer"
		>
			{children}
		</button>
	)
}

export default function ResetPasswordPage() {
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

			<Suspense fallback={
				<div className="w-full max-w-md border-2 border-foreground p-6 bg-background font-mono text-xs uppercase tracking-wider text-muted-foreground animate-pulse text-center">
					LOADING SECURE ENVELOPE...
				</div>
			}>
				<ResetPasswordForm />
			</Suspense>
		</div>
	)
}
