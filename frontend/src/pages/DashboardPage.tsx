import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { getPortfolio, removeFromPortfolio, addToPortfolio } from '../services/api'
import { useAuth } from '../context/AuthContext'

export default function DashboardPage() {
  const [portfolio, setPortfolio] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [newSymbol, setNewSymbol] = useState('')
  const { user, logout } = useAuth()

  useEffect(() => {
    loadPortfolio()
  }, [])

  const loadPortfolio = async () => {
    try {
      const data = await getPortfolio()
        setPortfolio(data || []) // Handle null response if portfolio is empty
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  const handleRemove = async (symbol: string) => {
    try {
      await removeFromPortfolio(symbol)
      loadPortfolio()
    } catch (e) {
      alert('Failed to remove stock')
    }
  }

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newSymbol) return
    try {
      await addToPortfolio(newSymbol.toUpperCase())
      setNewSymbol('')
      loadPortfolio()
    } catch (e) {
      alert('Failed to add stock')
    }
  }

  if (loading) return <div>Loading...</div>

  return (
    <div style={{ padding: 20, maxWidth: 800, margin: '0 auto' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1>Welcome, {user?.email}</h1>
        <button onClick={logout}>Logout</button>
      </div>

      <div style={{ marginTop: 20 }}>
        <h3>Your Portfolio</h3>
        <ul style={{ listStyle: 'none', padding: 0 }}>
          {portfolio.map(symbol => (
            <li key={symbol} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '10px 0', borderBottom: '1px solid #eee' }}>
              <Link to={`/quote/${symbol}`} style={{ fontWeight: 'bold', fontSize: '1.2em' }}>{symbol}</Link>
              <button onClick={() => handleRemove(symbol)} style={{ marginLeft: 'auto', background: '#ff4444', color: 'white', border: 'none', padding: '5px 10px', cursor: 'pointer' }}>Remove</button>
            </li>
          ))}
          {portfolio.length === 0 && <p>No stocks in portfolio.</p>}
        </ul>
      </div>

      <div style={{ marginTop: 30 }}>
        <h3>Add Stock</h3>
        <form onSubmit={handleAdd} style={{ display: 'flex', gap: 10 }}>
            <input 
                value={newSymbol}
                onChange={e => setNewSymbol(e.target.value)}
                placeholder="Symbol (e.g. MSFT)"
            />
            <button type="submit">Add</button>
        </form>
      </div>
    </div>
  )
}
