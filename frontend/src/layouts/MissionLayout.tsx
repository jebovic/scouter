import { Navigate, Outlet, useParams } from 'react-router-dom'
import { Topnav, BottomNav } from '../components/scouter'
import { useMission } from '../hooks'

export function MissionLayout() {
  const { slug } = useParams<{ slug: string }>()
  if (!slug) return <Navigate to="/" replace />
  const { mission } = useMission(slug)

  return (
    <div className="page grid-bg scanlines">
      <Topnav missionSlug={slug} missionName={mission?.name} />
      <Outlet />
      <BottomNav missionSlug={slug} />
    </div>
  )
}
