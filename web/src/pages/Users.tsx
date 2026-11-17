import { useEffect, useState } from "react";
import { api } from "../api";

export default function Users() {
  const [users, setUsers] = useState<any[]>([]);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState("user");

  async function load() { setUsers(await api("/api/v1/users")); }
  useEffect(() => { load().catch(() => {}); }, []);

  async function create() {
    await api("/api/v1/users", { method: "POST", body: JSON.stringify({ username, password, role }) });
    setUsername(""); setPassword("");
    load();
  }
  async function remove(id: number) {
    await api(`/api/v1/users/${id}`, { method: "DELETE" });
    load();
  }

  return (
    <div>
      <h2>Users</h2>
      <div className="card">
        <h3>Create user</h3>
        <div className="row">
          <input className="input" placeholder="username" value={username} onChange={e => setUsername(e.target.value)} />
          <input className="input" placeholder="password" type="password" value={password} onChange={e => setPassword(e.target.value)} />
          <select className="input" style={{ maxWidth: 120 }} value={role} onChange={e => setRole(e.target.value)}>
            <option value="user">user</option>
            <option value="admin">admin</option>
            <option value="guest">guest</option>
          </select>
          <button className="btn" onClick={create} disabled={!username || !password}>Create</button>
        </div>
      </div>
      <div className="card">
        <h3>Accounts</h3>
        {users.length === 0 && <p className="muted">No users.</p>}
        {users.map((u: any) => (
          <div key={u.id} className="list-item">
            <div>
              <strong>{u.username}</strong>
              <span className={`badge ${u.role === "admin" ? "warning" : "healthy"}`} style={{ marginLeft: 8 }}>{u.role}</span>
            </div>
            <button className="btn danger" onClick={() => remove(u.id)}>Delete</button>
          </div>
        ))}
      </div>
    </div>
  );
}
