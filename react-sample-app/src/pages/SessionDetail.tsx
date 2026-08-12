import { Link, useParams } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { ProblemList } from '../components/Training/ProblemList';
import { SessionResponse } from '../types/Training';
import { ApiError, trainingApi } from '../util/TrainingApi';

export function SessionDetail({ result = false }: { result?: boolean }) {
  const { id = '' } = useParams();
  const [data, setData] = useState<SessionResponse | null>(null);
  const [notFound, setNotFound] = useState(false);
  useEffect(() => {
    const request = result ? trainingApi.sync(id) : trainingApi.get(id);
    request.then(setData).catch((error: unknown) => {
      if (error instanceof ApiError && error.status === 404) setNotFound(true);
    });
  }, [id, result]);
  if (notFound) return <section><h1>セッションが見つかりません</h1><Link to="/history">履歴へ戻る</Link></section>;
  if (!data) return <p className="loading">結果を読み込んでいます…</p>;
  const { session } = data;
  const accepted = session.problems.filter(problem => problem.acceptedAt).length;
  const categories = ['D', 'E', 'F'].map(index => {
    const problems = session.problems.filter(problem => problem.problemIndex === index);
    return { index, accepted: problems.filter(problem => problem.acceptedAt).length, total: problems.length };
  });
  const solveTime = (acceptedAt?: string): string => {
    if (!acceptedAt) return '-';
    const seconds = Math.max(0, Math.floor((new Date(acceptedAt).getTime() - new Date(session.startedAt).getTime()) / 1000));
    return `${Math.floor(seconds / 60).toString().padStart(2, '0')}:${(seconds % 60).toString().padStart(2, '0')}`;
  };
  return (
    <section className="result-page">
      <p className="eyebrow">{result ? 'SESSION RESULT' : 'TRAINING DETAIL'}</p>
      <h1>{session.status === 'ABORTED' ? '中断したセッション' : '75分の結果'}</h1>
      <div className="result-score"><strong>{accepted}</strong><span> / {session.problems.length} AC</span></div>
      <div className="category-scores" aria-label="カテゴリ別結果">
        {categories.map(category => <span key={category.index}><strong>{category.index}</strong> {category.accepted} / {category.total}</span>)}
      </div>
      <p className="muted">{new Date(session.startedAt).toLocaleString('ja-JP')} 開始 · 難易度緩和レベル {session.fallbackLevel}</p>
      <ProblemList problems={session.problems} readonly />
      <div className="solve-times" aria-label="解答時間">
        {session.problems.map(problem => <span key={problem.id}>{problem.slot}: {solveTime(problem.acceptedAt)}</span>)}
      </div>
      <Link className="button primary" to="/">次のトレーニングへ</Link>
      <Link className="history-link" to="/history">履歴を見る →</Link>
    </section>
  );
}
