import { useState } from 'react'
import { useDecision, useRunDecision, useUpdateMission } from '../../hooks'
import type { Mission, WeightProfile } from '../../types'

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
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
      <div
        style={{
          flex: 1,
          height: 5,
          borderRadius: 3,
          background: 'var(--border)',
          overflow: 'hidden',
          position: 'relative',
        }}
      >
        <div
          style={{
            width: `${score}%`,
            height: '100%',
            background: `linear-gradient(90deg, ${scoreColor(score)}88, ${scoreColor(score)})`,
            borderRadius: 3,
            animation: `bar-fill 0.7s cubic-bezier(0.4,0,0.2,1) ${delay}ms both`,
            boxShadow: `0 0 8px ${scoreColor(score)}60`,
          }}
        />
      </div>
      <span
        style={{
          fontFamily: 'var(--font-mono)',
          fontSize: '0.75rem',
          color: scoreColor(score),
          minWidth: 28,
          textAlign: 'right',
          letterSpacing: '0.05em',
        }}
      >
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
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
      <span
        style={{
          fontFamily: 'var(--font-mono)',
          fontSize: '0.65rem',
          color: value > 0 ? color : 'var(--text-dim)',
          letterSpacing: '0.1em',
          textTransform: 'uppercase',
          minWidth: 56,
          transition: 'color 0.2s',
        }}
      >
        {label}
      </span>
      <div style={{ flex: 1, position: 'relative' }}>
        {/* custom track fill */}
        <div
          style={{
            position: 'absolute',
            top: '50%',
            left: 0,
            width: `${(value / 10) * 100}%`,
            height: 3,
            background: color,
            borderRadius: 2,
            transform: 'translateY(-50%)',
            opacity: 0.4,
            pointerEvents: 'none',
            transition: 'width 0.15s ease',
          }}
        />
        <input
          type="range"
          min={0}
          max={10}
          step={1}
          value={value}
          onChange={(e) => onChange(Number(e.target.value))}
          style={{ width: '100%', accentColor: color, cursor: 'pointer', position: 'relative' }}
        />
      </div>
      <span
        style={{
          fontFamily: 'var(--font-mono)',
          fontSize: '0.75rem',
          color: value > 0 ? color : 'var(--text-dim)',
          minWidth: 16,
          textAlign: 'right',
          transition: 'color 0.2s',
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
    // Save weights first, then run
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
    <div
      style={{
        background: 'var(--surface)',
        borderRadius: 'var(--radius)',
        border: `1px solid ${winner ? 'var(--green)' : 'var(--border)'}`,
        padding: '1.5rem',
        boxShadow: winner ? '0 0 20px rgba(0,214,143,0.08)' : 'none',
      }}
    >
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '1.25rem',
        }}
      >
        <h3
          style={{
            fontFamily: 'var(--font-display)',
            color: 'var(--text-mid)',
            fontSize: '0.8rem',
            letterSpacing: '0.1em',
          }}
        >
          🏆 DECISION ENGINE
        </h3>
        <button
          onClick={handleRun}
          disabled={isPending}
          style={{
            padding: '6px 16px',
            borderRadius: 'var(--radius-sm)',
            border: '1px solid var(--green)',
            background: isPending ? 'var(--surface)' : 'rgba(0,214,143,0.08)',
            color: 'var(--green)',
            cursor: isPending ? 'not-allowed' : 'pointer',
            fontFamily: 'var(--font-display)',
            fontWeight: 700,
            fontSize: '0.75rem',
            letterSpacing: '0.06em',
            opacity: isPending ? 0.6 : 1,
            transition: 'all 0.15s',
          }}
        >
          {isPending ? '⚙ Analyzing...' : '⚙ Run Decision'}
        </button>
      </div>

      {/* Weight sliders */}
      <div
        style={{
          background: 'var(--raised)',
          borderRadius: 'var(--radius-sm)',
          padding: '1rem',
          marginBottom: '1.25rem',
          display: 'flex',
          flexDirection: 'column',
          gap: 8,
        }}
      >
        <div
          style={{
            fontFamily: 'var(--font-display)',
            fontSize: '0.7rem',
            color: 'var(--text-dim)',
            letterSpacing: '0.1em',
            marginBottom: 4,
          }}
        >
          WEIGHT PROFILE
        </div>
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

      {/* Winner card */}
      {isLoading && (
        <div style={{ color: 'var(--text-dim)', fontSize: '0.85rem', textAlign: 'center', padding: '1rem' }}>
          Loading...
        </div>
      )}

      {!isLoading && !decision && (
        <div style={{ color: 'var(--text-dim)', fontSize: '0.85rem', textAlign: 'center', padding: '1rem' }}>
          No decision yet — configure weights and click Run Decision.
        </div>
      )}

      {winner && (
        <div
          style={{
            background: 'linear-gradient(135deg, rgba(0,214,143,0.08) 0%, rgba(0,214,143,0.03) 100%)',
            border: '1px solid var(--green)',
            borderRadius: 'var(--radius-sm)',
            padding: '1rem',
            marginBottom: '1rem',
            animation: 'glow-pulse 3s ease-in-out infinite',
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          {/* corner accent */}
          <div style={{
            position: 'absolute',
            top: 0,
            right: 0,
            width: 40,
            height: 40,
            borderBottom: '40px solid transparent',
            borderRight: '40px solid rgba(0,214,143,0.15)',
          }} />
          <div
            style={{
              fontFamily: 'var(--font-mono)',
              fontSize: '0.65rem',
              color: 'var(--green)',
              letterSpacing: '0.15em',
              marginBottom: 6,
              textTransform: 'uppercase',
            }}
          >
            ◆ TOP PICK
          </div>
          <div
            style={{
              fontFamily: 'var(--font-display)',
              fontWeight: 700,
              fontSize: '1.15rem',
              color: 'var(--text)',
              marginBottom: 8,
              letterSpacing: '0.04em',
            }}
          >
            {winner.optionName}
          </div>
          <ScoreBar score={winner.score} delay={100} />
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: '1fr 1fr 1fr',
              gap: '0.5rem',
              marginTop: 10,
            }}
          >
            {[
              { label: 'PRICE', val: winner.priceScore, color: 'var(--green)' },
              { label: 'QUAL', val: winner.qualScore, color: 'var(--cyan)' },
              { label: 'FEAT', val: winner.featScore, color: 'var(--gold)' },
            ].map(({ label, val, color }) => (
              <div key={label} style={{
                background: 'var(--raised)',
                borderRadius: 4,
                padding: '4px 6px',
                textAlign: 'center',
              }}>
                <div style={{ fontSize: '0.6rem', fontFamily: 'var(--font-mono)', color: 'var(--text-dim)', letterSpacing: '0.1em' }}>{label}</div>
                <div style={{ fontSize: '0.85rem', fontFamily: 'var(--font-mono)', color, fontWeight: 700 }}>{Math.round(val)}</div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Ranked list */}
      {topScores.length > 1 && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: '1rem' }}>
          {topScores.slice(1).map((s, i) => (
            <div
              key={s.optionId}
              style={{
                display: 'flex',
                flexDirection: 'column',
                gap: 4,
                animation: `card-enter 0.4s ease ${(i + 1) * 80}ms both`,
              }}
            >
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                  fontSize: '0.82rem',
                  color: 'var(--text-mid)',
                  fontFamily: 'var(--font-body)',
                }}
              >
                <span style={{
                  fontFamily: 'var(--font-mono)',
                  fontSize: '0.65rem',
                  color: 'var(--text-dim)',
                  minWidth: 18,
                }}>#{i + 2}</span>
                <span>{s.optionName}</span>
              </div>
              <ScoreBar score={s.score} delay={(i + 1) * 80 + 200} />
            </div>
          ))}
        </div>
      )}

      {/* Eliminated */}
      {eliminated.length > 0 && (
        <div style={{ marginBottom: '1rem' }}>
          <div
            style={{
              fontFamily: 'var(--font-display)',
              fontSize: '0.7rem',
              color: 'var(--coral)',
              letterSpacing: '0.08em',
              marginBottom: 4,
            }}
          >
            ELIMINATED
          </div>
          {eliminated.map((s) => (
            <div
              key={s.optionId}
              style={{
                display: 'flex',
                alignItems: 'baseline',
                gap: 8,
                fontSize: '0.78rem',
                fontFamily: 'var(--font-mono)',
                padding: '3px 0',
                opacity: 0.6,
              }}
            >
              <span style={{ color: 'var(--coral)', textDecoration: 'line-through', textDecorationColor: 'var(--coral)' }}>
                {s.optionName}
              </span>
              <span style={{ color: 'var(--text-dim)', fontSize: '0.7rem' }}>
                {s.eliminReason}
              </span>
            </div>
          ))}
        </div>
      )}

      {/* LLM Summary */}
      {decision?.summary && (
        <div
          style={{
            background: 'linear-gradient(135deg, var(--raised) 0%, rgba(0,229,255,0.04) 100%)',
            borderRadius: 'var(--radius-sm)',
            padding: '1rem',
            fontSize: '0.85rem',
            color: 'var(--text-mid)',
            lineHeight: 1.65,
            fontFamily: 'var(--font-body)',
            borderLeft: '2px solid var(--cyan)',
            animation: 'fade-in 0.6s ease both',
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              fontFamily: 'var(--font-mono)',
              fontSize: '0.65rem',
              color: 'var(--cyan)',
              letterSpacing: '0.15em',
              marginBottom: 8,
              textTransform: 'uppercase',
            }}
          >
            <span style={{ opacity: 0.6 }}>▸</span> ADVISOR
          </div>
          {decision.summary}
        </div>
      )}
    </div>
  )
}
