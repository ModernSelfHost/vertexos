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
      <h3>Alerts</h3>
      {alerts.map((a: any) => (
        <div key={a.id} style={{ display: "flex", justifyContent: "space-between", borderBottom: "1px solid #eee", padding: 4 }}>
          <span style={{ color: a.dismissed ? "#999" : "inherit" }}>{a.severity}: {a.message}</span>
          {!a.dismissed && <button onClick={() => dismiss(a.id)}>Dismiss</button>}
        </div>
      ))}
    </div>
  );
}
