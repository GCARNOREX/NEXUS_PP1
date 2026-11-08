import { useEffect, useState } from 'react'
import './App.css'

// Fetches `url` every `ms` milliseconds; pass null to pause.
function usePoll(url, ms) {
  const [data, setData] = useState(null)
  const [error, setError] = useState(null)
  useEffect(() => {
    if (!url) { setData(null); return }
    let alive = true
    const tick = async () => {
      try {
        const res = await fetch(url)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const json = await res.json()
        if (alive) { setData(json); setError(null) }
      } catch (e) {
        if (alive) setError(e.message)
      }
    }
    tick()
    const id = setInterval(tick, ms)
    return () => { alive = false; clearInterval(id) }
  }, [url, ms])
  return { data, error }
}

function Sparkline({ values }) {
  if (values.length < 2) return <div className="spark-empty">collecting…</div>
  const w = 300, h = 60
  const max = Math.max(5, ...values)
  const pts = values
    .map((v, i) => `${(i / (values.length - 1)) * w},${h - (v / max) * h}`)
    .join(' ')
  return (
    <svg viewBox={`0 0 ${w} ${h}`} className="spark">
      <polyline points={pts} fill="none" stroke="currentColor" strokeWidth="2" />
    </svg>
  )
}

export default function App() {
  const hosts = usePoll('/api/hosts', 3000)
  const [selected, setSelected] = useState('')
  const [sortKey, setSortKey] = useState('cpu_percent')
  const [cpuHistory, setCpuHistory] = useState([])

  useEffect(() => {
    if (!selected && hosts.data?.length) setSelected(hosts.data[0].host)
  }, [hosts.data, selected])

  const live = usePoll(selected ? `/api/hosts/${encodeURIComponent(selected)}` : null, 1000)

  useEffect(() => { setCpuHistory([]) }, [selected])
  useEffect(() => {
    if (live.data) setCpuHistory((h) => [...h.slice(-59), live.data.cpu_percent])
  }, [live.data])

  const status = hosts.data?.find((h) => h.host === selected)
  const online = status && status.last_seen_seconds_ago < 5
  const snap = live.data
  const usedKB = snap ? snap.mem.total_kb - snap.mem.available_kb : 0
  const memPct = snap ? (usedKB / snap.mem.total_kb) * 100 : 0

  const procs = [...(snap?.processes ?? [])].sort((a, b) =>
    sortKey === 'name' ? a.name.localeCompare(b.name) : b[sortKey] - a[sortKey]
  )

  return (
    <div className="app">
      <header>
        <h1>NEXUS</h1>
        <select value={selected} onChange={(e) => setSelected(e.target.value)}>
          {(hosts.data ?? []).map((h) => (
            <option key={h.host} value={h.host}>{h.host}</option>
          ))}
        </select>
        {selected && (
          <span className={online ? 'dot on' : 'dot off'}>{online ? 'online' : 'offline'}</span>
        )}
      </header>

      {hosts.error && <p className="err">Can't reach the API: {hosts.error}</p>}
      {!selected && !hosts.error && <p>No hosts yet. Start an agent and pipe its output to the API.</p>}

      {snap && (
        <>
          <section className="cards">
            <div className="card">
              <h3>CPU</h3>
              <div className="big">{snap.cpu_percent.toFixed(1)}%</div>
              <Sparkline values={cpuHistory} />
            </div>
            <div className="card">
              <h3>Memory</h3>
              <div className="big">{(usedKB / 1024).toFixed(0)} MB</div>
              <div className="bar"><div style={{ width: `${memPct}%` }} /></div>
              <small>{memPct.toFixed(0)}% of {(snap.mem.total_kb / 1024).toFixed(0)} MB</small>
            </div>
          </section>

          <table>
            <thead>
              <tr>
                <th>PID</th>
                <th onClick={() => setSortKey('name')}>Name</th>
                <th>State</th>
                <th>Threads</th>
                <th onClick={() => setSortKey('rss_kb')}>RSS (MB)</th>
                <th onClick={() => setSortKey('cpu_percent')}>CPU %</th>
              </tr>
            </thead>
            <tbody>
              {procs.map((p) => (
                <tr key={p.pid}>
                  <td>{p.pid}</td>
                  <td>{p.name}</td>
                  <td>{p.state}</td>
                  <td>{p.threads}</td>
                  <td>{(p.rss_kb / 1024).toFixed(0)}</td>
                  <td>{p.cpu_percent.toFixed(1)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      )}
    </div>
  )
}