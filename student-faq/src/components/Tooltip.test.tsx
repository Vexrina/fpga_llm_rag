import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import Tooltip from '../components/Tooltip'

describe('Tooltip', () => {
  it('renders children', () => {
    render(
      <Tooltip content="test tooltip">
        <span>child</span>
      </Tooltip>,
    )

    expect(screen.getByText('child')).toBeInTheDocument()
  })

  it('shows tooltip on hover', async () => {
    const user = userEvent.setup()
    render(
      <Tooltip content="test tooltip content">
        <span data-testid="trigger">?</span>
      </Tooltip>,
    )

    const trigger = screen.getByTestId('trigger')
    await user.hover(trigger)

    expect(screen.getByText('test tooltip content')).toBeInTheDocument()
  })

  it('hides tooltip on mouse leave', async () => {
    const user = userEvent.setup()
    render(
      <Tooltip content="test tooltip content">
        <span data-testid="trigger">?</span>
      </Tooltip>,
    )

    const trigger = screen.getByTestId('trigger')
    await user.hover(trigger)
    expect(screen.getByText('test tooltip content')).toBeInTheDocument()

    await user.unhover(trigger)
    expect(screen.queryByText('test tooltip content')).not.toBeInTheDocument()
  })
})
