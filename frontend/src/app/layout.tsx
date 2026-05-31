import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { Sidebar } from "@/components/sidebar";
import { ThemeToggle } from "@/components/theme-toggle";

const geistSans = Geist({ variable: "--font-geist-sans", subsets: ["latin"] });
const geistMono = Geist_Mono({ variable: "--font-geist-mono", subsets: ["latin"] });

export const metadata: Metadata = {
  title: "GoTest Agent — AI Testing Platform",
  description: "Self-hosted AI testing agent dashboard",
};

const themeScript = `(function(){try{var s=localStorage.getItem('theme');if(s==='dark'||(!s&&window.matchMedia('(prefers-color-scheme:dark)').matches)){document.documentElement.classList.add('dark')}}catch(e){}})();`;

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}>
      <head><script dangerouslySetInnerHTML={{ __html: themeScript }} /></head>
      <body className="h-full flex font-[family-name:var(--font-geist-sans)]">
        <Sidebar />
        <div className="flex-1 flex flex-col min-w-0">
          <header className="h-[52px] border-b border-[var(--border)] bg-[var(--bg-card)] flex items-center justify-end px-5 gap-2 sticky top-0 z-10 shadow-[var(--shadow-xs)]">
            <ThemeToggle />
          </header>
          <main className="flex-1 overflow-auto">
            <div className="max-w-[1120px] mx-auto px-6 py-6">{children}</div>
          </main>
        </div>
      </body>
    </html>
  );
}
