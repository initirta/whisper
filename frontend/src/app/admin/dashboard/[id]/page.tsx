import { Suspense } from "react";
import { ChatRoom } from "@/components/chat-room";

async function AdminConversationContent({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return <ChatRoom conversationID={id} adminMode />;
}

export default function AdminConversationPage({ params }: { params: Promise<{ id: string }> }) {
  return <Suspense fallback={<main className="app-page"><p className="status-line">Membuka percakapan…</p></main>}><AdminConversationContent params={params} /></Suspense>;
}
