import { render, screen } from '@testing-library/react';
import { ProblemList } from './ProblemList';

test('links to AtCoder and shows elapsed time with penalties in every context', () => {
  render(<ProblemList startedAt="2026-01-01T00:00:00Z" problems={[{
    id: 'p1', slot: 'D1', contestId: 'abc999', problemId: 'abc999_d', problemIndex: 'D', title: 'Example',
    acceptedAt: '2026-01-01T00:42:09Z', penaltyCount: 2,
    url: 'https://atcoder.jp/contests/abc999/tasks/abc999_d',
  }]} />);

  expect(screen.getByRole('link', { name: /abc999 D: Example/ })).toHaveAttribute('href', 'https://atcoder.jp/contests/abc999/tasks/abc999_d');
  expect(screen.getByText('42:09')).toBeInTheDocument();
  expect(screen.getByText('2ペナ')).toBeInTheDocument();
});
