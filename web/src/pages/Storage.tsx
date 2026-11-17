import { useEffect, useState } from "react";
import { api } from "../api";

export default function Storage() {
  const [drives, setDrives] = useState<any[]>([]);
  const [pools, setPools] = useState<any[]>([]);
  const [proposal, setProposal] = useState<any>(null);
  const [selected, setSelected] = useState<string[]>([]);
  const [poolName, setPoolName] = useState("tank");

  useEffect(() => {
    api("/api/v1/drives").then(setDrives).catch(() => {});
    api("/api/v1/pools").then(setPools).catch(() => {});
  }, []);

  async function propose() {
    const res = await api("/api/v1/pools/propose", { method: "POST", body: JSON.stringify({ drives: selected }) });
    setProposal(res);
  }
  async function create() {
    await api("/api/v1/pools/from-proposal", { method: "POST", body: JSON.stringify({ name: poolName, drives: selected }) });
    const ps = await api("/api/v1/pools");
    setPools(ps);
  }

  return (
    <div>
      <h2>Storage</h2>
      <h3>Drives</h3>
      {drives.map((d: any) => (
        <label key={d.device} style={{ display: "block" }}>
          <input type="checkbox" checked={selected.includes(d.device)} onChange={e => {
            setSelected(e.target.checked ? [...selected, d.device] : selected.filter(x => x !== d.device));
          }} />
          {d.name} — {d.model} — {(d.size_bytes / 1e9).toFixed(1)} GB — {d.health}
        </label>
      ))}
      <button onClick={propose} style={{ marginTop: 8 }}>Propose layout</button>
      {proposal && (
        <div style={{ border: "1px solid #ddd", padding: 8, marginTop: 8 }}>
          <p><strong>{proposal.layout}</strong> — {proposal.explanation}</p>
          <p>Usable: {(proposal.usable_bytes / 1e9).toFixed(1)} GB · survives {proposal.fault_tolerance} failure(s)</p>
          <input value={poolName} onChange={e => setPoolName(e.target.value)} placeholder="pool name" />
          <button onClick={create}>Create pool</button>
        </div>
      )}
      <h3>Pools</h3>
      {pools.map((p: any) => <div key={p.name}>{p.name} — {p.health}</div>)}
    </div>
  );
}
