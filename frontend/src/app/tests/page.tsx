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
import { AlertTriangle, CheckCircle2, Clock, FileText, Layers, MonitorPlay, PencilLine, PlayCircle, Plus, Search, Sparkles, TerminalSquare } from "lucide-react";
import { cn } from "@/lib/utils";

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
        setTests(testData);
        setMaintenance(maintenanceData);
      })
      .catch(() => {
        setTests([]);
        setMaintenance([]);
      })
      .finally(() => setLoading(false));
  }, []);

  const filteredTests = useMemo(() => tests.filter((test) => {
    const type = (test.type || "ui") as Tab;
    const haystack = `${test.title} ${test.feature} ${test.priority} ${test.tags?.join(" ")}`.toLowerCase();
    return type === activeTab && haystack.includes(query.toLowerCase());
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
    <div className="space-y-5 pb-10">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold">Test Library</h1>
          <p className="text-[13px] text-[var(--text-secondary)] mt-0.5">Approved UI and API test cases ready for repeatable execution</p>
        </div>
        <Link
          href="/projects"
          className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-[var(--radius-sm)] bg-[var(--accent)] text-white text-[12px] font-semibold hover:bg-[var(--accent-hover)] transition-colors shadow-sm"
        >
          <Plus className="w-3.5 h-3.5" />
          Generate Tests
        </Link>
      </div>

      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
          <input
            type="text"
            placeholder="Search test name, feature, priority, or tag..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="input pl-10"
          />
        </div>
      </div>

      {maintenance.length > 0 && (
        <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)] shadow-[var(--shadow-sm)] p-4">
          <div className="flex items-center justify-between gap-3 mb-3">
            <div className="flex items-center gap-2">
              <AlertTriangle className="w-4 h-4 text-[var(--warning)]" />
              <h2 className="text-sm font-bold">Maintenance Advisor</h2>
            </div>
            <span className="text-[11px] text-[var(--text-muted)]">{maintenance.length} items</span>
          </div>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-2">
            {maintenance.slice(0, 4).map((item) => (
              <div key={`${item.test_case_id}-${item.category}`} className="rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-subtle)] p-3">
                <div className="flex items-center gap-2">
                  <span className={cn("px-1.5 py-0.5 rounded text-[10px] font-semibold", item.severity === "high" ? "bg-[var(--danger-bg)] text-[var(--danger)]" : item.severity === "medium" ? "bg-[var(--warning-bg)] text-[var(--warning)]" : "bg-[var(--bg-card)] text-[var(--text-muted)]")}>{item.category.replace("_", " ")}</span>
                  <span className="text-[11px] text-[var(--text-muted)] truncate">{item.title}</span>
                </div>
                <p className="text-[12px] text-[var(--text-secondary)] mt-2">{item.reason}</p>
                <p className="text-[11px] text-[var(--text-muted)] mt-1">{item.action}</p>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="flex items-center gap-1 border-b border-[var(--border)]">
        <TabButton active={activeTab === "ui"} onClick={() => setActiveTab("ui")} icon={<MonitorPlay className="w-4 h-4" />} label="UI Testing" />
        <TabButton active={activeTab === "api"} onClick={() => setActiveTab("api")} icon={<TerminalSquare className="w-4 h-4" />} label="API Testing" />
      </div>

      {filteredTests.length === 0 ? (
        <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)]">
          <EmptyState
            icon={activeTab === "ui" ? <MonitorPlay className="w-6 h-6" /> : <TerminalSquare className="w-6 h-6" />}
            title={`No approved ${activeTab.toUpperCase()} tests`}
            description="Create a project, generate a plan, and approve cases to populate this library."
            action={
              <Link href="/projects" className="inline-flex items-center gap-1.5 px-4 py-2 rounded-[var(--radius-sm)] bg-[var(--accent)] text-white text-[12px] font-semibold hover:bg-[var(--accent-hover)] transition-colors shadow-sm">
                Open Projects
              </Link>
            }
          />
        </div>
      ) : (
        <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)] shadow-[var(--shadow-sm)] overflow-hidden divide-y divide-[var(--border)]">
          {filteredTests.map((test) => (
            <div key={test.id} className="w-full flex items-center justify-between gap-4 px-4 py-3 hover:bg-[var(--bg-hover)] transition-colors group">
              <div className="flex items-center gap-4 min-w-0">
                <span className={cn("w-2 h-2 rounded-full shrink-0", test.priority === "high" ? "bg-[var(--danger)]" : test.priority === "medium" ? "bg-[var(--warning)]" : "bg-[var(--success)]")} />
                <div className="min-w-0">
                  <h3 className="text-sm font-semibold text-[var(--text-primary)] group-hover:text-[var(--accent)] transition-colors truncate">
                    {test.title}
                  </h3>
                  <div className="flex flex-wrap items-center gap-3 mt-1">
                    <span className="flex items-center gap-1 text-[11px] text-[var(--text-muted)]">
                      <Layers className="w-3 h-3" /> {(test.type || "ui").toUpperCase()}
                    </span>
                    <span className="flex items-center gap-1 text-[11px] text-[var(--text-muted)] min-w-0">
                      <FileText className="w-3 h-3" /> <span className="truncate max-w-[280px]">{test.feature || "No feature"}</span>
                    </span>
                    <span className="flex items-center gap-1 text-[11px] text-[var(--text-muted)]">
                      <Clock className="w-3 h-3" /> v{test.version}
                    </span>
                  </div>
                </div>
              </div>
              <button
                onClick={() => handleRun(test)}
                disabled={runningId === test.id}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-[var(--radius-sm)] bg-[var(--accent)] text-white text-[11px] font-semibold hover:bg-[var(--accent-hover)] disabled:opacity-50 shrink-0"
              >
                <PlayCircle className="w-3.5 h-3.5" />
                {runningId === test.id ? "Starting" : "Run"}
              </button>
              <button
                onClick={() => openRefine(test)}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-card)] text-[11px] font-semibold hover:bg-[var(--bg-hover)] shrink-0"
              >
                <PencilLine className="w-3.5 h-3.5" />
                Refine
              </button>
            </div>
          ))}
        </div>
      )}

      {selected && (
        <div className="fixed inset-0 z-40 bg-black/25" onClick={() => setSelected(null)}>
          <div className="absolute right-0 top-0 h-full w-full max-w-[520px] bg-[var(--bg-card)] border-l border-[var(--border)] shadow-xl p-5 overflow-y-auto" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <h2 className="text-base font-bold truncate">{selected.title}</h2>
                <p className="text-[12px] text-[var(--text-muted)] mt-0.5">v{selected.version} - {(selected.type || "ui").toUpperCase()} - {selected.feature || "No feature"}</p>
              </div>
              <button onClick={() => setSelected(null)} className="text-[12px] font-semibold text-[var(--text-muted)] hover:text-[var(--text-primary)]">Close</button>
            </div>

            <div className="mt-5 rounded-lg border border-[var(--border)] bg-[var(--bg-subtle)] p-4">
              <label className="text-[11px] font-semibold uppercase tracking-wider text-[var(--text-muted)]">Refinement Prompt</label>
              <textarea
                value={refinePrompt}
                onChange={(e) => setRefinePrompt(e.target.value)}
                className="input mt-2 min-h-24 resize-none"
                placeholder="Add edge case, improve assertion, cover validation error..."
              />
              <button onClick={submitRefinement} disabled={refining || !refinePrompt.trim()} className="mt-3 inline-flex items-center gap-1.5 px-3 py-2 rounded-[var(--radius-sm)] bg-[var(--accent)] text-white text-[12px] font-semibold hover:bg-[var(--accent-hover)] disabled:opacity-50">
                <Sparkles className="w-3.5 h-3.5" />
                {refining ? "Creating proposal" : "Create Proposal"}
              </button>
            </div>

            <div className="mt-5 space-y-3">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-bold">Change Proposals</h3>
                <span className="text-[11px] text-[var(--text-muted)]">{proposals.length} total</span>
              </div>
              {proposals.length === 0 ? (
                <p className="text-[12px] text-[var(--text-muted)]">No proposals for this test yet.</p>
              ) : proposals.map((proposal) => (
                <div key={proposal.id} className="rounded-lg border border-[var(--border)] bg-[var(--bg-card)] p-4">
                  <div className="flex items-center justify-between gap-3">
                    <span className={cn("px-1.5 py-0.5 rounded text-[10px] font-semibold", proposal.status === "pending" ? "bg-[var(--warning-bg)] text-[var(--warning)]" : "bg-[var(--success-bg)] text-[var(--success)]")}>{proposal.status}</span>
                    <span className="text-[10px] text-[var(--text-muted)]">{new Date(proposal.created_at).toLocaleString()}</span>
                  </div>
                  <p className="text-[12px] text-[var(--text-secondary)] mt-3">{proposal.rationale}</p>
                  <div className="grid grid-cols-2 gap-2 mt-3 text-[11px]">
                    <DiffStat label="Steps" before={proposal.original.steps.length} after={proposal.proposed.steps.length} />
                    <DiffStat label="Assertions" before={proposal.original.assertions.length} after={proposal.proposed.assertions.length} />
                  </div>
                  {proposal.status === "pending" && (
                    <button onClick={() => approveProposal(proposal)} disabled={approvingId === proposal.id} className="mt-3 inline-flex items-center gap-1.5 px-3 py-1.5 rounded-[var(--radius-sm)] bg-[var(--success)] text-white text-[11px] font-semibold disabled:opacity-50">
                      <CheckCircle2 className="w-3.5 h-3.5" />
                      {approvingId === proposal.id ? "Approving" : "Approve"}
                    </button>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function DiffStat({ label, before, after }: { label: string; before: number; after: number }) {
  return (
    <div className="rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-subtle)] px-2 py-1.5">
      <span className="text-[var(--text-muted)]">{label}</span>
      <span className="ml-2 font-semibold text-[var(--text-primary)]">{before}{" -> "}{after}</span>
    </div>
  );
}

function TabButton({ active, onClick, icon, label }: { active: boolean; onClick: () => void; icon: ReactNode; label: string }) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "flex items-center gap-2 px-4 py-2.5 text-[13px] font-medium border-b-2 -mb-px transition-colors",
        active ? "border-[var(--accent)] text-[var(--accent)]" : "border-transparent text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
      )}
    >
      {icon}
      {label}
    </button>
  );
}
