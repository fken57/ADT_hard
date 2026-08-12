import { TrainingProblem } from '../../types/Training';

const slotOrder = ['D1', 'E1', 'E2', 'E3', 'F1'];

export function orderedProblems(problems: TrainingProblem[]): TrainingProblem[] {
  return [...problems].sort((left, right) => {
    const leftIndex = slotOrder.indexOf(left.slot);
    const rightIndex = slotOrder.indexOf(right.slot);
    return (leftIndex === -1 ? 99 : leftIndex) - (rightIndex === -1 ? 99 : rightIndex);
  });
}

export function ProblemList({ problems, readonly = false }: { problems: TrainingProblem[]; readonly?: boolean }) {
  return (
    <ol className="problem-list">
      {orderedProblems(problems).map((problem) => (
        <li key={problem.id} className={problem.acceptedAt ? 'accepted' : ''}>
          <span className="slot">{problem.slot}</span>
          <span className="problem-title">{problem.contestId} {problem.problemIndex}: {problem.title}</span>
          {problem.acceptedAt && <span className="ac">AC</span>}
          {!readonly && (
            <a href={problem.url} target="_blank" rel="noreferrer" className="open-link">
              OPEN <span aria-hidden="true">↗</span>
            </a>
          )}
        </li>
      ))}
    </ol>
  );
}
