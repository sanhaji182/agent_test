"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Badge } from "@/components/ui/badge";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";
import { Bell, CheckCircle2, AlertTriangle, XCircle, Search, Filter } from "lucide-react";

export default function AlertsPage() {
  const [alerts, setAlerts] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState("all");
  const [query, setQuery] = useState("");

  useEffect(() => {
    // TODO: Fetch real alerts
    setTimeout(() => setLoading(false), 500);
  }, []);

  const filteredAlerts = alerts.filter(alert => {
    if (filter !== "all" && alert.type !== filter) return false;
    if (query && !alert.message.toLowerCase().includes(query.toLowerCase())) return false;
    return true;
  });

  if (loading) return <LoadingSkeleton rows={6} />;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-semibold tracking-tight">Alerts</h1>
          <p className="text-sm text-[var(--text-muted)] mt-1">Test failures, drift detections & system notifications</p>
        </div>
        <Button variant="secondary">Mark All as Read</Button>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
          <Input
            placeholder="Search alerts..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="pl-9"
          />
        </div>

        <div className="inline-flex rounded-lg border border-[var(--border-default)] p-0.5 bg-white">
          {[
            { value: "all", label: "All" },
            { value: "failure", label: "Failures" },
            { value: "drift", label: "Drift" },
            { value: "system", label: "System" },
          ].map((f) => (
            <button
              key={f.value}
              onClick={() => setFilter(f.value)}
              className={filter === f.value ? "px-3 py-1.5 text-xs font-medium rounded-md bg-blue-600 text-white" : "px-3 py-1.5 text-xs font-medium rounded-md text-[var(--text-secondary)] hover:bg-gray-50"}
            >
              {f.label} ({getCount(filter, f.value)})
            </button>
          ))}
        </div>
      </div>

      {/* Alerts List */}
      {filteredAlerts.length === 0 ? (
        <div className="rounded-lg border border-[var(--border-default)] bg-white p-8 text-center">
          <Bell className="w-12 h-12 text-[var(--text-muted)] mx-auto mb-4" />
          <h3 className="text-base font-semibold text-[var(--text-primary)] mb-2">No alerts found</h3>
          <p className="text-sm text-[var(--text-muted)] max-w-sm mx-auto">
            {query ? "Try adjusting your search criteria." : "Your test runs are healthy. No active alerts."}
          </p>
        </div>
      ) : (
        <TableContainer>
          <table className="w-full text-left">
            <thead className="bg-gray-50 border-b border-[var(--border-default)]">
              <tr>
                <Th>Status</Th>
                <Th>Message</Th>
                <Th>Type</Th>
                <Th>Time</Th>
                <Th align="right">Actions</Th>
              </tr>
            </thead>
            <tbody>
              {filteredAlerts.map((alert, i) => (
                <Tr key={i} hover>
                  <Td>
                    {alert.severity === "critical" ? (
                      <XCircle className="w-5 h-5 text-red-600" />
                    ) : alert.severity === "warning" ? (
                      <AlertTriangle className="w-5 h-5 text-yellow-600" />
                    ) : (
                      <CheckCircle2 className="w-5 h-5 text-green-600" />
                    )}
                  </Td>
                  <Td className="font-medium">{alert.message}</Td>
                  <Td>
                    <Badge 
                      variant={alert.type === "failure" ? "danger" : alert.type === "drift" ? "warning" : "info"}
                      size="sm"
                    >
                      {alert.type}
                    </Badge>
                  </Td>
                  <Td className="text-[var(--text-muted)] text-xs whitespace-nowrap">{alert.time}</Td>
                  <Td align="right" className="space-x-2">
                    <Button variant="ghost" size="sm">Acknowledge</Button>
                    <Button variant="ghost" size="sm">Dismiss</Button>
                  </Td>
                </Tr>
              ))}
            </tbody>
          </table>
        </TableContainer>
      )}
    </div>
  );
}

function getCount(current: string, value: string): number {
  // Placeholder for filtering counts
  return 0;
}
