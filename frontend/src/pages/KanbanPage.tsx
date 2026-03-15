import { Topnav, LoadingPulse, KanbanBoard as KanbanBoardComponent } from '../components/scouter'
import { useMissions } from '../hooks'
import styles from './KanbanPage.module.css'

export default function KanbanPage() {
  const { missions, isLoading } = useMissions()

  return (
    <div className="page grid-bg scanlines">
      <Topnav />
      <main className={styles.main}>
        <div className={styles.container}>
          <div className={styles.header}>
            <h1 className={styles.title}>Mission Board</h1>
            <p className={styles.subtitle}>Visualize your missions by status</p>
          </div>

          {isLoading ? (
            <div className={styles.loading}>
              <LoadingPulse label="Loading missions..." />
            </div>
          ) : (
            <KanbanBoardComponent missions={missions} />
          )}
        </div>
      </main>
    </div>
  )
}
