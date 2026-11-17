import { Link, Outlet } from "react-router-dom";
import { useAuth } from "./auth";

export default function Layout() {
  const { user, logout } = useAuth();
  return (
    <div style={{ fontFamily: "sans-serif", margin: 0 }}>
      <header style={{ background: "#1f2937", color: "#fff", padding: "0.75rem 1rem", display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <strong>Vertex OS</strong>
        <nav style={{ display: "flex", gap: "1rem" }}>
          <Link to="/" style={{ color: "#fff" }}>Dashboard</Link>
          <Link to="/storage" style={{ color: "#fff" }}>Storage</Link>
          <Link to="/shares" style={{ color: "#fff" }}>Shares</Link>
          {user?.role === "admin" && <Link to="/users" style={{ color: "#fff" }}>Users</Link>}
          {user?.role === "admin" && <Link to="/system" style={{ color: "#fff" }}>System</Link>}
        </nav>
        <div>
          {user ? <><span style={{ marginRight: "1rem" }}>{user.username}</span><button onClick={logout}>Logout</button></> : <Link to="/login" style={{ color: "#fff" }}>Login</Link>}
        </div>
      </header>
      <main style={{ padding: "1rem", maxWidth: 1100, margin: "0 auto" }}>
        <Outlet />
      </main>
    </div>
  );
}
