import type { Quote, Candle } from '../types'

const json = async <T>(res: Response) => {
  if (!res.ok) throw new Error(await res.text())
  return res.json() as Promise<T>
}

export const getHealth = () => fetch('/healthz').then(r => r.text())
export const getQuote = (symbol: string) => fetch(`/api/quote?symbol=${encodeURIComponent(symbol)}`).then(json<Quote>)
export const getIntraday = (symbol: string, interval = '1min') =>
  fetch(`/api/intraday?symbol=${encodeURIComponent(symbol)}&interval=${interval}`).then(json<Candle[]>)
