import { useState, useEffect } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'

interface HealthResponse {
  status: string
  timestamp: number
}

function App() {
  const [count, setCount] = useState(0)
  const [health, setHealth] = useState<HealthResponse | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetch('/api/health')
      .then(res => res.json())
      .then(data => setHealth(data))
      .catch(err => setError(err.message))
  }, [])

  return (
    <>
      <div>
        <a href="https://vite.dev" target="_blank">
          <img src={viteLogo} className="logo" alt="Vite logo" />
        </a>
        <a href="https://react.dev" target="_blank">
          <img src={reactLogo} className="logo react" alt="React logo" />
        </a>
      </div>
      <h1>Claude Proxy</h1>
      <div className="card">
        <h2>Backend Health Check</h2>
        {error && <p style={{color: 'red'}}>Error: {error}</p>}
        {health && (
          <div>
            <p>Status: <strong>{health.status}</strong></p>
            <p>Timestamp: {new Date(health.timestamp * 1000).toLocaleString()}</p>
          </div>
        )}
      </div>
      <div className="card">
        <button onClick={() => setCount((count) => count + 1)}>
          count is {count}
        </button>
        <p>
          Edit <code>src/App.tsx</code> and save to test HMR
        </p>
      </div>
      <p className="read-the-docs">
        Frontend running on port 5173, Backend on port 4000
      </p>
    </>
  )
}

export default App
