import request from '@/utils/request'

export interface UploadedFile {
  id: number
  tenantId: number
  userId: number
  fileName: string
  fileKey: string
  fileType: string
  fileSize: number
  url: string
  storage: string
  createdAt: string
}

export interface FileQuery {
  page: number
  size: number
}

export function uploadFile(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post('/files/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function getFileList(params: FileQuery) {
  return request.get('/files', { params })
}

export function deleteFile(id: number) {
  return request.delete(`/files/${id}`)
}
