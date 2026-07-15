import type { Project } from "./types"

export function canWrite(role: "owner" | "admin" | "member" | null): boolean {
  return role === "owner" || role === "admin"
}

export function canRead(role: "owner" | "admin" | "member" | null): boolean {
  return role !== null
}

export function getRoleBadgeColor(role: "owner" | "admin" | "member" | null): string {
  switch (role) {
    case "owner":
      return "text-green-500 border-green-500 bg-green-500/10"
    case "admin":
      return "text-blue-500 border-blue-500 bg-blue-500/10"
    case "member":
      return "text-gray-500 border-gray-500 bg-gray-500/10"
    default:
      return "text-muted-foreground border-muted-foreground bg-muted-foreground/10"
  }
}
