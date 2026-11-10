import { Suspense } from "react";
import { ChatRoom } from "@/components/chat-room";

async function ConversationContent({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return <ChatRoom conversationID={id} />;
}

export default function ConversationPage({ params }: { params: Promise<{ id: string }> }) {
  return <Suspense fallback={<main className="app-page"><p className="status-line">Membuka percakapan…</p></main>}><ConversationContent params={params} /></Suspense>;
}
