import { render, screen, act } from '@testing-library/react';
import { Countdown } from './Countdown';
import { useCountdown } from '../../hooks/Training/useCountdown';

function RunningCountdown() {
  return <Countdown seconds={useCountdown('2026-01-01T00:00:00.000Z', 75, 0)} />;
}

describe('Countdown', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date('2026-01-01T00:00:00.000Z'));
  });
  afterEach(() => jest.useRealTimers());

  it('counts down from the server-derived deadline', () => {
    render(<RunningCountdown />);
    expect(screen.getByText('01:15')).toBeInTheDocument();
    act(() => { jest.advanceTimersByTime(15_000); });
    expect(screen.getByText('01:00')).toBeInTheDocument();
  });
});
