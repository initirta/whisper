import { Suspense } from "react";
import { PublicProfile } from "@/components/public-profile";

async function ProfileContent({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  return <PublicProfile slug={slug} />;
}

export default function ProfilePage({ params }: { params: Promise<{ slug: string }> }) {
  return <Suspense fallback={<main className="app-page"><p className="status-line">Membuka profil…</p></main>}><ProfileContent params={params} /></Suspense>;
}
