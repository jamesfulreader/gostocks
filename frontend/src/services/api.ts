import type { Quote, Candle, User, AuthResponse } from '../types'

const json = async <T>(res: Response) => {
  if (!res.ok) throw new Error(await res.text())
  return res.json() as Promise<T>
}

export const getHealth = () => fetch('/healthz').then(r => r.text())
export const getQuote = (symbol: string) => fetch(`/api/quote?symbol=${encodeURIComponent(symbol)}`).then(json<Quote>)
export const getIntraday = (symbol: string, interval = '1min') =>
  fetch(`/api/intraday?symbol=${encodeURIComponent(symbol)}&interval=${interval}`).then(json<Candle[]>)

// Auth & Portfolio API
const getHeaders = (): HeadersInit => {
  const token = localStorage.getItem('token')
  if (token) {
    return { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' }
  }
  return { 'Content-Type': 'application/json' }
}

export const register = (email: string, pass: string) => 
  fetch('/api/register', { method: 'POST', body: JSON.stringify({ email, password: pass }) }).then(json<User>)

export const login = (email: string, pass: string) => 
  fetch('/api/login', { method: 'POST', body: JSON.stringify({ email, password: pass }) }).then(json<AuthResponse>)

export const getPortfolio = () => 
  fetch('/api/portfolio', { headers: getHeaders() }).then(json<string[]>)

export const addToPortfolio = (symbol: string) => 
  fetch('/api/portfolio', { method: 'POST', body: JSON.stringify({ symbol }), headers: getHeaders() }).then(json<{status: string}>)

export const removeFromPortfolio = (symbol: string) => 
  fetch(`/api/portfolio?symbol=${encodeURIComponent(symbol)}`, { method: 'DELETE', headers: getHeaders() }).then(json<{status: string}>)
