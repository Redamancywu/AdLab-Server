import client, { unwrap } from './client'

export interface Document {
  key: string
  title: string
  content: string
  updated_at: string
}

export async function getDoc(key: string): Promise<Document> {
  const res = await client.get(`/api/v1/docs/${key}`)
  return unwrap(res)
}

export async function saveDoc(key: string, data: { title: string; content: string }): Promise<Document> {
  const res = await client.post(`/admin/docs/${key}`, data)
  return unwrap(res)
}
