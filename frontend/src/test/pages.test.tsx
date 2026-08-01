import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const { mockApi, mockPush } = vi.hoisted(() => {
  const mockApi = {
    login: vi.fn(),
    createRun: vi.fn(),
    listRecordingSessions: vi.fn(),
    createRecordingSession: vi.fn(),
    deleteRecordingSession: vi.fn(),
    getAllReviews: vi.fn(),
    approveReview: vi.fn(),
    rejectReview: vi.fn(),
    getChangeProposals: vi.fn(),
    approveChangeProposal: vi.fn(),
    rejectChangeProposal: vi.fn(),
    createTestList: vi.fn(),
    createSchedule: vi.fn(),
    getTestCases: vi.fn(),
    getTestLists: vi.fn(),
    getSchedules: vi.fn(),
    runTestList: vi.fn(),
    runScheduleNow: vi.fn(),
    isReviewerOrAbove: vi.fn(),
    isAdmin: vi.fn(),
    getUserRole: vi.fn(),
  }
  return { mockApi, mockPush: vi.fn() }
})

vi.mock('@/lib/api', () => mockApi)
vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush }),
}))
vi.mock('next/link', () => ({
  default: ({ href, children, ...props }: { href: string; children: React.ReactNode }) => (
    <a href={href} {...props}>{children}</a>
  ),
}))

import CreateTestPage from '@/app/create/page'
import LoginPage from '@/app/login/page'
import RecordingsPage from '@/app/recordings/page'
import ReviewsPage from '@/app/reviews/page'
import SuitesPage from '@/app/suites/page'

const now = '2026-08-01T12:00:00.000Z'

const testCase = (id: string, title: string) => ({
  id,
  project_id: 'project-1',
  title,
  type: 'ui',
  feature: 'checkout',
  priority: 'high',
  steps: ['Open the page'],
  assertions: ['Page is visible'],
  tags: ['smoke'],
  version: 1,
  created_at: now,
  updated_at: now,
})

beforeEach(() => {
  for (const fn of Object.values(mockApi)) {
    fn.mockReset()
  }
  mockPush.mockReset()

  mockApi.login.mockResolvedValue({ status: 'ok', redirect: '/' })
  mockApi.createRun.mockResolvedValue({ run_id: 'run-123', state: 'idle' })
  mockApi.listRecordingSessions.mockResolvedValue([])
  mockApi.createRecordingSession.mockResolvedValue({
    id: 'recording-new',
    name: 'Login Flow Recording',
    project_path: '/workspace/app',
    base_url: 'https://app.example.com',
    status: 'recording',
    event_count: 0,
    created_at: now,
    updated_at: now,
  })
  mockApi.deleteRecordingSession.mockResolvedValue(undefined)
  mockApi.getAllReviews.mockResolvedValue([])
  mockApi.getChangeProposals.mockResolvedValue([])
  mockApi.approveReview.mockResolvedValue({ id: 'review-1', status: 'approved' })
  mockApi.rejectReview.mockResolvedValue({ id: 'review-1', status: 'rejected' })
  mockApi.approveChangeProposal.mockResolvedValue({ proposal: { id: 'proposal-1', status: 'approved' }, test_case: testCase('case-1', 'Checkout') })
  mockApi.rejectChangeProposal.mockResolvedValue({ id: 'proposal-1', status: 'rejected' })
  mockApi.getTestCases.mockResolvedValue([])
  mockApi.getTestLists.mockResolvedValue([])
  mockApi.getSchedules.mockResolvedValue([])
  mockApi.createTestList.mockResolvedValue({
    id: 'list-new',
    name: 'Smoke Regression',
    project_id: 'project-1',
    tags: ['manual', 'regression'],
    test_case_ids: ['case-1'],
    pinned: false,
    created_at: now,
    updated_at: now,
  })
  mockApi.createSchedule.mockResolvedValue({
    id: 'schedule-new',
    name: 'Nightly Regression',
    project_id: 'project-1',
    test_list_id: 'list-1',
    project_path: '',
    requirements: '',
    mode: 'approved_list',
    environment: '',
    base_url: '',
    frequency: 'daily',
    enabled: true,
    next_run_at: now,
    last_run_at: '',
    created_at: now,
    updated_at: now,
  })
  mockApi.runTestList.mockResolvedValue({ test_list_id: 'list-1', run_ids: ['run-1'] })
  mockApi.runScheduleNow.mockResolvedValue({ run_id: 'run-1', run_ids: ['run-1'] })
  mockApi.isReviewerOrAbove.mockReturnValue(true)
  mockApi.isAdmin.mockReturnValue(true)
  mockApi.getUserRole.mockReturnValue('admin')
})

describe('LoginPage', () => {
  it('submits a trimmed API key and shows invalid-key feedback', async () => {
    mockApi.login.mockResolvedValueOnce({ status: 'error' })

    render(<LoginPage />)

    fireEvent.change(screen.getByLabelText(/api key/i), { target: { value: '  bad-key  ' } })
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => expect(mockApi.login).toHaveBeenCalledWith('bad-key'))
    expect(await screen.findByText(/invalid api key/i)).toBeInTheDocument()
  })
})

describe('CreateTestPage', () => {
  it('walks the wizard and creates a run from the configured target', async () => {
    render(<CreateTestPage />)

    expect(screen.getByRole('heading', { name: /create tests/i })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /continue/i }))
    expect(screen.getByText('Configuration')).toBeInTheDocument()

    fireEvent.change(screen.getByPlaceholderText('https://app.example.com'), {
      target: { value: 'https://app.example.com' },
    })
    fireEvent.click(screen.getByRole('button', { name: /continue/i }))
    expect(screen.getByText('Exploration Scope')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /continue/i }))
    expect(screen.getByText('Plan Review')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /generate and run/i }))

    await waitFor(() => {
      expect(mockApi.createRun).toHaveBeenCalledWith(expect.objectContaining({
        project_path: 'https://app.example.com',
        test_type: 'ui',
        mode: 'simple',
      }))
    })
    expect(mockPush).toHaveBeenCalledWith('/runs/run-123')
  })
})

