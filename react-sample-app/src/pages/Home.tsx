import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ApiError, trainingApi } from '../util/TrainingApi';
import { TrainingSession } from '../types/Training';

export function Home() {
  const navigate = useNavigate();
  const [active, setActive] = useState<TrainingSession | null>(null);
  const [loading, setLoading] = useState(true);
  const [starting, setStarting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    trainingApi.active()
      .then(({ session }) => { if (!cancelled) setActive(session); })
      .catch((cause: unknown) => {
        if (!cancelled && !(cause instanceof ApiError && cause.status === 404)) {
          setError('進行中のセッションを確認できませんでした。');
        }
      })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, []);

  const start = async () => {
    setStarting(true);
    setError(null);
    try {
      const { session } = await trainingApi.start();
      navigate(`/training/${session.id}`, { replace: true });
    } catch (cause) {
      setError(cause instanceof Error ? `開始できませんでした: ${cause.message}` : '開始できませんでした。');
      setStarting(false);
    }
  };

  return (
    <section className="home-page">
      <p className="eyebrow">ADT VIRTUAL CONTEST</p>
      <h1>毎日75分、D/E/Fを解く。</h1>
      <p className="lede">固定セットを引き当てるのではなく、あなたのAC履歴を除外して新しい5問を作成します。</p>
      <div className="session-config" aria-label="セッション設定">
        <div><strong>75</strong><span>分</span><small>制限時間</small></div>
        <div><strong>D1</strong><span>・E3・F1</span><small>出題構成</small></div>
        <div><strong>5</strong><span>問</span><small>同一ABCは重複なし</small></div>
      </div>
      {error && <p className="notice warning" role="alert">{error}</p>}
      {loading ? <p className="muted">セッションを確認中…</p> : active ? (
        <Link className="button primary" to={`/training/${active.id}`}>進行中のセッションを再開</Link>
      ) : (
        <button className="button primary" onClick={start} disabled={starting}>{starting ? '問題を選定中…' : 'START 75 MINUTES'}</button>
      )}
      <Link className="history-link" to="/history">過去のトレーニングを見る →</Link>
      <p className="privacy-note">問題は開始するまで表示されません。</p>
    </section>
  );
}
