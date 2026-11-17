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
    // create pool from proposal using all drives
    try {
      await api("/api/v1/pools/from-proposal", { method: "POST", body: JSON.stringify({ name: "tank", drives: [] }) });
      setMsg("Pool creation started.");
    } catch (e: any) { setMsg(e.message); }
    nav("/");
  }

  return (
    <div style={{ maxWidth: 480, margin: "2rem auto" }}>
      <h2>Setup Wizard</h2>
      {step === 0 && (
        <>
          <p>Welcome to Vertex OS.</p>
          <button onClick={() => setStep(1)}>Start</button>
        </>
      )}
      {step === 1 && (
        <>
          <h3>Name your server</h3>
          <input value={name} onChange={e => setName(e.target.value)} />
          <button onClick={() => setStep(2)}>Next</button>
        </>
      )}
      {step === 2 && (
        <>
          <h3>Create admin account</h3>
          <input placeholder="username" value={adminUser} onChange={e => setAdminUser(e.target.value)} />
          <input placeholder="password" type="password" value={adminPass} onChange={e => setAdminPass(e.target.value)} />
          <button onClick={() => setStep(3)}>Next</button>
        </>
      )}
      {step === 3 && (
        <>
          <h3>Create storage pool</h3>
          <p>Vertex will propose a layout from detected drives.</p>
          <button onClick={finish}>Create pool & finish</button>
          {msg && <p>{msg}</p>}
        </>
      )}
    </div>
  );
}
