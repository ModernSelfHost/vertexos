import { NavLink, Outlet } from "react-router-dom";
import { useAuth } from "./auth";
import Metaballs from "./Metaballs";

const links = [
  { to: "/", label: "Dashboard", icon: "dashboard" },
  { to: "/storage", label: "Storage", icon: "storage" },
  { to: "/shares", label: "Shares", icon: "folder_shared" },
  { to: "/users", label: "Users", icon: "group", admin: true },
  { to: "/system", label: "System", icon: "settings", admin: true },
];

export default function Layout() {
  const { user, logout } = useAuth();
  return (
    <div className="layout">
      <Metaballs />
      <aside className="sidebar">
        <div className="brand">Vertex</div>
        <nav>
          {links.map(l => {
            if (l.admin && user?.role !== "admin") return null;
            return (
              <NavLink key={l.to} to={l.to} end={l.to === "/"} className={({ isActive }) => "side-link" + (isActive ? " active" : "")}>
                <span className="material-symbols-outlined">{l.icon}</span>
                <span>{l.label}</span>
              </NavLink>
            );
          })}
        </nav>
        <div className="sidebar-footer">
          {user ? (
            <>
              <div className="muted" style={{ marginBottom: 8 }}>{user.username}</div>
              <button className="btn ghost" style={{ width: "100%" }} onClick={logout}>Logout</button>
            </>
          ) : (
            <NavLink to="/login" className="btn" style={{ display: "block", textAlign: "center" }}>Login</NavLink>
          )}
        </div>
      </aside>
      <main className="main">
        <Outlet />
      </main>
    </div>
  );
}
