import { cn } from "@/lib/utils";

/* ========== TABLE COMPONENTS ========== */

export function TableContainer({ children }: { children: React.ReactNode }) {
  return (
    <div className="overflow-hidden rounded-[var(--radius)] border border-[var(--border-default)] bg-white shadow-[var(--shadow-xs)]">
      <div className="overflow-x-auto">{children}</div>
    </div>
  );
}

export function Th({ children, align = "left" }: { children: React.ReactNode; align?: "left" | "center" | "right" }) {
  return (
    <th className={`px-5 py-3.5 text-[11px] font-semibold uppercase tracking-wider text-[var(--text-muted)] ${align === "right" ? "text-right" : ""}`}>
      {children}
    </th>
  );
}

export function Tr({ children, onClick, hover, className }: { 
  children: React.ReactNode; 
  onClick?: () => void; 
  hover?: boolean;
  className?: string;
}) {
  return (
    <tr 
      className={cn(
        "border-b border-[var(--border-default)] last:border-b-0",
        onClick ? "cursor-pointer" : "",
        hover && !onClick ? "hover:bg-[var(--bg-hover)] transition-colors" : "",
        className
      )}
      onClick={onClick}
    >
      {children}
    </tr>
  );
}

export function Td({ children, className, align = "left", colSpan }: { 
  children: React.ReactNode; 
  className?: string;
  align?: "left" | "center" | "right";
  colSpan?: number;
}) {
  return (
    <td 
      className={`px-5 py-3.5 text-sm text-[var(--text-primary)] ${className || ""} ${align === "right" ? "text-right" : ""}`}
      colSpan={colSpan}
    >
      {children}
    </td>
  );
}
