import { TrainingProblem } from '../../types/Training';

const slotOrder = ['Warmup', 'Stable', 'Main', 'Stretch', 'Challenge', 'D1', 'E1', 'E2', 'E3', 'F1'];

export function orderedProblems(problems: TrainingProblem[]): TrainingProblem[] {
  return [...problems].sort((left, right) => {
    const leftIndex = slotOrder.indexOf(left.slot);
    const rightIndex = slotOrder.indexOf(right.slot);
    return (leftIndex === -1 ? 99 : leftIndex) - (rightIndex === -1 ? 99 : rightIndex);
  });
}

export function elapsedTime(startedAt: string, acceptedAt?: string): string {
  if (!acceptedAt) return '—';
  const seconds = Math.max(0, Math.floor((new Date(acceptedAt).getTime() - new Date(startedAt).getTime()) / 1000));
  return `${Math.floor(seconds / 60)}:${(seconds % 60).toString().padStart(2, '0')}`;
}

export function ProblemList({ problems, startedAt }: { problems: TrainingProblem[]; startedAt: string }) {
  return (
    <ol className="problem-list">
      {orderedProblems(problems).map((problem) => (
        <li key={problem.id} className={problem.acceptedAt ? 'accepted' : ''}>
          <span className="slot">{problem.slot}<small>diff {problem.difficulty ?? '—'}</small></span>
          <a href={problem.url} target="_blank" rel="noreferrer" className="problem-title">
            {problem.contestId} {problem.problemIndex}: {problem.title}
          </a>
          <span className={`score-cell ${problem.acceptedAt ? 'solved' : ''}`}>
            <strong>{elapsedTime(startedAt, problem.acceptedAt)}</strong>
            {problem.acceptedAt && <small>{problem.penaltyCount}ペナ</small>}
          </span>
          <a href={problem.url} target="_blank" rel="noreferrer" className="open-link">
            問題を開く <span aria-hidden="true">↗</span>
          </a>
        </li>
      ))}
    </ol>
  );
}
