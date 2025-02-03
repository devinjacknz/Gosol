import { vi } from 'vitest';

const mockAxiosInstance = {
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
  create: vi.fn().mockReturnThis(),
  interceptors: {
    request: { use: vi.fn(), eject: vi.fn() },
    response: { use: vi.fn(), eject: vi.fn() }
  }
};

export default {
  ...mockAxiosInstance,
  create: () => mockAxiosInstance
};
