import { useEffect, useState } from "react";
import { api } from "../api";

export default function System() {
  const [alerts, setAlerts] = useState<any[]>([]);
  useEffect(() => { api("/api/v1/alerts").then(setAlerts).catch(() => {}); }, []);
  async function dismiss(id: number) {
    await api(`/api/v1/alerts/${id}/dismiss`, { method: "POST" });
    setAlerts(await api("/api/v1/alerts"));
  }
  return (
    <div>
      <h2>System</h2>
      <div className="card">
        <h3>Alerts</h3>
        {alerts.length === 0 && <p className="muted">No alerts.</p>}
        {alerts.map((a: any) => (
          <div key={a.id} className="list-item">
            <div>
              <span style={{ color: a.dismissed ? "var(--muted)" : a.severity === "critical" ? "var(--danger)" : "var(--warn)" }}>
                {a.severity.toUpperCase()}: {a.message}
              </span>
              <p className="muted" style={{ margin: 0 }}>{a.created_at}</p>
            </div>
            {!a.dismissed && <button className="btn ghost" onClick={() => dismiss(a.id)}>Dismiss</button>}
          </div>
        ))}
      </div>
    </div>
  );
}
