"use client";

import { useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import Link from "next/link";
import {
  approveTestPlan,
  createProject,
  createRun,
  extractProjectFeatures,
  generateProjectTestPlan,
  getProjects,
  updateTestPlanCase,
  parseApiDocs,
  regenerateTestPlan,
  type DraftPlan,
  type Project,
} from "@/lib/api";
import { EmptyState, LoadingSkeleton, Section } from "@/components/ui/section";
import {
  ArrowRight,
  CheckCircle2,
  FileText,
  FolderOpen,
  GitBranch,
  KeyRound,
  PlayCircle,
  Plus,
  Radar,
  TerminalSquare,
  Upload,
} from "lucide-react";
import { cn } from "@/lib/utils";

export default function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [selected, setSelected] = useState<Project | null>(null);
  const [plan, setPlan] = useState<DraftPlan | null>(null);
  const [approving, setApproving] = useState(false);
  const [form, setForm] = useState({
    name: "",
    test_type: "ui",
    base_url: "",
    environment: "staging",
    spec: "",
    api_docs: "",
    auth_type: "none",
    credentials: "",
    focus_hints: "",
    skip_hints: "",
  });

  useEffect(() => {
    getProjects().then((items) => {
      setProjects(items);
      setSelected(items[0] || null);
    }).catch(() => {}).finally(() => setLoading(false));
  }, []);

  const selectedFeatures = selected?.feature_map?.features || [];
  const canCreate = form.name.trim() && form.base_url.trim();

  const handleUpload = async (file?: File) => {
    if (!file) return;
    const spec = await file.text();
    setForm((prev) => ({ ...prev, spec }));
  };

  const handleCreate = async () => {
    if (!canCreate) return;
    setCreating(true);
    try {
      const project = await createProject(form);
      setProjects((prev) => [project, ...prev]);
      setSelected(project);
      setForm({
        name: "",
        test_type: "ui",
        base_url: "",
        environment: "staging",
        spec: "",
        api_docs: "",
        auth_type: "none",
        credentials: "",
        focus_hints: "",
        skip_hints: "",
      });
    } finally {
      setCreating(false);
    }
  };

  const selectProject = (project: Project) => {
    setSelected(project);
    setPlan(null);
  };

  const handleExtract = async () => {
    if (!selected) return;
    const featureMap = await extractProjectFeatures(selected.id);
    const updated = { ...selected, feature_map: featureMap };
    setSelected(updated);
    setProjects((prev) => prev.map((p) => (p.id === selected.id ? updated : p)));
  };

  const handleGeneratePlan = async () => {
    if (!selected) return;
    if (selected.test_type === "api") {
      const draft = await parseApiDocs(selected.id);
      setPlan(draft);
    } else {
      const draft = await generateProjectTestPlan(selected.id);
      setPlan(draft);
    }
  };

  const handleRegeneratePlan = async () => {
    if (!plan) return;
    const draft = await regenerateTestPlan(plan.id);
    setPlan(draft);
  };

  const handleToggleCase = async (draftCase: DraftPlan["cases"][number]) => {
    if (!plan) return;
    const updated = await updateTestPlanCase(plan.id, draftCase.id, { ...draftCase, enabled: !draftCase.enabled });
    setPlan(updated);
  };

  const handleApprovePlan = async () => {
    if (!plan) return;
    setApproving(true);
    try {
      await approveTestPlan(plan.id);
      setPlan({ ...plan, status: "approved" });
    } finally {
      setApproving(false);
    }
  };

  const handleRun = async () => {
    if (!selected) return;
    const { run_id } = await createRun({
      project_path: selected.base_url,
      requirements: selected.focus_hints || selected.spec,
      test_type: selected.test_type,
      prd: selected.spec,
      api_docs: selected.api_docs,
      auth_type: selected.auth_type,
      credentials: selected.credentials,
      focus_hints: selected.focus_hints,
      skip_hints: selected.skip_hints,
    });
    window.location.href = `/runs/${run_id}`;
  };

  if (loading) return <LoadingSkeleton rows={8} />;

  return (
    <div className="space-y-5 pb-10">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-[var(--text-primary)]">Projects</h1>
          <p className="text-sm text-[var(--text-secondary)] mt-0.5">Create self-hosted test projects from PRDs, URLs, API docs, and auth context</p>
        </div>
        <Link href="/create" className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-card)] text-[12px] font-semibold text-[var(--text-secondary)] hover:bg-[var(--bg-hover)]">
          Guided run <ArrowRight className="w-3.5 h-3.5" />
        </Link>
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-[420px_1fr] gap-5">
        <Section title="New Project" action={<Plus className="w-3.5 h-3.5 text-[var(--accent)]" />}>
          <div className="space-y-4">
            <Field label="Project name" icon={<FileText className="w-3.5 h-3.5" />}>
              <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} className="input" placeholder="Customer portal" />
            </Field>

            <div className="grid grid-cols-2 gap-2">
              <TypeButton active={form.test_type === "ui"} onClick={() => setForm({ ...form, test_type: "ui" })} icon={<Radar className="w-4 h-4" />} label="UI" />
              <TypeButton active={form.test_type === "api"} onClick={() => setForm({ ...form, test_type: "api" })} icon={<TerminalSquare className="w-4 h-4" />} label="API" />
            </div>

            <Field label={form.test_type === "ui" ? "Starting URL" : "Base URL"} icon={<FolderOpen className="w-3.5 h-3.5" />}>
              <input value={form.base_url} onChange={(e) => setForm({ ...form, base_url: e.target.value })} className="input" placeholder="https://app.example.com" />
            </Field>

            <Field label="PRD / specification" icon={<Upload className="w-3.5 h-3.5" />}>
              <textarea value={form.spec} onChange={(e) => setForm({ ...form, spec: e.target.value })} rows={6} className="input resize-none" placeholder="Paste PRD, user journeys, or acceptance criteria." />
              <label className="mt-2 inline-flex items-center gap-1.5 px-3 py-1.5 rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-subtle)] text-[12px] font-semibold text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] cursor-pointer">
                <Upload className="w-3.5 h-3.5" /> Upload .md/.txt
                <input type="file" accept=".md,.txt,.text" className="hidden" onChange={(e) => handleUpload(e.target.files?.[0])} />
              </label>
            </Field>

            {form.test_type === "api" && (
              <Field label="API docs" icon={<TerminalSquare className="w-3.5 h-3.5" />}>
                <textarea value={form.api_docs} onChange={(e) => setForm({ ...form, api_docs: e.target.value })} rows={4} className="input resize-none" placeholder="OpenAPI, Postman notes, endpoint docs, or expected schemas." />
              </Field>
            )}

            <Field label="Auth notes" icon={<KeyRound className="w-3.5 h-3.5" />}>
              <textarea value={form.credentials} onChange={(e) => setForm({ ...form, credentials: e.target.value })} rows={3} className="input resize-none" placeholder="Use secret reference, seeded test account, token flow, or login notes." />
            </Field>

            <div className="grid grid-cols-2 gap-2">
              <Field label="Focus" icon={<CheckCircle2 className="w-3.5 h-3.5" />}>
                <textarea value={form.focus_hints} onChange={(e) => setForm({ ...form, focus_hints: e.target.value })} rows={3} className="input resize-none" placeholder="Prioritize these flows." />
              </Field>
              <Field label="Skip" icon={<GitBranch className="w-3.5 h-3.5" />}>
                <textarea value={form.skip_hints} onChange={(e) => setForm({ ...form, skip_hints: e.target.value })} rows={3} className="input resize-none" placeholder="Avoid destructive areas." />
              </Field>
            </div>

            <button
              onClick={handleCreate}
              disabled={!canCreate || creating}
              className="inline-flex items-center justify-center gap-1.5 w-full px-4 py-2 rounded-[var(--radius-sm)] bg-[var(--accent)] text-white text-[13px] font-semibold hover:bg-[var(--accent-hover)] disabled:opacity-50"
            >
              {creating ? "Creating..." : "Create and Extract Features"}
            </button>
          </div>
        </Section>

        <div className="space-y-5">
          {projects.length === 0 ? (
            <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)]">
              <EmptyState
                icon={<FolderOpen className="w-6 h-6" />}
                title="No projects registered"
                description="Create a project from a PRD or API docs to start building a reusable test inventory."
              />
            </div>
          ) : (
            <ProjectWorkspace
              projects={projects}
              selected={selected}
              selectedFeatures={selectedFeatures}
              plan={plan}
              approving={approving}
              onSelect={selectProject}
              onExtract={handleExtract}
              onGeneratePlan={handleGeneratePlan}
              onRegeneratePlan={handleRegeneratePlan}
              onToggleCase={handleToggleCase}
              onApprovePlan={handleApprovePlan}
              onRun={handleRun}
            />
          )}
        </div>
      </div>
    </div>
  );
}

