"use client";

import { useState } from "react";
import { docs, categories, type Lang } from "@/lib/docs";
import { Globe, Book } from "lucide-react";

export default function DocsPage() {
  const [lang, setLang] = useState<Lang>("en");
  const [active, setActive] = useState("introduction");

  const page = docs.find((d) => d.slug === active);

  return (
    <div className="flex gap-6 -mx-6 -my-6 min-h-[calc(100vh-52px)]">
      {/* Sidebar */}
      <aside className="w-[200px] shrink-0 border-r border-[var(--border)] bg-[var(--bg-card)] p-4 overflow-y-auto">
        <div className="flex items-center gap-2 mb-4">
          <Book className="w-4 h-4 text-[var(--accent)]" />
          <span className="text-[12px] font-bold">Docs</span>
        </div>

        {/* Language toggle */}
        <div className="flex items-center gap-1 mb-4 p-1 rounded-[var(--radius-sm)] bg-[var(--bg-subtle)] border border-[var(--border)]">
          <button onClick={() => setLang("en")} className={`flex-1 px-2 py-1 rounded text-[10px] font-semibold transition-colors ${lang === "en" ? "bg-[var(--bg-card)] text-[var(--accent)] shadow-[var(--shadow-xs)]" : "text-[var(--text-muted)]"}`}>EN</button>
          <button onClick={() => setLang("id")} className={`flex-1 px-2 py-1 rounded text-[10px] font-semibold transition-colors ${lang === "id" ? "bg-[var(--bg-card)] text-[var(--accent)] shadow-[var(--shadow-xs)]" : "text-[var(--text-muted)]"}`}>ID</button>
        </div>

        {/* Nav */}
        {categories.map((cat) => (
          <div key={cat.key} className="mb-3">
            <p className="text-[9px] font-bold uppercase tracking-wider text-[var(--text-muted)] px-2 mb-1">{cat.label[lang]}</p>
            {docs.filter((d) => d.category === cat.key).map((d) => (
              <button
                key={d.slug}
                onClick={() => setActive(d.slug)}
                className={`w-full text-left px-2 py-1.5 rounded-[var(--radius-sm)] text-[11px] font-medium transition-colors ${active === d.slug ? "bg-[var(--accent-bg)] text-[var(--accent)]" : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]"}`}
              >
                {d.title[lang]}
              </button>
            ))}
          </div>
        ))}
      </aside>

      {/* Content */}
      <main className="flex-1 py-6 pr-6 overflow-y-auto">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2 text-[11px] text-[var(--text-muted)]">
            <Globe className="w-3.5 h-3.5" />
            <span>{lang === "en" ? "English" : "Bahasa Indonesia"}</span>
          </div>
        </div>
        {page && <DocContent content={page.content[lang]} />}
      </main>
    </div>
  );
}

// Simple markdown-like renderer
function DocContent({ content }: { content: string }) {
  const lines = content.split("\n");
  const elements: React.ReactNode[] = [];
  let inCode = false;
  let codeBlock: string[] = [];
  let inTable = false;
  let tableRows: string[][] = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];

    if (line.startsWith("```")) {
      if (inCode) {
        elements.push(<pre key={i} className="px-4 py-3 rounded-[var(--radius-sm)] bg-[var(--bg-subtle)] border border-[var(--border)] text-[11px] leading-relaxed overflow-x-auto my-3 text-[var(--text-secondary)]">{codeBlock.join("\n")}</pre>);
        codeBlock = [];
        inCode = false;
      } else {
        inCode = true;
      }
      continue;
    }
    if (inCode) { codeBlock.push(line); continue; }

    if (line.startsWith("|") && line.includes("|")) {
      if (!inTable) { inTable = true; tableRows = []; }
      if (line.includes("---")) continue; // separator
      tableRows.push(line.split("|").filter(Boolean).map((c) => c.trim()));
      if (i + 1 >= lines.length || !lines[i + 1]?.startsWith("|")) {
        elements.push(
          <table key={i} className="w-full text-[11px] my-3 border border-[var(--border)] rounded-[var(--radius-sm)] overflow-hidden">
            <thead><tr className="bg-[var(--bg-subtle)]">{tableRows[0]?.map((h, j) => <th key={j} className="px-3 py-2 text-left font-semibold text-[var(--text-primary)]">{h}</th>)}</tr></thead>
            <tbody>{tableRows.slice(1).map((row, ri) => <tr key={ri} className="border-t border-[var(--border)]">{row.map((c, ci) => <td key={ci} className="px-3 py-2 text-[var(--text-secondary)]">{c}</td>)}</tr>)}</tbody>
          </table>
        );
        inTable = false;
      }
      continue;
    }

    if (line.startsWith("# ")) { elements.push(<h1 key={i} className="text-xl font-bold mb-2 mt-6 first:mt-0">{line.slice(2)}</h1>); continue; }
    if (line.startsWith("## ")) { elements.push(<h2 key={i} className="text-[15px] font-bold mb-1.5 mt-5 text-[var(--text-primary)]">{line.slice(3)}</h2>); continue; }
    if (line.startsWith("### ")) { elements.push(<h3 key={i} className="text-[13px] font-semibold mb-1 mt-4 text-[var(--text-primary)]">{line.slice(4)}</h3>); continue; }
    if (line.startsWith("- ")) { elements.push(<li key={i} className="text-[12px] text-[var(--text-secondary)] ml-4 list-disc leading-relaxed">{renderInline(line.slice(2))}</li>); continue; }
    if (line.trim() === "") { elements.push(<div key={i} className="h-2" />); continue; }
    elements.push(<p key={i} className="text-[12px] text-[var(--text-secondary)] leading-relaxed">{renderInline(line)}</p>);
  }

  return <div className="max-w-2xl">{elements}</div>;
}

function renderInline(text: string): React.ReactNode {
  // Bold, code, links
  const parts = text.split(/(\*\*[^*]+\*\*|`[^`]+`)/g);
  return parts.map((part, i) => {
    if (part.startsWith("**") && part.endsWith("**")) return <strong key={i} className="font-semibold text-[var(--text-primary)]">{part.slice(2, -2)}</strong>;
    if (part.startsWith("`") && part.endsWith("`")) return <code key={i} className="px-1 py-0.5 rounded bg-[var(--bg-subtle)] text-[var(--accent)] text-[11px] font-mono">{part.slice(1, -1)}</code>;
    return <span key={i}>{part}</span>;
  });
}
