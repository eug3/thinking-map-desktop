import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import HomePage from '@/features/home/HomePage';
import WorkspaceLayout from '@/features/workspace/WorkspaceLayout';
import { Toaster } from '@/components/ui/sonner';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/history" element={<HomePage />} />
        <Route path="/workspace/:id" element={<WorkspaceLayout />} />
        <Route path="/map/:id" element={<Navigate to={`/workspace/:id`} replace />} />
      </Routes>
      <Toaster />
    </BrowserRouter>
  );
}

export default App;
