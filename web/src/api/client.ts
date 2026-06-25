import type { Bookmark, VerifyLoginResult } from '../types'

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api/v1'

type ApiError = { error: string }

async function request<T>(
  path: string,
  options: RequestInit = {},
  token?: string | null,
): Promise<T> {
  const headers = new Headers(options.headers)
  headers.set('Content-Type', 'application/json')
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
  })

  if (!response.ok) {
    let message = `request failed (${response.status})`
    try {
      const body = (await response.json()) as ApiError
      if (body.error) message = body.error
    } catch {
      // ignore
    }
    throw new Error(message)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}

export function requestLogin(email: string) {
  return request<{ message: string }>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email }),
  })
}

export function verifyLogin(email: string, code: string) {
  return request<VerifyLoginResult>('/auth/verify', {
    method: 'POST',
    body: JSON.stringify({ email, code }),
  })
}

export function listBookmarks(token: string) {
  return request<Bookmark[]>('/bookmarks', { method: 'GET' }, token)
}

export function createBookmark(token: string, url: string) {
  return request<Bookmark>(
    '/bookmarks',
    {
      method: 'POST',
      body: JSON.stringify({ url }),
    },
    token,
  )
}

export function getBookmark(token: string, id: string) {
  return request<Bookmark>(`/bookmarks/${id}`, { method: 'GET' }, token)
}

export function deleteBookmark(token: string, id: string) {
  return request<void>(`/bookmarks/${id}`, { method: 'DELETE' }, token)
}
