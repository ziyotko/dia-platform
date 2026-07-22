import request from '@/utils/request'

export function uploadFile(file: File, dir?: string) {
  const formData = new FormData()
  formData.append('file', file)
  if (dir) {
    formData.append('dir', dir)
  }
  return request.post('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
