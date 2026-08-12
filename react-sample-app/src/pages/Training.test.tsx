import { act, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { Training } from './Training';

const session = {
  id: 'session-1', atcoderUserId: 'fken_prime_57', startedAt: '2026-01-01T00:00:00.000Z', durationSeconds: 4500,
  status: 'ACTIVE' as const, fallbackLevel: 0, createdAt: '2026-01-01T00:00:00.000Z', updatedAt: '2026-01-01T00:00:00.000Z',
  problems: [{ id: 'p1', slot: 'D1', contestId: 'ABC999', problemId: 'abc999_d', problemIndex: 'D', title: 'Visible after start', penaltyCount: 0, url: 'https://atcoder.jp/contests/abc999/tasks/abc999_d' }],
};

function json(status: number, body?: unknown): Promise<Response> {
  return Promise.resolve({ ok: status >= 200 && status < 300, status, json: async () => body } as Response);
}

describe('Training', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date('2026-01-01T00:00:05.000Z'));
    global.fetch = jest.fn((input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith('/sync')) return json(503, { message: 'temporary outage' });
      return json(200, { session, serverNow: '2026-01-01T00:00:05.000Z' });
    });
  });
  afterEach(() => jest.useRealTimers());

  it('keeps the problem and countdown visible when submission sync fails', async () => {
    render(<MemoryRouter initialEntries={['/training/session-1']}><Routes><Route path="/training/:id" element={<Training />} /></Routes></MemoryRouter>);
    await waitFor(() => expect(screen.getByText(/Visible after start/)).toBeInTheDocument());
    act(() => { jest.advanceTimersByTime(15_000); });
    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('タイマーは継続しています'));
    expect(screen.getByText(/Visible after start/)).toBeInTheDocument();
    expect(screen.getByLabelText(/残り/)).toBeInTheDocument();
  });

  it('performs a final sync when the countdown reaches zero', async () => {
    jest.setSystemTime(new Date('2026-01-01T01:14:59.000Z'));
    const fetchMock = global.fetch as jest.Mock;
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith('/sync')) return json(503, { message: 'temporary outage' });
      return json(200, { session, serverNow: '2026-01-01T01:14:59.000Z' });
    });
    render(<MemoryRouter initialEntries={['/training/session-1']}><Routes>
      <Route path="/training/:id" element={<Training />} />
      <Route path="/results/:id" element={<p>result</p>} />
    </Routes></MemoryRouter>);
    await waitFor(() => expect(screen.getByText(/Visible after start/)).toBeInTheDocument());
    act(() => { jest.advanceTimersByTime(1_100); });
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      expect.stringMatching(/\/training\/sessions\/session-1\/sync$/),
      expect.objectContaining({ method: 'POST' }),
    ));
    expect(await screen.findByText('result')).toBeInTheDocument();
  });
});
