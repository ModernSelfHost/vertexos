import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../auth";

export default function Login() {
  const { login } = useAuth();
  const nav = useNavigate();
  const [u, setU] = useState("");
  const [p, setP] = useState("");
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);
  async function submit(e: any) {
    e.preventDefault();
    setLoading(true);
    setErr("");
    try { await login(u, p); nav("/"); }
    catch (e: any) { setErr(e.message || "Login failed"); }
    finally { setLoading(false); }
  }
  return (
    <div className="card login-box">
      <h2 style={{ textAlign: "center", marginTop: 0 }}>Vertex OS</h2>
      <p className="muted" style={{ textAlign: "center" }}>Sign in to continue</p>
      <form onSubmit={submit}>
        <input className="input" placeholder="username" value={u} onChange={e => setU(e.target.value)} />
        <input className="input" placeholder="password" type="password" value={p} onChange={e => setP(e.target.value)} />
        <button className="btn" type="submit" style={{ width: "100%" }} disabled={loading}>
          {loading ? "Signing in..." : "Login"}
        </button>
        {err && <p style={{ color: "var(--danger)", marginTop: 8 }}>{err}</p>}
      </form>
    </div>
  );
}
