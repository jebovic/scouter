import { Component, type ErrorInfo, type ReactNode } from 'react'

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
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: '100vh',
            padding: '2rem',
            background: 'var(--bg)',
            color: 'var(--text)',
          }}
        >
          <div
            style={{
              background: 'var(--surface)',
              border: '1px solid var(--coral)',
              borderRadius: 'var(--radius)',
              padding: '2.5rem',
              maxWidth: 480,
              width: '100%',
              textAlign: 'center',
            }}
          >
            <div
              style={{
                fontFamily: 'var(--font-mono)',
                fontSize: '0.65rem',
                color: 'var(--coral)',
                letterSpacing: '0.15em',
                marginBottom: '1rem',
              }}
            >
              [ SYSTEM ERROR ]
            </div>
            <h2
              style={{
                fontFamily: 'var(--font-display)',
                color: 'var(--coral)',
                fontSize: '1.2rem',
                letterSpacing: '0.08em',
                marginBottom: '0.75rem',
              }}
            >
              SOMETHING WENT WRONG
            </h2>
            <p style={{ color: 'var(--text-dim)', fontSize: '0.85rem', marginBottom: '1.5rem', fontFamily: 'var(--font-mono)' }}>
              {this.state.error?.message ?? 'An unexpected error occurred'}
            </p>
            <button
              onClick={this.handleReset}
              style={{
                padding: '8px 24px',
                borderRadius: 'var(--radius-sm)',
                border: '1px solid var(--cyan)',
                background: 'var(--cyan-dim)',
                color: 'var(--cyan)',
                cursor: 'pointer',
                fontFamily: 'var(--font-display)',
                fontSize: '0.8rem',
                letterSpacing: '0.06em',
              }}
            >
              RELOAD
            </button>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
