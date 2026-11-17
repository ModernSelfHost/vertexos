import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../auth";

export default function Login() {
  const { login } = useAuth();
  const nav = useNavigate();
  const [u, setU] = useState("");
  const [p, setP] = useState("");
  const [err, setErr] = useState("");
  async function submit(e: any) {
    e.preventDefault();
    try { await login(u, p); nav("/"); }
    catch (e: any) { setErr(e.message); }
  }
  return (
    <form onSubmit={submit} style={{ maxWidth: 320, margin: "2rem auto" }}>
      <h2>Login</h2>
      <input placeholder="username" value={u} onChange={e => setU(e.target.value)} style={{ display: "block", width: "100%", marginBottom: 8 }} />
      <input placeholder="password" type="password" value={p} onChange={e => setP(e.target.value)} style={{ display: "block", width: "100%", marginBottom: 8 }} />
      <button type="submit">Login</button>
      {err && <p style={{ color: "red" }}>{err}</p>}
    </form>
  );
}
