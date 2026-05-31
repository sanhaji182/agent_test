"use client";

import { useEffect, useState } from "react";
import { getNotifications, type Notification } from "@/lib/api";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Bell, CheckCircle2, XCircle } from "lucide-react";

export default function AlertsPage() {
  const [notifs, setNotifs] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getNotifications().then(setNotifs).catch(() => {}).finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="space-y-6"><LoadingSkeleton rows={4} /></div>;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold">Alerts & Notifications</h1>
        <p className="text-sm text-[var(--text-secondary)] mt-0.5">Failure alerts and notification history</p>
      </div>

      <Section title="Notification History">
        {notifs.length === 0 ? (
          <EmptyState
            icon={<Bell className="w-6 h-6" />}
            title="No notifications"
            description="Alerts are triggered when scheduled runs fail. Configure notify_on_fail in your schedules."
          />
        ) : (
          <div className="space-y-2">
            {notifs.map((n) => (
              <div key={n.id} className="flex items-start gap-3 p-3 rounded-lg border border-[var(--border)] bg-[var(--bg-subtle)]">
                {n.delivered ? (
                  <CheckCircle2 className="w-4 h-4 text-[var(--success)] mt-0.5 shrink-0" />
                ) : (
                  <XCircle className="w-4 h-4 text-[var(--danger)] mt-0.5 shrink-0" />
                )}
                <div className="flex-1 min-w-0">
                  <p className="text-xs font-medium">{n.message}</p>
                  <p className="text-[11px] text-[var(--text-muted)] mt-0.5">
                    {n.type} · {new Date(n.created_at).toLocaleString()} · {n.delivered ? "Delivered" : "Pending"}
                  </p>
                </div>
              </div>
            ))}
          </div>
        )}
      </Section>
    </div>
  );
}