function ProjectWorkspace({
  projects,
  selected,
  selectedFeatures,
  plan,
  approving,
  onSelect,
  onExtract,
  onGeneratePlan,
  onRegeneratePlan,
  onToggleCase,
  onApprovePlan,
  onRun,
}: {
  projects: Project[];
  selected: Project | null;
  selectedFeatures: NonNullable<Project["feature_map"]>["features"];
  plan: DraftPlan | null;
  approving: boolean;
  onSelect: (project: Project) => void;
  onExtract: () => void;
  onGeneratePlan: () => void;
  onRegeneratePlan: () => void;
  onToggleCase: (draftCase: DraftPlan["cases"][number]) => void;
  onApprovePlan: () => void;
  onRun: () => void;
}) {
  const stats = useMemo(() => ({
    total: projects.length,
    ui: projects.filter((p) => p.test_type === "ui").length,
    api: projects.filter((p) => p.test_type === "api").length,
  }), [projects]);

  return (
    <>
      <div className="grid grid-cols-3 gap-3">
        <MiniStat label="Projects" value={stats.total} />
        <MiniStat label="UI" value={stats.ui} />
        <MiniStat label="API" value={stats.api} />
      </div>

      <Section title="Project Inventory" action={<span className="text-[11px] text-[var(--text-muted)]">{projects.length} total</span>}>
        <div className="space-y-2">
          {projects.map((project) => (
            <button
              key={project.id}
              onClick={() => onSelect(project)}
              className={cn(
                "w-full text-left rounded-[var(--radius-sm)] border p-3 transition-colors",
                selected?.id === project.id ? "border-[var(--accent)]/40 bg-[var(--accent-bg)]" : "border-[var(--border)] bg-[var(--bg-subtle)] hover:bg-[var(--bg-hover)]"
              )}
            >
              <div className="flex items-center justify-between gap-3">
                <div className="min-w-0">
                  <p className="text-[13px] font-semibold text-[var(--text-primary)] truncate">{project.name}</p>
                  <p className="text-[11px] text-[var(--text-muted)] truncate">{project.base_url}</p>
                </div>
                <span className="rounded px-1.5 py-0.5 text-[10px] font-bold bg-[var(--bg-card)] border border-[var(--border)] text-[var(--text-muted)]">
                  {(project.test_type || "ui").toUpperCase()}
                </span>
              </div>
            </button>
          ))}
        </div>
      </Section>

      {selected && (
        <Section
          title={selected.test_type === "api" ? "API Docs" : "Feature Map"}
          action={
            <div className="flex items-center gap-2">
              {selected.test_type !== "api" && (
                <button onClick={onExtract} className="text-[11px] font-semibold text-[var(--accent)] hover:underline">Regenerate Map</button>
              )}
              <button onClick={onGeneratePlan} className="text-[11px] font-semibold text-[var(--accent)] hover:underline">
                {selected.test_type === "api" ? "Parse API & Generate Plan" : "Generate Plan"}
              </button>
              <button onClick={onRun} className="inline-flex items-center gap-1 px-2.5 py-1 rounded-[var(--radius-sm)] bg-[var(--success)] text-white text-[11px] font-semibold">
                <PlayCircle className="w-3 h-3" /> Run
              </button>
            </div>
          }
        >
          {selected.test_type === "api" ? (
             <div className="rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-subtle)] p-4 text-[13px] text-[var(--text-secondary)] whitespace-pre-wrap font-mono">
               {selected.api_docs ? selected.api_docs : <EmptyState icon={<GitBranch className="w-6 h-6" />} title="No API Docs" description="Add API documentation during creation to parse endpoints." />}
             </div>
          ) : selectedFeatures.length === 0 ? (
            <EmptyState icon={<GitBranch className="w-6 h-6" />} title="No feature map yet" description="Regenerate extraction after adding a PRD/spec." />
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {selectedFeatures.map((feature, index) => (
                <div key={`${feature.name}-${index}`} className="rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-subtle)] p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <span className="w-6 h-6 rounded-[var(--radius-sm)] bg-[var(--accent-bg)] text-[var(--accent)] flex items-center justify-center text-[11px] font-bold">{index + 1}</span>
                    <h3 className="text-[13px] font-semibold text-[var(--text-primary)]">{feature.name}</h3>
                  </div>
                  <ul className="space-y-1.5">
                    {feature.use_cases.map((useCase) => (
                      <li key={useCase} className="flex items-start gap-2 text-[12px] text-[var(--text-secondary)]">
                        <CheckCircle2 className="w-3.5 h-3.5 text-[var(--success)] mt-0.5 shrink-0" />
                        <span>{useCase}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              ))}
            </div>
          )}
        </Section>
      )}

      {selected && plan && (
        <Section
          title="Plan Review"
          action={
            <div className="flex items-center gap-3">
              {plan.status !== "approved" && (
                <button onClick={onRegeneratePlan} className="text-[11px] font-semibold text-[var(--accent)] hover:underline">
                  Regenerate Plan
                </button>
              )}
              <button
                onClick={onApprovePlan}
                disabled={plan.status === "approved" || approving}
                className="inline-flex items-center gap-1 px-2.5 py-1 rounded-[var(--radius-sm)] bg-[var(--accent)] text-white text-[11px] font-semibold disabled:opacity-50"
              >
                <CheckCircle2 className="w-3 h-3" /> {plan.status === "approved" ? "Approved" : approving ? "Approving..." : "Approve"}
              </button>
            </div>
          }
        >
          <div className="space-y-2">
            {plan.cases.map((draftCase) => (
              <div key={draftCase.id} className="rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-subtle)] p-3">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => onToggleCase(draftCase)}
                        className={cn(
                          "w-4 h-4 rounded border flex items-center justify-center",
                          draftCase.enabled ? "bg-[var(--accent)] border-[var(--accent)] text-white" : "border-[var(--border-strong)] bg-[var(--bg-card)]"
                        )}
                        title={draftCase.enabled ? "Disable case" : "Enable case"}
                      >
                        {draftCase.enabled && <CheckCircle2 className="w-3 h-3" />}
                      </button>
                      <h3 className="text-[13px] font-semibold text-[var(--text-primary)] truncate">{draftCase.title}</h3>
                    </div>
                    <p className="text-[11px] text-[var(--text-muted)] mt-1">{draftCase.feature} · {draftCase.priority} · {Math.round(draftCase.confidence * 100)}% confidence</p>
                  </div>
                  <span className="rounded px-1.5 py-0.5 text-[10px] font-bold bg-[var(--bg-card)] border border-[var(--border)] text-[var(--text-muted)]">
                    {draftCase.type.toUpperCase()}
                  </span>
                </div>
                <div className="mt-3 grid grid-cols-1 md:grid-cols-2 gap-3">
                  <PlanList label="Steps" items={draftCase.steps} />
                  <PlanList label="Assertions" items={draftCase.assertions} />
                </div>
              </div>
            ))}
          </div>
        </Section>
      )}
    </>
  );
}

function Field({ label, icon, children }: { label: string; icon: ReactNode; children: ReactNode }) {
  return (
    <label className="block">
      <span className="flex items-center gap-1.5 text-[12px] font-semibold mb-1.5 text-[var(--text-primary)]">
        <span className="text-[var(--text-muted)]">{icon}</span>
        {label}
      </span>
      {children}
    </label>
  );
}

function TypeButton({ active, onClick, icon, label }: { active: boolean; onClick: () => void; icon: ReactNode; label: string }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-[var(--radius-sm)] border px-3 py-2 text-[12px] font-semibold",
        active ? "border-[var(--accent)]/40 bg-[var(--accent-bg)] text-[var(--accent)]" : "border-[var(--border)] bg-[var(--bg-subtle)] text-[var(--text-secondary)] hover:bg-[var(--bg-hover)]"
      )}
    >
      {icon}
      {label}
    </button>
  );
}

function MiniStat({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-card)] p-3">
      <p className="text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)]">{label}</p>
      <p className="text-lg font-bold text-[var(--text-primary)] mt-1">{value}</p>
    </div>
  );
}

function PlanList({ label, items }: { label: string; items: string[] }) {
  return (
    <div>
      <p className="text-[10px] font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-1">{label}</p>
      <ul className="space-y-1">
        {items.map((item) => (
          <li key={item} className="text-[11px] text-[var(--text-secondary)]">{item}</li>
        ))}
      </ul>
    </div>
  );
}
