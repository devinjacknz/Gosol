import type { AxiosError, AxiosResponse, InternalAxiosRequestConfig } from 'axios';
import type { ApiErrorResponse } from '@/types/api';

export const createApiError = (status: number, message: string, code = 'ERROR'): AxiosError => {
  const error = new Error(message) as AxiosError;
  error.isAxiosError = true;
  error.code = code;
  error.config = {
    url: '',
    method: 'get',
    baseURL: 'http://localhost:8080',
    timeout: 5000,
    headers: { 'Content-Type': 'application/json' }
  } as InternalAxiosRequestConfig;
  error.response = {
    status,
    statusText: message,
    headers: { 'content-type': 'application/json' },
    config: error.config,
    data: {
      success: false,
      error: message,
      code,
      details: {
        timestamp: new Date().toISOString(),
        path: error.config.url,
        method: error.config.method?.toUpperCase()
      }
    } as ApiErrorResponse
  } as AxiosResponse;
  return error;
};

export const createNetworkError = (message: string): AxiosError => {
  const error = new Error(message) as AxiosError;
  error.isAxiosError = true;
  error.code = 'NETWORK_ERROR';
  error.config = {
    url: '',
    method: 'get',
    baseURL: 'http://localhost:8080',
    timeout: 5000
  } as InternalAxiosRequestConfig;
  error.request = { status: 0, statusText: message };
  return error;
};

export const createTimeoutError = (message = 'Request timeout'): AxiosError => {
  const error = createNetworkError(message);
  error.code = 'ECONNABORTED';
  return error;
};

export const createConnectionError = (message = 'Connection refused'): AxiosError => {
  const error = createNetworkError(message);
  error.code = 'ECONNREFUSED';
  return error;
};
