# Mission & Option Management Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add delete/archive/edit actions for missions and options, fix pin-as-shortlist workflow, and surface the shortlist in the buying phase.

**Architecture:** Pure frontend changes — all required backend endpoints already exist. New components (`MissionActionBar`, `MissionEditModal`, `OptionEditModal`, `ShortlistPanel`) are added alongside existing pages. Mutations go through React Query hooks following existing patterns in `useMission.ts` and `useOptions.ts`.

**Tech Stack:** React 19, TypeScript, Vite, Tanstack Query v5, Zod, react-i18next, Vitest + Testing Library

**Spec:** `docs/superpowers/specs/2026-03-16-mission-option-management-design.md`

---

## Chunk 1: Foundation — i18n + archive API + hooks

### Task 1: Add i18n keys

**Files:**
- Modify: `frontend/src/i18n/en.json`
- Modify: `frontend/src/i18n/fr.json`

- [ ] **Step 1: Add mission action keys to en.json**

Open `frontend/src/i18n/en.json`. Find the `"mission"` object and add these keys (place in alphabetical order within the object):

```json
"actions": {
  "archive": "Archive",
  "archiveConfirm": "Archive this mission?",
  "archivedBadge": "Archived",
  "delete": "Delete mission",
  "deleteConfirm": "This will permanently delete the mission and all its options.",
  "deleteTypeToConfirm": "Type the mission name to confirm",
  "edit": "Edit mission",
  "showArchived": "Show archived",
  "unarchive": "Unarchive",
  "unarchiveConfirm": "Restore this mission to your dashboard?"
}
```

- [ ] **Step 2: Add option action keys to en.json**

In en.json, there is an existing `"options"` object — but the i18n keys for option management use the singular `option` namespace (e.g. `option.actions.edit`). Add a **new** top-level `"option"` object alongside the existing `"options"` object:

```json
"option": {
  "actions": {
    "badgeLabel": "Badge",
    "delete": "Delete option",
    "deleteConfirm": "Remove this option permanently?",
    "edit": "Edit option",
    "unpinAll": "Unpin all ({{count}})",
    "unpinAllError": "Could not unpin {{count}} option(s). Please try again."
  }
}
```

- [ ] **Step 3: Add shortlist keys to en.json**

Add a new top-level `"shortlist"` key:

```json
"shortlist": {
  "empty": "Pin your top picks in Options to build your shortlist.",
  "emptyLink": "Go to Options",
  "replaceConfirm": "A purchase is already recorded. Replace with this option?",
  "select": "Select",
  "title": "Your Shortlist"
}
```

- [ ] **Step 4: Mirror all keys in fr.json**

Open `frontend/src/i18n/fr.json` and add the same structure with French translations, using proper nested JSON matching the exact en.json nesting.

**1) Add to the `"mission"` object in fr.json:**

```json
"actions": {
  "archive": "Archiver",
  "archiveConfirm": "Archiver cette mission ?",
  "archivedBadge": "Archivée",
  "delete": "Supprimer la mission",
  "deleteConfirm": "Cela supprimera définitivement la mission et toutes ses options.",
  "deleteTypeToConfirm": "Tapez le nom de la mission pour confirmer",
  "edit": "Modifier la mission",
  "showArchived": "Afficher les archivées",
  "unarchive": "Désarchiver",
  "unarchiveConfirm": "Restaurer cette mission dans votre tableau de bord ?"
}
```

**2) Add a new top-level `"option"` object in fr.json** (alongside the existing `"options"` object):

```json
"option": {
  "actions": {
    "badgeLabel": "Badge",
    "delete": "Supprimer l'option",
    "deleteConfirm": "Supprimer cette option définitivement ?",
    "edit": "Modifier l'option",
    "unpinAll": "Tout désépingler ({{count}})",
    "unpinAllError": "Impossible de désépingler {{count}} option(s). Veuillez réessayer."
  }
}
```

**3) Add a new top-level `"shortlist"` object in fr.json:**

```json
"shortlist": {
  "empty": "Épinglez vos meilleures options pour construire votre liste courte.",
  "emptyLink": "Aller aux options",
  "replaceConfirm": "Un achat est déjà enregistré. Remplacer par cette option ?",
  "select": "Sélectionner",
  "title": "Votre liste courte"
}
```

- [ ] **Step 5: Commit**

```bash
cd /home/jibei/projects/scouter
git add frontend/src/i18n/en.json frontend/src/i18n/fr.json
git commit -m "feat(i18n): add mission/option management and shortlist keys"
```

---

### Task 2: Add archive/unarchive API functions

**Files:**
- Modify: `frontend/src/api/missions.ts`

The file at `frontend/src/api/missions.ts` currently has `listMissions()`, `getMission()`, `createMission()`, `updateMission()`, `deleteMission()`, etc. It does NOT have archive/unarchive functions or `include_archived` support on list.

- [ ] **Step 1: Write failing test**

Create `frontend/src/api/missions.test.ts` (new file):

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'
import * as client from './client'

vi.mock('./client', () => ({ apiFetch: vi.fn() }))
const mockApiFetch = vi.mocked(client.apiFetch)

beforeEach(() => {
  mockApiFetch.mockReset()
})

describe('archiveMission', () => {
  it('POSTs to /api/missions/:id/archive', async () => {
    mockApiFetch.mockResolvedValueOnce(undefined)
    const { archiveMission } = await import('./missions')
    await archiveMission('mission-uuid-123')
    expect(mockApiFetch).toHaveBeenCalledWith(
      '/api/missions/mission-uuid-123/archive',
      expect.objectContaining({ method: 'POST' })
    )
  })
})

describe('unarchiveMission', () => {
  it('POSTs to /api/missions/:id/unarchive', async () => {
    mockApiFetch.mockResolvedValueOnce(undefined)
    const { unarchiveMission } = await import('./missions')
    await unarchiveMission('mission-uuid-123')
    expect(mockApiFetch).toHaveBeenCalledWith(
      '/api/missions/mission-uuid-123/unarchive',
      expect.objectContaining({ method: 'POST' })
    )
  })
})

