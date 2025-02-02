import React from 'react'
import ReactDOM from 'react-dom/client'
import { Provider } from 'react-redux'
import { ConfigProvider } from 'antd'
import { store } from '@/store'
import App from './App'
import { lightTheme } from '@/styles/theme'
import '@/styles/global.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Provider store={store}>
      <ConfigProvider theme={lightTheme}>
        <App />
      </ConfigProvider>
    </Provider>
  </React.StrictMode>
) 