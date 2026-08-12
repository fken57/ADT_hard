import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { SessionHistoryResponse } from '../types/Training';
import { trainingApi } from '../util/TrainingApi';

export function History() {
  const [data, setData] = useState<SessionHistoryResponse | null>(null);
  const [page, setPage] = useState(1);
  const [error, setError] = useState(false);
  useEffect(() => {
    setError(false);
    trainingApi.history(page).then(setData).catch(() => setError(true));
  }, [page]);
  if (error) return <section><h1>履歴を読み込めませんでした</h1><Link to="/">ホームへ戻る</Link></section>;
  if (!data) return <p className="loading">履歴を読み込んでいます…</p>;
  return (
    <section className="history-page">
      <p className="eyebrow">HISTORY</p><h1>トレーニング履歴</h1>
      {data.sessions.length === 0 ? <p className="muted">まだ終了したセッションはありません。</p> : <ul className="history-list">
        {data.sessions.map(session => <li key={session.id}><Link to={`/history/${session.id}`}>
          <span>{new Date(session.startedAt).toLocaleDateString('ja-JP')}</span>
          <strong>{session.problems.filter(problem => problem.acceptedAt).length}/{session.problems.length} AC</strong>
          <em>{session.status === 'ABORTED' ? '中断' : '終了'}</em>
        </Link></li>)}
      </ul>}
      <div className="pagination">
        <button disabled={page <= 1} onClick={() => setPage(current => current - 1)}>前へ</button>
        <span>{data.page} / {Math.max(1, Math.ceil(data.total / data.pageSize))}</span>
        <button disabled={page * data.pageSize >= data.total} onClick={() => setPage(current => current + 1)}>次へ</button>
      </div>
    </section>
  );
}
