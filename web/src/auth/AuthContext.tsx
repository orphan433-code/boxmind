import { createContext, useContext, useMemo, useState, type ReactNode } from 'react'
import type { User } from '../types'

const TOKEN_KEY = 'boxmind-token'
const USER_KEY = 'boxmind-user'

type AuthContextValue = {
  token: string | null
  user: User | null
  login: (token: string, user: User) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

function readUser(): User | null {
  const raw = localStorage.getItem(USER_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as User
  } catch {
    return null
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem(TOKEN_KEY))
  const [user, setUser] = useState<User | null>(() => readUser())

  const value = useMemo(
    () => ({
      token,
      user,
      login(nextToken: string, nextUser: User) {
        localStorage.setItem(TOKEN_KEY, nextToken)
        localStorage.setItem(USER_KEY, JSON.stringify(nextUser))
        setToken(nextToken)
        setUser(nextUser)
        window.dispatchEvent(new CustomEvent('boxmind-auth-change'))
      },
      logout() {
        localStorage.removeItem(TOKEN_KEY)
        localStorage.removeItem(USER_KEY)
        setToken(null)
        setUser(null)
        window.dispatchEvent(new CustomEvent('boxmind-auth-change'))
      },
    }),
    [token, user],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
