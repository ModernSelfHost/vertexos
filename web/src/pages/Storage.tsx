import { useEffect, useState } from "react";
import { api } from "../api";

export default function Storage() {
  const [drives, setDrives] = useState<any[]>([]);
  const [pools, setPools] = useState<any[]>([]);
  const [proposal, setProposal] = useState<any>(null);
  const [selected, setSelected] = useState<string[]>([]);
  const [poolName, setPoolName] = useState("tank");
  const [creating, setCreating] = useState(false);

  async function load() {
    setDrives(await api("/api/v1/drives"));
    setPools(await api("/api/v1/pools"));
  }
  useEffect(() => { load().catch(() => {}); }, []);

  async function propose() {
    const res = await api("/api/v1/pools/propose", { method: "POST", body: JSON.stringify({ drives: selected }) });
    setProposal(res);
  }
  async function create() {
    setCreating(true);
    try {
      await api("/api/v1/pools/from-proposal", { method: "POST", body: JSON.stringify({ name: poolName, drives: selected }) });
      await load();
      setProposal(null);
    } finally { setCreating(false); }
  }

  return (
    <div>
      <h2>Storage</h2>
      <div className="card">
        <h3>Drives</h3>
        {drives.map((d: any) => (
          <label key={d.device} className="list-item" style={{ cursor: "pointer" }}>
            <div className="row">
              <input type="checkbox" checked={selected.includes(d.device)} onChange={e => {
                setSelected(e.target.checked ? [...selected, d.device] : selected.filter(x => x !== d.device));
              }} />
              <div>
                <strong>{d.name}</strong> <span className="muted">{d.model}</span>
              </div>
            </div>
            <div className="row">
              <span className="muted">{(d.size_bytes / 1e9).toFixed(1)} GB</span>
              <span className={`badge ${d.health}`}>{d.health}</span>
            </div>
          </label>
        ))}
        <div className="row" style={{ marginTop: 12 }}>
          <button className="btn" onClick={propose} disabled={selected.length === 0}>Propose layout</button>
        </div>
      </div>

      {proposal && (
        <div className="card">
          <h3>Proposed layout</h3>
          <p><strong style={{ color: "var(--accent-2)" }}>{proposal.layout}</strong></p>
          <p>{proposal.explanation}</p>
          <p className="muted">Usable: {(proposal.usable_bytes / 1e9).toFixed(1)} GB · survives {proposal.fault_tolerance} failure(s)</p>
          <div className="row">
            <input className="input" style={{ maxWidth: 200 }} value={poolName} onChange={e => setPoolName(e.target.value)} placeholder="pool name" />
            <button className="btn" onClick={create} disabled={creating}>{creating ? "Creating..." : "Create pool"}</button>
          </div>
        </div>
      )}

      <div className="card">
        <h3>Pools</h3>
        {pools.length === 0 && <p className="muted">No pools.</p>}
        {pools.map((p: any) => (
          <div key={p.name} className="list-item">
            <strong>{p.name}</strong>
            <span className={`badge ${p.health}`}>{p.health}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
