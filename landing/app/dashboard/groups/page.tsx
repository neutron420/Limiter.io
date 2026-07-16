"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { Plus, Save, Trash2, Folder, Users } from "lucide-react"

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
} from "@/components/dashboard/kit"
import { api, ApiError } from "@/lib/api"
import type { OrganizationGroup } from "@/lib/types"

export default function GroupsPage() {
  const [orgId, setOrgId] = React.useState<string | null>(null)
  const [groups, setGroups] = React.useState<OrganizationGroup[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [saving, setSaving] = React.useState(false)
  const [showModal, setShowModal] = React.useState(false)
  const [name, setName] = React.useState("")
  const [desc, setDesc] = React.useState("")

  const fetchData = React.useCallback(async () => {
    setLoading(true)
    try {
      const orgs = await api.get<any[]>("/organizations")
      if (orgs.length > 0) {
        setOrgId(orgs[0].id)
        const g = await api.get<OrganizationGroup[]>(`/organizations/${orgs[0].id}/groups`)
        setGroups(g)
      }
    } catch {
      // no org
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => { fetchData() }, [fetchData])

  const handleCreate = async () => {
    if (!orgId) return
    setSaving(true)
    setError("")
    try {
      await api.post(`/organizations/${orgId}/groups`, { name, description: desc })
      setName("")
      setDesc("")
      setShowModal(false)
      fetchData()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to create group")
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!orgId) return
    try {
      await api.del(`/organizations/${orgId}/groups/${id}`)
      fetchData()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to delete group")
    }
  }

  return (
    <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-heading tracking-tight">GROUPS</h1>
          <p className="text-sm text-muted-foreground mt-1">Organize members into groups</p>
        </div>
        {orgId && <BrutalButton onClick={() => setShowModal(true)} icon={Plus}>CREATE GROUP</BrutalButton>}
      </div>

      {loading ? (
        <Spinner label="LOADING GROUPS" />
      ) : !orgId ? (
        <EmptyState icon={Folder} title="No organization" hint="Create an organization first" />
      ) : groups.length === 0 ? (
        <EmptyState icon={Users} title="No groups" hint="Create your first group" />
      ) : (
        <div className="space-y-3">
          {groups.map((g) => (
            <Panel key={g.id}>
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-mono text-sm font-bold">{g.name}</p>
                  {g.description && <p className="text-xs text-muted-foreground">{g.description}</p>}
                </div>
                <BrutalButton onClick={() => handleDelete(g.id)} variant="destructive" size="sm">
                  <Trash2 className="size-4" />
                </BrutalButton>
              </div>
            </Panel>
          ))}
        </div>
      )}

      {error && <InlineError message={error} />}

      <Modal open={showModal} onClose={() => setShowModal(false)} title="Create Group">
        <div className="space-y-4">
          <Field label="Name">
            <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={name} onChange={(e) => setName(e.target.value)} />
          </Field>
          <Field label="Description">
            <textarea className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={desc} onChange={(e) => setDesc(e.target.value)} />
          </Field>
          <SubmitButton onClick={handleCreate} loading={saving} icon={Save}>CREATE</SubmitButton>
        </div>
      </Modal>
    </motion.div>
  )
}
