"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { API_BASE } from "@/lib/api";

export function PublicProfile({ slug }: { slug: string }) {
  const [content, setContent] = useState("");
  const [error, setError] = useState("");
  const [sending, setSending] = useState(false);
  const router = useRouter();
  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault(); setError(""); setSending(true);
    try {
      const response = await fetch(`${API_BASE}/public/${encodeURIComponent(slug)}/conversations`, { method: "POST", headers: { "Content-Type": "application/json" } });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error ?? "Belum bisa membuat percakapan. Coba lagi sebentar.");
      const message = await fetch(`${API_BASE}/conversations/${data.conversation.id}/messages`, { method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${data.access_token}` }, body: JSON.stringify({ content }) });
      if (!message.ok) { const result = await message.json().catch(()=>({})); throw new Error(`Percakapan dibuat, tetapi pesan belum terkirim (${message.status}): ${result.error ?? "coba buka link percakapannya lagi"}.`); }
      router.push(data.conversation_url);
    } catch (e) { setError(e instanceof Error ? e.message : "Terjadi kesalahan."); }
    finally { setSending(false); }
  }
  return <main className="app-page"><header className="app-topbar"><Link className="brand" href="/"><span className="brand-mark">w</span> whisper</Link><Link className="back-link" href="/">Ruang untuk bicara jujur</Link></header><section className="content-card"><div className="profile-avatar">{slug.slice(0,1).toUpperCase()}</div><span className="eyebrow">RUANG PESAN ANONIM</span><h1>Kirim pesan ke {slug}</h1><p>Tulis apa yang ingin kamu sampaikan. Pesanmu tidak menampilkan identitasmu.</p><form onSubmit={submit}><label className="field-label" htmlFor="message">Pesanmu</label><textarea id="message" className="message-input" maxLength={5000} minLength={1} required placeholder="Mulai dengan apa yang ingin kamu katakan…" value={content} onChange={e=>setContent(e.target.value)} /><div className="form-hint">{content.length}/5000 karakter</div>{error&&<p className="form-error">{error}</p>}<button className="button button-primary full-width" disabled={sending}>{sending?"Mengirim…":"Kirim pesan anonim"}<span>→</span></button></form><p className="muted-note">Whisper tidak meminta nama atau email. Hindari membagikan informasi pribadi.</p></section></main>;
}
