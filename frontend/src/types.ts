export type Quote = {
  symbol: string
  price: number
  open: number
  high: number
  low: number
  previousClose: number
  change: number
  changePercent: number
  timestamp?: string
}

export type Candle = {
  time: string | Date
  open: number
  high: number
  low: number
  close: number
  volume: number
}

export type User = {
  id: number
  email: string
  created_at: string
}

export type AuthResponse = {
  token: string
  user: User
}
