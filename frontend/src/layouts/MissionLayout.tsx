import { Outlet, useParams } from 'react-router-dom'
import { Topnav } from '../components/scouter'
import { useMission } from '../hooks'

export function MissionLayout() {
  const { slug } = useParams<{ slug: string }>()
  const { mission } = useMission(slug!)

  return (
    <div className="page grid-bg scanlines">
      <Topnav missionSlug={slug} missionName={mission?.name} />
      <Outlet />
    </div>
  )
}
