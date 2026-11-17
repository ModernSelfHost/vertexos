import { useEffect, useState } from "react";
import { api } from "../api";

export default function Dashboard() {
  const [drives, setDrives] = useState<any[]>([]);
  const [pools, setPools] = useState<any[]>([]);
  const [alerts, setAlerts] = useState<any[]>([]);
  useEffect(() => {
    api("/api/v1/drives").then(setDrives).catch(() => {});
    api("/api/v1/pools").then(setPools).catch(() => {});
    api("/api/v1/alerts").then(setAlerts).catch(() => {});
  }, []);
  return (
    <div>
      <h2>Dashboard</h2>
      <section>
        <h3>Storage pools</h3>
        {pools.length === 0 && <p>No pools.</p>}
        {pools.map((p: any) => (
          <div key={p.name} style={{ border: "1px solid #ddd", padding: 8, marginBottom: 8, borderRadius: 4 }}>
            <strong>{p.name}</strong> — {p.health} — size {p.size} alloc {p.alloc} free {p.free}
          </div>
        ))}
      </section>
      <section>
        <h3>Drives</h3>
        {drives.map((d: any) => (
          <div key={d.device} style={{ display: "flex", justifyContent: "space-between", borderBottom: "1px solid #eee", padding: 4 }}>
            <span>{d.name} ({d.model})</span>
            <span>{(d.size_bytes / 1e9).toFixed(1)} GB — {d.health}</span>
          </div>
        ))}
      </section>
      <section>
        <h3>Alerts</h3>
        {alerts.filter((a: any) => !a.dismissed).map((a: any) => (
          <div key={a.id} style={{ color: a.severity === "critical" ? "red" : "orange" }}>{a.message}</div>
        ))}
        {alerts.filter((a: any) => !a.dismissed).length === 0 && <p>No alerts.</p>}
      </section>
    </div>
  );
}
