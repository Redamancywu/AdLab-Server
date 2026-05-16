import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import AppLayout from './components/Layout'
import Dashboard from './pages/Dashboard'
import AppList from './pages/apps/AppList'
import PlacementList from './pages/placements/PlacementList'
import SourceList from './pages/sources/SourceList'
import DSPConfigList from './pages/dsp-configs/DSPConfigList'
import MaterialList from './pages/materials/MaterialList'
import MockAdList from './pages/mock-ads/MockAdList'
import BidLogList from './pages/logs/BidLogList'
import ScenarioSwitch from './pages/scenarios/ScenarioSwitch'
import StatsPage from './pages/stats/StatsPage'
import ChangeLogList from './pages/change-logs/ChangeLogList'
import SettingsPage from './pages/settings/SettingsPage'
import AdPlayer from './pages/ad-player/AdPlayer'

export default function App() {
  return (
    <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
      <Routes>
        <Route path="/" element={<AppLayout />}>
          <Route index element={<Dashboard />} />
          <Route path="apps" element={<AppList />} />
          <Route path="placements" element={<PlacementList />} />
          <Route path="sources" element={<SourceList />} />
          <Route path="dsp-configs" element={<DSPConfigList />} />
          <Route path="materials" element={<MaterialList />} />
          <Route path="mock-ads" element={<MockAdList />} />
          <Route path="logs" element={<BidLogList />} />
          <Route path="scenarios" element={<ScenarioSwitch />} />
          <Route path="stats" element={<StatsPage />} />
          <Route path="change-logs" element={<ChangeLogList />} />
          <Route path="settings" element={<SettingsPage />} />
          <Route path="ad-player" element={<AdPlayer />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
