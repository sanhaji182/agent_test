"use client";

import { useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import Link from "next/link";
import {
  approveChangeProposal,
  getTestCaseProposals,
  getTestCaseMaintenance,
  getTestCases,
  refineTestCase,
  runTestCase,
  type ChangeProposal,
  type MaintenanceItem,
  type TestCase,
} from "@/lib/api";
import { EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { AlertTriangle, CheckCircle2, PlayCircle, Plus, Search, Sparkles } from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";

type Tab = "ui" | "api";

export default function TestLibraryPage() {
  const [tests, setTests] = useState<TestCase[]>([]);
  const [maintenance, setMaintenance] = useState<MaintenanceItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<Tab>("ui");
  const [query, setQuery] = useState("");
  const [runningId, setRunningId] = useState<string | null>(null);
  const [selected, setSelected] = useState<TestCase | null>(null);
  const [proposals, setProposals] = useState<ChangeProposal[]>([]);
  const [refinePrompt, setRefinePrompt] = useState("");
  const [refining, setRefining] = useState(false);
  const [approvingId, setApprovingId] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([getTestCases(), getTestCaseMaintenance()])
      .then(([testData, maintenanceData]) => {
        setTests(testData || []);
        setMaintenance(maintenanceData || []);
      })
      .catch(() => {
        setTests([]);
        setMaintenance([]);
      })
      .finally(() => setLoading(false));
  }, []);

  const filteredTests = useMemo(() => tests.filter((test) => {
    const type = (test.type || "ui") as Tab;
    if (type !== activeTab) return false;
    const haystack = `${test.title} ${test.feature} ${test.priority} ${test.tags?.join(" ")}`.toLowerCase();
    return haystack.includes(query.toLowerCase());
  }), [tests, activeTab, query]);

  const handleRun = async (test: TestCase) => {
    setRunningId(test.id);
    try {
      const res = await runTestCase(test.id);
      window.location.href = `/runs/${res.run_id}`;
    } finally {
      setRunningId(null);
    }
  };

  const openRefine = async (test: TestCase) => {
    setSelected(test);
    setRefinePrompt("");
    setProposals(await getTestCaseProposals(test.id).catch(() => []));
  };

  const submitRefinement = async () => {
    if (!selected || !refinePrompt.trim()) return;
    setRefining(true);
    try {
      const proposal = await refineTestCase(selected.id, refinePrompt);
      setProposals((prev) => [proposal, ...prev]);
      setRefinePrompt("");
      getTestCaseMaintenance().then((m) => setMaintenance(m || [])).catch(() => {});
    } finally {
      setRefining(false);
    }
  };

  const approveProposal = async (proposal: ChangeProposal) => {
    setApprovingId(proposal.id);
    try {
      const res = await approveChangeProposal(proposal.id, { reviewer: "self-hosted", comment: "Approved from Test Library" });
      setTests((prev) => prev.map((test) => test.id === res.test_case.id ? res.test_case : test));
      setSelected(res.test_case);
      setProposals((prev) => prev.map((item) => item.id === res.proposal.id ? res.proposal : item));
      getTestCaseMaintenance().then((m) => setMaintenance(m || [])).catch(() => {});
    } finally {
      setApprovingId(null);
    }
  };

  if (loading) return <LoadingSkeleton rows={7} />;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-semibold tracking-tight">Test Library</h1>
          <p className="text-sm text-[var(--text-muted)] mt-1">Approved test cases ready for execution</p>
        </div>
        <Link href="/create">
          <Button>New Test</Button>
        </Link>
      </div>

      {/* Tab Controls - Modern Segmented Control */}
      <div className="inline-flex rounded-lg border border-[var(--border-default)] p-0.5 bg-white">
        {[
          { value: "ui", label: "UI Tests" },
          { value: "api", label: "API Tests" },
        ].map((tab) => (
          <button
            key={tab.value}
            onClick={() => setActiveTab(tab.value as Tab)}
            className={cn(
              "px-4 py-2 text-sm font-medium rounded-md transition-colors",
              activeTab === tab.value
                ? "bg-blue-600 text-white shadow-sm"
                : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-gray-50"
            )}
          >
            {tab.label} ({tests.filter(t => t.type === tab.value).length})
          </button>
        ))}
      </div>

      {/* Search */}
      <div className="relative max-w-md">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
        <Input
          placeholder="Search by name, feature, priority, or tag..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="pl-9"
        />
      </div>

      {/* Maintenance Panel - Collapsible Style */}
      {maintenance.length > 0 && (
        <div className="rounded-lg border border-yellow-200 bg-yellow-50 p-4">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <AlertTriangle className="w-4 h-4 text-yellow-700" />
              <h2 className="text-sm font-semibold text-yellow-800">Maintenance Needed</h2>
            </div>
            <Badge variant="warning">{maintenance.length} items</Badge>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
            {maintenance.slice(0, 4).map((item) => (
              <MaintenanceItemCard key={`${item.test_case_id}-${item.category}`} item={item} />
            ))}
          </div>
        </div>
      )}

      {/* Tests List */}
      <TableContainer>
        <table className="w-full text-left">
          <thead className="bg-gray-50 border-b border-[var(--border-default)]">
            <tr>
              <Th>Name</Th>
              <Th>Priority</Th>
              <Th align="right">Actions</Th>
            </tr>
          </thead>
          <tbody>
            {filteredTests.length === 0 ? (
              <Tr>
                <Td colSpan={3}>
                  <EmptyState 
                    icon={<Sparkles className="w-6 h-6" />}
                    title={!query ? "No tests yet" : "No matching tests"}
                    description={!query ? "Generate your first test to populate the library." : "Try adjusting your search criteria."}
                  />
                </Td>
              </Tr>
            ) : (
              filteredTests.map((test) => {
                const deterministic = !!test.executable_content;
                return (
                  <Tr key={test.id} hover>
                    <Td className="font-medium">
                      <span className="block truncate max-w-[180px]">{test.title}</span>
                      <span className="inline-flex items-center gap-1.5 text-xs text-[var(--text-muted)]">
                        {deterministic && (
                          <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-[var(--success-bg)] border border-[var(--success)]/20 text-[10px] font-semibold text-[var(--success)]" title="Test case ini bisa di-run ulang secara deterministik tanpa AI — persis seperti rekam-putar Katalon">
                            <CheckCircle2 className="w-3 h-3" /> Deterministic
                          </span>
                        )}
                        {test.steps?.length > 0 && <span>{test.steps.length} steps</span>}
                        {test.feature && <span>· {test.feature}</span>}
                      </span>
                    </Td>
                    <Td><PriorityBadge priority={test.priority} /></Td>
                    <Td className="text-right space-x-2">
                      <Button
                        variant="primary"
                        size="sm"
                        isLoading={runningId === test.id}
                        onClick={() => handleRun(test)}
                      >
                        <PlayCircle className="w-3.5 h-3.5" />
                        Run
                      </Button>
                      <Link href={`/tests/${test.id}/edit`}>
                        <Button variant="secondary" size="sm">Edit</Button>
                      </Link>
                    </Td>
                  </Tr>
                );
              })
            )}
          </tbody>
        </table>
      </TableContainer>

      {/* Refinement Modal */}
      {selected && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="bg-white rounded-lg shadow-xl max-w-lg w-full mx-4 p-6">
            <h3 className="text-lg font-semibold mb-4">Refine Test Case</h3>
            <p className="text-sm text-[var(--text-muted)] mb-4">{selected.title}</p>
            
            <Input
              placeholder="Describe what changes you'd like to make..."
              value={refinePrompt}
              onChange={(e) => setRefinePrompt(e.target.value)}
              autoFocus
            />
            
            <div className="mt-4 flex justify-end gap-2">
              <Button variant="secondary" onClick={() => setSelected(null)}>Cancel</Button>
              <Button 
                onClick={submitRefinement} 
                isLoading={refining}
                disabled={!refinePrompt.trim()}
              >
                Generate Changes
              </Button>
            </div>

            {/* Proposals */}
            {proposals.length > 0 && (
              <div className="mt-4 pt-4 border-t">
                <h4 className="text-sm font-semibold mb-3">Generated Proposals:</h4>
                {proposals.map((proposal) => (
                  <div key={proposal.id} className="p-3 rounded-lg bg-gray-50 mb-2">
                    <p className="text-sm font-medium mb-2">{proposal.rationale}</p>
                    <div className="flex gap-2">
                      <Button 
                        size="sm"
                        isLoading={approvingId === proposal.id}
                        onClick={() => approveProposal(proposal)}
                      >
                        Approve
                      </Button>
                      <Button size="sm" variant="secondary">Decline</Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

// === Components ===

function PriorityBadge({ priority }: { priority: string }) {
  const colors: Record<string, string> = {
    high: "bg-red-100 text-red-700",
    medium: "bg-yellow-100 text-yellow-700",
    low: "bg-green-100 text-green-700",
  };
  
  return (
    <span className={cn("inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize", colors[priority] || "bg-gray-100 text-gray-700")}>
      {priority}
    </span>
  );
}

function MaintenanceItemCard({ item }: { item: MaintenanceItem }) {
  const severityColors: Record<string, string> = {
    high: "text-red-700 border-red-200 bg-red-50",
    medium: "text-yellow-700 border-yellow-200 bg-yellow-50",
    low: "text-gray-700 border-gray-200 bg-gray-50",
  };
  
  return (
    <div className={cn(
      "p-3 rounded-lg border text-sm",
      severityColors[item.severity] || severityColors.low
    )}>
      <div className="flex items-start justify-between mb-1">
        <span className="font-medium truncate flex-1">{item.title}</span>
        <Badge variant={item.severity === "high" ? "danger" : item.severity === "medium" ? "warning" : "default"} size="sm">
          {item.severity}
        </Badge>
      </div>
      <p className="text-xs opacity-75 mb-1">{item.reason}</p>
      <p className="text-xs opacity-60">{item.action}</p>
    </div>
  );
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  const diffHour = Math.floor(diffMs / 3600000);
  const diffDay = Math.floor(diffMs / 86400000);
  
  if (diffMin < 60) return `${diffMin}m ago`;
  if (diffHour < 24) return `${diffHour}h ago`;
  if (diffDay < 7) return `${diffDay}d ago`;
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}
