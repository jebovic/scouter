const API_BASE = import.meta.env.VITE_API_URL ?? ''

type ExportFormat = 'json' | 'markdown' | 'pdf'

export async function downloadMissionExport(missionId: string, format: ExportFormat): Promise<void> {
  const url = `${API_BASE}/api/missions/${missionId}/export?format=${format}`
  const res = await fetch(url)
  if (!res.ok) throw new Error(`Export failed: ${res.status}`)

  const blob = await res.blob()
  const disposition = res.headers.get('Content-Disposition') ?? ''
  const match = disposition.match(/filename="([^"]+)"/)
  const filename = match?.[1] ?? `mission-export.${format}`

  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(link.href)
}
