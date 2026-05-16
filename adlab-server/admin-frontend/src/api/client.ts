import axios, { AxiosError, AxiosResponse } from 'axios'
import { msg } from '../hooks/useMessage'
import type { ApiResponse } from '../types'

const client = axios.create({
  baseURL: '/',
  timeout: 15000,
})

client.interceptors.request.use((config) => {
  const token = localStorage.getItem('admin_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

client.interceptors.response.use(
  (res: AxiosResponse<ApiResponse>) => {
    if (res.config.responseType === 'blob') {
      return res
    }

    const data = res.data
    if (data.code !== 0) {
      const message = data.message || '请求失败'
      msg.error(message)
      return Promise.reject(new Error(message))
    }
    return res
  },
  (err: AxiosError<ApiResponse>) => {
    if (err.code === 'ECONNABORTED' || err.message?.includes('timeout')) {
      msg.error('请求超时，请检查后端服务是否启动')
    } else if (err.response?.status === 401) {
      msg.error('未授权，请检查 Admin Token 配置')
    } else if (err.response?.status === 404) {
      // Keep 404s quiet by default; callers can decide whether to surface them.
    } else if ((err.response?.status ?? 0) >= 500) {
      msg.error(`服务器错误 (${err.response?.status})`)
    } else if (!err.response) {
      msg.error('无法连接到后端服务，请确认服务已启动（默认端口 8080）')
    } else {
      msg.error(err.response.data?.message || err.message || '请求失败')
    }

    return Promise.reject(err)
  },
)

export function unwrap<T>(res: AxiosResponse<ApiResponse<T>>): T {
  return res.data.data
}

export default client
