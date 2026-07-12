import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import MainLayout from './layouts/MainLayout';
import Dashboard from './pages/Dashboard';
import Containers from './pages/Containers';
import Experiments from './pages/Experiments';
import Metrics from './pages/Metrics';
import Runtime from './pages/Runtime';
import Logs from './pages/Logs';
import Settings from './pages/Settings';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 2,
    },
  },
});

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<MainLayout />}>
            <Route index element={<Dashboard />} />
            <Route path="containers" element={<Containers />} />
            <Route path="experiments" element={<Experiments />} />
            <Route path="metrics" element={<Metrics />} />
            <Route path="runtime" element={<Runtime />} />
            <Route path="logs" element={<Logs />} />
            <Route path="settings" element={<Settings />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
