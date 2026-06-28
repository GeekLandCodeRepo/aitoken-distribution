import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { Layout } from './components/layout/Layout'
import { Toaster } from './components/ui/sonner'
import LoginPage from './pages/LoginPage'
import { DashboardPage } from './pages/DashboardPage'
import { KeysPage } from './pages/KeysPage'
import { UsagePage } from './pages/UsagePage'
import { ModelPlazaPage } from './pages/ModelPlazaPage'
import { RedeemPage } from './pages/RedeemPage'
import { SettingsPage } from './pages/SettingsPage'
import { UsersPage } from './pages/UsersPage'
import { AdminDashboardPage } from './pages/AdminDashboardPage'
import { ChannelsPage } from './pages/ChannelsPage'
import { ModelManagementPage } from './pages/ModelManagementPage'
import { RedeemCodesPage } from './pages/RedeemCodesPage'
import { LogsPage } from './pages/LogsPage'
import { SiteSettingsPage } from './pages/SiteSettingsPage'

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        
        <Route path="/" element={<Layout />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<DashboardPage />} />
          
          {/* User routes */}
          <Route path="keys" element={<KeysPage />} />
          <Route path="models" element={<ModelPlazaPage />} />
          <Route path="usage" element={<UsagePage />} />
          <Route path="redeem" element={<RedeemPage />} />
          <Route path="settings" element={<SettingsPage />} />
          
          {/* Admin routes */}
          <Route path="admin/dashboard" element={<AdminDashboardPage />} />
          <Route path="admin/users" element={<UsersPage />} />
          <Route path="admin/channels" element={<ChannelsPage />} />
          <Route path="admin/models" element={<ModelManagementPage />} />
          <Route path="admin/redeem-codes" element={<RedeemCodesPage />} />
          <Route path="admin/logs" element={<LogsPage />} />
          <Route path="admin/settings" element={<SiteSettingsPage />} />
        </Route>
      </Routes>
      <Toaster position="top-right" richColors />
    </Router>
  )
}

export default App
