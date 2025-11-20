import { useState } from 'react'
import { getQuote, getIntraday } from './services/api'
import type { Quote, Candle } from './types'

export default function App() {
  const [symbol, setSymbol] = useState('AAPL')
  const [quote, setQuote] = useState<Quote | null>(null)
  const [candles, setCandles] = useState<Candle[] | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const run = async () => {
    setLoading(true); setError(null)
    try {
      const [q, c] = await Promise.all([
        getQuote(symbol),
        getIntraday(symbol, '1min')
      ])
      setQuote(q)
      setCandles(c)
    } catch (e:any) {
      setError(e.message ?? String(e))
    } finally { setLoading(false) }
  }

  return (
    <div style={{ padding: 24, fontFamily: 'system-ui, sans-serif', maxWidth: 900, margin: '0 auto' }}>
      <h1>GoStocks</h1>
      <p style={{ opacity: 0.7 }}>Enter a ticker and fetch a real-time quote and recent intraday candles.</p>
      <div style={{ display: 'flex', gap: 8 }}>
        <input value={symbol} onChange={e => setSymbol(e.target.value.toUpperCase())} placeholder="Symbol (e.g. AAPL)" />
        <button onClick={run} disabled={loading}>{loading ? 'Loading…' : 'Fetch'}</button>
      </div>
      {error && <p style={{ color: 'crimson' }}>{error}</p>}

      {quote && (
        <div style={{ marginTop: 16, padding: 12, border: '1px solid #ddd', borderRadius: 8 }}>
          <h2>{quote.symbol} — ${quote.price.toFixed(2)}</h2>
          <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap' }}>
            <span>Open: {quote.open}</span>
            <span>High: {quote.high}</span>
            <span>Low: {quote.low}</span>
            <span>Prev Close: {quote.previousClose}</span>
            <span>Change: {quote.change} ({quote.changePercent}%)</span>
          </div>
        </div>
      )}

      {candles && (
        <div style={{ marginTop: 16 }}>
          <h3>Recent candles</h3>
          <div style={{ maxHeight: 240, overflow: 'auto', border: '1px solid #eee' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr>
                  {['Time','Open','High','Low','Close','Volume'].map(h => <th key={h} style={{ textAlign: 'left', padding: 6, borderBottom: '1px solid #ddd' }}>{h}</th>)}
                </tr>
              </thead>
              <tbody>
                {candles.map((c, i) => (
                  <tr key={i}>
                    <td style={{ padding: 6, borderBottom: '1px solid #f3f3f3' }}>{String(c.time)}</td>
                    <td style={{ padding: 6, borderBottom: '1px solid #f3f3f3' }}>{c.open}</td>
                    <td style={{ padding: 6, borderBottom: '1px solid #f3f3f3' }}>{c.high}</td>
                    <td style={{ padding: 6, borderBottom: '1px solid #f3f3f3' }}>{c.low}</td>
                    <td style={{ padding: 6, borderBottom: '1px solid #f3f3f3' }}>{c.close}</td>
                    <td style={{ padding: 6, borderBottom: '1px solid #f3f3f3' }}>{c.volume}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
