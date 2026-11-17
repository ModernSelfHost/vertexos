import { useEffect, useState } from "react";
import { api } from "../api";

export default function Apps() {
  const [apps, setApps] = useState<any[]>([]);
  const [filter, setFilter] = useState("");
  useEffect(() => { api("/api/v1/apps").then(setApps).catch(() => {}); }, []);

  async function install(id: string) {
    try { await api(`/api/v1/apps/${id}/install`, { method: "POST" }); alert("Install job started (runtime not configured in v1)."); }
    catch (e: any) { alert(e.message); }
  }

  const filtered = apps.filter((a: any) => !filter || a.name.toLowerCase().includes(filter.toLowerCase()) || a.category.includes(filter.toLowerCase()));

  return (
    <div>
      <h2>App Store</h2>
      <p className="muted">Curated catalog from openappstore. Runtime integration is post-v1.</p>
      <input className="input" placeholder="Search apps..." value={filter} onChange={e => setFilter(e.target.value)} style={{ maxWidth: 320, marginBottom: 12 }} />
      <div className="grid grid-3">
        {filtered.map((a: any) => (
          <div key={a.id} className="card">
            <div className="row" style={{ justifyContent: "space-between" }}>
              <strong>{a.name}</strong>
              <span className="badge">{a.category}</span>
            </div>
            <p className="muted" style={{ margin: "8px 0" }}>{a.description}</p>
            <p className="muted" style={{ margin: 0 }}>v{a.version} · {a.runtime}</p>
            <div className="row" style={{ marginTop: 12 }}>
              <button className="btn" onClick={() => install(a.id)}>Install</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
