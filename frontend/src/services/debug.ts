type LogLevel = 'debug' | 'info' | 'warn' | 'error'

interface LogEntry {
  level: LogLevel
  message: string
  timestamp: string
  data?: any
}

class DebugService {
  private static instance: DebugService
  private logs: LogEntry[] = []
  private maxLogs: number = 1000
  private isDebugMode: boolean = process.env.NODE_ENV === 'development'

  private constructor() {
    window.addEventListener('error', this.handleGlobalError.bind(this))
    window.addEventListener('unhandledrejection', this.handlePromiseError.bind(this))
  }

  static getInstance(): DebugService {
    if (!DebugService.instance) {
      DebugService.instance = new DebugService()
    }
    return DebugService.instance
  }

  private handleGlobalError(event: ErrorEvent) {
    this.error('Global Error', {
      message: event.message,
      filename: event.filename,
      lineno: event.lineno,
      colno: event.colno,
      stack: event.error?.stack,
    })
  }

  private handlePromiseError(event: PromiseRejectionEvent) {
    this.error('Unhandled Promise Rejection', {
      reason: event.reason,
    })
  }

  private addLog(level: LogLevel, message: string, data?: any) {
    const log: LogEntry = {
      level,
      message,
      timestamp: new Date().toISOString(),
      data,
    }

    this.logs.unshift(log)
    if (this.logs.length > this.maxLogs) {
      this.logs.pop()
    }

    if (this.isDebugMode) {
      const consoleMethod = level === 'error' ? 'error' : level === 'warn' ? 'warn' : 'log'
      console[consoleMethod](`[${log.timestamp}] ${level.toUpperCase()}: ${message}`, data || '')
    }
  }

  debug(message: string, data?: any) {
    this.addLog('debug', message, data)
  }

  info(message: string, data?: any) {
    this.addLog('info', message, data)
  }

  warn(message: string, data?: any) {
    this.addLog('warn', message, data)
  }

  error(message: string, data?: any) {
    this.addLog('error', message, data)
  }

  getLogs(level?: LogLevel): LogEntry[] {
    return level ? this.logs.filter(log => log.level === level) : this.logs
  }

  clearLogs() {
    this.logs = []
  }

  downloadLogs() {
    const logText = this.logs
      .map(log => `[${log.timestamp}] ${log.level.toUpperCase()}: ${log.message} ${JSON.stringify(log.data || '')}`)
      .join('\n')

    const blob = new Blob([logText], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `debug-logs-${new Date().toISOString()}.txt`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }
}

export const debugService = DebugService.getInstance()
