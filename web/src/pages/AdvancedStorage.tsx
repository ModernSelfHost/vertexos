import { useEffect, useState } from "react";
import { api } from "../api";

export default function AdvancedStorage() {
  const [pools, setPools] = useState<any[]>([]);
  const [datasets, setDatasets] = useState<any[]>([]);
  const [snapshots, setSnapshots] = useState<any[]>([]);
  const [schedules, setSchedules] = useState<any[]>([]);
  const [status, setStatus] = useState("");
  const [propTarget, setPropTarget] = useState("");
  const [propName, setPropName] = useState("");
  const [propValue, setPropValue] = useState("");
  const [snapDataset, setSnapDataset] = useState("");
  const [snapName, setSnapName] = useState("");
  const [schedPool, setSchedPool] = useState("");
  const [schedCron, setSchedCron] = useState("0 2 * * 0");
  const [msg, setMsg] = useState("");

  async function load() {
    try { setPools(await api("/api/v1/pools")); } catch {}
    try { setDatasets(await api("/api/v1/datasets")); } catch {}
    try { setSnapshots(await api("/api/v1/snapshots")); } catch {}
    try { setSchedules(await api("/api/v1/scrub-schedules")); } catch {}
  }
  useEffect(() => { load(); }, []);

  async function loadStatus(name: string) {
    try { setStatus(await api(`/api/v1/pools/${name}/status`)); } catch (e: any) { setStatus(e.message); }
  }
  async function setProperty(kind: "pool" | "dataset", target: string) {
    setMsg("");
    try {
      const path = kind === "pool" ? `/api/v1/pools/${target}/properties` : `/api/v1/datasets/${target}/properties`;
      await api(path, { method: "PUT", body: JSON.stringify({ property: propName, value: propValue }) });
      setMsg("Property update queued");
    } catch (e: any) { setMsg(e.message); }
  }
  async function createSnapshot() {
    setMsg("");
    try {
      await api("/api/v1/snapshots", { method: "POST", body: JSON.stringify({ dataset: snapDataset, name: snapName }) });
      setMsg("Snapshot creation queued");
      load();
    } catch (e: any) { setMsg(e.message); }
  }
  async function destroySnapshot(name: string) {
    setMsg("");
    try {
      await api(`/api/v1/snapshots/${encodeURIComponent(name)}`, { method: "DELETE" });
      setMsg("Snapshot destroy queued");
      load();
    } catch (e: any) { setMsg(e.message); }
  }
  async function addSchedule() {
    setMsg("");
    try {
      await api("/api/v1/scrub-schedules", { method: "POST", body: JSON.stringify({ pool: schedPool, cron: schedCron, enabled: true }) });
      setMsg("Schedule added");
      load();
    } catch (e: any) { setMsg(e.message); }
  }

  return (
    <div>
      <h2>Advanced Storage</h2>
      {msg && <p>{msg}</p>}

      <div className="card">
        <h3>Pools</h3>
        {pools.map((p: any) => (
          <div key={p.name} className="list-item">
            <div><strong>{p.name}</strong> <span className={`badge ${p.health}`}>{p.health}</span></div>
            <button className="btn ghost" onClick={() => loadStatus(p.name)}>Status</button>
          </div>
        ))}
        {status && <pre style={{ whiteSpace: "pre-wrap", maxHeight: 240, overflow: "auto", marginTop: 12 }}>{status}</pre>}
      </div>

      <div className="card">
        <h3>Datasets</h3>
        {datasets.map((d: any) => (
          <div key={d.name} className="list-item">
            <div><strong>{d.name}</strong></div>
            <div className="muted">used {d.used} · avail {d.avail}</div>
          </div>
        ))}
      </div>

      <div className="card">
        <h3>Property editor</h3>
        <div className="row">
          <select className="input" style={{ maxWidth: 200 }} value={propTarget} onChange={e => setPropTarget(e.target.value)}>
            <option value="">target</option>
            {pools.map((p: any) => <option key={p.name} value={p.name}>{p.name} (pool)</option>)}
            {datasets.map((d: any) => <option key={d.name} value={d.name}>{d.name} (dataset)</option>)}
          </select>
          <input className="input" placeholder="property" value={propName} onChange={e => setPropName(e.target.value)} />
          <input className="input" placeholder="value" value={propValue} onChange={e => setPropValue(e.target.value)} />
        </div>
        <div className="row" style={{ marginTop: 8 }}>
          <button className="btn" onClick={() => setProperty(pools.some(p => p.name === propTarget) ? "pool" : "dataset", propTarget)}>Apply</button>
        </div>
      </div>

      <div className="card">
        <h3>Snapshots</h3>
        <div className="row">
          <input className="input" placeholder="dataset" value={snapDataset} onChange={e => setSnapDataset(e.target.value)} />
          <input className="input" placeholder="snapshot name" value={snapName} onChange={e => setSnapName(e.target.value)} />
          <button className="btn" onClick={createSnapshot}>Create</button>
        </div>
        {snapshots.map((s: any) => (
          <div key={s.name} className="list-item">
            <div><strong>{s.name}</strong></div>
            <div className="muted">used {s.used} · refer {s.refer}</div>
            <button className="btn danger" onClick={() => destroySnapshot(s.name)}>Destroy</button>
          </div>
        ))}
      </div>

      <div className="card">
        <h3>Scrub schedules</h3>
        <div className="row">
          <input className="input" placeholder="pool" value={schedPool} onChange={e => setSchedPool(e.target.value)} />
          <input className="input" placeholder="cron" value={schedCron} onChange={e => setSchedCron(e.target.value)} />
          <button className="btn" onClick={addSchedule}>Add</button>
        </div>
        {schedules.map((s: any) => (
          <div key={s.id} className="list-item">
            <div><strong>{s.pool}</strong> <span className="muted">{s.cron}</span></div>
            <span className={`badge ${s.enabled ? "healthy" : ""}`}>{s.enabled ? "enabled" : "disabled"}</span>
          </div>
        ))}
      </div>
    </div>
  );
}