"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { API_BASE } from "@/lib/api";

export function AdminLogin() {
  const [username,setUsername]=useState("");const [password,setPassword]=useState("");const [error,setError]=useState("");const [busy,setBusy]=useState(false);const router=useRouter();
  async function submit(e:React.FormEvent<HTMLFormElement>){e.preventDefault();setBusy(true);setError("");try{const response=await fetch(`${API_BASE}/admin/auth/login`,{method:"POST",credentials:"include",headers:{"Content-Type":"application/json"},body:JSON.stringify({username,password})});const data=await response.json();if(!response.ok)throw new Error(data.error??"Tidak dapat masuk.");router.push("/admin/dashboard");}catch(e){setError(e instanceof Error?e.message:"Tidak dapat masuk.");}finally{setBusy(false);}}
  return <main className="app-page"><header className="app-topbar"><Link className="brand" href="/"><span className="brand-mark">w</span> whisper</Link><Link className="back-link" href="/">Kembali ke beranda</Link></header><section className="content-card"><div className="profile-avatar">w</div><span className="eyebrow">RUANG PRIBADIMU</span><h1>Selamat datang kembali</h1><p>Masuk untuk melihat dan membalas pesan anonim.</p><form onSubmit={submit}><label className="field-label" htmlFor="username">Username</label><input className="text-input" id="username" autoComplete="username" required minLength={3} maxLength={50} value={username} onChange={e=>setUsername(e.target.value)} placeholder="username"/><label className="field-label" htmlFor="password">Password</label><input className="text-input" id="password" type="password" autoComplete="current-password" required value={password} onChange={e=>setPassword(e.target.value)} placeholder="Password akun Whisper"/>{error&&<p className="form-error">{error}</p>}<button className="button button-primary full-width" disabled={busy}>{busy?"Memeriksa…":"Masuk"}<span>→</span></button></form><p className="muted-note">Akun dibuat oleh admin sistem. Tidak ada pendaftaran publik.</p></section></main>;
}
