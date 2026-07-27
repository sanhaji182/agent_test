import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { StatusBadge, PriorityBadge } from '@/components/ui/badge'

describe('StatusBadge', () => {
  it('renders done state', () => {
    render(<StatusBadge state="done" />)
    expect(screen.getByText('done')).toBeInTheDocument()
  })

  it('renders failed state', () => {
    render(<StatusBadge state="failed" />)
    expect(screen.getByText('failed')).toBeInTheDocument()
  })

  it('renders running state with underscore replaced', () => {
    render(<StatusBadge state="plan_generated" />)
    expect(screen.getByText('plan generated')).toBeInTheDocument()
  })

  it('renders idle state as default for unknown states', () => {
    render(<StatusBadge state="simulated" />)
    expect(screen.getByText('simulated')).toBeInTheDocument()
  })

  it('has capitalized text', () => {
    render(<StatusBadge state="done" />)
    expect(screen.getByText('done')).toHaveClass('capitalize')
  })
})

describe('PriorityBadge', () => {
  it('renders high priority', () => {
    render(<PriorityBadge priority="high" />)
    expect(screen.getByText('high')).toBeInTheDocument()
  })

  it('renders medium priority', () => {
    render(<PriorityBadge priority="medium" />)
    expect(screen.getByText('medium')).toBeInTheDocument()
  })
})
