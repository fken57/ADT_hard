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
    trainingApi.sync(id).then(setData).catch((error: unknown) => {
      if (error instanceof ApiError && error.status === 404) setNotFound(true);
    });
  }, [id]);
  if (notFound) return <section><h1>セッションが見つかりません</h1><Link to="/history">履歴へ戻る</Link></section>;
  if (!data) return <p className="loading">結果を読み込んでいます…</p>;
  const { session } = data;
  const accepted = session.problems.filter(problem => problem.acceptedAt).length;
  const penalties = session.problems.reduce((total, problem) => total + (problem.acceptedAt ? problem.penaltyCount : 0), 0);
  const profileLabel = { STANDARD: '標準', LIGHT: '軽め', HEAVY: '重め', LEGACY: '旧構成' }[session.difficultyProfile] || session.difficultyProfile;
  const categories = ['D', 'E', 'F'].map(index => {
    const problems = session.problems.filter(problem => problem.problemIndex === index);
    return { index, accepted: problems.filter(problem => problem.acceptedAt).length, total: problems.length };
  });
  return (
    <section className="result-page">
      <p className="eyebrow">{result ? 'SESSION RESULT' : 'TRAINING DETAIL'}</p>
      <h1>{session.status === 'ABORTED' ? '中断したセッション' : '75分の結果'}</h1>
      <div className="result-score"><strong>{accepted}</strong><span> / {session.problems.length} AC</span><em>{penalties} ペナ</em></div>
      <div className="category-scores" aria-label="カテゴリ別結果">
        {categories.map(category => <span key={category.index}><strong>{category.index}</strong> {category.accepted} / {category.total}</span>)}
      </div>
      <p className="muted">{new Date(session.startedAt).toLocaleString('ja-JP')} 開始 · {profileLabel} · 難易度緩和レベル {session.fallbackLevel}</p>
      <ProblemList problems={session.problems} startedAt={session.startedAt} />
      <Link className="button primary" to="/">次のトレーニングへ</Link>
      <Link className="history-link" to="/history">履歴を見る →</Link>
    </section>
  );
}
