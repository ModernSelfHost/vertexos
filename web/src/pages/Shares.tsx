import { useEffect, useState } from "react";
import { api } from "../api";

export default function Shares() {
  const [shares, setShares] = useState<any[]>([]);
  const [name, setName] = useState("");
  const [guest, setGuest] = useState(false);

  async function load() { setShares(await api("/api/v1/shares")); }
  useEffect(() => { load().catch(() => {}); }, []);

  async function create() {
    await api("/api/v1/shares", { method: "POST", body: JSON.stringify({ name, guest_accessible: guest }) });
    setName(""); setGuest(false);
    load();
  }
  async function remove(id: number) {
    await api(`/api/v1/shares/${id}`, { method: "DELETE" });
    load();
  }

  return (
    <div>
      <h2>Shares</h2>
      <div className="card">
        <h3>Create share</h3>
        <div className="row">
          <input className="input" placeholder="share name" value={name} onChange={e => setName(e.target.value)} />
          <label className="row"><input type="checkbox" checked={guest} onChange={e => setGuest(e.target.checked)} /> Guest access</label>
          <button className="btn" onClick={create} disabled={!name}>Create</button>
        </div>
      </div>
      <div className="card">
        <h3>Existing shares</h3>
        {shares.length === 0 && <p className="muted">No shares yet.</p>}
        {shares.map((s: any) => (
          <div key={s.id} className="list-item">
            <div>
              <strong>{s.name}</strong>
              <p className="muted" style={{ margin: 0 }}>{s.dataset}</p>
            </div>
            <div className="row">
              {s.guest_accessible ? <span className="badge healthy">guest</span> : <span className="badge">private</span>}
              <button className="btn danger" onClick={() => remove(s.id)}>Delete</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
