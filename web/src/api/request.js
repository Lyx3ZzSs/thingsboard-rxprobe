import axios from 'axios'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 30000
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  response => {
    const res = response.data
    if (res.code !== 0) {
      console.error('API Error:', res.message)
      return Promise.reject(new Error(res.message))
    }
    return res
  },
  error => {
    console.error('Request Error:', error.message)
    return Promise.reject(error)
  }
)

export default request
