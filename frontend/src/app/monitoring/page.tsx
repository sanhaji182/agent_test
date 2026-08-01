"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Badge } from "@/components/ui/badge";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";
import { Plus, Search, Bell, CheckCircle2, AlertTriangle, XCircle, Clock } from "lucide-react";

export default function MonitoringPage() {
  const [metrics, setMetrics] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // TODO: Fetch real metrics from /api/v1/metrics endpoint
    setTimeout(() => setLoading(false), 500);
  }, []);

  if (loading) return <LoadingSkeleton rows={6} />;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-semibold tracking-tight">Monitoring</h1>
          <p className="text-sm text-[var(--text-muted)] mt-1">System health & performance metrics</p>
        </div>
        <Button variant="secondary">Refresh</Button>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard label="Test Success Rate" value="94%" trend="+2%" positive />
        <StatCard label="Avg Response Time" value="1.2s" trend="-15%" positive />
        <StatCard label="Active Runs" value="12" color="blue" />
        <StatCard label="Failed Tests" value="6" danger />
      </div>

      {/* Performance Chart Placeholder */}
      <Section title="Test Performance Over Time">
        <div className="h-64 bg-gray-50 rounded-lg flex items-center justify-center text-[var(--text-muted)]">
          <p className="text-sm">Performance chart visualization would appear here</p>
        </div>
      </Section>

      {/* Recent Activity */}
      <Section title="Recent Activity">
        <ActivityItem icon={<CheckCircle2 className="w-4 h-4 text-green-600" />} title="Test completed successfully" time="2m ago" />
        <ActivityItem icon={<AlertTriangle className="w-4 h-4 text-yellow-600" />} title="Performance degradation detected" time="15m ago" warning />
        <ActivityItem icon={<XCircle className="w-4 h-4 text-red-600" />} title="Build failed" time="1h ago" danger />
        <ActivityItem icon={<Clock className="w-4 h-4 text-blue-600" />} title="Deployment initiated" time="3h ago" />
      </Section>
    </div>
  );
}

function StatCard({ 
  label, 
  value, 
  trend, 
  positive, 
  danger, 
  color = "default" 
}: { 
  label: string; 
  value: string; 
  trend?: string;
  positive?: boolean;
  danger?: boolean;
  color?: string;
}) {
  const textColor = danger ? "text-red-600" : positive ? "text-green-600" : color === "blue" ? "text-blue-600" : "";
  
  return (
    <div className="bg-white rounded-lg p-4 border border-[var(--border-default)]">
      <p className="text-xs text-[var(--text-muted)] font-medium uppercase tracking-wide">{label}</p>
      <div className="mt-2 flex items-baseline gap-2">
        <span className={`text-2xl font-semibold ${textColor}`}>{value}</span>
        {trend && <span className={`text-xs ${positive ? "text-green-600" : ""}`}>{trend}</span>}
      </div>
    </div>
  );
}

function ActivityItem({ 
  icon, 
  title, 
  time, 
  warning, 
  danger 
}: { 
  icon: React.ReactNode; 
  title: string; 
  time: string;
  warning?: boolean;
  danger?: boolean;
}) {
  const bgColor = warning ? "bg-yellow-50" : danger ? "bg-red-50" : "bg-gray-50";
  const borderColor = warning ? "border-yellow-200" : danger ? "border-red-200" : "border-gray-200";
  
  return (
    <div className={`flex items-start gap-3 p-3 rounded-lg border ${bgColor} ${borderColor}`}>
      <div className="shrink-0 mt-0.5">{icon}</div>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-[var(--text-primary)]">{title}</p>
        <p className="text-xs text-[var(--text-muted)] mt-0.5">{time}</p>
      </div>
    </div>
  );
}
