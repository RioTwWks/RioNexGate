import axios from 'axios';

const API_KEY_STORAGE = 'rionexgate_api_key';

const api = axios.create({
  baseURL: '/api',
});

api.interceptors.request.use((config) => {
  const key = localStorage.getItem(API_KEY_STORAGE);
  if (key) {
    config.headers['X-API-Key'] = key;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      clearApiKey();
      if (!window.location.pathname.startsWith('/login')) {
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  },
);

export function setApiKey(key: string) {
  localStorage.setItem(API_KEY_STORAGE, key);
}

export function getApiKey(): string | null {
  return localStorage.getItem(API_KEY_STORAGE);
}

export function clearApiKey() {
  localStorage.removeItem(API_KEY_STORAGE);
}

export default api;
