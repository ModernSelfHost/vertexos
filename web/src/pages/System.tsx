import { useEffect, useState } from "react";
import { api } from "../api";

export default function System() {
  const [alerts, setAlerts] = useState<any[]>([]);
  const [logs, setLogs] = useState("");
  const [updates, setUpdates] = useState<string[]>([]);
  useEffect(() => {
    api("/api/v1/alerts").then(setAlerts).catch(() => {});
  }, []);
  async function dismiss(id: number) {
    await api(`/api/v1/alerts/${id}/dismiss`, { method: "POST" });
    setAlerts(await api("/api/v1/alerts"));
  }
  async function loadLogs() {
    try { setLogs(await api("/api/v1/logs", { method: "GET" })); } catch (e: any) { setLogs(e.message); }
  }
  async function checkUpdates() {
    try { const r: any = await api("/api/v1/updates"); setUpdates(r.updates || []); } catch (e: any) { setUpdates([e.message]); }
  }
  async function applyUpdates() {
    try { await api("/api/v1/updates/apply", { method: "POST" }); alert("Update started"); } catch (e: any) { alert(e.message); }
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

      <div className="card">
        <h3>Logs</h3>
        <button className="btn" onClick={loadLogs}>Load recent logs</button>
        {logs && <pre style={{ whiteSpace: "pre-wrap", maxHeight: 300, overflow: "auto", marginTop: 12 }}>{logs}</pre>}
      </div>

      <div className="card">
        <h3>OS updates</h3>
        <div className="row">
          <button className="btn" onClick={checkUpdates}>Check for updates</button>
          <button className="btn" onClick={applyUpdates}>Apply updates</button>
        </div>
        {updates.length > 0 && (
          <pre style={{ whiteSpace: "pre-wrap", maxHeight: 200, overflow: "auto", marginTop: 12 }}>{updates.join("\n")}</pre>
        )}
      </div>
    </div>
  );
}