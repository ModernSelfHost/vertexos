import { Link, Outlet } from "react-router-dom";
import { useAuth } from "./auth";

export default function Layout() {
  const { user, logout } = useAuth();
  return (
    <div>
      <header className="app-header">
        <div className="brand">Vertex OS</div>
        <nav className="nav">
          <Link to="/">Dashboard</Link>
          <Link to="/storage">Storage</Link>
          <Link to="/shares">Shares</Link>
          {user?.role === "admin" && <Link to="/users">Users</Link>}
          {user?.role === "admin" && <Link to="/system">System</Link>}
        </nav>
        <div className="row">
          {user ? (
            <>
              <span className="muted">{user.username}</span>
              <button className="btn ghost" onClick={logout}>Logout</button>
            </>
          ) : (
            <Link to="/login" className="btn">Login</Link>
          )}
        </div>
      </header>
      <main>
        <Outlet />
      </main>
    </div>
  );
}
