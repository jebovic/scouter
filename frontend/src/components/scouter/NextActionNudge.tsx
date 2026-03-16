import { Link } from 'react-router-dom'
import { ArrowRight, Search, GitCompare, ShoppingCart } from 'lucide-react'
import type { Mission } from '../../types'
import styles from './NextActionNudge.module.css'

interface NextActionNudgeProps {
  mission: Mission
}

interface NudgeConfig {
  icon: React.ReactNode
  message: string
  action: string
  href: string
  color: string
}

function getNudge(mission: Mission): NudgeConfig | null {
  if (mission.archivedAt) return null

  const base = `/missions/${mission.slug}`

  switch (mission.phase) {
    case 'researching':
      return {
        icon: <Search size={14} />,
        message: 'Lance une recherche pour trouver des options.',
        action: 'Rechercher',
        href: `${base}/options`,
        color: 'var(--cyan)',
      }
    case 'comparing':
      return {
        icon: <GitCompare size={14} />,
        message: 'Compare les options pour choisir la meilleure.',
        action: 'Comparer',
        href: `${base}/options`,
        color: 'var(--gold)',
      }
    case 'buying':
      return {
        icon: <ShoppingCart size={14} />,
        message: 'Vous êtes prêt·e à acheter ! Vérifiez les prix.',
        action: 'Acheter',
        href: `${base}/shopping`,
        color: 'var(--green)',
      }
    default:
      return null
  }
}

export function NextActionNudge({ mission }: NextActionNudgeProps) {
  const nudge = getNudge(mission)
  if (!nudge) return null

  return (
    <div className={styles.strip} style={{ '--nudge-color': nudge.color } as React.CSSProperties}>
      <span className={styles.icon}>{nudge.icon}</span>
      <span className={styles.message}>{nudge.message}</span>
      <Link to={nudge.href} className={styles.action}>
        {nudge.action}
        <ArrowRight size={12} className={styles.arrow} />
      </Link>
    </div>
  )
}
