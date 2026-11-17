import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import Layout from "./Layout";
import Login from "./pages/Login";
import Dashboard from "./pages/Dashboard";
import Storage from "./pages/Storage";
import Shares from "./pages/Shares";
import Users from "./pages/Users";
import System from "./pages/System";
import Wizard from "./pages/Wizard";
import { useAuth } from "./auth";

function Guard({ children }: { children: any }) {
  const { user, loading } = useAuth();
  if (loading) return <div>Loading...</div>;
  if (!user) return <Navigate to="/login" />;
  return children;
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/wizard" element={<Wizard />} />
        <Route element={<Layout />}>
          <Route path="/" element={<Guard><Dashboard /></Guard>} />
          <Route path="/storage" element={<Guard><Storage /></Guard>} />
          <Route path="/shares" element={<Guard><Shares /></Guard>} />
          <Route path="/users" element={<Guard><Users /></Guard>} />
          <Route path="/system" element={<Guard><System /></Guard>} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
