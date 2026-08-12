import { Route, Routes } from 'react-router-dom';
import { Layout } from './components/Layout';
import { History } from './pages/History';
import { Home } from './pages/Home';
import { NotFound } from './pages/NotFound';
import { SessionDetail } from './pages/SessionDetail';
import { Training } from './pages/Training';

export default function App() {
  return <Layout><Routes>
    <Route path="/" element={<Home />} />
    <Route path="/training/:id" element={<Training />} />
    <Route path="/results/:id" element={<SessionDetail result />} />
    <Route path="/history" element={<History />} />
    <Route path="/history/:id" element={<SessionDetail />} />
    <Route path="*" element={<NotFound />} />
  </Routes></Layout>;
}
