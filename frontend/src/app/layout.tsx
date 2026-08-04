import type { Metadata } from "next";
import "./globals.css";
import { SidebarProvider } from "@/components/layout/sidebar-context";
import { AppShell } from "@/components/layout/app-shell";

export const metadata: Metadata = {
  title: "GoTest Agent — AI Testing Platform",
  description: "Self-hosted AI testing agent dashboard",
};

const themeScript = `(function(){try{var s=localStorage.getItem('theme');if(s==='dark'||(!s&&window.matchMedia('(prefers-color-scheme:dark)').matches)){document.documentElement.classList.add('dark')}else{document.documentElement.classList.remove('dark')}}catch(e){}})();`;

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        {/* Theme script - moved to client-only for proper hydration */}
        <script
          dangerouslySetInnerHTML={{
            __html: themeScript
          }}
        />
      </head>
      <body className={`h-full antialiased`}>
        <SidebarProvider>
          <AppShell>{children}</AppShell>
        </SidebarProvider>
      </body>
    </html>
  );
}
