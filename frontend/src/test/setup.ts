import '@testing-library/jest-dom'
import { vi } from 'vitest'
import { TextEncoder, TextDecoder } from 'util'

// Mock WebSocket
class MockWebSocket {
  onopen: (() => void) | null = null
  onclose: (() => void) | null = null
  onmessage: ((data: any) => void) | null = null
  onerror: ((error: any) => void) | null = null
  readyState = WebSocket.CONNECTING

  constructor() {
    setTimeout(() => {
      this.readyState = WebSocket.OPEN
      this.onopen?.()
    }, 100)
  }

  send(data: string) {
    console.log('WebSocket send:', data)
  }

  close() {
    this.readyState = WebSocket.CLOSED
    this.onclose?.()
  }
}

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
}

// Mock ResizeObserver
class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}

// Mock IntersectionObserver
class IntersectionObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}

// Mock Canvas
HTMLCanvasElement.prototype.getContext = vi.fn(() => ({
  fillRect: vi.fn(),
  clearRect: vi.fn(),
  getImageData: vi.fn(() => ({
    data: new Array(4),
  })),
  putImageData: vi.fn(),
  createImageData: vi.fn(),
  setTransform: vi.fn(),
  drawImage: vi.fn(),
  save: vi.fn(),
  restore: vi.fn(),
  scale: vi.fn(),
  rotate: vi.fn(),
  translate: vi.fn(),
  transform: vi.fn(),
  beginPath: vi.fn(),
  moveTo: vi.fn(),
  lineTo: vi.fn(),
  stroke: vi.fn(),
  fill: vi.fn(),
  arc: vi.fn(),
}))

// Setup global mocks
global.WebSocket = MockWebSocket as any
global.localStorage = localStorageMock as any
global.ResizeObserver = ResizeObserver
global.IntersectionObserver = IntersectionObserver as any
global.TextEncoder = TextEncoder
global.TextDecoder = TextDecoder

// Mock matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

// Mock requestAnimationFrame
global.requestAnimationFrame = vi.fn(callback => setTimeout(callback, 0))
global.cancelAnimationFrame = vi.fn(id => clearTimeout(id))

// Mock console methods
console.error = vi.fn()
console.warn = vi.fn()
console.log = vi.fn() 