import { formatDuration } from '../../hooks/Training/useCountdown';

export function Countdown({ seconds }: { seconds: number }) {
  return <time className="countdown" aria-label={`残り ${formatDuration(seconds)}`}>{formatDuration(seconds)}</time>;
}
