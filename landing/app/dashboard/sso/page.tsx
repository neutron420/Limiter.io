"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { Save, RefreshCw, Shield, Key } from "lucide-react"

import {
  Panel,
  PanelHeader,
  BrutalButton,
  Field,
  InlineError,
  Spinner,
  StatusBadge,
  SubmitButton,
} from "@/components/dashboard/kit"
import { api, ApiError } from "@/lib/api"
import type { SAMLConfig, OIDCConfig } from "@/lib/types"

export default function SSOPage() {
  const [orgId, setOrgId] = React.useState<string | null>(null)
  const [saml, setSaml] = React.useState<SAMLConfig | null>(null)
  const [oidc, setOidc] = React.useState<OIDCConfig | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [saving, setSaving] = React.useState(false)
  const [error, setError] = React.useState("")
  const [success, setSuccess] = React.useState("")

  const [tab, setTab] = React.useState<"saml" | "oidc">("saml")

  const [idpEntityId, setIdpEntityId] = React.useState("")
  const [idpSsoUrl, setIdpSsoUrl] = React.useState("")
  const [spEntityId, setSpEntityId] = React.useState("")
  const [spAcsUrl, setSpAcsUrl] = React.useState("")

  const [issuerUrl, setIssuerUrl] = React.useState("")
  const [clientId, setClientId] = React.useState("")
  const [redirectUrl, setRedirectUrl] = React.useState("")
  const [scopes, setScopes] = React.useState("openid profile email")

  const fetchConfig = React.useCallback(async () => {
    setLoading(true)
    try {
      const orgs = await api.get<any[]>("/organizations")
      if (orgs.length === 0) { setLoading(false); return }
      setOrgId(orgs[0].id)
      const s = await api.get<SAMLConfig>(`/organizations/${orgs[0].id}/sso/saml`).catch(() => null)
      setSaml(s)
      if (s) {
        setIdpEntityId(s.idp_entity_id)
        setIdpSsoUrl(s.idp_sso_url)
        setSpEntityId(s.sp_entity_id)
        setSpAcsUrl(s.sp_acs_url)
      }
      const o = await api.get<OIDCConfig>(`/organizations/${orgs[0].id}/sso/oidc`).catch(() => null)
      setOidc(o)
      if (o) {
        setIssuerUrl(o.issuer_url)
        setClientId(o.client_id)
        setRedirectUrl(o.redirect_url)
        setScopes(o.scopes)
      }
    } catch {
      // not configured
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => { fetchConfig() }, [fetchConfig])

  const handleSaveSaml = async () => {
    if (!orgId) return
    setSaving(true)
    setError("")
    setSuccess("")
    try {
      await api.post(`/organizations/${orgId}/sso/saml`, {
        idp_entity_id: idpEntityId,
        idp_sso_url: idpSsoUrl,
        sp_entity_id: spEntityId,
        sp_acs_url: spAcsUrl,
        enabled: true,
      })
      setSuccess("SAML configuration saved")
      fetchConfig()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to save SAML config")
    } finally {
      setSaving(false)
    }
  }

  const handleSaveOidc = async () => {
    if (!orgId) return
    setSaving(true)
    setError("")
    setSuccess("")
    try {
      await api.post(`/organizations/${orgId}/sso/oidc`, {
        issuer_url: issuerUrl,
        client_id: clientId,
        redirect_url: redirectUrl,
        enabled: true,
        scopes,
      })
      setSuccess("OIDC configuration saved")
      fetchConfig()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to save OIDC config")
    } finally {
      setSaving(false)
    }
  }

  return (
    <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-heading tracking-tight">SINGLE SIGN-ON</h1>
          <p className="text-sm text-muted-foreground mt-1">Configure SAML or OIDC authentication</p>
        </div>
      </div>

      <div className="flex gap-2">
        <BrutalButton onClick={() => setTab("saml")} variant={tab === "saml" ? "default" : "secondary"}>
          <Shield className="size-4" /> SAML
        </BrutalButton>
        <BrutalButton onClick={() => setTab("oidc")} variant={tab === "oidc" ? "default" : "secondary"}>
          <Key className="size-4" /> OIDC
        </BrutalButton>
      </div>

      {loading ? (
        <Spinner label="LOADING SSO CONFIG" />
      ) : !orgId ? (
        <p className="text-sm text-muted-foreground">Create an organization to configure SSO</p>
      ) : tab === "saml" ? (
        <Panel>
          <PanelHeader icon={Shield} title="SAML 2.0" action={
            saml?.enabled ? <StatusBadge status="enabled" /> : <StatusBadge status="disabled" />
          } />
          <div className="space-y-4">
            <Field label="IdP Entity ID">
              <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={idpEntityId} onChange={(e) => setIdpEntityId(e.target.value)} />
            </Field>
            <Field label="IdP SSO URL">
              <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={idpSsoUrl} onChange={(e) => setIdpSsoUrl(e.target.value)} />
            </Field>
            <Field label="SP Entity ID">
              <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={spEntityId} onChange={(e) => setSpEntityId(e.target.value)} />
            </Field>
            <Field label="SP ACS URL">
              <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={spAcsUrl} onChange={(e) => setSpAcsUrl(e.target.value)} />
            </Field>
            <SubmitButton onClick={handleSaveSaml} loading={saving} icon={Save}>SAVE SAML</SubmitButton>
          </div>
        </Panel>
      ) : (
        <Panel>
          <PanelHeader icon={Key} title="OpenID Connect" action={
            oidc?.enabled ? <StatusBadge status="enabled" /> : <StatusBadge status="disabled" />
          } />
          <div className="space-y-4">
            <Field label="Issuer URL">
              <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={issuerUrl} onChange={(e) => setIssuerUrl(e.target.value)} />
            </Field>
            <Field label="Client ID">
              <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={clientId} onChange={(e) => setClientId(e.target.value)} />
            </Field>
            <Field label="Redirect URL">
              <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={redirectUrl} onChange={(e) => setRedirectUrl(e.target.value)} />
            </Field>
            <Field label="Scopes">
              <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={scopes} onChange={(e) => setScopes(e.target.value)} />
            </Field>
            <SubmitButton onClick={handleSaveOidc} loading={saving} icon={Save}>SAVE OIDC</SubmitButton>
          </div>
        </Panel>
      )}

      {error && <InlineError message={error} />}
      {success && <p className="text-sm text-green-600 font-mono">{success}</p>}
    </motion.div>
  )
}
