import { useNavigate } from 'react-router-dom'
import styles from './NotFoundPage.module.css'

export default function NotFoundPage() {
  const navigate = useNavigate()

  return (
    <div className={styles.root}>
      <div className={styles.grid} aria-hidden="true" />
      <div className={styles.scanline} aria-hidden="true" />

      <div className={styles.content}>
        <div className={styles.tag}>[ SECTOR_NOT_FOUND ]</div>

        <div className={styles.codeWrap} aria-label="404">
          <div className={styles.bloom} aria-hidden="true" />
          <span className={styles.digit}>4</span>
          <span className={styles.digitGlitch} data-text="0">0</span>
          <span className={styles.digit}>4</span>
        </div>

        <h1 className={styles.title}>Route Inconnue</h1>
        <p className={styles.description}>
          La page que vous cherchez n'existe pas ou a été déplacée.
          Vérifiez l'URL ou revenez au quartier général.
        </p>

        <div className={styles.divider} aria-hidden="true">
          <span />
          <span className={styles.dividerDot} />
          <span />
        </div>

        <div className={styles.actions}>
          <button className={styles.btnPrimary} onClick={() => navigate('/')}>
            ← Retour au QG
          </button>
          <button className={styles.btnGhost} onClick={() => navigate(-1)}>
            Page précédente
          </button>
        </div>
      </div>

      <div className={styles.cornerTL} aria-hidden="true" />
      <div className={styles.cornerBR} aria-hidden="true" />
    </div>
  )
}
