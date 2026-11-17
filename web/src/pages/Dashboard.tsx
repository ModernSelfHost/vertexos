import { useEffect, useState } from "react";
import { api } from "../api";

export default function Dashboard() {
  const [drives, setDrives] = useState<any[]>([]);
  const [pools, setPools] = useState<any[]>([]);
  const [alerts, setAlerts] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    Promise.all([
      api("/api/v1/drives").then(setDrives).catch(() => {}),
      api("/api/v1/pools").then(setPools).catch(() => {}),
      api("/api/v1/alerts").then(setAlerts).catch(() => {}),
    ]).finally(() => setLoading(false));
  }, []);
  const activeAlerts = alerts.filter((a: any) => !a.dismissed);
  return (
    <div>
      <h2>Dashboard</h2>
      {loading && <div className="spinner" />}
      <div className="grid grid-3">
        <div className="card">
          <h3>Pools</h3>
          <div className="stat">{pools.length}</div>
          <p className="muted">configured pools</p>
        </div>
        <div className="card">
          <h3>Drives</h3>
          <div className="stat">{drives.length}</div>
          <p className="muted">detected drives</p>
        </div>
        <div className="card">
          <h3>Alerts</h3>
          <div className="stat" style={{ color: activeAlerts.length ? "var(--danger)" : "var(--success)" }}>{activeAlerts.length}</div>
          <p className="muted">active alerts</p>
        </div>
      </div>

      <div className="card">
        <h3>Storage pools</h3>
        {pools.length === 0 && <p className="muted">No pools configured.</p>}
        {pools.map((p: any) => (
          <div key={p.name} className="list-item">
            <div>
              <strong>{p.name}</strong>
              <span className={`badge ${p.health}`} style={{ marginLeft: 8 }}>{p.health}</span>
            </div>
            <div className="muted">size {p.size} · alloc {p.alloc} · free {p.free}</div>
          </div>
        ))}
      </div>

      <div className="card">
        <h3>Drives</h3>
        {drives.map((d: any) => (
          <div key={d.device} className="list-item">
            <div>
              <strong>{d.name}</strong>
              <span className="muted" style={{ marginLeft: 8 }}>{d.model}</span>
            </div>
            <div>
              <span className="muted">{(d.size_bytes / 1e9).toFixed(1)} GB</span>
              <span className={`badge ${d.health}`} style={{ marginLeft: 8 }}>{d.health}</span>
            </div>
          </div>
        ))}
      </div>

      <div className="card">
        <h3>Alerts</h3>
        {activeAlerts.length === 0 && <p className="muted">No active alerts.</p>}
        {activeAlerts.map((a: any) => (
          <div key={a.id} className="list-item">
            <span style={{ color: a.severity === "critical" ? "var(--danger)" : "var(--warn)" }}>{a.severity.toUpperCase()}: {a.message}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
