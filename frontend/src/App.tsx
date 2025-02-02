import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { Layout } from 'antd'
import MainLayout from '@/layouts/MainLayout'
import TradingView from '@/pages/TradingView'
import Analysis from '@/pages/Analysis'
import Monitoring from '@/pages/Monitoring'

const App = () => {
  return (
    <Router>
      <Layout style={{ minHeight: '100vh' }}>
        <Routes>
          <Route path="/" element={<MainLayout />}>
            <Route index element={<TradingView />} />
            <Route path="analysis" element={<Analysis />} />
            <Route path="monitoring" element={<Monitoring />} />
          </Route>
        </Routes>
      </Layout>
    </Router>
  )
}

export default App 