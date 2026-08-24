import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { Home } from './Home';

function json(status: number, body?: unknown): Promise<Response> {
  return Promise.resolve({ ok: status >= 200 && status < 300, status, json: async () => body } as Response);
}

describe('Home', () => {
  beforeEach(() => { global.fetch = jest.fn(() => json(404)); });

  it('does not expose any selected problem before start', async () => {
    render(<MemoryRouter><Home /></MemoryRouter>);
    await waitFor(() => expect(screen.getByRole('button', { name: 'START 75 MINUTES' })).toBeEnabled());
    expect(screen.queryByText(/ABC999/)).not.toBeInTheDocument();
    expect(screen.getByText('問題は開始するまで表示されません。')).toBeInTheDocument();
  });
});
