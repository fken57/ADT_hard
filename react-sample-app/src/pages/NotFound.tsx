import { Link } from 'react-router-dom';

export function NotFound() {
  return <section className="not-found"><p className="eyebrow">404</p><h1>ページが見つかりません</h1><Link className="button primary" to="/">ホームへ戻る</Link></section>;
}
