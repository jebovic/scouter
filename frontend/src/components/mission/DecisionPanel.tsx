import { useState } from 'react'
import { useDecision, useRunDecision, useUpdateMission } from '../../hooks'
import type { Mission, WeightProfile } from '../../types'
import styles from './DecisionPanel.module.css'

interface DecisionPanelProps {
  mission: Mission
}

const DEFAULT_WEIGHTS: WeightProfile = { price: 1, quality: 1, feature: 1 }

function scoreColor(score: number): string {
  if (score >= 75) return 'var(--green)'
  if (score >= 50) return 'var(--gold)'
  if (score >= 25) return 'var(--orange)'
  return 'var(--coral)'
}

function ScoreBar({ score, delay = 0 }: { score: number; delay?: number }) {
  const color = scoreColor(score)
  return (
    <div className={styles.scoreBarRow}>
      <div className={styles.scoreTrack}>
        <div
          className={styles.scoreFill}
          style={{
            '--bar-width': `${score}%`,
            '--bar-gradient': `linear-gradient(90deg, ${color}88, ${color})`,
            '--bar-shadow': `0 0 8px ${color}60`,
            '--bar-delay': `${delay}ms`,
          } as React.CSSProperties}
        />
      </div>
      <span className={styles.scoreValue} style={{ color }}>
        {Math.round(score)}
      </span>
    </div>
  )
}

function WeightSlider({
  label,
  value,
  onChange,
  color = 'var(--cyan)',
}: {
  label: string
  value: number
  onChange: (v: number) => void
  color?: string
}) {
  const activeColor = value > 0 ? color : 'var(--text-dim)'
  return (
    <div className={styles.sliderRow}>
      <span className={styles.sliderLabel} style={{ color: activeColor }}>
        {label}
      </span>
      <div className={styles.sliderTrackWrap}>
        <div
          className={styles.sliderFill}
          style={{
            '--fill-width': `${(value / 10) * 100}%`,
            '--fill-color': color,
          } as React.CSSProperties}
        />
        <input
          type="range"
          min={0}
          max={10}
          step={1}
          value={value}
          onChange={(e) => onChange(Number(e.target.value))}
          className={styles.sliderInput}
          style={{ '--slider-color': color } as React.CSSProperties}
        />
      </div>
      <span
        className={styles.sliderValue}
        style={{
          color: activeColor,
          fontWeight: value > 5 ? 700 : 400,
        }}
      >
        {value}
      </span>
    </div>
  )
}

export function DecisionPanel({ mission }: DecisionPanelProps) {
  const { decision, isLoading } = useDecision(mission.id)
  const { runDecision, isPending } = useRunDecision(mission.id)
  const { updateMission } = useUpdateMission(mission.slug)

  const saved = mission.weightProfile
  const hasWeights = saved.price > 0 || saved.quality > 0 || saved.feature > 0
  const [weights, setWeights] = useState<WeightProfile>(
    hasWeights ? saved : DEFAULT_WEIGHTS,
  )

  async function handleRun() {
    await updateMission({ weightProfile: weights })
    await runDecision()
  }

  const topScores = decision
    ? [...decision.scores]
        .filter((s) => !s.eliminated)
        .sort((a, b) => b.score - a.score)
        .slice(0, 5)
    : []

  const eliminated = decision
    ? decision.scores.filter((s) => s.eliminated)
    : []

  const winner = topScores[0]

  return (
    <div className={`${styles.panel} ${winner ? styles.panelWinner : styles.panelDefault}`}>
      <div className={styles.panelHeader}>
        <h3 className={styles.panelTitle}>🏆 DECISION ENGINE</h3>
        <button
          onClick={handleRun}
          disabled={isPending}
          className={`${styles.runBtn} ${isPending ? styles.runBtnPending : styles.runBtnActive}`}
        >
          {isPending ? '⚙ Analyzing...' : '⚙ Run Decision'}
        </button>
      </div>

      {/* Weight sliders */}
      <div className={styles.weightsBox}>
        <div className={styles.weightsLabel}>WEIGHT PROFILE</div>
        <WeightSlider
          label="Price"
          value={weights.price}
          onChange={(v) => setWeights((w) => ({ ...w, price: v }))}
          color="var(--green)"
        />
        <WeightSlider
          label="Quality"
          value={weights.quality}
          onChange={(v) => setWeights((w) => ({ ...w, quality: v }))}
          color="var(--cyan)"
        />
        <WeightSlider
          label="Features"
          value={weights.feature}
          onChange={(v) => setWeights((w) => ({ ...w, feature: v }))}
          color="var(--gold)"
        />
      </div>

      {/* Loading / empty states */}
      {isLoading && (
        <div className={styles.emptyState}>Loading...</div>
      )}

      {!isLoading && !decision && (
        <div className={styles.emptyState}>
          No decision yet — configure weights and click Run Decision.
        </div>
      )}

      {/* Winner card */}
      {winner && (
        <div className={styles.winnerCard}>
          <div className={styles.winnerCorner} />
          <div className={styles.topPickLabel}>◆ TOP PICK</div>
          <div className={styles.winnerName}>{winner.optionName}</div>
          <ScoreBar score={winner.score} delay={100} />
          <div className={styles.scoreGrid}>
            {[
              { label: 'PRICE', val: winner.priceScore, color: 'var(--green)' },
              { label: 'QUAL', val: winner.qualScore, color: 'var(--cyan)' },
              { label: 'FEAT', val: winner.featScore, color: 'var(--gold)' },
            ].map(({ label, val, color }) => (
              <div key={label} className={styles.scoreCell}>
                <div className={styles.scoreCellLabel}>{label}</div>
                <div className={styles.scoreCellValue} style={{ color }}>
                  {Math.round(val)}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Ranked list */}
      {topScores.length > 1 && (
        <div className={styles.rankedList}>
          {topScores.slice(1).map((s, i) => (
            <div
              key={s.optionId}
              className={styles.rankedItem}
              style={{ animation: `card-enter 0.4s ease ${(i + 1) * 80}ms both` }}
            >
              <div className={styles.rankedItemHeader}>
                <span className={styles.rankedRank}>#{i + 2}</span>
                <span>{s.optionName}</span>
              </div>
              <ScoreBar score={s.score} delay={(i + 1) * 80 + 200} />
            </div>
          ))}
        </div>
      )}

      {/* Eliminated */}
      {eliminated.length > 0 && (
        <div className={styles.eliminatedSection}>
          <div className={styles.eliminatedLabel}>ELIMINATED</div>
          {eliminated.map((s) => (
            <div key={s.optionId} className={styles.eliminatedRow}>
              <span className={styles.eliminatedName}>{s.optionName}</span>
              <span className={styles.eliminatedReason}>{s.eliminReason}</span>
            </div>
          ))}
        </div>
      )}

      {/* LLM Summary */}
      {decision?.summary && (
        <div className={styles.summary}>
          <div className={styles.summaryHeader}>
            <span className={styles.summaryArrow}>▸</span> ADVISOR
          </div>
          {decision.summary}
        </div>
      )}
    </div>
  )
}
