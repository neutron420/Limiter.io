"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { Boxes } from "lucide-react"

import { BrutalButton, EmptyState, Spinner } from "./kit"
import { useProject } from "@/lib/project-context"
import type { Project } from "@/lib/types"

/**
 * Renders children only once a project is selected. Otherwise shows a loader
 * or a prompt to create one. Passes the active project down via render prop.
 */
export function RequireProject({ children }: { children: (project: Project) => React.ReactNode }) {
  const router = useRouter()
  const { current, loading, projects } = useProject()

  if (loading) return <Spinner label="LOADING PROJECT" />

  if (projects.length === 0 || !current) {
    return (
      <EmptyState
        icon={Boxes}
        title="No project selected"
        hint="Create or select a project first — keys, policies and analytics are all scoped to a project."
        action={
          <BrutalButton variant="primary" onClick={() => router.push("/dashboard/projects")}>
            Go to projects
          </BrutalButton>
        }
      />
    )
  }

  return <>{children(current)}</>
}
