import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor
api.interceptors.request.use(
  (config) => {
    // Add auth token if available
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Handle unauthorized
      localStorage.removeItem('auth_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default api

// API functions
export const jobsApi = {
  list: () => api.get('/jobs'),
  get: (id: string) => api.get(`/jobs/${id}`),
  create: (data: any) => api.post('/jobs', data),
  update: (id: string, data: any) => api.put(`/jobs/${id}`, data),
  delete: (id: string) => api.delete(`/jobs/${id}`),
  trigger: (id: string, params?: any) => api.post(`/jobs/${id}/trigger`, params),
}

export const buildsApi = {
  list: (params?: any) => api.get('/builds', { params }),
  get: (id: string) => api.get(`/builds/${id}`),
  cancel: (id: string) => api.post(`/builds/${id}/cancel`),
  logs: (id: string) => api.get(`/builds/${id}/logs`),
  artifacts: (id: string) => api.get(`/builds/${id}/artifacts`),
}

export const workersApi = {
  list: () => api.get('/workers'),
  get: (id: string) => api.get(`/workers/${id}`),
  update: (id: string, data: any) => api.put(`/workers/${id}`, data),
  drain: (id: string) => api.post(`/workers/${id}/drain`),
}

export const deploymentsApi = {
  list: (params?: any) => api.get('/deployments', { params }),
  get: (id: string) => api.get(`/deployments/${id}`),
  create: (data: any) => api.post('/deployments', data),
  rollback: (id: string) => api.post(`/deployments/${id}/rollback`),
}

export const pluginsApi = {
  list: () => api.get('/plugins'),
  get: (id: string) => api.get(`/plugins/${id}`),
  install: (data: any) => api.post('/plugins', data),
}
