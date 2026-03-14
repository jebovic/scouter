import { Component, type ErrorInfo, type ReactNode } from 'react'
import styles from './ErrorBoundary.module.css'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('[ErrorBoundary]', error, info.componentStack)
  }

  handleReset = () => {
    window.location.reload()
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) return this.props.fallback
      return (
        <div className={styles.fullPage}>
          <div className={styles.card}>
            <div className={styles.label}>[ SYSTEM ERROR ]</div>
            <h2 className={styles.title}>SOMETHING WENT WRONG</h2>
            <p className={styles.message}>
              {this.state.error?.message ?? 'An unexpected error occurred'}
            </p>
            <button onClick={this.handleReset} className={styles.reloadBtn}>
              RELOAD
            </button>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
