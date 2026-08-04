"use client";

import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Badge } from "@/components/ui/badge";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";
import { Bell, CheckCircle2, AlertTriangle, XCircle, Search, CheckCheck, EyeOff, Check } from "lucide-react";

import { getAlerts, acknowledgeAlert, dismissAlert, markAllAlertsRead, type Alert } from "@/lib/api";

export default function AlertsPage() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState("all");
  const [query, setQuery] = useState("");
  const [busyId, setBusyId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const loadAlerts = useCallback(() => {
    return getAlerts()
      .then((data) => setAlerts(data.map((n) => ({ ...n, severity: n.severity || severityFor(n.type), category: n.category || typeCategory(n.type) }))))
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    loadAlerts();
  }, [loadAlerts]);

  const activeAlerts = alerts.filter((a) => !a.dismissed);

  const filteredAlerts = activeAlerts.filter((alert) => {
    if (filter !== "all" && alert.category !== filter) return false;
    if (query && !alert.message.toLowerCase().includes(query.toLowerCase())) return false;
    return true;
  });

  const handleAck = async (id: string) => {
    setBusyId(id);
    setError(null);
    try {
      await acknowledgeAlert(id);
      await loadAlerts();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusyId(null);
    }
  };

  const handleDismiss = async (id: string) => {
    setBusyId(id);
    setError(null);
    try {
      await dismissAlert(id);
      await loadAlerts();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusyId(null);
    }
  };

  const handleMarkAll = async () => {
    setError(null);
    try {
      await markAllAlertsRead();
      await loadAlerts();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    }
  };

  if (loading) return <LoadingSkeleton rows={6} />;

  const unreadCount = activeAlerts.filter((a) => !a.acknowledged).length;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-semibold tracking-tight">Alerts</h1>
          <p className="text-sm text-[var(--text-muted)] mt-1">Test failures, drift detections & system notifications</p>
        </div>
        <Button variant="secondary" onClick={handleMarkAll} disabled={unreadCount === 0}>
          <CheckCheck className="w-4 h-4 mr-2" />
          Mark All as Read{unreadCount > 0 ? ` (${unreadCount})` : ""}
        </Button>
      </div>

      {/* Error Message */}
      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 p-4">
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}

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
              {f.label} ({getCount(activeAlerts, f.value)})
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
              {filteredAlerts.map((alert) => (
                <Tr key={alert.id} hover className={alert.acknowledged ? "opacity-60" : ""}>
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
                      variant={alert.category === "failure" ? "danger" : alert.category === "drift" ? "warning" : "info"}
                      size="sm"
                    >
                      {alert.type}
                    </Badge>
                  </Td>
                  <Td className="text-[var(--text-muted)] text-xs whitespace-nowrap">{formatTime(alert.created_at)}</Td>
                  <Td align="right" className="space-x-2">
                    {alert.acknowledged ? (
                      <span className="inline-flex items-center gap-1 text-xs text-[var(--text-muted)]">
                        <Check className="w-3.5 h-3.5" /> Read
                      </span>
                    ) : (
                      <Button variant="ghost" size="sm" disabled={busyId === alert.id} onClick={() => handleAck(alert.id)}>
                        Acknowledge
                      </Button>
                    )}
                    <Button variant="ghost" size="sm" disabled={busyId === alert.id} onClick={() => handleDismiss(alert.id)}>
                      <EyeOff className="w-3.5 h-3.5 mr-1" />
                      Dismiss
                    </Button>
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

function severityFor(type: string): string {
  if (type === "failure") return "critical";
  if (type === "flake" || type === "degradation") return "warning";
  return "info";
}

function typeCategory(type: string): string {
  if (type === "failure") return "failure";
  if (type === "flake" || type === "degradation") return "drift";
  return "system";
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  const diffHour = Math.floor(diffMs / 3600000);
  const diffDay = Math.floor(diffMs / 86400000);

  if (diffMin < 1) return "just now";
  if (diffMin < 60) return `${diffMin}m ago`;
  if (diffHour < 24) return `${diffHour}h ago`;
  if (diffDay < 7) return `${diffDay}d ago`;
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

function getCount(alerts: Alert[], value: string): number {
  if (value === "all") return alerts.length;
  return alerts.filter((a) => a.category === value).length;
}