describe('RecordingsPage', () => {
  it('renders recording sessions and filters by search query', async () => {
    mockApi.listRecordingSessions.mockResolvedValueOnce([
      {
        id: 'rec-1',
        name: 'Checkout Flow',
        project_path: '/workspace/checkout',
        base_url: 'https://shop.example.com',
        status: 'completed',
        event_count: 12,
        created_at: now,
        updated_at: now,
      },
      {
        id: 'rec-2',
        name: 'Login Session',
        project_path: '/workspace/auth',
        base_url: 'https://auth.example.com',
        status: 'recording',
        event_count: 3,
        created_at: now,
        updated_at: now,
      },
    ])

    render(<RecordingsPage />)

    expect(await screen.findByText('Checkout Flow')).toBeInTheDocument()
    expect(screen.getByText('Login Session')).toBeInTheDocument()

    fireEvent.change(screen.getByPlaceholderText(/search sessions/i), { target: { value: 'login' } })

    expect(screen.queryByText('Checkout Flow')).not.toBeInTheDocument()
    expect(screen.getByText('Login Session')).toBeInTheDocument()
  })

  it('creates a new recording session from the modal form', async () => {
    render(<RecordingsPage />)

    await screen.findByText('No recording sessions')
    fireEvent.click(screen.getByRole('button', { name: /new recording/i }))

    fireEvent.change(screen.getByPlaceholderText(/login flow recording/i), {
      target: { value: ' Login Flow Recording ' },
    })
    fireEvent.change(screen.getByPlaceholderText(/\/path\/to\/project/i), {
      target: { value: ' /workspace/app ' },
    })
    fireEvent.change(screen.getByPlaceholderText('https://example.com'), {
      target: { value: ' https://app.example.com ' },
    })
    fireEvent.click(screen.getByRole('button', { name: /create session/i }))

    await waitFor(() => {
      expect(mockApi.createRecordingSession).toHaveBeenCalledWith({
        name: 'Login Flow Recording',
        project_path: '/workspace/app',
        base_url: 'https://app.example.com',
      })
    })
  })
})

describe('ReviewsPage', () => {
  it('renders pending reviews and proposals and sends review decisions', async () => {
    mockApi.getAllReviews.mockResolvedValue([
      {
        id: 'review-1',
        run_id: 'run-aaaaaaaa',
        type: 'test_plan',
        status: 'pending',
        created_at: now,
      },
    ])
    mockApi.getChangeProposals.mockResolvedValue([
      {
        id: 'proposal-1',
        test_case_id: 'case-bbbbbbbb',
        status: 'pending',
        prompt: 'Improve selectors',
        rationale: 'Retry flaky locator with accessible role',
        original: testCase('case-bbbbbbbb', 'Old checkout test'),
        proposed: testCase('case-bbbbbbbb', 'Improved checkout test'),
        created_at: now,
        updated_at: now,
      },
    ])

    render(<ReviewsPage />)

    expect(await screen.findByText('test plan')).toBeInTheDocument()
    expect(screen.getByText(/retry flaky locator/i)).toBeInTheDocument()

    fireEvent.click(screen.getAllByTitle('Approve')[0])
    await waitFor(() => expect(mockApi.approveReview).toHaveBeenCalledWith('review-1', 'admin', 'Approved'))

    fireEvent.click(screen.getAllByTitle('Reject')[1])
    await waitFor(() => {
      expect(mockApi.rejectChangeProposal).toHaveBeenCalledWith('proposal-1', {
        reviewer: 'admin',
        comment: 'Rejected from Reviews',
      })
    })
  })
})

describe('SuitesPage', () => {
  it('renders lists, creates a selected test list, and schedules recurring runs', async () => {
    const cases = [testCase('case-1', 'Checkout happy path'), testCase('case-2', 'Login happy path')]
    mockApi.getTestCases.mockResolvedValue(cases)
    mockApi.getTestLists.mockResolvedValue([
      {
        id: 'list-1',
        name: 'Existing Smoke',
        project_id: 'project-1',
        tags: ['smoke'],
        test_case_ids: ['case-1'],
        pinned: true,
        created_at: now,
        updated_at: now,
      },
    ])
    mockApi.getSchedules.mockResolvedValue([])

    render(<SuitesPage />)

    expect(await screen.findAllByText('Existing Smoke')).toHaveLength(2)
    expect(screen.getAllByText('Checkout happy path').length).toBeGreaterThan(0)

    fireEvent.click(screen.getByRole('button', { name: /create test list/i }))
    await waitFor(() => {
      expect(mockApi.createTestList).toHaveBeenCalledWith(expect.objectContaining({
        name: 'Smoke Regression',
        project_id: 'project-1',
        test_case_ids: ['case-1', 'case-2'],
        pinned: false,
      }))
    })

    const recurringRuns = screen.getByText('Recurring Runs').closest('section') ?? document.body
    fireEvent.click(within(recurringRuns as HTMLElement).getByRole('button', { name: /schedule/i }))

    await waitFor(() => {
      expect(mockApi.createSchedule).toHaveBeenCalledWith(expect.objectContaining({
        name: 'Nightly Regression',
        project_id: 'project-1',
        test_list_id: 'list-1',
        frequency: 'daily',
        mode: 'approved_list',
        enabled: true,
      }))
    })
  })
})
