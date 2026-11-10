"use client";

import Link from "next/link";
import { useCallback, useEffect, useRef, useState } from "react";
import { API_BASE, Message, websocketURL } from "@/lib/api";

type Incoming = { type: string; message?: Message; messages?: Message[]; content?: string; error?: string };

export function ChatRoom({ conversationID, adminMode = false }: { conversationID: string; adminMode?: boolean }) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [text, setText] = useState("");
  const [status, setStatus] = useState("Menghubungkan…");
  const [hasAccess, setHasAccess] = useState(adminMode);
  const socket = useRef<WebSocket | null>(null);
  const tokenRef = useRef("");
  const bottom = useRef<HTMLDivElement>(null);
  const addMessage = useCallback((message: Message) => setMessages(current => current.some(m=>m.id===message.id)?current:[...current,message].sort((a,b)=>a.created_at.localeCompare(b.created_at))), []);

  useEffect(() => {
    const accessToken = new URLSearchParams(window.location.hash.slice(1)).get("k") ?? "";
    tokenRef.current = accessToken;
    if (!adminMode && !accessToken) {
      const timer = setTimeout(() => setStatus("Kunci akses tidak ditemukan. Buka link percakapan lengkap untuk melanjutkan."), 0);
      return () => clearTimeout(timer);
    }
    if (!adminMode) setTimeout(() => setHasAccess(true), 0);
    let alive=true; let reconnectTimer: ReturnType<typeof setTimeout>;
    const connect=()=>{
      if(!alive)return;
      let rejected=false;
      const ws=new WebSocket(websocketURL(conversationID)); socket.current=ws;
      ws.onopen=()=>{ws.send(JSON.stringify({type:"authenticate",...(adminMode?{}:{token:accessToken})}));setStatus("Mengamankan percakapan…");};
      ws.onmessage=event=>{const data=JSON.parse(event.data) as Incoming;if(data.type==="auth.success"){setStatus("Percakapan privat · terhubung");}else if(data.type==="auth.failed"){rejected=true;setStatus(data.error??"Akses ditolak.");ws.close();}else if(data.type==="message.sync"){setMessages(data.messages??[]);}else if((data.type==="message.new"||data.type==="message.ack")&&data.message){addMessage(data.message);}else if(data.type==="error"){setStatus(data.error??"Koneksi mengalami kendala.");}};
      ws.onclose=()=>{if(alive&&!rejected){setStatus("Koneksi terputus · mencoba menyambung kembali…");reconnectTimer=setTimeout(connect,1800);}};
      ws.onerror=()=>ws.close();
    };
    connect(); return()=>{alive=false;clearTimeout(reconnectTimer);socket.current?.close();};
  },[conversationID,addMessage,adminMode]);

  useEffect(()=>{bottom.current?.scrollIntoView({behavior:"smooth"});},[messages]);
  async function send(event: React.FormEvent<HTMLFormElement>){event.preventDefault();const content=text.trim();if(!content||!hasAccess||(!adminMode&&!tokenRef.current))return;setText("");const clientID=crypto.randomUUID();const ws=socket.current;
    if(ws?.readyState===WebSocket.OPEN){ws.send(JSON.stringify({type:"message.send",content,client_message_id:clientID}));return;}
    try{const response=await fetch(`${API_BASE}${adminMode?"/admin":""}/conversations/${conversationID}/messages`,{method:"POST",credentials:"include",headers:{"Content-Type":"application/json",...(adminMode?{}:{Authorization:`Bearer ${tokenRef.current}`})},body:JSON.stringify({content,client_message_id:clientID})});const data=await response.json();if(!response.ok)throw new Error(data.error??"Pesan gagal dikirim");addMessage(data.message);}catch(e){setStatus(e instanceof Error?e.message:"Pesan gagal dikirim");setText(content);}
  }
  return <main className="app-page"><header className="app-topbar"><Link className="brand" href={adminMode?"/admin/dashboard":"/"}><span className="brand-mark">w</span> whisper</Link><span className="back-link">Ruang percakapan privat</span></header><section className="content-card chat-card"><div className="chat-head"><div className="avatar avatar-lilac">✦</div><div><h1>Percakapan anonim</h1><p>{status}</p></div><span className="online-dot" /></div><div className="chat-scroll">{messages.length===0&&<div className="status-line">{hasAccess?"Percakapan dimulai. Kamu bisa membalas kapan saja.":"Akses privat diperlukan untuk melihat pesan."}</div>}{messages.map(message=><div className={`chat-bubble ${(message.sender_type===(adminMode?"admin":"anonymous"))?"outgoing":"incoming"}`} key={message.id}>{message.content}<span className="chat-meta">{new Date(message.created_at).toLocaleTimeString("id-ID",{hour:"2-digit",minute:"2-digit"})}</span></div>)}<div ref={bottom}/></div><form className="chat-compose" onSubmit={send}><textarea aria-label="Tulis balasan" disabled={!hasAccess} maxLength={5000} placeholder={hasAccess?"Tulis balasan…":"Kunci akses diperlukan"} value={text} onChange={e=>setText(e.target.value)} onKeyDown={e=>{if(e.key==="Enter"&&!e.shiftKey){e.preventDefault();e.currentTarget.form?.requestSubmit();}}}/><button className="send-button" aria-label="Kirim pesan" disabled={!hasAccess||!text.trim()}>↑</button></form></section><p className="muted-note">{adminMode?"Percakapan privat dengan pengirim anonim.":"Simpan link ini untuk membuka percakapan lagi. Kunci akses ada di alamat privat browser."}</p></main>;
}
