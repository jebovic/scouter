import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MissionTimeline } from './MissionTimeline'
import * as timelineHook from '../../hooks/useTimeline'
import type { TimelineEvent } from '../../hooks/useTimeline'

describe('MissionTimeline', () => {
  const missionId = 'test-mission-id'
  const missionSlug = 'test-mission'

  const mockEvents: TimelineEvent[] = [
    {
      id: 'mission-1',
      type: 'mission_created',
      date: new Date('2026-03-01T10:00:00Z'),
      title: 'Mission créée',
      icon: '🚀',
    },
    {
      id: 'option-1',
      type: 'option_added',
      date: new Date('2026-03-02T11:00:00Z'),
      title: 'Option ajoutée: Laptop Pro',
      description: 'electronics',
      icon: '📦',
      entityId: 'option-1',
    },
    {
      id: 'item-1',
      type: 'item_added',
      date: new Date('2026-03-03T12:00:00Z'),
      title: 'Article ajouté: Laptop',
      description: 'Amazon',
      icon: '🛒',
      entityId: 'item-1',
    },
    {
      id: 'item-purchased-1',
      type: 'item_purchased',
      date: new Date('2026-03-04T13:00:00Z'),
      title: 'Article acheté: Laptop',
      description: 'Amazon',
      icon: '✅',
      entityId: 'item-1',
    },
    {
      id: 'option-2',
      type: 'option_added',
      date: new Date('2026-03-05T14:00:00Z'),
      title: 'Option ajoutée: Monitor',
      description: 'electronics',
      icon: '📦',
      entityId: 'option-2',
    },
    {
      id: 'option-3',
      type: 'option_added',
      date: new Date('2026-03-06T15:00:00Z'),
      title: 'Option ajoutée: Keyboard',
      description: 'accessories',
      icon: '📦',
      entityId: 'option-3',
    },
  ]

  beforeEach(() => {
    vi.spyOn(timelineHook, 'useTimeline').mockReturnValue({
      events: mockEvents,
      isLoading: false,
    })
  })

  it('renders timeline component', () => {
    render(<MissionTimeline missionId={missionId} missionSlug={missionSlug} />)
    const container = screen.getByRole('region')
    expect(container).toBeInTheDocument()
  })

  it('shows last 5 events by default', () => {
    render(<MissionTimeline missionId={missionId} missionSlug={missionSlug} />)
    const events = screen.getAllByTestId(/^timeline-event-/)
    expect(events.length).toBe(5)
  })

  it('displays all events when expanded', async () => {
    const user = userEvent.setup()
    render(<MissionTimeline missionId={missionId} missionSlug={missionSlug} />)

    const expandButton = screen.getByRole('button', { name: /Voir tout/i })
    await user.click(expandButton)

    const events = screen.getAllByTestId(/^timeline-event-/)
    expect(events.length).toBe(mockEvents.length)
  })

  it('displays expand button with event count', () => {
    render(<MissionTimeline missionId={missionId} missionSlug={missionSlug} />)
    expect(screen.getByRole('button', { name: /Voir tout \(6\)/i })).toBeInTheDocument()
  })

  it('renders event titles', () => {
    render(<MissionTimeline missionId={missionId} missionSlug={missionSlug} />)
    expect(screen.getByText('Mission créée')).toBeInTheDocument()
    expect(screen.getByText(/Option ajoutée: Laptop Pro/)).toBeInTheDocument()
  })

  it('renders event icons', () => {
    render(<MissionTimeline missionId={missionId} missionSlug={missionSlug} />)
    expect(screen.getByText('🚀')).toBeInTheDocument()
    const packageIcons = screen.getAllByText('📦')
    expect(packageIcons.length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('🛒')).toBeInTheDocument()
    expect(screen.getByText('✅')).toBeInTheDocument()
  })

  it('renders formatted dates in fr-FR locale', () => {
    render(<MissionTimeline missionId={missionId} missionSlug={missionSlug} />)
    const dateFormatter = new Intl.DateTimeFormat('fr-FR', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
    const expectedDate = dateFormatter.format(mockEvents[0].date)
    expect(screen.getByText(expectedDate)).toBeInTheDocument()
  })

  it('renders empty state when no events', () => {
    vi.spyOn(timelineHook, 'useTimeline').mockReturnValue({
      events: [],
      isLoading: false,
    })

    render(<MissionTimeline missionId={missionId} missionSlug={missionSlug} />)
    expect(screen.getByText('Aucune activité enregistrée')).toBeInTheDocument()
  })

  it('renders loading state', () => {
    vi.spyOn(timelineHook, 'useTimeline').mockReturnValue({
      events: [],
      isLoading: true,
    })

    render(<MissionTimeline missionId={missionId} missionSlug={missionSlug} />)
    expect(screen.getByText("Chargement de l'historique...")).toBeInTheDocument()
  })

  it('collapses expanded timeline when collapse button clicked', async () => {
    const user = userEvent.setup()
    render(<MissionTimeline missionId={missionId} missionSlug={missionSlug} />)

    const expandButton = screen.getByRole('button', { name: /Voir tout/i })
    await user.click(expandButton)

    let events = screen.getAllByTestId(/^timeline-event-/)
    expect(events.length).toBe(mockEvents.length)

    const collapseButton = screen.getByRole('button', { name: /Voir moins/i })
    await user.click(collapseButton)

    events = screen.getAllByTestId(/^timeline-event-/)
    expect(events.length).toBe(5)
  })
})
