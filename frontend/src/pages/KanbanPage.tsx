import { useTranslation } from 'react-i18next'
import { Topnav, LoadingPulse, KanbanBoard as KanbanBoardComponent } from '../components/scouter'
import { useMissions } from '../hooks'
import styles from './KanbanPage.module.css'

export default function KanbanPage() {
  const { t } = useTranslation()
  const { missions, isLoading } = useMissions()

  return (
    <div className="page grid-bg scanlines">
      <Topnav />
      <main className={styles.main}>
        <div className={styles.container}>
          <div className={styles.header}>
            <h1 className={styles.title}>{t('kanban.title')}</h1>
            <p className={styles.subtitle}>{t('kanban.subtitle')}</p>
          </div>

          {isLoading ? (
            <div className={styles.loading}>
              <LoadingPulse label={t('kanban.loading')} />
            </div>
          ) : (
            <KanbanBoardComponent missions={missions} />
          )}
        </div>
      </main>
    </div>
  )
}
