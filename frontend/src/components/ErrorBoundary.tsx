import React, { Component, ErrorInfo, ReactNode } from 'react'
import { Card, Typography, Button, Space } from 'antd'
import { debugService } from '@/services/debug'

const { Title, Text } = Typography

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
  errorInfo: ErrorInfo | null
}

class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
    errorInfo: null,
  }

  public static getDerivedStateFromError(error: Error): State {
    return {
      hasError: true,
      error,
      errorInfo: null,
    }
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    debugService.error('React Component Error', {
      error,
      errorInfo,
      componentStack: errorInfo.componentStack,
    })
  }

  private handleReload = () => {
    window.location.reload()
  }

  private handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    })
  }

  private handleDownloadLogs = () => {
    debugService.downloadLogs()
  }

  public render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }

      return (
        <Card style={{ maxWidth: 800, margin: '32px auto' }}>
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            <Title level={3}>组件错误</Title>
            <Text type="danger">
              {this.state.error?.message || '发生未知错误'}
            </Text>
            {process.env.NODE_ENV === 'development' && (
              <div style={{ 
                background: '#f5f5f5', 
                padding: 16, 
                borderRadius: 4,
                maxHeight: 200,
                overflow: 'auto'
              }}>
                <pre style={{ margin: 0 }}>
                  {this.state.error?.stack}
                </pre>
              </div>
            )}
            <Space>
              <Button type="primary" onClick={this.handleReset}>
                重试
              </Button>
              <Button onClick={this.handleReload}>
                刷新页面
              </Button>
              <Button onClick={this.handleDownloadLogs}>
                下载日志
              </Button>
            </Space>
          </Space>
        </Card>
      )
    }

    return this.props.children
  }
}

export default ErrorBoundary
