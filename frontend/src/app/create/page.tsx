"use client";

import { useMemo, useState } from "react";
import type { ReactNode } from "react";
import { useRouter } from "next/navigation";
import { createRun } from "@/lib/api";
import { Section } from "@/components/ui/section";
import {
  ArrowRight,
  CheckCircle2,
  ChevronRight,
  FileText,
  FolderOpen,
  GitBranch,
  KeyRound,
  ListChecks,
  PlayCircle,
  Radar,
  Settings,
  TerminalSquare,
  Upload,
} from "lucide-react";
import { cn } from "@/lib/utils";

type Step = 1 | 2 | 3 | 4;
type TestType = "ui" | "api";

export default function CreateTestPage() {
  const router = useRouter();
  const [step, setStep] = useState<Step>(1);
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    name: "New Autonomous Test",
    testType: "ui" as TestType,
    projectPath: "",
    prd: "",
    requirements: "",
    apiDocs: "",
    authType: "none",
    credentials: "",
    focusHints: "",
    skipHints: "",
    mode: "simple",
  });

  const featureMap = useMemo(
    () => deriveFeatureMap(formData.prd || formData.requirements),
    [formData.prd, formData.requirements]
  );

  const canContinue = step === 1
    ? formData.name.trim().length > 0
    : step === 2
      ? formData.projectPath.trim().length > 0
      : true;

  const readUpload = async (file?: File) => {
    if (!file) return;
    const text = await file.text();
    setFormData((prev) => ({ ...prev, prd: text }));
  };

  const handleRun = async () => {
    setLoading(true);
    try {
      const { run_id } = await createRun({
        project_path: formData.projectPath,
        requirements: formData.requirements || formData.prd,
        mode: formData.mode,
        test_type: formData.testType,
        prd: formData.prd,
        api_docs: formData.apiDocs,
        auth_type: formData.authType,
        credentials: formData.credentials,
        focus_hints: formData.focusHints,
        skip_hints: formData.skipHints,
      });
      router.push(`/runs/${run_id}`);
    } catch (err) {
      console.error("Failed to create run", err);
      router.push("/runs");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-5xl mx-auto space-y-5 pb-12">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold">Create Tests</h1>
          <p className="text-[13px] text-[var(--text-secondary)] mt-1 max-w-2xl">
            Build a spec-driven test run from product intent, live app configuration, exploration scope, and an editable execution plan.
          </p>
        </div>
        <div className="hidden sm:flex items-center gap-1.5 rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-card)] px-2.5 py-1.5 text-[11px] text-[var(--text-muted)]">
          <Radar className="w-3.5 h-3.5" />
          Self-hosted lifecycle
        </div>
      </div>

      <Stepper current={step} />

      {step === 1 && (
        <Section title="Project Setup and PRD">
          <div className="grid grid-cols-1 lg:grid-cols-[1.1fr_0.9fr] gap-5">
            <div className="space-y-4">
              <Field label="Project name" icon={<FileText className="w-3.5 h-3.5" />}>
                <input
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="input"
                  placeholder="Customer portal regression"
                />
              </Field>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <TypeButton
                  active={formData.testType === "ui"}
                  icon={<Radar className="w-4 h-4" />}
                  title="UI testing"
                  desc="Browser flows, forms, auth, visual states"
                  onClick={() => setFormData({ ...formData, testType: "ui" })}
                />
                <TypeButton
                  active={formData.testType === "api"}
                  icon={<TerminalSquare className="w-4 h-4" />}
                  title="API testing"
                  desc="Endpoints, schema, auth, data integrity"
                  onClick={() => setFormData({ ...formData, testType: "api" })}
                />
              </div>

              <Field label="PRD or product spec" icon={<Upload className="w-3.5 h-3.5" />}>
                <div className="space-y-2">
                  <textarea
                    value={formData.prd}
                    onChange={(e) => setFormData({ ...formData, prd: e.target.value })}
                    rows={8}
                    className="input resize-none"
                    placeholder="Paste product requirements, user journeys, acceptance criteria, or endpoint behavior here."
                  />
                  <label className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-subtle)] text-[12px] font-semibold text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] cursor-pointer">
                    <Upload className="w-3.5 h-3.5" />
                    Upload .md or .txt
                    <input type="file" accept=".md,.txt,.text" className="hidden" onChange={(e) => readUpload(e.target.files?.[0])} />
                  </label>
                </div>
              </Field>
            </div>

            <FeatureMapPreview features={featureMap} source={formData.prd ? "PRD" : "requirements"} />
          </div>
        </Section>
      )}

      {step === 2 && (
        <Section title="Configuration">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
            <div className="space-y-4">
              <Field label={formData.testType === "ui" ? "Starting URL" : "Base URL"} icon={<FolderOpen className="w-3.5 h-3.5" />}>
                <input
                  value={formData.projectPath}
                  onChange={(e) => setFormData({ ...formData, projectPath: e.target.value })}
                  className="input"
                  placeholder={formData.testType === "ui" ? "https://app.example.com" : "https://api.example.com"}
                />
              </Field>

              <Field label="Test account or auth notes" icon={<KeyRound className="w-3.5 h-3.5" />}>
                <textarea
                  value={formData.credentials}
                  onChange={(e) => setFormData({ ...formData, credentials: e.target.value })}
                  rows={4}
                  className="input resize-none"
                  placeholder="User role, login path, test credentials storage notes, token instructions, or auth constraints."
                />
              </Field>

              <Field label="Auth type" icon={<Settings className="w-3.5 h-3.5" />}>
                <select
                  value={formData.authType}
                  onChange={(e) => setFormData({ ...formData, authType: e.target.value })}
                  className="input"
                >
                  <option value="none">None</option>
                  <option value="login">Username/password login</option>
                  <option value="bearer">Bearer token</option>
                  <option value="api_key">API key</option>
                  <option value="custom">Custom</option>
                </select>
              </Field>
            </div>

            <div className="space-y-4">
              {formData.testType === "api" && (
                <Field label="OpenAPI, Swagger, Postman, or API docs" icon={<TerminalSquare className="w-3.5 h-3.5" />}>
                  <textarea
                    value={formData.apiDocs}
                    onChange={(e) => setFormData({ ...formData, apiDocs: e.target.value })}
                    rows={8}
                    className="input resize-none"
                    placeholder="Paste endpoint docs, OpenAPI fragments, Postman notes, or expected schemas."
                  />
                </Field>
              )}

              <Field label="Focus hints" icon={<ListChecks className="w-3.5 h-3.5" />}>
                <textarea
                  value={formData.focusHints}
                  onChange={(e) => setFormData({ ...formData, focusHints: e.target.value })}
                  rows={4}
                  className="input resize-none"
                  placeholder="Flows, pages, endpoints, roles, or edge cases to prioritize."
                />
              </Field>

              <Field label="Skip hints" icon={<GitBranch className="w-3.5 h-3.5" />}>
                <textarea
                  value={formData.skipHints}
                  onChange={(e) => setFormData({ ...formData, skipHints: e.target.value })}
                  rows={4}
                  className="input resize-none"
                  placeholder="Dangerous actions, external providers, destructive endpoints, or unstable areas to avoid."
                />
              </Field>
            </div>
          </div>
        </Section>
      )}

      {step === 3 && (
        <Section title={formData.testType === "ui" ? "Exploration Scope" : "Discovery Scope"}>
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-3">
            {featureMap.map((feature, index) => (
              <div key={feature.name} className="rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-subtle)] p-4">
                <div className="flex items-center gap-2 mb-3">
                  <span className="w-6 h-6 rounded-[var(--radius-sm)] bg-[var(--accent-bg)] text-[var(--accent)] flex items-center justify-center text-[11px] font-bold">
                    {index + 1}
                  </span>
                  <h2 className="text-[13px] font-semibold text-[var(--text-primary)]">{feature.name}</h2>
                </div>
                <ul className="space-y-2">
                  {feature.useCases.map((useCase) => (
                    <li key={useCase} className="flex items-start gap-2 text-[12px] text-[var(--text-secondary)]">
                      <CheckCircle2 className="w-3.5 h-3.5 text-[var(--success)] mt-0.5 shrink-0" />
                      <span>{useCase}</span>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
          <div className="mt-4 rounded-[var(--radius-sm)] border border-[var(--info)]/15 bg-[var(--info-bg)] p-3">
            <p className="text-[12px] text-[var(--text-secondary)]">
              This self-hosted build stores the extracted feature map and uses it to ground the generated plan. Deep crawling and endpoint probing can be added behind this same step.
            </p>
          </div>
        </Section>
      )}

      {step === 4 && (
        <Section title="Plan Review">
          <div className="grid grid-cols-1 lg:grid-cols-[1fr_0.85fr] gap-5">
            <div className="rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-subtle)] overflow-hidden">
              <SummaryRow label="Type" value={formData.testType === "ui" ? "UI browser test" : "API backend test"} />
              <SummaryRow label="Target" value={formData.projectPath || "Not configured"} mono />
              <SummaryRow label="Plan source" value={formData.prd ? "PRD/spec + live configuration" : "Requirements + live configuration"} />
              <SummaryRow label="Focus" value={formData.focusHints || "Explore primary product behavior"} />
              <SummaryRow label="Safety skips" value={formData.skipHints || "No skip hints provided"} />
            </div>

            <div className="space-y-3">
              <Field label="Run instructions" icon={<FileText className="w-3.5 h-3.5" />}>
                <textarea
                  value={formData.requirements}
                  onChange={(e) => setFormData({ ...formData, requirements: e.target.value })}
                  rows={7}
                  className="input resize-none"
                  placeholder="Optional final instruction. Example: prioritize checkout, auth reset, and billing plan changes."
                />
              </Field>
              <Field label="Execution mode" icon={<Settings className="w-3.5 h-3.5" />}>
                <select value={formData.mode} onChange={(e) => setFormData({ ...formData, mode: e.target.value })} className="input">
                  <option value="simple">Standard generate and run</option>
                  <option value="advanced">Advanced multi-agent</option>
                </select>
              </Field>
            </div>
          </div>
        </Section>
      )}

      <div className="flex items-center justify-between">
        <button
          onClick={() => setStep((prev) => Math.max(1, prev - 1) as Step)}
          disabled={step === 1 || loading}
          className="px-4 py-2 text-[13px] font-semibold text-[var(--text-secondary)] hover:text-[var(--text-primary)] disabled:opacity-40 transition-colors"
        >
          Back
        </button>
        {step < 4 ? (
          <button
            onClick={() => setStep((prev) => Math.min(4, prev + 1) as Step)}
            disabled={!canContinue || loading}
            className="inline-flex items-center gap-1.5 px-4 py-2 rounded-[var(--radius-sm)] bg-[var(--accent)] text-white text-[13px] font-semibold hover:bg-[var(--accent-hover)] transition-colors disabled:opacity-50"
          >
            Continue <ArrowRight className="w-4 h-4" />
          </button>
        ) : (
          <button
            onClick={handleRun}
            disabled={loading || !formData.projectPath}
            className="inline-flex items-center gap-1.5 px-5 py-2 rounded-[var(--radius-sm)] bg-[var(--success)] text-white text-[13px] font-bold hover:bg-[#047857] transition-colors shadow-sm disabled:opacity-50"
          >
            {loading ? "Starting..." : <><PlayCircle className="w-4 h-4" /> Generate and Run</>}
          </button>
        )}
      </div>
    </div>
  );
}

function Stepper({ current }: { current: Step }) {
  const steps = [
    { id: 1, label: "Spec" },
    { id: 2, label: "Configure" },
    { id: 3, label: "Explore" },
    { id: 4, label: "Review" },
  ] as const;

  return (
    <div className="flex flex-wrap items-center gap-2 text-[12px] font-medium">
      {steps.map((item, index) => (
        <div key={item.id} className="flex items-center gap-2">
          <div className={cn("flex items-center gap-1.5", current >= item.id ? "text-[var(--accent)]" : "text-[var(--text-muted)]")}>
            <span className={cn(
              "w-5 h-5 rounded-full flex items-center justify-center text-[10px]",
              current >= item.id ? "bg-[var(--accent-bg)]" : "bg-[var(--bg-subtle)]"
            )}>
              {item.id}
            </span>
            {item.label}
          </div>
          {index < steps.length - 1 && <ChevronRight className="w-3.5 h-3.5 text-[var(--text-muted)]" />}
        </div>
      ))}
    </div>
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

function TypeButton({
  active,
  icon,
  title,
  desc,
  onClick,
}: {
  active: boolean;
  icon: ReactNode;
  title: string;
  desc: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "text-left rounded-[var(--radius-sm)] border p-3 transition-colors",
        active ? "border-[var(--accent)]/40 bg-[var(--accent-bg)]" : "border-[var(--border)] bg-[var(--bg-subtle)] hover:bg-[var(--bg-hover)]"
      )}
    >
      <span className={cn("flex items-center gap-2 text-[13px] font-semibold", active ? "text-[var(--accent)]" : "text-[var(--text-primary)]")}>
        {icon}
        {title}
      </span>
      <span className="block mt-1 text-[11px] text-[var(--text-muted)]">{desc}</span>
    </button>
  );
}

function FeatureMapPreview({ features, source }: { features: FeaturePreview[]; source: string }) {
  return (
    <div className="rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-subtle)] p-4">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <GitBranch className="w-4 h-4 text-[var(--accent)]" />
          <h2 className="text-[13px] font-semibold">Feature Map</h2>
        </div>
        <span className="text-[10px] font-semibold uppercase text-[var(--text-muted)]">{source}</span>
      </div>
      <div className="space-y-2">
        {features.map((feature) => (
          <div key={feature.name} className="rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--bg-card)] p-3">
            <p className="text-[12px] font-semibold text-[var(--text-primary)]">{feature.name}</p>
            <p className="text-[11px] text-[var(--text-muted)] mt-1">{feature.useCases[0]}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

function SummaryRow({ label, value, mono = false }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="px-4 py-3 border-b border-[var(--border)] last:border-b-0">
      <div className="text-[10px] font-semibold text-[var(--text-muted)] uppercase tracking-wider mb-1">{label}</div>
      <div className={cn("text-[13px] text-[var(--text-secondary)]", mono && "font-mono break-all")}>{value}</div>
    </div>
  );
}

type FeaturePreview = { name: string; useCases: string[] };

function deriveFeatureMap(source: string): FeaturePreview[] {
  const text = source.trim();
  if (!text) {
    return [
      { name: "Primary product flow", useCases: ["Create a baseline plan from the target URL and user-provided run instructions."] },
      { name: "Failure handling", useCases: ["Capture clear evidence, replay context, and next action for any failed step."] },
    ];
  }

  const lines = text
    .split("\n")
    .map((line) => line.replace(/^[-*#\d.\s]+/, "").trim())
    .filter((line) => line.length > 8)
    .slice(0, 5);

  const selected = lines.length > 0 ? lines : [text.slice(0, 140)];
  return selected.map((line, index) => ({
    name: line.length > 54 ? `${line.slice(0, 51)}...` : line || `Feature ${index + 1}`,
    useCases: [line.length > 120 ? `${line.slice(0, 117)}...` : line],
  }));
}
