import styles from './VoteSummaryPanel.module.css'
import { useVoteSummary } from '../../hooks/useVoteSummary'
import type { VotedItem } from '../../api/wishlistvote'

interface Props {
  missionId: string
}

function sentimentClass(sentiment: string): string {
  switch (sentiment) {
    case 'forte approbation': return styles.sentimentStrong
    case 'approuvé':          return styles.sentimentApproved
    case 'mitigé':            return styles.sentimentMixed
    case 'désapprouvé':       return styles.sentimentDisapproved
    default:                  return styles.sentimentNone
  }
}

function netScoreClass(score: number): string {
  if (score > 0) return styles.netPositive
  if (score < 0) return styles.netNegative
  return styles.netNeutral
}

function VoteBar({ ups, downs, totalVotes }: { ups: number; downs: number; totalVotes: number }) {
  const upPct  = totalVotes > 0 ? (ups  / totalVotes) * 100 : 0
  const downPct = totalVotes > 0 ? (downs / totalVotes) * 100 : 0
  return (
    <div className={styles.voteBarWrap}>
      <div className={styles.voteBar}>
        <div className={styles.voteBarUp}   style={{ width: `${upPct}%` }} />
        <div className={styles.voteBarDown} style={{ width: `${downPct}%` }} />
      </div>
      <div className={styles.voteCounts}>
        <span className={styles.voteUp}>+{ups}</span>
        <span className={styles.voteDown}>-{downs}</span>
      </div>
    </div>
  )
}

function HighlightCard({
  item,
  label,
  variant,
}: {
  item: VotedItem
  label: string
  variant: 'approved' | 'controversial'
}) {
  const cls = variant === 'approved' ? styles.highlightApproved : styles.highlightControversial
  const fmt = new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' })
  return (
    <div className={`${styles.highlight} ${cls}`}>
      <div className={styles.highlightLabel}>{label}</div>
      <div className={styles.highlightName} title={item.name}>{item.name}</div>
      <div className={styles.highlightMeta}>
        {fmt.format(item.price)} · {item.totalVotes} vote{item.totalVotes !== 1 ? 's' : ''} · {Math.round(item.approvalRate * 100)}% approbation
      </div>
    </div>
  )
}

export default function VoteSummaryPanel({ missionId }: Props) {
  const { data, isLoading, error } = useVoteSummary(missionId)

  if (isLoading) {
    return (
      <div className={styles.card}>
        <div className={styles.title}>🗳️ Votes communautaires</div>
        <div className={styles.skeleton} />
        <div className={styles.skeleton} />
        <div className={styles.skeleton} />
      </div>
    )
  }

  if (error) {
    return (
      <div className={styles.card}>
        <div className={styles.title}>🗳️ Votes communautaires</div>
        <p className={styles.errorState}>Impossible de charger les votes.</p>
      </div>
    )
  }

  const hasVotes = data && data.items.some(it => it.totalVotes > 0)

  if (!data || data.items.length === 0) {
    return (
      <div className={styles.card}>
        <div className={styles.header}>
          <span className={styles.title}>🗳️ Votes communautaires</span>
        </div>
        <div className={styles.empty}>
          <div className={styles.emptyIcon}>🗳️</div>
          <div className={styles.emptyText}>Aucun article à afficher.</div>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <span className={styles.title}>🗳️ Votes communautaires</span>
      </div>

      {/* Most approved / most controversial highlights */}
      {hasVotes && (data.mostApproved || data.mostControversial) && (
        <div className={styles.highlights}>
          {data.mostApproved && (
            <HighlightCard
              item={data.mostApproved}
              label="Le plus approuvé"
              variant="approved"
            />
          )}
          {data.mostControversial && (
            <HighlightCard
              item={data.mostControversial}
              label="Le plus controversé"
              variant="controversial"
            />
          )}
        </div>
      )}

      {/* Ranked list */}
      {!hasVotes && (
        <div className={styles.empty}>
          <div className={styles.emptyIcon}>🗳️</div>
          <div className={styles.emptyText}>
            Aucun vote enregistré pour cette mission. Invitez des collaborateurs pour voter sur les articles.
          </div>
        </div>
      )}

      {hasVotes && (
        <div className={styles.list}>
          {data.items.map((item, idx) => (
            <div key={item.itemId} className={styles.row}>
              <span className={`${styles.rank} ${idx === 0 ? styles.rankFirst : ''}`}>
                {idx + 1}
              </span>

              <div className={styles.itemInfo}>
                <div className={styles.itemName} title={item.name}>{item.name}</div>
                <div className={styles.itemPrice}>
                  {new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' }).format(item.price)}
                </div>
              </div>

              <VoteBar ups={item.ups} downs={item.downs} totalVotes={item.totalVotes} />

              <span className={`${styles.netScore} ${netScoreClass(item.netScore)}`}>
                {item.netScore > 0 ? '+' : ''}{item.netScore}
              </span>

              <span className={`${styles.sentiment} ${sentimentClass(item.sentiment)}`}>
                {item.sentiment}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
