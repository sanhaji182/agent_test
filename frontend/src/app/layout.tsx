import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { Sidebar } from "@/components/sidebar";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "GoTest Agent — AI Testing Platform",
  description: "Self-hosted AI testing agent dashboard",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="h-full flex font-[family-name:var(--font-geist-sans)]">
        <Sidebar />
        <main className="flex-1 overflow-auto">
          <div className="max-w-6xl mx-auto px-6 py-8">{children}</div>
        </main>
      </body>
    </html>
  );
}
