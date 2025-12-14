import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Jobs from './pages/Jobs'
import Builds from './pages/Builds'
import Workers from './pages/Workers'
import Deployments from './pages/Deployments'
import Plugins from './pages/Plugins'
import BuildDetail from './pages/BuildDetail'
import JobDetail from './pages/JobDetail'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="jobs" element={<Jobs />} />
          <Route path="jobs/:id" element={<JobDetail />} />
          <Route path="builds" element={<Builds />} />
          <Route path="builds/:id" element={<BuildDetail />} />
          <Route path="workers" element={<Workers />} />
          <Route path="deployments" element={<Deployments />} />
          <Route path="plugins" element={<Plugins />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
