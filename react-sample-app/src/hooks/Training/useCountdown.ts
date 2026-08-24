import { useEffect, useMemo, useState } from 'react';

export function secondsRemaining(startedAt: string, durationSeconds: number, serverOffsetMs: number, nowMs = Date.now()): number {
  const deadline = new Date(startedAt).getTime() + durationSeconds * 1000;
  return Math.max(0, Math.ceil((deadline - (nowMs + serverOffsetMs)) / 1000));
}

export function useCountdown(startedAt: string, durationSeconds: number, serverOffsetMs: number): number {
  const calculate = () => secondsRemaining(startedAt, durationSeconds, serverOffsetMs);
  const [remaining, setRemaining] = useState(calculate);

  useEffect(() => {
    setRemaining(calculate());
    const interval = window.setInterval(() => setRemaining(calculate()), 250);
    return () => window.clearInterval(interval);
  // Inputs intentionally reset the reference clock when a freshly synced session arrives.
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [startedAt, durationSeconds, serverOffsetMs]);

  return remaining;
}

export function formatDuration(totalSeconds: number): string {
  const minutes = Math.floor(totalSeconds / 60).toString().padStart(2, '0');
  const seconds = (totalSeconds % 60).toString().padStart(2, '0');
  return `${minutes}:${seconds}`;
}

export function useServerOffset(serverNow?: string): number {
  return useMemo(() => serverNow ? new Date(serverNow).getTime() - Date.now() : 0, [serverNow]);
}
