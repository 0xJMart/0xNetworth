import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import './index.css'
import Layout from './components/Layout.tsx'
import Dashboard from './pages/Dashboard.tsx'
import WorkflowReviewPage from './pages/WorkflowReviewPage.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="workflows" element={<WorkflowReviewPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  </StrictMode>,
)

