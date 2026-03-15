import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useFlightSearch, useTrainSearch } from '../../hooks'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { FlightOffer, TrainOffer } from '../../api/travel'
import styles from './TravelSearchWidget.module.css'

type Mode = 'flights' | 'trains'

function formatTime(iso: string): string {
  try {
    return new Date(iso).toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' })
  } catch {
    return iso
  }
}

function durationLabel(departAt: string, arriveAt: string): string {
  const diff = new Date(arriveAt).getTime() - new Date(departAt).getTime()
  if (isNaN(diff)) return ''
  const h = Math.floor(diff / 3_600_000)
  const m = Math.floor((diff % 3_600_000) / 60_000)
  return h > 0 ? `${h}h${m > 0 ? m + 'm' : ''}` : `${m}m`
}

function FlightRow({ offer }: { offer: FlightOffer }) {
  const { fmt } = useFormatCurrency()
  return (
    <div className={styles.resultRow}>
      <div className={styles.carrier}>
        <span className={styles.badge}>{offer.airline}</span>
        <span style={{ marginLeft: '0.4rem', fontSize: '0.72rem', color: 'var(--text-dim)' }}>
          {offer.flightNum}
        </span>
      </div>
      <div className={styles.route}>
        {offer.departure}
        <span className={styles.arrow}>→</span>
        {offer.arrival}
      </div>
      <div className={styles.times}>
        {formatTime(offer.departAt)} – {formatTime(offer.arriveAt)}
      </div>
      <div className={styles.duration}>{durationLabel(offer.departAt, offer.arriveAt)}</div>
      <div className={styles.price}>{fmt(offer.priceEur)}</div>
    </div>
  )
}

function TrainRow({ offer }: { offer: TrainOffer }) {
  const { fmt } = useFormatCurrency()
  const durH = Math.floor(offer.durationMin / 60)
  const durM = offer.durationMin % 60
  const durLabel = durH > 0 ? `${durH}h${durM > 0 ? durM + 'm' : ''}` : `${durM}m`
  return (
    <div className={styles.resultRow}>
      <div className={styles.carrier}>
        <span className={styles.badge}>{offer.operator}</span>
        <span style={{ marginLeft: '0.4rem', fontSize: '0.72rem', color: 'var(--text-dim)' }}>
          {offer.trainNum}
        </span>
      </div>
      <div className={styles.route}>
        {offer.departure}
        <span className={styles.arrow}>→</span>
        {offer.arrival}
      </div>
      <div className={styles.times}>
        {formatTime(offer.departAt)} – {formatTime(offer.arriveAt)}
      </div>
      <div className={styles.duration}>{durLabel}</div>
      <div className={styles.price}>{fmt(offer.priceEur)}</div>
    </div>
  )
}

export default function TravelSearchWidget() {
  const { t } = useTranslation()
  const [mode, setMode] = useState<Mode>('flights')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [date, setDate] = useState('')
  const [query, setQuery] = useState({ from: '', to: '', date: '' })
  const [submitted, setSubmitted] = useState(false)

  const { flights, isLoading: flightsLoading } = useFlightSearch(
    mode === 'flights' && submitted ? query.from : '',
    mode === 'flights' && submitted ? query.to : '',
    mode === 'flights' && submitted ? query.date : '',
  )

  const { trains, isLoading: trainsLoading } = useTrainSearch(
    mode === 'trains' && submitted ? query.from : '',
    mode === 'trains' && submitted ? query.to : '',
    mode === 'trains' && submitted ? query.date : '',
  )

  const isLoading = mode === 'flights' ? flightsLoading : trainsLoading

  function handleSearch() {
    if (!from || !to || !date) return
    setQuery({ from, to, date })
    setSubmitted(true)
  }

  const results: (FlightOffer | TrainOffer)[] = mode === 'flights' ? flights : trains

  return (
    <div className={styles.widget}>
      <div className={styles.title}>{t('travel.title')}</div>

      <div className={styles.modeTabs}>
        <button
          className={`${styles.modeTab} ${mode === 'flights' ? styles.modeTabActive : ''}`}
          onClick={() => { setMode('flights'); setSubmitted(false) }}
        >
          ✈ {t('travel.flights')}
        </button>
        <button
          className={`${styles.modeTab} ${mode === 'trains' ? styles.modeTabActive : ''}`}
          onClick={() => { setMode('trains'); setSubmitted(false) }}
        >
          🚄 {t('travel.trains')}
        </button>
      </div>

      <div className={styles.form}>
        <div className={styles.field}>
          <label className={styles.label}>{t('travel.from')}</label>
          <input
            className={styles.input}
            placeholder={mode === 'flights' ? 'CDG' : 'Paris Gare de Lyon'}
            value={from}
            onChange={(e) => setFrom(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
          />
        </div>
        <div className={styles.field}>
          <label className={styles.label}>{t('travel.to')}</label>
          <input
            className={styles.input}
            placeholder={mode === 'flights' ? 'NCE' : 'Lyon Part-Dieu'}
            value={to}
            onChange={(e) => setTo(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
          />
        </div>
        <div className={styles.field}>
          <label className={styles.label}>{t('travel.date')}</label>
          <input
            className={styles.input}
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
          />
        </div>
        <button
          className={styles.searchBtn}
          disabled={!from || !to || !date || isLoading}
          onClick={handleSearch}
        >
          {isLoading ? t('travel.searching') : t('travel.search')}
        </button>
      </div>

      {submitted && (
        <div className={styles.results}>
          {isLoading && (
            <div className={styles.empty}>{t('travel.searching')}</div>
          )}
          {!isLoading && results.length === 0 && (
            <div className={styles.empty}>{t('travel.noResults')}</div>
          )}
          {!isLoading && mode === 'flights' && (flights as FlightOffer[]).map((f, i) => (
            <FlightRow key={`${f.flightNum}-${i}`} offer={f} />
          ))}
          {!isLoading && mode === 'trains' && (trains as TrainOffer[]).map((tr, i) => (
            <TrainRow key={`${tr.trainNum}-${i}`} offer={tr} />
          ))}
        </div>
      )}
    </div>
  )
}
