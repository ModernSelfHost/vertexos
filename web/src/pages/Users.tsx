import { useEffect, useState } from "react";
import { api } from "../api";

export default function Users() {
  const [users, setUsers] = useState<any[]>([]);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState("user");

  async function load() { setUsers(await api("/api/v1/users")); }
  useEffect(() => { load(); }, []);

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
      <div style={{ display: "flex", gap: 8, marginBottom: 12 }}>
        <input placeholder="username" value={username} onChange={e => setUsername(e.target.value)} />
        <input placeholder="password" type="password" value={password} onChange={e => setPassword(e.target.value)} />
        <select value={role} onChange={e => setRole(e.target.value)}>
          <option value="user">user</option>
          <option value="admin">admin</option>
          <option value="guest">guest</option>
        </select>
        <button onClick={create}>Create</button>
      </div>
      {users.map((u: any) => (
        <div key={u.id} style={{ display: "flex", justifyContent: "space-between", borderBottom: "1px solid #eee", padding: 4 }}>
          <span>{u.username} ({u.role})</span>
          <button onClick={() => remove(u.id)}>Delete</button>
        </div>
      ))}
    </div>
  );
}
