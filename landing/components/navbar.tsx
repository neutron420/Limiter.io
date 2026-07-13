"use client"

import { Cpu, LogOut, LayoutDashboard } from "lucide-react"
import { motion } from "framer-motion"
import { ThemeToggle } from "@/components/theme-toggle"
import { useAuth } from "@/lib/auth"

export function Navbar() {
  const { user, logout } = useAuth()

  return (
    <motion.div
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
      className="w-full px-4 pt-4 lg:px-6 lg:pt-6"
    >
      <nav className="w-full border border-foreground/20 bg-background/80 backdrop-blur-sm px-6 py-3 lg:px-8">
        <div className="flex items-center justify-between">
          {/* Logo */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.2, duration: 0.4 }}
            className="flex items-center gap-3 cursor-pointer"
            onClick={() => window.location.href = "/"}
          >
            <Cpu size={16} strokeWidth={1.5} />
            <span className="text-xs font-mono tracking-[0.15em] uppercase font-bold text-foreground">
              Limiter.io
            </span>
          </motion.div>

          {/* Center nav links */}
          <div className="hidden md:flex items-center gap-8">
            {["Platform", "Enterprise", "Pricing", "Resources", "Company"].map((link, i) => {
              const href = link === "Pricing" 
                ? "https://riteshsingh.lemonsqueezy.com/checkout/buy/0ea6a298-bf86-4434-8b95-f1b367d9857b"
                : `/${link.toLowerCase()}`
              return (
                <motion.a
                  key={link}
                  href={href}
                  target={link === "Pricing" ? "_blank" : undefined}
                  rel={link === "Pricing" ? "noopener noreferrer" : undefined}
                  initial={{ opacity: 0, y: -8 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.3 + i * 0.06, duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
                  className={`text-xs font-mono tracking-widest uppercase hover:text-foreground transition-colors duration-200 ${
                    link === "Pricing" ? "text-orange-500 font-bold animate-pulse" : "text-muted-foreground"
                  }`}
                >
                  {link}
                </motion.a>
              )
            })}
          </div>

          {/* Right side: Login / Profile + Console */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.5, duration: 0.4 }}
            className="flex items-center gap-4 animate-fade-in"
          >
            <ThemeToggle />
            
            {user ? (
              <div className="flex items-center gap-4">
                <span className="hidden lg:inline text-[10px] font-mono uppercase tracking-wider text-muted-foreground bg-muted px-2 py-1 border border-foreground/10">
                  {user.email}
                </span>
                
                <motion.button
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  onClick={() => window.location.href = "/dashboard"}
                  className="bg-foreground text-background px-3 py-1.5 text-xs font-mono tracking-widest uppercase cursor-pointer flex items-center gap-1.5"
                >
                  <LayoutDashboard size={12} />
                  Console
                </motion.button>
                
                <motion.button
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  onClick={() => logout()}
                  className="text-muted-foreground hover:text-red-500 transition-colors cursor-pointer p-1"
                  title="Log Out"
                >
                  <LogOut size={14} />
                </motion.button>
              </div>
            ) : (
              <>
                <a
                  href="/login"
                  className="hidden sm:block text-xs font-mono tracking-widest uppercase text-muted-foreground hover:text-foreground transition-colors duration-200"
                >
                  Log In
                </a>
                <motion.button
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  onClick={() => window.location.href = "/register"}
                  className="bg-foreground text-background px-4 py-2 text-xs font-mono tracking-widest uppercase cursor-pointer"
                >
                  Register
                </motion.button>
              </>
            )}
          </motion.div>
        </div>
      </nav>
    </motion.div>
  )
}
