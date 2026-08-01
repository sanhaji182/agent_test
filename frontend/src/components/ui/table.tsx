import { cn } from "@/lib/utils";

/* ========== TABLE COMPONENTS ========== */

export function TableContainer({ children }: { children: React.ReactNode }) {
  return (
    <div className="overflow-hidden rounded-lg border border-[var(--border-default)] bg-white">
      <div className="overflow-x-auto">{children}</div>
    </div>
  );
}

export function Th({ children, align = "left" }: { children: React.ReactNode; align?: "left" | "center" | "right" }) {
  return (
    <th className={`px-4 py-3 text-xs font-semibold uppercase tracking-wider text-[var(--text-muted)] ${align === "right" ? "text-right" : ""}`}>
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
        "border-b border-[var(--border-default)]",
        onClick ? "cursor-pointer" : "",
        hover && !onClick ? "hover:bg-gray-50 transition-colors" : "",
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
      className={`px-4 py-3 text-sm ${className || ""} ${align === "right" ? "text-right" : ""}`}
      colSpan={colSpan}
    >
      {children}
    </td>
  );
}
