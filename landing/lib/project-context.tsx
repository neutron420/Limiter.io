"use client"

import * as React from "react"

import { api } from "./api"
import { useAuth } from "./auth"
import type { Project } from "./types"

interface ProjectContextValue {
  projects: Project[]
  current: Project | null
  role: "owner" | "admin" | "member" | null
  loading: boolean
  error: string | null
  select: (projectId: string) => void
  refresh: () => Promise<void>
}

const ProjectContext = React.createContext<ProjectContextValue | null>(null)

const CURRENT_KEY = "limiter_current_project"

export function ProjectProvider({ children }: { children: React.ReactNode }) {
  const { user, ready } = useAuth()
  const [projects, setProjects] = React.useState<Project[]>([])
  const [currentId, setCurrentId] = React.useState<string | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)

  const refresh = React.useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const list = await api.get<Project[]>("/projects")
      const arr = list ?? []
      setProjects(arr)
      setCurrentId((prev) => {
        const stored = typeof window !== "undefined" ? localStorage.getItem(CURRENT_KEY) : null
        const candidate = prev ?? stored
        if (candidate && arr.some((p) => p.id === candidate)) return candidate
        return arr[0]?.id ?? null
      })
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load projects")
      setProjects([])
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => {
    if (ready && user) refresh()
    else if (ready && !user) setLoading(false)
  }, [ready, user, refresh])

  const select = React.useCallback((projectId: string) => {
    setCurrentId(projectId)
    if (typeof window !== "undefined") localStorage.setItem(CURRENT_KEY, projectId)
  }, [])

  const current = React.useMemo(
    () => projects.find((p) => p.id === currentId) ?? null,
    [projects, currentId],
  )

  const role = React.useMemo(
    () => current?.role ?? null,
    [current],
  )

  const value = React.useMemo(
    () => ({ projects, current, role, loading, error, select, refresh }),
    [projects, current, role, loading, error, select, refresh],
  )

  return <ProjectContext.Provider value={value}>{children}</ProjectContext.Provider>
}

export function useProject() {
  const ctx = React.useContext(ProjectContext)
  if (!ctx) throw new Error("useProject must be used within a ProjectProvider")
  return ctx
}
