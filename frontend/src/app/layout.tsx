import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Whisper",
  description: "Platform pesan anonim real-time",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="id" className="h-full antialiased">
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