describe('listMissions', () => {
  it('passes include_archived param when true', async () => {
    mockApiFetch.mockResolvedValueOnce({ items: [] })
    const { listMissions } = await import('./missions')
    await listMissions({ includeArchived: true })
    expect(mockApiFetch).toHaveBeenCalledWith(
      expect.stringContaining('include_archived=true')
    )
  })

  it('omits include_archived param by default', async () => {
    mockApiFetch.mockResolvedValueOnce({ items: [] })
    const { listMissions } = await import('./missions')
    await listMissions()
    const url = mockApiFetch.mock.calls[0][0] as string
    expect(url).not.toContain('include_archived')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/api/missions.test.ts
```

Expected: FAIL — `archiveMission is not a function` (or similar)

- [ ] **Step 3: Add archive/unarchive functions to api/missions.ts**

Open `frontend/src/api/missions.ts`. Note: `apiFetch` is already imported from `'./client'` at the top of the file — no new import needed. At the end of the file, add:

```typescript
export async function archiveMission(missionId: string): Promise<void> {
  await apiFetch<void>(`/api/missions/${missionId}/archive`, { method: 'POST' })
}

export async function unarchiveMission(missionId: string): Promise<void> {
  await apiFetch<void>(`/api/missions/${missionId}/unarchive`, { method: 'POST' })
}
```

Also update the `listMissions` function signature to accept options. The actual current code uses `apiFetch`, so the before/after looks like:

```typescript
// Before (actual code):
export async function listMissions(): Promise<Mission[]> {
  const data = await apiFetch<unknown>('/api/missions')
  const { items } = z.object({ items: z.array(MissionSchema) }).parse(data)
  return items as Mission[]
}

// After:
export async function listMissions(opts?: { includeArchived?: boolean }): Promise<Mission[]> {
  const params = opts?.includeArchived ? '?include_archived=true' : ''
  const data = await apiFetch<unknown>(`/api/missions${params}`)
  const { items } = z.object({ items: z.array(MissionSchema) }).parse(data)
  return items as Mission[]
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/api/missions.test.ts
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /home/jibei/projects/scouter
git add frontend/src/api/missions.ts frontend/src/api/missions.test.ts
git commit -m "feat(api): add archiveMission, unarchiveMission, listMissions includeArchived param"
```

---

### Task 3: Add archive/unarchive hooks to useMission.ts

**Files:**
- Modify: `frontend/src/hooks/useMission.ts`

- [ ] **Step 1: Write failing test**

Create `frontend/src/hooks/useMission.test.ts` (new file):

```typescript
import { renderHook, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi } from 'vitest'
import React from 'react'

vi.mock('../api', () => ({
  listMissions: vi.fn().mockResolvedValue([]),
  getMission: vi.fn(),
  createMission: vi.fn(),
  updateMission: vi.fn(),
  deleteMission: vi.fn(),
  archiveMission: vi.fn().mockResolvedValue(undefined),
  unarchiveMission: vi.fn().mockResolvedValue(undefined),
  duplicateMission: vi.fn(),
  cloneMission: vi.fn(),
  triggerResearch: vi.fn(),
  triggerPricing: vi.fn(),
}))

function wrapper({ children }: { children: React.ReactNode }) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return React.createElement(QueryClientProvider, { client: queryClient }, children)
}

describe('useArchiveMission', () => {
  it('exports useArchiveMission hook', async () => {
    const { useArchiveMission } = await import('./useMission')
    expect(typeof useArchiveMission).toBe('function')
  })

  it('calls archiveMission API on mutate', async () => {
    const { archiveMission } = await import('../api')
    const { useArchiveMission } = await import('./useMission')
    const { result } = renderHook(() => useArchiveMission(), { wrapper })
    await act(async () => {
      await result.current.archiveMission('mission-uuid-123')
    })
    expect(archiveMission).toHaveBeenCalledWith('mission-uuid-123')
  })
})

describe('useUnarchiveMission', () => {
  it('exports useUnarchiveMission hook', async () => {
    const { useUnarchiveMission } = await import('./useMission')
    expect(typeof useUnarchiveMission).toBe('function')
  })
})

describe('useMissions with includeArchived', () => {
  it('exports useMissions hook', async () => {
    const { useMissions } = await import('./useMission')
    expect(typeof useMissions).toBe('function')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/hooks/useMission.test.ts
```

Expected: FAIL — `useArchiveMission is not a function`

- [ ] **Step 3: Add hooks to useMission.ts**

Open `frontend/src/hooks/useMission.ts`. The file follows this pattern for each mutation (note: the real `useDeleteMission` uses `qc` for the QueryClient, `mutateAsync`/`isPending` destructuring, and returns a named object — match this exactly):

```typescript
export function useDeleteMission() {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: (slug: string) => deleteMission(slug),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.missions.all() })
      toast('Mission deleted', 'success')
    },
    onError: (err: unknown) => toast(`Failed to delete mission: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { deleteMission: mutateAsync, isPending }
}
```

Add after `useDeleteMission`:

```typescript
export function useArchiveMission() {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { t } = useTranslation()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: (missionId: string) => archiveMission(missionId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.missions.all() })
      toast(t('mission.actions.archive'), 'success')
    },
    onError: (_err: unknown) => toast(t('common.error'), 'error'),
  })
  return { archiveMission: mutateAsync, isPending }
}

export function useUnarchiveMission() {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { t } = useTranslation()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: (missionId: string) => unarchiveMission(missionId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.missions.all() })
      toast(t('mission.actions.unarchive'), 'success')
    },
    onError: (_err: unknown) => toast(t('common.error'), 'error'),
  })
  return { unarchiveMission: mutateAsync, isPending }
}
```

Also update `useMissions` to accept `includeArchived`. Use a stable query key — when `includeArchived` is falsy the key stays the same so existing `invalidateQueries` callers are not broken:

```typescript
// Before (actual code — note `queryFn: listMissions` bare reference, no lambda):
export function useMissions() {
  const {
    data: missions = [],
    isLoading,
    error,
  } = useQuery({
    queryKey: queryKeys.missions.all(),
    queryFn: listMissions,
  })
  return { missions, isLoading, error }
}

// After:
export function useMissions(opts?: { includeArchived?: boolean }) {
  const { data: missions = [], isLoading, error } = useQuery({
    queryKey: opts?.includeArchived
      ? [...queryKeys.missions.all(), 'archived']
      : queryKeys.missions.all(),
    // NOTE: lambda required here because opts must be passed as an argument;
    // the bare-reference convention (queryFn: listMissions) only applies when there are no arguments.
    queryFn: () => listMissions(opts),
  })
  return { missions, isLoading, error }
}
```

Add `archiveMission, unarchiveMission` to the existing import from `'../api'` (the barrel) in `useMission.ts`. The updated import line should be:
```typescript
import { ..., archiveMission, unarchiveMission } from '../api'
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/hooks/useMission.test.ts
```

Expected: PASS

- [ ] **Step 5: Typecheck**

```bash
cd /home/jibei/projects/scouter/frontend
npm run typecheck
```

Expected: no errors

- [ ] **Step 6: Commit**

```bash
cd /home/jibei/projects/scouter
git add frontend/src/hooks/useMission.ts frontend/src/hooks/useMission.test.ts
git commit -m "feat(hooks): add useArchiveMission, useUnarchiveMission, includeArchived to useMissions"
```

---

## Chunk 2: Mission management UI

### Task 4: MissionActionBar component

**Files:**
- Create: `frontend/src/components/mission/MissionActionBar.tsx`
- Create: `frontend/src/components/mission/MissionActionBar.module.css`
- Create: `frontend/src/components/mission/MissionActionBar.test.tsx`

The action bar renders three buttons: Edit, Archive, Delete. Delete requires typing the mission name to confirm.

- [ ] **Step 1: Write failing tests**

Create `frontend/src/components/mission/MissionActionBar.test.tsx`:

```typescript
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { MissionActionBar } from './MissionActionBar'

const defaultProps = {
  mission: { id: 'abc', slug: 'test-mission', name: 'Test Mission' },
  onEdit: vi.fn(),
  onArchive: vi.fn(),
  onDelete: vi.fn(),
}

describe('MissionActionBar', () => {
  it('renders Edit, Archive, and Delete buttons', () => {
    render(<MissionActionBar {...defaultProps} />)
    expect(screen.getByRole('button', { name: /edit/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /archive/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /delete/i })).toBeInTheDocument()
  })

  it('calls onEdit when Edit is clicked', () => {
    render(<MissionActionBar {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /edit/i }))
    expect(defaultProps.onEdit).toHaveBeenCalled()
  })

  it('shows archive confirmation dialog when Archive is clicked', () => {
    render(<MissionActionBar {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /archive/i }))
    expect(screen.getByText(/archive this mission/i)).toBeInTheDocument()
  })

  it('calls onArchive when archive is confirmed', async () => {
    render(<MissionActionBar {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /archive/i }))
    fireEvent.click(screen.getByRole('button', { name: /confirm/i }))
    await waitFor(() => expect(defaultProps.onArchive).toHaveBeenCalled())
  })

  it('shows delete confirmation with name input when Delete is clicked', () => {
    render(<MissionActionBar {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /delete/i }))
    expect(screen.getByPlaceholderText(/test mission/i)).toBeInTheDocument()
  })

  it('keeps delete confirm button disabled until name is typed correctly', () => {
    render(<MissionActionBar {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /delete/i }))
    const confirmBtn = screen.getByRole('button', { name: /confirm delete/i })
    expect(confirmBtn).toBeDisabled()
    fireEvent.change(screen.getByPlaceholderText(/test mission/i), {
      target: { value: 'Test Mission' },
    })
    expect(confirmBtn).not.toBeDisabled()
  })

  it('calls onDelete when delete is confirmed with correct name', async () => {
    render(<MissionActionBar {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /delete/i }))
    fireEvent.change(screen.getByPlaceholderText(/test mission/i), {
      target: { value: 'Test Mission' },
    })
    fireEvent.click(screen.getByRole('button', { name: /confirm delete/i }))
    await waitFor(() => expect(defaultProps.onDelete).toHaveBeenCalled())
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/components/mission/MissionActionBar.test.tsx
```

Expected: FAIL — `MissionActionBar is not defined`

- [ ] **Step 3: Implement MissionActionBar.tsx**

Create `frontend/src/components/mission/MissionActionBar.tsx`:

```typescript
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import styles from './MissionActionBar.module.css'

interface Mission {
  id: string
  slug: string
  name: string
}

interface MissionActionBarProps {
  mission: Mission
  onEdit: () => void
  onArchive: () => void
  onDelete: () => void
}

type Dialog = 'none' | 'archive' | 'delete'

export function MissionActionBar({ mission, onEdit, onArchive, onDelete }: MissionActionBarProps) {
  const { t } = useTranslation()
  const [dialog, setDialog] = useState<Dialog>('none')
  const [deleteConfirmName, setDeleteConfirmName] = useState('')

  function handleArchiveConfirm() {
    onArchive()
    setDialog('none')
  }

  function handleDeleteConfirm() {
    if (deleteConfirmName !== mission.name) return
    onDelete()
    setDialog('none')
    setDeleteConfirmName('')
  }

  function handleClose() {
    setDialog('none')
    setDeleteConfirmName('')
  }

  return (
    <>
      <div className={styles.bar}>
        <button className={styles.editBtn} onClick={onEdit}>
          {t('mission.actions.edit')}
        </button>
        <button className={styles.archiveBtn} onClick={() => setDialog('archive')}>
          {t('mission.actions.archive')}
        </button>
        <button className={styles.deleteBtn} onClick={() => setDialog('delete')}>
          {t('mission.actions.delete')}
        </button>
      </div>

      {dialog === 'archive' && (
        <div className={styles.overlay} onClick={handleClose}>
          <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
            <p>{t('mission.actions.archiveConfirm')}</p>
            <div className={styles.dialogActions}>
              <button onClick={handleClose}>{t('common.cancel')}</button>
              <button className={styles.confirmBtn} onClick={handleArchiveConfirm}>
                {t('common.confirm')}
              </button>
            </div>
          </div>
        </div>
      )}

      {dialog === 'delete' && (
        <div className={styles.overlay} onClick={handleClose}>
          <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
            <p>{t('mission.actions.deleteConfirm')}</p>
            <p className={styles.deleteHint}>{t('mission.actions.deleteTypeToConfirm')}</p>
            <input
              className={styles.deleteInput}
              placeholder={mission.name}
              value={deleteConfirmName}
              onChange={(e) => setDeleteConfirmName(e.target.value)}
              autoFocus
            />
            <div className={styles.dialogActions}>
              <button onClick={handleClose}>{t('common.cancel')}</button>
              <button
                aria-label="confirm delete"
                className={styles.deleteConfirmBtn}
                onClick={handleDeleteConfirm}
                disabled={deleteConfirmName !== mission.name}
              >
                {t('common.delete')}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
```

- [ ] **Step 4: Create MissionActionBar.module.css**

Create `frontend/src/components/mission/MissionActionBar.module.css`:

```css
.bar {
  display: flex;
  gap: 8px;
  align-items: center;
}

.editBtn,
.archiveBtn {
  background: transparent;
  border: 1px solid var(--border);
  color: var(--text-dim);
  border-radius: 6px;
  padding: 6px 12px;
  font-size: 12px;
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s;
}

.editBtn:hover {
  border-color: var(--accent);
  color: var(--text);
}

.archiveBtn:hover {
  border-color: var(--warning, #ff8f00);
  color: var(--warning, #ff8f00);
}

.deleteBtn {
  background: transparent;
  border: 1px solid transparent;
  color: var(--danger, #ef5350);
  border-radius: 6px;
  padding: 6px 12px;
  font-size: 12px;
  cursor: pointer;
  transition: border-color 0.15s;
}

.deleteBtn:hover {
  border-color: var(--danger, #ef5350);
}

.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
}

.dialog {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 24px;
  max-width: 400px;
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.deleteHint {
  font-size: 12px;
  color: var(--text-dim);
}

.deleteInput {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 8px 12px;
  color: var(--text);
  font-size: 14px;
  width: 100%;
}

.dialogActions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  margin-top: 8px;
}

.confirmBtn {
  background: var(--accent);
  color: var(--bg);
  border: none;
  border-radius: 6px;
  padding: 8px 16px;
  font-weight: 600;
  cursor: pointer;
}

.deleteConfirmBtn {
  background: var(--danger, #ef5350);
  color: #fff;
  border: none;
  border-radius: 6px;
  padding: 8px 16px;
  font-weight: 600;
  cursor: pointer;
}

.deleteConfirmBtn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/components/mission/MissionActionBar.test.tsx
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
cd /home/jibei/projects/scouter
git add frontend/src/components/mission/MissionActionBar.tsx \
        frontend/src/components/mission/MissionActionBar.module.css \
        frontend/src/components/mission/MissionActionBar.test.tsx
git commit -m "feat(ui): add MissionActionBar with edit/archive/delete actions"
```

---

### Task 5: MissionEditModal component

**Files:**
- Create: `frontend/src/components/mission/MissionEditModal.tsx`
- Create: `frontend/src/components/mission/MissionEditModal.module.css`
- Create: `frontend/src/components/mission/MissionEditModal.test.tsx`

The edit modal wraps the existing `MissionForm` component in an overlay, passing `initialValues` and wiring the submit to the update API.

- [ ] **Step 1: Write failing tests**

Create `frontend/src/components/mission/MissionEditModal.test.tsx`:

```typescript
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { MissionEditModal } from './MissionEditModal'
import type { Mission } from '../../types'

const mockMission: Partial<Mission> = {
  id: 'abc',
  slug: 'test-mission',
  name: 'Test Mission',
  budget: 2500,
  currency: 'EUR',
  category: 'computing',
  constraints: [],
}

describe('MissionEditModal', () => {
  it('renders a form with the mission name pre-filled', () => {
    render(
      <MissionEditModal
        mission={mockMission as Mission}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByDisplayValue('Test Mission')).toBeInTheDocument()
  })

  it('calls onClose when Cancel is clicked', () => {
    const onClose = vi.fn()
    render(
      <MissionEditModal
        mission={mockMission as Mission}
        onSave={vi.fn()}
        onClose={onClose}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }))
    expect(onClose).toHaveBeenCalled()
  })

  it('calls onSave with updated data on submit', () => {
    const onSave = vi.fn()
    render(
      <MissionEditModal
        mission={mockMission as Mission}
        onSave={onSave}
        onClose={vi.fn()}
      />
    )
    fireEvent.submit(screen.getByRole('form'))
    expect(onSave).toHaveBeenCalled()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/components/mission/MissionEditModal.test.tsx
```

Expected: FAIL

- [ ] **Step 3: Implement MissionEditModal.tsx**

Create `frontend/src/components/mission/MissionEditModal.tsx`:

```typescript
import { useTranslation } from 'react-i18next'
import { MissionForm } from './MissionForm'
import type { Mission } from '../../types'
import type { MissionCreateRequest } from '../../types'
import styles from './MissionEditModal.module.css'

interface MissionEditModalProps {
  mission: Mission
  onSave: (updates: Partial<MissionCreateRequest>) => void
  onClose: () => void
  loading?: boolean
  error?: string
}

export function MissionEditModal({ mission, onSave, onClose, loading, error }: MissionEditModalProps) {
  const { t } = useTranslation()

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.container} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <h2 className={styles.title}>{t('mission.actions.edit')}</h2>
          <button className={styles.closeBtn} onClick={onClose} aria-label={t('common.cancel')}>
            ×
          </button>
        </div>
        <MissionForm
          initialValues={{
            name: mission.name,
            icon: mission.icon,
            category: mission.category,
            budget: mission.budget,
            currency: mission.currency,
            constraints: mission.constraints,
            envelopeId: mission.envelopeId,
          }}
          onSubmit={onSave}
          onCancel={onClose}
          loading={loading}
          error={error}
        />
      </div>
    </div>
  )
}
```

Create `frontend/src/components/mission/MissionEditModal.module.css`:

```css
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  padding: 16px;
}

.container {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 16px;
  width: 100%;
  max-width: 560px;
  max-height: 90vh;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px 0;
}

.title {
  font-size: 16px;
  font-weight: 700;
  color: var(--text);
  margin: 0;
}

.closeBtn {
  background: none;
  border: none;
  color: var(--text-dim);
  font-size: 20px;
  cursor: pointer;
  padding: 4px 8px;
  line-height: 1;
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/components/mission/MissionEditModal.test.tsx
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /home/jibei/projects/scouter
git add frontend/src/components/mission/MissionEditModal.tsx \
        frontend/src/components/mission/MissionEditModal.module.css \
        frontend/src/components/mission/MissionEditModal.test.tsx
git commit -m "feat(ui): add MissionEditModal wrapping MissionForm"
```

---

### Task 6: Wire MissionActionBar into MissionOverview

**Files:**
- Modify: `frontend/src/pages/MissionOverview.tsx`

`MissionOverview.tsx` is 559 lines. The header section is around lines 165–186. We need to:
1. Import `MissionActionBar` and `MissionEditModal`
2. Import `useArchiveMission`, `useUnarchiveMission` from hooks
3. Add `showEditModal` state
4. Add action handlers (handleArchive, handleDelete, handleEdit)
5. Render `<MissionActionBar>` in the header
6. Render `<MissionEditModal>` conditionally
7. Navigate to `/` after delete

- [ ] **Step 1: Add imports and state**

Open `frontend/src/pages/MissionOverview.tsx`.

Add to existing imports (near the top):
```typescript
import { MissionActionBar } from '../components/mission/MissionActionBar'
import { MissionEditModal } from '../components/mission/MissionEditModal'
```

`MissionOverview.tsx` line 10 imports from `'../hooks'` — add `useDeleteMission` and `useArchiveMission` to the existing destructure on that line. The line should now include:
```typescript
import { useMission, useShopping, useResearch, usePriceIntel, useUpdateMission, useKeyboardShortcuts, usePurchaseRecord, useSuggestCategory, useDeleteMission, useArchiveMission } from '../hooks'
```

Find where `useUpdateMission` is called (around line 30–50). Add alongside it:
```typescript
const { mutate: archiveMission } = useArchiveMission()
const { mutate: unarchiveMission } = useUnarchiveMission()
const [showEditModal, setShowEditModal] = useState(false)
```

- [ ] **Step 2: Add action handlers**

Find the `handlePhase` function. Add after it:
```typescript
function handleArchive() {
  if (!mission?.id) return
  archiveMission(mission.id, {
    onSuccess: () => navigate('/'),
  })
}

function handleDelete() {
  if (!mission?.slug) return
  deleteMission(mission.slug, {
    onSuccess: () => navigate('/'),
  })
}
```

Note: `deleteMission` already exists from `useDeleteMission`. Check if it's already destructured from a hook — if not, add:
```typescript
const { mutate: deleteMission } = useDeleteMission()
```
And import `useDeleteMission` if not already imported.

- [ ] **Step 3: Add MissionActionBar to header**

Find the mission header section (around lines 165–186). It renders the mission name and status. Add `<MissionActionBar>` immediately after the mission name block:

```typescript
{mission && (
  <MissionActionBar
    mission={mission}
    onEdit={() => setShowEditModal(true)}
    onArchive={handleArchive}
    onDelete={handleDelete}
  />
)}
```

- [ ] **Step 4: Add MissionEditModal conditional render**

Near the bottom of the return statement (before the closing tag), add:

```typescript
{showEditModal && mission && (
  <MissionEditModal
    mission={mission}
    onSave={(updates) => {
      updateMission(updates, { onSuccess: () => setShowEditModal(false) })
    }}
    onClose={() => setShowEditModal(false)}
  />
)}
```

Note: `updateMission` already exists from `useUpdateMission`. Check the existing destructure pattern and match it.

- [ ] **Step 5: Typecheck and run existing tests**

```bash
cd /home/jibei/projects/scouter/frontend
npm run typecheck
npm run test -- --run
```

Expected: typecheck passes, existing tests pass

- [ ] **Step 6: Commit**

```bash
cd /home/jibei/projects/scouter
git add frontend/src/pages/MissionOverview.tsx
git commit -m "feat(mission): wire MissionActionBar and MissionEditModal into MissionOverview"
```

---

### Task 7: MissionCard ⋯ menu + HQ Dashboard archived filter

**Files:**
- Modify: `frontend/src/components/mission/MissionCard.tsx`
- Modify: `frontend/src/components/mission/MissionCard.module.css`
- Modify: `frontend/src/pages/HQDashboard.tsx`

MissionCard is 133 lines. It currently has swipe-to-archive. We're adding a `⋯` menu with Archive/Delete (and Unarchive for archived missions).

- [ ] **Step 1: Write failing tests**

Create `frontend/src/components/mission/MissionCard.test.tsx` (new file):

```typescript
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import { MissionCard } from './MissionCard'
import type { Mission } from '../../types'

vi.mock('../../hooks/useMission', () => ({
  useArchiveMission: () => ({ mutate: vi.fn() }),
  useUnarchiveMission: () => ({ mutate: vi.fn() }),
  useDeleteMission: () => ({ mutate: vi.fn() }),
}))

const mockMission: Partial<Mission> = {
  id: 'abc',
  slug: 'test-mission',
  name: 'Test Mission',
  phase: 'researching',
  budget: 2500,
  currency: 'EUR',
  category: 'computing',
  constraints: [],
  archivedAt: null,
}

function renderCard(props = {}) {
  return render(
    <MemoryRouter>
      <MissionCard mission={mockMission as Mission} {...props} />
    </MemoryRouter>
  )
}

describe('MissionCard ⋯ menu', () => {
  it('renders the overflow menu button', () => {
    renderCard()
    expect(screen.getByRole('button', { name: /more options/i })).toBeInTheDocument()
  })

  it('opens menu on click showing Archive and Delete', () => {
    renderCard()
    fireEvent.click(screen.getByRole('button', { name: /more options/i }))
    expect(screen.getByText(/archive/i)).toBeInTheDocument()
    expect(screen.getByText(/delete/i)).toBeInTheDocument()
  })

  it('shows Unarchive instead of Archive when mission is archived', () => {
    renderCard({ mission: { ...mockMission, archivedAt: '2026-01-01T00:00:00Z' } as Mission })
    fireEvent.click(screen.getByRole('button', { name: /more options/i }))
    expect(screen.getByText(/unarchive/i)).toBeInTheDocument()
    expect(screen.queryByText(/^archive$/i)).not.toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/components/mission/MissionCard.test.tsx
```

Expected: FAIL

- [ ] **Step 3: Add ⋯ menu to MissionCard.tsx**

Open `frontend/src/components/mission/MissionCard.tsx`.

Add imports at top:
```typescript
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useArchiveMission, useUnarchiveMission, useDeleteMission } from '../../hooks/useMission'
```

Note: `MissionCard.tsx` currently has NO `useTranslation` import and NO `t()` calls. Add `const { t } = useTranslation()` unconditionally inside the component body.

Inside the component body, add hooks and menu state:
```typescript
const { t } = useTranslation()
const { mutate: archiveMission } = useArchiveMission()
const { mutate: unarchiveMission } = useUnarchiveMission()
const { mutate: deleteMission } = useDeleteMission()
const [menuOpen, setMenuOpen] = useState(false)
const [confirmDelete, setConfirmDelete] = useState(false)
const [deleteConfirmName, setDeleteConfirmName] = useState('')
const isArchived = !!mission.archivedAt
```

**Swipe gesture note:** Keep the existing `onArchive` prop for the swipe gesture — it continues to work as before. The new `⋯` menu calls `useArchiveMission` directly (the hook internally). Both paths coexist: swipe = archive via `onArchive` prop callback, menu = archive via the hook.

**Dialog positioning note:** The `.swipeWrapper` in MissionCard has a CSS `transform` (`--swipe-x`) applied, which breaks `position: fixed` overlays inside it. For the delete confirmation dialog, use `position: absolute; inset: 0` on the overlay and ensure the `.swipeWrapper` container has `position: relative`. This keeps the dialog within the card's stacking context instead of breaking out via fixed positioning.

Add the `⋯` button and its dropdown into the card's header area. After the existing status badge or clone button, add:

```typescript
<div className={styles.menuContainer}>
  <button
    className={styles.menuBtn}
    aria-label="more options"
    onClick={(e) => { e.stopPropagation(); setMenuOpen((o) => !o) }}
  >
    ⋯
  </button>
  {menuOpen && (
    <div className={styles.menu} onClick={(e) => e.stopPropagation()}>
      {isArchived ? (
        <button onClick={() => { unarchiveMission(mission.id); setMenuOpen(false) }}>
          {t('mission.actions.unarchive')}
        </button>
      ) : (
        <button onClick={() => { archiveMission(mission.id); setMenuOpen(false) }}>
          {t('mission.actions.archive')}
        </button>
      )}
      <button
        className={styles.menuDeleteBtn}
        onClick={() => { setConfirmDelete(true); setMenuOpen(false) }}
      >
        {t('mission.actions.delete')}
      </button>
    </div>
  )}
</div>
```

Add the delete confirmation dialog (before the closing JSX):
```typescript
{confirmDelete && (
  <div className={styles.overlay} onClick={() => setConfirmDelete(false)}>
    <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
      <p>{t('mission.actions.deleteConfirm')}</p>
      <p className={styles.deleteHint}>{t('mission.actions.deleteTypeToConfirm')}</p>
      <input
        placeholder={mission.name}
        value={deleteConfirmName}
        onChange={(e) => setDeleteConfirmName(e.target.value)}
        autoFocus
      />
      <div className={styles.dialogActions}>
        <button onClick={() => setConfirmDelete(false)}>{t('common.cancel')}</button>
        <button
          disabled={deleteConfirmName !== mission.name}
          onClick={() => {
            deleteMission(mission.slug)
            setConfirmDelete(false)
            setDeleteConfirmName('')
          }}
        >
          {t('common.delete')}
        </button>
      </div>
    </div>
  </div>
)}
```

Add the `t` import if not already present: `const { t } = useTranslation()`

Add CSS classes to `MissionCard.module.css`:
```css
.menuContainer {
  position: relative;
}

.menuBtn {
  background: none;
  border: none;
  color: var(--text-dim);
  cursor: pointer;
  padding: 4px 8px;
  font-size: 16px;
  border-radius: 4px;
}

.menuBtn:hover {
  background: var(--surface-hover, rgba(255,255,255,0.05));
}

.menu {
  position: absolute;
  top: 100%;
  right: 0;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 4px;
  min-width: 140px;
  z-index: 50;
  display: flex;
  flex-direction: column;
}

.menu button {
  background: none;
  border: none;
  color: var(--text);
  text-align: left;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.menu button:hover {
  background: var(--surface-hover, rgba(255,255,255,0.05));
}

.menuDeleteBtn {
  color: var(--danger, #ef5350) !important;
}
```

Also add `.swipeWrapper` with `position: relative` (or ensure it already has it) and the overlay using `position: absolute` (not `fixed`) to avoid breaking out of the CSS transform context:

```css
/* Ensure the swipe wrapper creates a containing block */
.swipeWrapper {
  position: relative;
}

/* Delete confirmation overlay — absolute (not fixed) to stay within the swipe transform context */
.overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  border-radius: inherit;
}

.dialog {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 20px;
  max-width: 320px;
  width: 90%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.deleteHint {
  font-size: 12px;
  color: var(--text-dim);
}

.dialogActions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}
```

- [ ] **Step 4: Add archived badge to MissionCard**

Find where the mission name/status is displayed and add an "Archived" badge when `isArchived`:

```typescript
{isArchived && (
  <span className={styles.archivedBadge}>{t('mission.actions.archivedBadge')}</span>
)}
```

Add CSS:
```css
.archivedBadge {
  background: var(--text-dim);
  color: var(--bg);
  border-radius: 4px;
  padding: 2px 6px;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
```

- [ ] **Step 5: Add "Show archived" filter to HQDashboard**

Open `frontend/src/pages/HQDashboard.tsx`.

Add state:
```typescript
const [showArchived, setShowArchived] = useState(false)
```

Update the `useMissions` call. Note: after Task 3, `useMissions` returns `{ missions, isLoading, error }` — NOT `{ data: missions }`. Use the correct destructure:
```typescript
// Before:
const { missions } = useMissions()

// After:
const { missions } = useMissions({ includeArchived: showArchived })
```

In the dashboard header (find where the "New Mission" button is rendered), add a toggle:
```typescript
<button
  className={showArchived ? styles.filterActive : styles.filterBtn}
  onClick={() => setShowArchived((s) => !s)}
>
  {t('mission.actions.showArchived')}
</button>
```

- [ ] **Step 6: Run tests and typecheck**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/components/mission/MissionCard.test.tsx
npm run typecheck
```

Expected: tests PASS, no type errors

- [ ] **Step 7: Commit**

```bash
cd /home/jibei/projects/scouter
git add frontend/src/components/mission/MissionCard.tsx \
        frontend/src/components/mission/MissionCard.module.css \
        frontend/src/components/mission/MissionCard.test.tsx \
        frontend/src/pages/HQDashboard.tsx
git commit -m "feat(ui): add mission card overflow menu and HQ Dashboard archived filter"
```

---

## Chunk 3: Option management UI

### Task 8: OptionEditModal component

**Files:**
- Create: `frontend/src/components/options/OptionEditModal.tsx`
- Create: `frontend/src/components/options/OptionEditModal.module.css`
- Create: `frontend/src/components/options/OptionEditModal.test.tsx`

The form edits: name, badge (dropdown), price range min/max, notes, warnings. Attributes shown read-only.

Valid badge values: `'recommended' | 'alternative' | 'watch' | 'rejected'`

- [ ] **Step 1: Write failing tests**

Create `frontend/src/components/options/OptionEditModal.test.tsx`:

```typescript
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { OptionEditModal } from './OptionEditModal'
import type { Option } from '../../types'

const mockOption: Partial<Option> = {
  id: 'opt-1',
  missionId: 'mission-1',
  name: 'MacBook Pro M4',
  badge: 'recommended',
  priceRange: { min: 2000, max: 2500, best: 2200 },
  notes: 'Great laptop',
  warnings: 'Expensive',
  attributes: [],
}

describe('OptionEditModal', () => {
  it('renders with option name pre-filled', () => {
    render(
      <OptionEditModal option={mockOption as Option} onSave={vi.fn()} onClose={vi.fn()} />
    )
    expect(screen.getByDisplayValue('MacBook Pro M4')).toBeInTheDocument()
  })

  it('shows badge dropdown with current badge selected', () => {
    render(
      <OptionEditModal option={mockOption as Option} onSave={vi.fn()} onClose={vi.fn()} />
    )
    const select = screen.getByRole('combobox')
    expect((select as HTMLSelectElement).value).toBe('recommended')
  })

  it('calls onClose when Cancel is clicked', () => {
    const onClose = vi.fn()
    render(
      <OptionEditModal option={mockOption as Option} onSave={vi.fn()} onClose={onClose} />
    )
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }))
    expect(onClose).toHaveBeenCalled()
  })

  it('calls onSave with updated fields on submit', () => {
    const onSave = vi.fn()
    render(
      <OptionEditModal option={mockOption as Option} onSave={onSave} onClose={vi.fn()} />
    )
    fireEvent.change(screen.getByDisplayValue('MacBook Pro M4'), {
      target: { value: 'MacBook Pro M4 Pro' },
    })
    fireEvent.submit(screen.getByRole('form'))
    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({ name: 'MacBook Pro M4 Pro' })
    )
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/components/options/OptionEditModal.test.tsx
```

Expected: FAIL

- [ ] **Step 3: Implement OptionEditModal.tsx**

Create `frontend/src/components/options/OptionEditModal.tsx`:

```typescript
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { Option } from '../../types'
import styles from './OptionEditModal.module.css'

const BADGES = ['recommended', 'alternative', 'watch', 'rejected'] as const
type Badge = typeof BADGES[number]

interface OptionEditModalProps {
  option: Option
  onSave: (updates: Partial<Option>) => void
  onClose: () => void
  loading?: boolean
  error?: string
}

export function OptionEditModal({ option, onSave, onClose, loading, error }: OptionEditModalProps) {
  const { t } = useTranslation()
  const [name, setName] = useState(option.name)
  const [badge, setBadge] = useState<Badge>(option.badge as Badge)
  const [priceMin, setPriceMin] = useState(String(option.priceRange?.min ?? ''))
  const [priceMax, setPriceMax] = useState(String(option.priceRange?.max ?? ''))
  const [notes, setNotes] = useState(option.notes ?? '')
  const [warnings, setWarnings] = useState(option.warnings ?? '')

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    onSave({
      name,
      badge,
      priceRange: {
        min: parseFloat(priceMin) || 0,
        max: parseFloat(priceMax) || 0,
        best: option.priceRange?.best ?? parseFloat(priceMin) || 0,
      },
      notes,
      warnings,
    })
  }

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.container} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <h2 className={styles.title}>{t('option.actions.edit')}</h2>
          <button className={styles.closeBtn} onClick={onClose} aria-label={t('common.cancel')}>×</button>
        </div>

        <form aria-label="form" onSubmit={handleSubmit} className={styles.form}>
          {error && <p className={styles.error}>{error}</p>}

          <label className={styles.label}>{t('common.name')}</label>
          <input
            className={styles.input}
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />

          <label className={styles.label}>{t('option.actions.badgeLabel')}</label>
          <select className={styles.select} value={badge} onChange={(e) => setBadge(e.target.value as Badge)}>
            {BADGES.map((b) => (
              <option key={b} value={b}>{t(`options.badge.${b}`)}</option>
            ))}
          </select>

          <div className={styles.priceRow}>
            <div>
              <label className={styles.label}>{t('common.priceMin', 'Min price')}</label>
              <input
                className={styles.input}
                type="number"
                value={priceMin}
                onChange={(e) => setPriceMin(e.target.value)}
              />
            </div>
            <div>
              <label className={styles.label}>{t('common.priceMax', 'Max price')}</label>
              <input
                className={styles.input}
                type="number"
                value={priceMax}
                onChange={(e) => setPriceMax(e.target.value)}
              />
            </div>
          </div>

          <label className={styles.label}>{t('common.notes', 'Notes')}</label>
          <textarea className={styles.textarea} value={notes} onChange={(e) => setNotes(e.target.value)} rows={3} />

          <label className={styles.label}>{t('common.warnings', 'Warnings')}</label>
          <textarea className={styles.textarea} value={warnings} onChange={(e) => setWarnings(e.target.value)} rows={2} />

          {option.attributes.length > 0 && (
            <div className={styles.attributesSection}>
              <p className={styles.attributesLabel}>{t('common.attributes', 'Attributes')} ({t('common.readOnly', 'read-only')})</p>
              <ul className={styles.attributeList}>
                {option.attributes.map((attr) => (
                  <li key={attr.key} className={styles.attributeItem}>
                    <span className={styles.attrLabel}>{attr.label}</span>
                    <span className={styles.attrValue}>{String(attr.value)}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          <div className={styles.actions}>
            <button type="button" onClick={onClose}>{t('common.cancel')}</button>
            <button type="submit" className={styles.saveBtn} disabled={loading}>
              {loading ? t('common.saving', 'Saving…') : t('common.save')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
```

Create `frontend/src/components/options/OptionEditModal.module.css`:

```css
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  padding: 16px;
}

.container {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 16px;
  width: 100%;
  max-width: 480px;
  max-height: 90vh;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px 0;
}

.title {
  font-size: 16px;
  font-weight: 700;
  color: var(--text);
  margin: 0;
}

.closeBtn {
  background: none;
  border: none;
  color: var(--text-dim);
  font-size: 20px;
  cursor: pointer;
  padding: 4px 8px;
}

.form {
  padding: 20px 24px 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-dim);
  font-weight: 600;
}

.input,
.select,
.textarea {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 8px 12px;
  color: var(--text);
  font-size: 14px;
  width: 100%;
}

.textarea {
  resize: vertical;
  font-family: inherit;
}

.priceRow {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.attributesSection {
  border-top: 1px solid var(--border);
  padding-top: 12px;
}

.attributesLabel {
  font-size: 11px;
  text-transform: uppercase;
  color: var(--text-dim);
  margin-bottom: 8px;
}

.attributeList {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.attributeItem {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--text-dim);
}

.attrValue {
  color: var(--text);
}

.actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  padding-top: 8px;
  border-top: 1px solid var(--border);
}

.saveBtn {
  background: var(--accent);
  color: var(--bg);
  border: none;
  border-radius: 6px;
  padding: 8px 16px;
  font-weight: 600;
  cursor: pointer;
}

.saveBtn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.error {
  color: var(--danger, #ef5350);
  font-size: 13px;
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/components/options/OptionEditModal.test.tsx
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /home/jibei/projects/scouter
git add frontend/src/components/options/OptionEditModal.tsx \
        frontend/src/components/options/OptionEditModal.module.css \
        frontend/src/components/options/OptionEditModal.test.tsx
git commit -m "feat(ui): add OptionEditModal with name/badge/price/notes/warnings form"
```

---

### Task 9: OptionCard ⋯ menu (edit/delete) + wire into OptionsExplorer

**Files:**
- Modify: `frontend/src/components/options/OptionCard.tsx`
- Modify: `frontend/src/components/options/OptionCard.module.css`
- Modify: `frontend/src/pages/OptionsExplorer.tsx`

OptionCard currently receives `onPin`, `onReject`, `onUnreject` as props. We'll add `onEdit` and `onDelete` props.

- [ ] **Step 1: Write failing tests**

Create `frontend/src/components/options/OptionCard.menu.test.tsx`:

```typescript
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { OptionCard } from './OptionCard'
import type { Option } from '../../types'

const mockOption: Partial<Option> = {
  id: 'opt-1',
  missionId: 'mission-1',
  name: 'MacBook Pro',
  badge: 'recommended',
  priceRange: { min: 2000, max: 2500, best: 2200 },
  attributes: [],
  pinned: false,
  rejected: false,
}

function renderCard(props = {}) {
  return render(
    <OptionCard
      option={mockOption as Option}
      missionId="mission-1"
      onEdit={vi.fn()}
      onDelete={vi.fn()}
      onPin={vi.fn()}
      onReject={vi.fn()}
      onUnreject={vi.fn()}
      {...props}
    />
  )
}

describe('OptionCard ⋯ menu', () => {
  it('renders the overflow menu button', () => {
    renderCard()
    expect(screen.getByRole('button', { name: /more options/i })).toBeInTheDocument()
  })

  it('opens menu showing Edit and Delete', () => {
    renderCard()
    fireEvent.click(screen.getByRole('button', { name: /more options/i }))
    expect(screen.getByText(/edit option/i)).toBeInTheDocument()
    expect(screen.getByText(/delete option/i)).toBeInTheDocument()
  })

  it('calls onEdit when Edit is clicked', () => {
    const onEdit = vi.fn()
    renderCard({ onEdit })
    fireEvent.click(screen.getByRole('button', { name: /more options/i }))
    fireEvent.click(screen.getByText(/edit option/i))
    expect(onEdit).toHaveBeenCalledWith('opt-1')
  })

  it('shows delete confirmation when Delete is clicked', () => {
    renderCard()
    fireEvent.click(screen.getByRole('button', { name: /more options/i }))
    fireEvent.click(screen.getByText(/delete option/i))
    expect(screen.getByText(/remove this option/i)).toBeInTheDocument()
  })

  it('calls onDelete when delete is confirmed', () => {
    const onDelete = vi.fn()
    renderCard({ onDelete })
    fireEvent.click(screen.getByRole('button', { name: /more options/i }))
    fireEvent.click(screen.getByText(/delete option/i))
    fireEvent.click(screen.getByRole('button', { name: /confirm/i }))
    expect(onDelete).toHaveBeenCalledWith('opt-1')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/components/options/OptionCard.menu.test.tsx
```

Expected: FAIL

- [ ] **Step 3: Add ⋯ menu to OptionCard.tsx**

Open `frontend/src/components/options/OptionCard.tsx`. The existing props interface (lines 16–29) has `onPin`, `onReject`, `onUnreject`. Add:

```typescript
onEdit?: (optionId: string) => void
onDelete?: (optionId: string) => void
```

Add state inside the component:
```typescript
const [menuOpen, setMenuOpen] = useState(false)
const [confirmDelete, setConfirmDelete] = useState(false)
```

Import `useState` if not already imported.

Add the menu button in the card header area (alongside the existing pin button, around line 252):

```typescript
<div className={styles.menuWrapper}>
  <button
    className={styles.menuBtn}
    aria-label="more options"
    onClick={(e) => { e.stopPropagation(); setMenuOpen((o) => !o) }}
  >
    ⋯
  </button>
  {menuOpen && (
    <div className={styles.menu}>
      {onEdit && (
        <button className={styles.menuItem} onClick={() => { onEdit(option.id); setMenuOpen(false) }}>
          {t('option.actions.edit')}
        </button>
      )}
      {onDelete && (
        <button
          className={`${styles.menuItem} ${styles.menuItemDanger}`}
          onClick={() => { setConfirmDelete(true); setMenuOpen(false) }}
        >
          {t('option.actions.delete')}
        </button>
      )}
    </div>
  )}
</div>
```

Add delete confirmation at the end of the return (before closing tag):

```typescript
{confirmDelete && (
  <div className={styles.overlay} onClick={() => setConfirmDelete(false)}>
    <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
      <p>{t('option.actions.deleteConfirm')}</p>
      <div className={styles.dialogActions}>
        <button onClick={() => setConfirmDelete(false)}>{t('common.cancel')}</button>
        <button
          aria-label="confirm"
          className={styles.deleteConfirmBtn}
          onClick={() => { onDelete?.(option.id); setConfirmDelete(false) }}
        >
          {t('common.delete')}
        </button>
      </div>
    </div>
  </div>
)}
```

Add the following CSS to `OptionCard.module.css`:

```css
.menuWrapper {
  position: relative;
}

.menuBtn {
  background: none;
  border: none;
  color: var(--text-dim);
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 1rem;
  line-height: 1;
}

.menuBtn:hover {
  color: var(--text);
  background: var(--surface-hover);
}

.menu {
  position: absolute;
  top: 100%;
  right: 0;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
  z-index: 100;
  min-width: 160px;
  overflow: hidden;
}

.menuItem {
  display: block;
  width: 100%;
  padding: 10px 14px;
  background: none;
  border: none;
  text-align: left;
  cursor: pointer;
  font-size: 0.875rem;
  color: var(--text);
}

.menuItem:hover {
  background: var(--surface-hover);
}

.menuItemDanger {
  color: var(--status-crisis);
}

.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
}

.dialog {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 24px;
  max-width: 360px;
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.dialogActions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.deleteConfirmBtn {
  background: var(--status-crisis);
  color: #fff;
  border: none;
  border-radius: 6px;
  padding: 8px 16px;
  font-weight: 600;
  cursor: pointer;
}
```

- [ ] **Step 4: Wire edit/delete into OptionsExplorer**

Open `frontend/src/pages/OptionsExplorer.tsx`.

Find where `useOptions` and mutations are imported/used. Add:
```typescript
const { mutate: deleteOption } = useDeleteOption(missionId)
```

Add state for the edit modal:
```typescript
const [editingOption, setEditingOption] = useState<Option | null>(null)
```

Import `OptionEditModal` at the top:
```typescript
import { OptionEditModal } from '../components/options/OptionEditModal'
```

Find where `<OptionCard>` is rendered (around line 209). Add the new props:
```typescript
<OptionCard
  ...existing props...
  onEdit={(id) => setEditingOption(options.find((o) => o.id === id) ?? null)}
  onDelete={(id) => deleteOption(id)}
/>
```

At the end of the return, add:
```typescript
{editingOption && (
  <OptionEditModal
    option={editingOption}
    onSave={(updates) => {
      updateOption({ optionId: editingOption.id, ...updates }, {
        onSuccess: () => setEditingOption(null),
      })
    }}
    onClose={() => setEditingOption(null)}
  />
)}
```

Note: `useUpdateOption` should already exist in `useOptions.ts` (it does, line 20). Check the mutation call signature and match it. If `updateOption` from `useUpdateOption` is not already used in OptionsExplorer, add:
```typescript
const { mutate: updateOption } = useUpdateOption(missionId)
```

- [ ] **Step 5: Run tests and typecheck**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/components/options/OptionCard.menu.test.tsx
npm run typecheck
```

Expected: tests PASS, no type errors

- [ ] **Step 6: Commit**

```bash
cd /home/jibei/projects/scouter
git add frontend/src/components/options/OptionCard.tsx \
        frontend/src/components/options/OptionCard.module.css \
        frontend/src/components/options/OptionCard.menu.test.tsx \
        frontend/src/pages/OptionsExplorer.tsx
git commit -m "feat(ui): add option card overflow menu with edit/delete, wire OptionEditModal"
```

---

### Task 10: Unpin All (fix "Clear Pinned")

**Files:**
- Modify: `frontend/src/hooks/useOptions.ts`
- Modify: `frontend/src/pages/OptionsExplorer.tsx`

The current `useDeletePinnedOptions` calls `DELETE /api/missions/{id}/options/pinned` which deletes the options. We need a new `useUnpinAllOptions` hook that calls `pinOption` for each pinned option in parallel to toggle the pin off.

- [ ] **Step 1: Write failing test**

Create `frontend/src/hooks/useUnpinAll.test.tsx`:

```typescript
import { renderHook, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi } from 'vitest'
import React from 'react'

vi.mock('../api/options', () => ({
  listOptions: vi.fn().mockResolvedValue([]),
  pinOption: vi.fn().mockResolvedValue({ id: 'opt-1', pinned: false }),
  deleteOption: vi.fn(),
  updateOption: vi.fn(),
  rejectOption: vi.fn(),
  unrejectOption: vi.fn(),
  deletePinnedOptions: vi.fn(),
  retranslateOption: vi.fn(),
  getOption: vi.fn(),
  createOption: vi.fn(),
  getOptionsExportURL: vi.fn(),
}))

function wrapper({ children }: { children: React.ReactNode }) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return React.createElement(QueryClientProvider, { client: queryClient }, children)
}

describe('useUnpinAllOptions', () => {
  it('exports useUnpinAllOptions', async () => {
    const { useUnpinAllOptions } = await import('./useOptions')
    expect(typeof useUnpinAllOptions).toBe('function')
  })

  it('calls pinOption for each provided pinned option id', async () => {
    const { pinOption } = await import('../api/options')
    const { useUnpinAllOptions } = await import('./useOptions')
    const { result } = renderHook(() => useUnpinAllOptions('mission-1'), { wrapper })
    await act(async () => {
      await result.current.unpinAllOptions(['opt-1', 'opt-2'])
    })
    expect(pinOption).toHaveBeenCalledWith('mission-1', 'opt-1')
    expect(pinOption).toHaveBeenCalledWith('mission-1', 'opt-2')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/hooks/useUnpinAll.test.tsx
```

Expected: FAIL

- [ ] **Step 3: Add useUnpinAllOptions to useOptions.ts**

Open `frontend/src/hooks/useOptions.ts`. Add `import { useTranslation } from 'react-i18next'` to the imports if not already present (it is NOT currently in `useOptions.ts` — add it). Add after `useDeletePinnedOptions`:

```typescript
export function useUnpinAllOptions(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { t } = useTranslation()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: async (pinnedIds: string[]) => {
      const results = await Promise.allSettled(
        pinnedIds.map(id => pinOption(missionId, id))
      )
      const failed = results.filter(r => r.status === 'rejected').length
      if (failed > 0) {
        toast(t('option.actions.unpinAllError', { count: failed }), 'error')
      }
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.options.all(missionId) })
    },
  })
  return { unpinAllOptions: mutateAsync, isPending }
}
```

Note: uses `Promise.allSettled` (not `Promise.all`) so partial failures are reported via toast rather than throwing. The return value is `unpinAllOptions` (not `unpinAll`) — update call sites accordingly.

Import `pinOption` if not already in the import from `'../api'` barrel (it should already be there as `usePinOption` uses it).

- [ ] **Step 4: Update OptionsExplorer to use "Unpin All"**

Open `frontend/src/pages/OptionsExplorer.tsx`.

Find the "Clear Pinned" button (around line 119–127):
```typescript
// Before:
{pinnedCount > 0 && (
  <button onClick={() => deletePinnedOptions()} className={styles.clearPinnedBtn}>
    {t('options.clearPinned', { count: pinnedCount })}
  </button>
)}
```

Add the new hook:
```typescript
const { unpinAllOptions } = useUnpinAllOptions(missionId)
```

Update the button:
```typescript
// After:
{pinnedCount > 0 && (
  <button
    onClick={() => unpinAllOptions(options.filter((o) => o.pinned).map((o) => o.id))}
    className={styles.clearPinnedBtn}
  >
    {t('option.actions.unpinAll', { count: pinnedCount })}
  </button>
)}
```

- [ ] **Step 5: Run tests and typecheck**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/hooks/useUnpinAll.test.tsx
npm run typecheck
```

Expected: PASS, no type errors

- [ ] **Step 6: Commit**

```bash
cd /home/jibei/projects/scouter
git add frontend/src/hooks/useOptions.ts \
        frontend/src/hooks/useUnpinAll.test.tsx \
        frontend/src/pages/OptionsExplorer.tsx
git commit -m "feat(ui): rename Clear Pinned to Unpin All, change behavior to unpin not delete"
```

---

## Chunk 4: Shortlist panel (buying phase)

### Task 11: ShortlistPanel component

**Files:**
- Create: `frontend/src/components/mission/ShortlistPanel.tsx`
- Create: `frontend/src/components/mission/ShortlistPanel.module.css`
- Create: `frontend/src/components/mission/ShortlistPanel.test.tsx`

The panel shows pinned options. Each has a "Select" button that calls `onSelect(option)`. If no pinned options, shows an empty state with a link to Options.

- [ ] **Step 1: Write failing tests**

Create `frontend/src/components/mission/ShortlistPanel.test.tsx`:

```typescript
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import { ShortlistPanel } from './ShortlistPanel'
import type { Option } from '../../types'

const pinnedOption: Partial<Option> = {
  id: 'opt-1',
  name: 'MacBook Pro',
  badge: 'recommended',
  priceRange: { min: 2000, max: 2500, best: 2200 },
  pinned: true,
}

function renderPanel(options: Partial<Option>[] = [], onSelect = vi.fn()) {
  return render(
    <MemoryRouter>
      <ShortlistPanel
        options={options as Option[]}
        missionSlug="test-mission"
        onSelect={onSelect}
      />
    </MemoryRouter>
  )
}

describe('ShortlistPanel', () => {
  it('shows the panel title', () => {
    renderPanel([pinnedOption])
    expect(screen.getByText(/your shortlist/i)).toBeInTheDocument()
  })

  it('renders pinned options as cards', () => {
    renderPanel([pinnedOption])
    expect(screen.getByText('MacBook Pro')).toBeInTheDocument()
  })

  it('renders a Select button per option', () => {
    renderPanel([pinnedOption])
    expect(screen.getByRole('button', { name: /select/i })).toBeInTheDocument()
  })

  it('calls onSelect with the option when Select is clicked', () => {
    const onSelect = vi.fn()
    renderPanel([pinnedOption], onSelect)
    fireEvent.click(screen.getByRole('button', { name: /select/i }))
    expect(onSelect).toHaveBeenCalledWith(pinnedOption)
  })

  it('shows empty state when no options are pinned', () => {
    renderPanel([])
    expect(screen.getByText(/pin your top picks/i)).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /go to options/i })).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/components/mission/ShortlistPanel.test.tsx
```

Expected: FAIL

- [ ] **Step 3: Implement ShortlistPanel.tsx**

Create `frontend/src/components/mission/ShortlistPanel.tsx`:

```typescript
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import type { Option } from '../../types'
import styles from './ShortlistPanel.module.css'

interface ShortlistPanelProps {
  options: Option[]
  missionSlug: string
  onSelect: (option: Option) => void
}

export function ShortlistPanel({ options, missionSlug, onSelect }: ShortlistPanelProps) {
  const { t } = useTranslation()
  const pinned = options.filter((o) => o.pinned)

  return (
    <div className={styles.panel}>
      <h3 className={styles.title}>{t('shortlist.title')}</h3>

      {pinned.length === 0 ? (
        <div className={styles.empty}>
          <p>{t('shortlist.empty')}</p>
          <Link to={`/missions/${missionSlug}/options`} className={styles.emptyLink}>
            {t('shortlist.emptyLink')}
          </Link>
        </div>
      ) : (
        <div className={styles.list}>
          {pinned.map((option) => (
            <div key={option.id} className={styles.card}>
              <div className={styles.cardInfo}>
                <span className={styles.optionName}>{option.name}</span>
                {option.priceRange && (
                  <span className={styles.priceRange}>
                    {option.priceRange.min}–{option.priceRange.max}
                  </span>
                )}
                {option.badge && (
                  <span className={`${styles.badge} ${styles[option.badge]}`}>
                    {option.badge}
                  </span>
                )}
              </div>
              <button
                className={styles.selectBtn}
                onClick={() => onSelect(option)}
              >
                {t('shortlist.select')}
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
```

Create `frontend/src/components/mission/ShortlistPanel.module.css`:

```css
.panel {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 16px 20px;
  margin-bottom: 16px;
}

.title {
  font-size: 13px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--accent);
  margin: 0 0 12px;
}

.empty {
  display: flex;
  flex-direction: column;
  gap: 8px;
  color: var(--text-dim);
  font-size: 13px;
}

.emptyLink {
  color: var(--accent);
  font-size: 13px;
}

.list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 8px;
}

.cardInfo {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  min-width: 0;
}

.optionName {
  font-weight: 600;
  font-size: 13px;
  color: var(--text);
}

.priceRange {
  font-size: 12px;
  color: var(--text-dim);
}

.badge {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 600;
}

.recommended { background: rgba(102, 187, 106, 0.15); color: #66bb6a; }
.alternative { background: rgba(79, 195, 247, 0.15); color: #4fc3f7; }
.watch { background: rgba(186, 104, 200, 0.15); color: #ba68c8; }
.rejected { background: rgba(239, 83, 80, 0.15); color: #ef5350; }

.selectBtn {
  background: var(--accent);
  color: var(--bg);
  border: none;
  border-radius: 6px;
  padding: 6px 14px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  flex-shrink: 0;
}

.selectBtn:hover {
  opacity: 0.9;
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/components/mission/ShortlistPanel.test.tsx
```

Expected: PASS

- [ ] **Step 5: Export ShortlistPanel from the mission barrel**

Open `frontend/src/components/mission/index.ts`. Add the export line at the end:

```typescript
export { ShortlistPanel } from './ShortlistPanel'
```

- [ ] **Step 6: Commit**

```bash
cd /home/jibei/projects/scouter
git add frontend/src/components/mission/ShortlistPanel.tsx \
        frontend/src/components/mission/ShortlistPanel.module.css \
        frontend/src/components/mission/ShortlistPanel.test.tsx \
        frontend/src/components/mission/index.ts
git commit -m "feat(ui): add ShortlistPanel showing pinned options with Select action"
```

---

### Task 12: Wire ShortlistPanel into MissionOverview (buying phase)

**Files:**
- Modify: `frontend/src/pages/MissionOverview.tsx`

The `PurchaseForm` component (already in MissionOverview at lines 467–472) accepts a `prefill` prop of type `PurchaseFormPrefill`:
```typescript
interface PurchaseFormPrefill {
  merchant?: string
  finalPrice?: number
  purchasedAt?: string
}
```

When the user clicks "Select" on a shortlist card, we set a `prefill` state that feeds into `PurchaseForm`. If a purchase record already exists, we show a confirmation before overwriting the prefill.

- [ ] **Step 1: Write failing test**

Create `frontend/src/pages/MissionOverview.shortlist.test.tsx`:

```typescript
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import React from 'react'

// Mock entire '../hooks' barrel — MissionOverview.tsx line 10 imports all hooks from here
vi.mock('../hooks', () => ({
  useMission: () => ({
    data: {
      id: 'mission-1',
      slug: 'test',
      name: 'Test',
      phase: 'buying',
      budget: 2500,
      currency: 'EUR',
      constraints: [],
      category: 'computing',
    },
    isLoading: false,
  }),
  useUpdateMission: () => ({ updateMission: vi.fn(), isPending: false }),
  useDeleteMission: () => ({ deleteMission: vi.fn(), isPending: false }),
  useArchiveMission: () => ({ archiveMission: vi.fn(), isPending: false }),
  useUnarchiveMission: () => ({ unarchiveMission: vi.fn(), isPending: false }),
  useOptions: () => ({
    data: [{ id: 'opt-1', name: 'MacBook Pro', badge: 'recommended', pinned: true, priceRange: { min: 2000, max: 2500, best: 2200 }, attributes: [] }],
  }),
  useUnpinAllOptions: () => ({ unpinAllOptions: vi.fn(), isPending: false }),
  useShopping: () => ({ data: [] }),
  usePurchaseRecord: () => ({ data: null }),
  useResearch: () => ({ data: null, isLoading: false }),
  usePriceIntel: () => ({ data: null }),
  useKeyboardShortcuts: () => undefined,
  useSuggestCategory: () => ({ data: null }),
}))

// Individual mocks for hooks imported from their own files (lines 6–7 of MissionOverview.tsx)
vi.mock('../hooks/useBudgetAlerts', () => ({
  useBudgetAlerts: () => ({ alerts: [] }),
}))

vi.mock('../hooks/useScorecard', () => ({
  useScorecard: () => ({ data: null, isLoading: false }),
}))

// Mock all exports from '../components/mission' barrel (MissionOverview.tsx line 8 + line 11)
// Includes all pre-existing exports PLUS new components added by this plan
vi.mock('../components/mission', () => ({
  CategoryTemplate: () => null,
  DecisionPanel: () => null,
  MissionTimeline: () => null,
  PurchaseForm: () => <div data-testid="purchase-form" />,
  LessonsField: () => null,
  CollaboratorsPanel: () => null,
  TravelSearchWidget: () => null,
  TimingAdvisorCard: () => null,
  ExportPanel: () => null,
  ReceiptScanner: () => null,
  SummaryReport: () => null,
  CoachPanel: () => null,
  HealthScoreCard: () => null,
  MissionSummaryCard: () => null,
  CommentThread: () => null,
  CategoryBadge: () => null,
  MissionGoalTracker: () => null,
  BudgetRecommendations: () => null,
  SalesCalendar: () => null,
  EcoScorePanel: () => null,
  MissionProgressWidget: () => null,
  GiftFinderWidget: () => null,
  LoyaltySummaryPanel: () => null,
  MissionROICard: () => null,
  InflationTrackerPanel: () => null,
  DecisionMatrixTable: () => null,
  SmartAlertsPanel: () => null,
  VoteSummaryPanel: () => null,
  MissionReportButton: () => null,
  ReorderSuggestionsPanel: () => null,
  NegotiationOutcomePanel: () => null,
  BundleDealsPanel: () => null,
  BurnRateCard: () => null,
  RegretAnalyzerCard: () => null,
  ListOptimizerPanel: () => null,
  CashbackSummaryPanel: () => null,
  PriceDropWatchlist: () => null,
  SeasonalCalendarPanel: () => null,
  BudgetAdvisorPanel: () => null,
  ComparisonScorePanel: () => null,
  PriceAlertDigestPanel: () => null,
  SpendingVelocityCard: () => null,
  ExpenseCategoryPanel: () => null,
  // New components added by this plan:
  MissionActionBar: () => null,
  MissionEditModal: () => null,
  ShortlistPanel: ({ options, onSelect }: { options: { id: string; name: string; pinned: boolean }[]; onSelect: (o: unknown) => void }) => (
    <div>
      <span>Your Shortlist</span>
      {options.filter(o => o.pinned).map(o => (
        <div key={o.id}>
          <span>{o.name}</span>
          <button onClick={() => onSelect(o)}>Select</button>
        </div>
      ))}
    </div>
  ),
}))

vi.mock('../components/forecast', () => ({
  ForecastPanel: () => null,
}))

vi.mock('../components/scouter', () => ({
  LoadingPulse: () => null,
  BudgetBar: () => null,
  StatusBadge: () => null,
  ToastContainer: () => null,
  useToasts: () => ({ toasts: [], toast: vi.fn() }),
  NextActionNudge: () => null,
}))

vi.mock('../components/scouter/LastVisitCard', () => ({
  setLastVisitedMission: vi.fn(),
}))

vi.mock('./mission-overview/PhaseSection', () => ({
  PhaseSection: () => null,
}))

vi.mock('./mission-overview/QuickNavSection', () => ({
  QuickNavSection: () => null,
}))

import MissionOverview from './MissionOverview'

function renderOverview() {
  const queryClient = new QueryClient()
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={['/missions/test']}>
        <Routes>
          <Route path="/missions/:slug" element={<MissionOverview />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('MissionOverview shortlist (buying phase)', () => {
  it('renders ShortlistPanel when phase is buying', () => {
    renderOverview()
    expect(screen.getByText(/your shortlist/i)).toBeInTheDocument()
  })

  it('shows pinned option in shortlist', () => {
    renderOverview()
    expect(screen.getByText('MacBook Pro')).toBeInTheDocument()
  })
})
```

Note: This test requires careful mocking of the many hooks used in MissionOverview. If the mock setup is complex, write the minimal version that tests the ShortlistPanel is rendered. The main logic (pre-filling) is tested in ShortlistPanel's own tests.

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/pages/MissionOverview.shortlist.test.tsx
```

Expected: FAIL

- [ ] **Step 3: Wire ShortlistPanel into MissionOverview**

Open `frontend/src/pages/MissionOverview.tsx`.

Add imports:
```typescript
import { ShortlistPanel } from '../components/mission'
```

Note: `import type { PurchaseFormPrefill }` is already on line 13 of `MissionOverview.tsx` — do NOT add it again.

Add `useOptions` to the EXISTING import from `'../hooks'` on line 10 (do NOT add a new import line). The updated line should include:
```typescript
import { useMission, useShopping, useResearch, usePriceIntel, useUpdateMission, useKeyboardShortcuts, usePurchaseRecord, useSuggestCategory, useDeleteMission, useArchiveMission, useOptions } from '../hooks'
```

Add hooks and state (near the top of the component, with other hooks):
```typescript
const { data: options = [] } = useOptions(mission?.id ?? '')
const [shortlistPrefill, setShortlistPrefill] = useState<PurchaseFormPrefill | undefined>(undefined)
const [shortlistKey, setShortlistKey] = useState(0)
const [replacePrefillOption, setReplacePrefillOption] = useState<Option | null>(null)
```

Add the "Select" handler. Note: `MissionOverview` already uses `scanPrefill` and `scanKey` for the receipt scanner prefill — the new shortlist prefill must coexist with these. When the user clicks "Select", increment `shortlistKey` to re-mount PurchaseForm with the new prefill.

**Scope note:** The spec says "Option name → purchase item name field" but the existing `PurchaseFormPrefill` type only has `merchant`, `finalPrice`, and `purchasedAt` — there is no `itemName` field. We pre-fill only `finalPrice` (the price midpoint). Extending `PurchaseFormPrefill` with an item name field is deferred to a follow-up task.

```typescript
function handleSelectShortlistOption(option: Option) {
  if (purchaseRecord) {
    // Purchase already exists — confirm before overwriting prefill
    setReplacePrefillOption(option)
  } else {
    setShortlistPrefill({
      finalPrice: option.priceRange
        ? Math.round((option.priceRange.min + option.priceRange.max) / 2)
        : undefined,
    })
    setShortlistKey(k => k + 1)
  }
}

function handleConfirmReplace() {
  if (!replacePrefillOption) return
  setShortlistPrefill({
    finalPrice: replacePrefillOption.priceRange
      ? Math.round((replacePrefillOption.priceRange.min + replacePrefillOption.priceRange.max) / 2)
      : undefined,
  })
  setShortlistKey(k => k + 1)
  setReplacePrefillOption(null)
}
```

Find where the purchase form is rendered (around line 454):
```typescript
{(mission.phase === 'buying' || mission.phase === 'done') && (
```

Above the purchase form (but inside this conditional), add ShortlistPanel for the buying phase only:
```typescript
{mission.phase === 'buying' && (
  <ShortlistPanel
    options={options}
    missionSlug={mission.slug}
    onSelect={handleSelectShortlistOption}
  />
)}
```

Pass the merged prefill to `PurchaseForm`. Shortlist prefill takes priority over scan prefill (user explicitly clicked Select). Keep the existing `key={scanKey}` behavior by using `shortlistKey + scanKey` (addition ensures any increment from either source triggers a re-mount):
```typescript
<PurchaseForm
  key={shortlistKey + scanKey}
  missionId={mission.id}
  existingRecord={purchaseRecord ?? undefined}
  prefill={shortlistPrefill ?? scanPrefill}
/>
```

Add the replace-confirm dialog:
```typescript
{replacePrefillOption && (
  <div className={styles.overlay} onClick={() => setReplacePrefillOption(null)}>
    <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
      <p>{t('shortlist.replaceConfirm')}</p>
      <div className={styles.dialogActions}>
        <button onClick={() => setReplacePrefillOption(null)}>{t('common.cancel')}</button>
        <button className={styles.confirmBtn} onClick={handleConfirmReplace}>{t('common.confirm')}</button>
      </div>
    </div>
  </div>
)}
```

Add the following CSS classes to `MissionOverview.module.css` (these mirror the overlay pattern used in MissionActionBar):

```css
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
}

.dialog {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 24px;
  max-width: 400px;
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.dialogActions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  margin-top: 8px;
}

.confirmBtn {
  background: var(--accent);
  color: var(--bg);
  border: none;
  border-radius: 6px;
  padding: 8px 16px;
  font-weight: 600;
  cursor: pointer;
}
```

- [ ] **Step 4: Run tests and typecheck**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run src/pages/MissionOverview.shortlist.test.tsx
npm run typecheck
```

Expected: PASS, no type errors

- [ ] **Step 5: Run full test suite**

```bash
cd /home/jibei/projects/scouter/frontend
npm run test -- --run
npm run build
```

Expected: all tests pass, build succeeds

- [ ] **Step 6: Final commit**

```bash
cd /home/jibei/projects/scouter
git add frontend/src/pages/MissionOverview.tsx \
        frontend/src/pages/MissionOverview.module.css \
        frontend/src/pages/MissionOverview.shortlist.test.tsx
git commit -m "feat(mission): wire ShortlistPanel into MissionOverview buying phase with prefill"
```

---

## Summary

| Task | Scope | New files |
|------|-------|-----------|
| 1 | i18n keys (en + fr) | — |
| 2 | Archive/unarchive API functions | `missions.test.ts` |
| 3 | Archive/unarchive hooks | `useMission.test.ts` |
| 4 | MissionActionBar component | `MissionActionBar.tsx`, `.css`, `.test.tsx` |
| 5 | MissionEditModal component | `MissionEditModal.tsx`, `.css`, `.test.tsx` |
| 6 | Wire actions into MissionOverview | — |
| 7 | MissionCard ⋯ menu + HQ archived filter | `MissionCard.test.tsx` |
| 8 | OptionEditModal component | `OptionEditModal.tsx`, `.css`, `.test.tsx` |
| 9 | OptionCard ⋯ menu + OptionsExplorer wiring | `OptionCard.menu.test.tsx` |
| 10 | Unpin All (replace Clear Pinned) | `useUnpinAll.test.tsx` |
| 11 | ShortlistPanel component | `ShortlistPanel.tsx`, `.css`, `.test.tsx` |
| 12 | Wire ShortlistPanel into MissionOverview | `MissionOverview.shortlist.test.tsx` |
