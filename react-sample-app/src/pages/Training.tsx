import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Countdown } from '../components/Training/Countdown';
import { ProblemList } from '../components/Training/ProblemList';
import { secondsRemaining, useCountdown, useServerOffset } from '../hooks/Training/useCountdown';
import { SessionResponse } from '../types/Training';
import { trainingApi } from '../util/TrainingApi';

export function Training() {
  const { id = '' } = useParams();
  const navigate = useNavigate();
  const [data, setData] = useState<SessionResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [aborting, setAborting] = useState(false);
  const isMounted = useRef(true);
  const finishing = useRef(false);
  const offset = useServerOffset(data?.serverNow);
  const remaining = useCountdown(data?.session.startedAt || new Date().toISOString(), data?.session.durationSeconds || 0, offset);
  const status = data?.session.status;

  const apply = useCallback((response: SessionResponse) => {
    if (!isMounted.current) return;
    setData(response);
    if (response.submissionSync?.status === 'STALE' || response.submissionSync?.status === 'FAILED') {
      setError(response.submissionSync.message || '提出結果の更新が遅れています。タイマーは継続しています。');
    }
  }, []);

  useEffect(() => {
    isMounted.current = true;
    trainingApi.get(id).then(apply).catch(() => setError('セッションを読み込めませんでした。再読み込みしてください。'));
    return () => { isMounted.current = false; };
  }, [id, apply]);

  useEffect(() => {
    if (status !== 'ACTIVE') return;
    const sync = () => trainingApi.sync(id).then(response => { apply(response); }).catch(() => {
      if (isMounted.current) setError('提出結果を更新できませんでした。タイマーは継続しています。');
    });
    const timer = window.setInterval(sync, 15_000);
    return () => window.clearInterval(timer);
  }, [status, id, apply]);

  useEffect(() => {
    if (!data || finishing.current) return;
    if (data.session.status !== 'ACTIVE') {
      navigate(`/results/${id}`, { replace: true });
      return;
    }
    const expired = secondsRemaining(data.session.startedAt, data.session.durationSeconds, offset) === 0;
    if (remaining !== 0 || !expired) return;
    finishing.current = true;
    // One final synchronization captures submissions made in the last polling interval.
    trainingApi.sync(id)
      .then(() => navigate(`/results/${id}`, { replace: true }))
      .catch(() => navigate(`/results/${id}`, { replace: true }));
  }, [data, remaining, offset, id, navigate]);

  const abort = async () => {
    if (!window.confirm('このセッションを中断しますか？\n中断後は、元の終了予定時刻まで新しいセッションを開始できません。')) return;
    setAborting(true);
    try {
      const response = await trainingApi.abort(id);
      navigate(`/results/${response.session.id}`, { replace: true });
    } catch {
      setError('中断できませんでした。もう一度お試しください。');
      setAborting(false);
    }
  };

  if (!data) return <p className="loading">セッションを読み込んでいます…</p>;
  const { session } = data;
  return (
    <section className="training-page">
      <div className="training-topline"><p>TRAINING IN PROGRESS</p><Countdown seconds={remaining} /></div>
      <h1>集中して、1問ずつ。</h1>
      <p className="muted">{session.atcoderUserId} · {session.problems.filter(problem => problem.acceptedAt).length} / {session.problems.length} AC</p>
      {error && <p className="notice warning" role="alert">{error}</p>}
      <ProblemList problems={session.problems} startedAt={session.startedAt} />
      <button className="button text-button" disabled={aborting} onClick={abort}>{aborting ? '中断中…' : 'このセッションを中断する'}</button>
    </section>
  );
}
