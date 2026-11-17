import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api";

export default function Wizard() {
  const nav = useNavigate();
  const [step, setStep] = useState(0);
  const [name, setName] = useState("vertex");
  const [adminUser, setAdminUser] = useState("admin");
  const [adminPass, setAdminPass] = useState("");
  const [msg, setMsg] = useState("");

  async function finish() {
    try {
      await api("/api/v1/pools/from-proposal", { method: "POST", body: JSON.stringify({ name: "tank", drives: [] }) });
      setMsg("Pool creation started.");
    } catch (e: any) { setMsg(e.message); }
    nav("/");
  }

  return (
    <div className="card" style={{ maxWidth: 520, margin: "2rem auto" }}>
      <h2 style={{ marginTop: 0 }}>Setup Wizard</h2>
      {step === 0 && (
        <>
          <p>Welcome to Vertex OS. Let's get your NAS running.</p>
          <button className="btn" onClick={() => setStep(1)}>Start</button>
        </>
      )}
      {step === 1 && (
        <>
          <h3>Name your server</h3>
          <input className="input" value={name} onChange={e => setName(e.target.value)} />
          <div className="row" style={{ marginTop: 12 }}>
            <button className="btn" onClick={() => setStep(2)}>Next</button>
          </div>
        </>
      )}
      {step === 2 && (
        <>
          <h3>Create admin account</h3>
          <input className="input" placeholder="username" value={adminUser} onChange={e => setAdminUser(e.target.value)} />
          <input className="input" placeholder="password" type="password" value={adminPass} onChange={e => setAdminPass(e.target.value)} style={{ marginTop: 8 }} />
          <div className="row" style={{ marginTop: 12 }}>
            <button className="btn" onClick={() => setStep(3)}>Next</button>
          </div>
        </>
      )}
      {step === 3 && (
        <>
          <h3>Create storage pool</h3>
          <p className="muted">Vertex will propose a layout from detected drives.</p>
          <button className="btn" onClick={finish}>Create pool & finish</button>
          {msg && <p>{msg}</p>}
        </>
      )}
    </div>
  );
}
