import { useEffect, useState } from "react";
import { api } from "../api";

export default function Shares() {
  const [shares, setShares] = useState<any[]>([]);
  const [name, setName] = useState("");
  const [guest, setGuest] = useState(false);

  async function load() { setShares(await api("/api/v1/shares")); }
  useEffect(() => { load(); }, []);

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
      <div style={{ display: "flex", gap: 8, marginBottom: 12 }}>
        <input placeholder="share name" value={name} onChange={e => setName(e.target.value)} />
        <label><input type="checkbox" checked={guest} onChange={e => setGuest(e.target.checked)} /> guest</label>
        <button onClick={create}>Create</button>
      </div>
      {shares.map((s: any) => (
        <div key={s.id} style={{ border: "1px solid #ddd", padding: 8, marginBottom: 8, borderRadius: 4, display: "flex", justifyContent: "space-between" }}>
          <div>
            <strong>{s.name}</strong> — {s.dataset} — quota {s.quota_bytes} — guest {String(s.guest_accessible)}
          </div>
          <button onClick={() => remove(s.id)}>Delete</button>
        </div>
      ))}
    </div>
  );
}
