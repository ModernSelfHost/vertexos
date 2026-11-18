import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api";
import { useAuth } from "../auth";

type Drive = { device: string; name: string; model: string; size_bytes: number; health: string; stable_id?: string };

export default function Wizard() {
  const nav = useNavigate();
  const { login } = useAuth();
  const [step, setStep] = useState(0);
  const [serverName, setServerName] = useState("vertex");
  const [adminUser, setAdminUser] = useState("admin");
  const [adminPass, setAdminPass] = useState("");
  const [drives, setDrives] = useState<Drive[]>([]);
  const [selected, setSelected] = useState<string[]>([]);
  const [proposal, setProposal] = useState<any>(null);
  const [poolName, setPoolName] = useState("tank");
  const [shareName, setShareName] = useState("shared");
  const [guest, setGuest] = useState(false);
  const [msg, setMsg] = useState("");
  const [busy, setBusy] = useState(false);
  const [needsSetup, setNeedsSetup] = useState<boolean | null>(null);

  useEffect(() => {
    api("/api/v1/setup/status").then((r: any) => {
      setNeedsSetup(!!r.needs_setup);
      if (!r.needs_setup) nav("/");
    }).catch(() => setNeedsSetup(true));
  }, [nav]);

  useEffect(() => {
    if (step === 3) {
      api("/api/v1/drives").then((d: any) => setDrives(d || [])).catch(() => {});
    }
  }, [step]);

  async function waitForJob(id: string) {
    for (;;) {
      const j: any = await api(`/api/v1/jobs/${id}`);
      if (j.Status === "succeeded") return j;
      if (j.Status === "failed" || j.Status === "cancelled") throw new Error(j.Message || "job " + j.Status);
      await new Promise((r) => setTimeout(r, 1000));
    }
  }

  async function saveServerName() {
    setBusy(true); setMsg("");
    try {
      await api("/api/v1/setup/server-name", { method: "POST", body: JSON.stringify({ name: serverName }) });
      setStep(2);
    } catch (e: any) { setMsg(e.message || "Failed"); }
    finally { setBusy(false); }
  }

  async function createAdmin() {
    setBusy(true); setMsg("");
    try {
      await api("/api/v1/setup/admin", { method: "POST", body: JSON.stringify({ username: adminUser, password: adminPass }) });
      await login(adminUser, adminPass);
      setStep(3);
    } catch (e: any) { setMsg(e.message || "Failed"); }
    finally { setBusy(false); }
  }

  async function propose() {
    setBusy(true); setMsg("");
    try {
      const res = await api("/api/v1/pools/propose", { method: "POST", body: JSON.stringify({ drives: selected }) });
      setProposal(res);
    } catch (e: any) { setMsg(e.message || "Failed"); }
    finally { setBusy(false); }
  }

  async function createPool() {
    setBusy(true); setMsg("");
    try {
      if (selected.length === 0) throw new Error("Select at least one drive");
      const job: any = await api("/api/v1/pools/from-proposal", { method: "POST", body: JSON.stringify({ name: poolName, drives: selected }) });
      await waitForJob(job.ID);
      setProposal(null);
      setStep(4);
    } catch (e: any) { setMsg(e.message || "Failed"); }
    finally { setBusy(false); }
  }

  async function createShare() {
    setBusy(true); setMsg("");
    try {
      const job: any = await api("/api/v1/shares", { method: "POST", body: JSON.stringify({ name: shareName, guest_accessible: guest }) });
      await waitForJob(job.ID);
      setStep(5);
    } catch (e: any) { setMsg(e.message || "Failed"); }
    finally { setBusy(false); }
  }

  if (needsSetup === null) return <div className="card" style={{ maxWidth: 520, margin: "2rem auto" }}><p>Loading...</p></div>;
  if (!needsSetup) return null;

  return (
    <div className="card" style={{ maxWidth: 560, margin: "2rem auto" }}>
      <h2 style={{ marginTop: 0 }}>Setup Wizard</h2>
      {msg && <p style={{ color: "var(--danger)" }}>{msg}</p>}

      {step === 0 && (
        <>
          <p>Welcome to Vertex OS. Let's get your NAS running.</p>
          <button className="btn" onClick={() => setStep(1)}>Start</button>
        </>
      )}

      {step === 1 && (
        <>
          <h3>Name your server</h3>
          <input className="input" value={serverName} onChange={e => setServerName(e.target.value)} />
          <div className="row" style={{ marginTop: 12 }}>
            <button className="btn" onClick={saveServerName} disabled={busy || !serverName}>{busy ? "Saving..." : "Next"}</button>
          </div>
        </>
      )}

      {step === 2 && (
        <>
          <h3>Create admin account</h3>
          <input className="input" placeholder="username" value={adminUser} onChange={e => setAdminUser(e.target.value)} />
          <input className="input" placeholder="password" type="password" value={adminPass} onChange={e => setAdminPass(e.target.value)} style={{ marginTop: 8 }} />
          <div className="row" style={{ marginTop: 12 }}>
            <button className="btn" onClick={createAdmin} disabled={busy || !adminUser || !adminPass}>{busy ? "Creating..." : "Next"}</button>
          </div>
        </>
      )}

      {step === 3 && (
        <>
          <h3>Create storage pool</h3>
          <p className="muted">Vertex will propose a layout from detected drives.</p>
          <div className="card" style={{ padding: 12 }}>
            <h4 style={{ marginTop: 0 }}>Drives</h4>
            {drives.length === 0 && <p className="muted">No drives detected.</p>}
            {drives.filter(d => d.device && !d.device.startsWith("loop") && !d.name?.startsWith("zram")).map((d) => (
              <label key={d.device} className="list-item" style={{ cursor: "pointer" }}>
                <div className="row">
                  <input type="checkbox" checked={selected.includes(d.device)} onChange={e => {
                    setSelected(e.target.checked ? [...selected, d.device] : selected.filter(x => x !== d.device));
                  }} />
                  <div>
                    <strong>{d.name}</strong> <span className="muted">{d.model}</span>
                  </div>
                </div>
                <div className="row">
                  <span className="muted">{(d.size_bytes / 1e9).toFixed(1)} GB</span>
                  <span className={`badge ${d.health}`}>{d.health}</span>
                </div>
              </label>
            ))}
            <div className="row" style={{ marginTop: 12 }}>
              <button className="btn" onClick={propose} disabled={busy || selected.length === 0}>{busy ? "..." : "Propose layout"}</button>
            </div>
          </div>

          {proposal && (
            <div className="card" style={{ padding: 12, marginTop: 12 }}>
              <h4 style={{ marginTop: 0 }}>Proposed layout</h4>
              <p><strong style={{ color: "var(--accent-2)" }}>{proposal.layout}</strong></p>
              <p>{proposal.explanation}</p>
              <p className="muted">Usable: {(proposal.usable_bytes / 1e9).toFixed(1)} GB · survives {proposal.fault_tolerance} failure(s)</p>
              <div className="row">
                <input className="input" style={{ maxWidth: 200 }} value={poolName} onChange={e => setPoolName(e.target.value)} placeholder="pool name" />
                <button className="btn" onClick={createPool} disabled={busy}>{busy ? "Creating..." : "Create pool"}</button>
              </div>
            </div>
          )}
        </>
      )}

      {step === 4 && (
        <>
          <h3>Create your first share</h3>
          <input className="input" placeholder="share name" value={shareName} onChange={e => setShareName(e.target.value)} />
          <label className="row" style={{ marginTop: 8 }}><input type="checkbox" checked={guest} onChange={e => setGuest(e.target.checked)} /> Guest access</label>
          <div className="row" style={{ marginTop: 12 }}>
            <button className="btn" onClick={createShare} disabled={busy || !shareName}>{busy ? "Creating..." : "Create share"}</button>
            <button className="btn ghost" onClick={() => setStep(5)} disabled={busy}>Skip</button>
          </div>
        </>
      )}

      {step === 5 && (
        <>
          <h3>You're all set!</h3>
          <p>Your server <strong>{serverName}</strong> is ready.</p>
          <p className="muted">Pool <strong>{poolName}</strong> and share <strong>{shareName}</strong> created.</p>
          <button className="btn" onClick={() => nav("/")}>Go to dashboard</button>
        </>
      )}
    </div>
  );
}
