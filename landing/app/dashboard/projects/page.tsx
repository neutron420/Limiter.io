"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { motion } from "framer-motion"
import { FolderPlus, Folder, Trash2, ArrowRight, Boxes } from "lucide-react"

import {
  Panel,
  PanelHeader,
  BrutalButton,
  Field,
  Modal,
  InlineError,
  Spinner,
  EmptyState,
  SubmitButton,
  Label,
} from "@/components/dashboard/kit"
import { api, ApiError } from "@/lib/api"
import { useProject } from "@/lib/project-context"
import type { Project } from "@/lib/types"

export default function ProjectsPage() {
  const router = useRouter()
  const { projects, current, loading, error, select, refresh } = useProject()

  const [createOpen, setCreateOpen] = React.useState(false)
  const [name, setName] = React.useState("")
  const [description, setDescription] = React.useState("")
  const [saving, setSaving] = React.useState(false)
  const [formError, setFormError] = React.useState<string | null>(null)

  const [toDelete, setToDelete] = React.useState<Project | null>(null)
  const [deleting, setDeleting] = React.useState(false)

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormError(null)
    if (name.trim().length < 3) {
      setFormError("Name must be at least 3 characters")
      return
    }
    setSaving(true)
    try {
      const created = await api.post<Project>("/projects", { name, description })
      await refresh()
      if (created?.id) select(created.id)
      setCreateOpen(false)
      setName("")
      setDescription("")
    } catch (err) {
      setFormError(err instanceof ApiError ? err.message : "Failed to create project")
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async () => {
    if (!toDelete) return
    setDeleting(true)
    try {
      await api.del(`/projects/${toDelete.id}`)
      await refresh()
      setToDelete(null)
    } catch (err) {
      setFormError(err instanceof ApiError ? err.message : "Failed to delete project")
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-bold uppercase tracking-widest">Projects</h1>
          <p className="mt-1 text-[10px] uppercase tracking-wider text-muted-foreground">
            Isolated workspaces grouping API keys and rate rules
          </p>
        </div>
        <BrutalButton variant="primary" icon={FolderPlus} onClick={() => setCreateOpen(true)}>
          New Project
        </BrutalButton>
      </div>

      <InlineError message={error} />

      {loading ? (
        <Spinner label="LOADING PROJECTS" />
      ) : projects.length === 0 ? (
        <EmptyState
          icon={Boxes}
          title="No projects yet"
          hint="Create your first project to start defining rate-limit policies and issuing API keys."
          action={
            <BrutalButton variant="primary" icon={FolderPlus} onClick={() => setCreateOpen(true)}>
              Create your first project
            </BrutalButton>
          }
        />
      ) : (
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3">
          {projects.map((p, i) => (
            <motion.div
              key={p.id}
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.25, delay: i * 0.04 }}
            >
              <Panel accent={current?.id === p.id} className="flex h-full flex-col">
                <PanelHeader
                  title={p.name}
                  subtitle={current?.id === p.id ? "● Selected" : undefined}
                  icon={Folder}
                />
                <div className="flex flex-1 flex-col gap-3 p-4">
                  <p className="min-h-[32px] text-xs text-muted-foreground">
                    {p.description || "No description provided."}
                  </p>
                  <Label>Created {new Date(p.created_at).toLocaleDateString()}</Label>
                  <div className="mt-auto flex items-center gap-2 pt-2">
                    <BrutalButton
                      variant="primary"
                      icon={ArrowRight}
                      className="flex-1"
                      onClick={() => {
                        select(p.id)
                        router.push("/dashboard")
                      }}
                    >
                      Open
                    </BrutalButton>
                    <BrutalButton
                      variant="danger"
                      icon={Trash2}
                      aria-label="Delete project"
                      onClick={() => setToDelete(p)}
                    />
                  </div>
                </div>
              </Panel>
            </motion.div>
          ))}
        </div>
      )}

      {/* Create modal */}
      <Modal open={createOpen} onClose={() => setCreateOpen(false)} title="Create Project">
        <form onSubmit={handleCreate} className="flex flex-col gap-4">
          <InlineError message={formError} />
          <Field
            label="Project Name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Production Gateway"
            hint="3–100 characters."
            autoFocus
          />
          <Field
            label="Description (optional)"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Public API edge rate limiting"
          />
          <SubmitButton loading={saving}>CREATE PROJECT</SubmitButton>
        </form>
      </Modal>

      {/* Delete confirm */}
      <Modal open={!!toDelete} onClose={() => setToDelete(null)} title="Delete Project">
        <div className="flex flex-col gap-4">
          <p className="text-xs uppercase tracking-wider text-muted-foreground">
            This permanently deletes{" "}
            <span className="font-bold text-foreground">{toDelete?.name}</span> and all its keys,
            rules and analytics. This cannot be undone.
          </p>
          <div className="flex gap-2">
            <BrutalButton variant="outline" className="flex-1" onClick={() => setToDelete(null)}>
              Cancel
            </BrutalButton>
            <BrutalButton variant="danger" className="flex-1" loading={deleting} onClick={handleDelete}>
              Delete
            </BrutalButton>
          </div>
        </div>
      </Modal>
    </div>
  )
}
