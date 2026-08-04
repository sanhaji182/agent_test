"use client";

import React, { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import {
  getTestCase,
  updateTestCase,
  runTestCase,
  type TestCase,
} from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import {
  ArrowLeft,
  ArrowUpDown,
  CheckCircle2,
  ChevronDown,
  ChevronUp,
  Clock,
  Keyboard,
  List,
  Monitor,
  MousePointer,
  MousePointerClick,
  Navigation,
  Play,
  Plus,
  Save,
  Trash2,
  Type,
} from "lucide-react";
import { cn } from "@/lib/utils";

/* ========== Browser action schema ========== */

interface BrowserAction {
  action: string;
  url?: string;
  selector?: string;
  value?: string;
  key?: string;
  assert?: string;
  text?: string;
  y?: number;
  ms?: number;
}

const ACTION_TYPES = [
  "goto",
  "fill",
  "click",
  "scroll",
  "wait",
  "hover",
  "select",
  "press",
  "assert",
  "screenshot",
] as const;

const ASSERT_TYPES = [
  "visible",
  "hidden",
  "text_contains",
  "url_contains",
  "title_contains",
  "count",
  "attribute",
] as const;

const actionIcons: Record<string, React.ElementType> = {
  goto: Navigation,
  fill: Type,
  click: MousePointerClick,
  scroll: ArrowUpDown,
  wait: Clock,
  hover: MousePointer,
  select: List,
  press: Keyboard,
  assert: CheckCircle2,
  screenshot: Monitor,
};

function defaultAction(action: string): BrowserAction {
  switch (action) {
    case "goto":
      return { action: "goto", url: "" };
    case "fill":
      return { action: "fill", selector: "", value: "" };
    case "wait":
      return { action: "wait", ms: 1000 };
    case "press":
      return { action: "press", key: "" };
    case "scroll":
      return { action: "scroll", y: 0 };
    case "assert":
      return { action: "assert", selector: "", assert: "visible", text: "" };
    case "select":
      return { action: "select", selector: "", value: "" };
    default:
      return { action: "click", selector: "" };
  }
}

/** Parse executable_content into BrowserAction[] or return null if empty/unparseable. */
function parseActions(raw?: string): BrowserAction[] | null {
  if (!raw || !raw.trim()) return null;
  try {
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return null;
    return parsed as BrowserAction[];
  } catch {
    return null;
  }
}

/* ========== Page ========== */

export default function EditTestCasePage() {
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;

  const [test, setTest] = useState<TestCase | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  // Section 1 — Info dasar
  const [title, setTitle] = useState("");
  const [priority, setPriority] = useState("medium");
  const [tagsInput, setTagsInput] = useState("");

  // Section 2 — Langkah Test
  const [steps, setSteps] = useState<string[]>([]);

  // Section 3 — Aksi Browser (deterministic)
  const [actions, setActions] = useState<BrowserAction[] | null>(null);

  // Feedback
  const [saving, setSaving] = useState(false);
  const [saveRunning, setSaveRunning] = useState(false);
  const [savedMsg, setSavedMsg] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  useEffect(() => {
    getTestCase(id)
      .then((t) => {
        setTest(t);
        setTitle(t.title || "");
        setPriority(t.priority || "medium");
        setTagsInput((t.tags || []).join(", "));
        setSteps(t.steps && t.steps.length > 0 ? [...t.steps] : [""]);
        setActions(parseActions(t.executable_content));
      })
      .catch((e) => setLoadError(e instanceof Error ? e.message : "Gagal memuat test case"))
      .finally(() => setLoading(false));
  }, [id]);

  // Auto-hide "Tersimpan ✓" after 3s
  useEffect(() => {
    if (!savedMsg) return;
    const timer = setTimeout(() => setSavedMsg(false), 3000);
    return () => clearTimeout(timer);
  }, [savedMsg]);

  const deterministic = useMemo(() => !!(test?.executable_content && test.executable_content.trim()), [test]);

  /* ----- Steps helpers ----- */
  const updateStep = (i: number, value: string) =>
    setSteps((prev) => prev.map((s, idx) => (idx === i ? value : s)));
  const addStep = () => setSteps((prev) => [...prev, ""]);
  const removeStep = (i: number) => setSteps((prev) => prev.filter((_, idx) => idx !== i));
  const moveStep = (i: number, dir: -1 | 1) =>
    setSteps((prev) => {
      const next = [...prev];
      const j = i + dir;
      if (j < 0 || j >= next.length) return prev;
      [next[i], next[j]] = [next[j], next[i]];
      return next;
    });

  /* ----- Actions helpers ----- */
  const updateAction = (i: number, patch: Partial<BrowserAction>) =>
    setActions((prev) => (prev ? prev.map((a, idx) => (idx === i ? { ...a, ...patch } : a)) : prev));
  const setActionType = (i: number, type: string) =>
    setActions((prev) => {
      if (!prev) return prev;
      return prev.map((a, idx) => {
        if (idx !== i) return a;
        // Preserve compatible fields from the old action; fields irrelevant to the new type are dropped.
        const fresh = defaultAction(type);
        const merged: BrowserAction = { ...fresh, action: type };
        if (fresh.url !== undefined && a.url !== undefined) merged.url = a.url;
        if (fresh.selector !== undefined && a.selector !== undefined) merged.selector = a.selector;
        if (fresh.value !== undefined && a.value !== undefined) merged.value = a.value;
        if (fresh.key !== undefined && a.key !== undefined) merged.key = a.key;
        if (fresh.assert !== undefined && a.assert !== undefined) merged.assert = a.assert;
        if (fresh.text !== undefined && a.text !== undefined) merged.text = a.text;
        if (fresh.y !== undefined && a.y !== undefined) merged.y = a.y;
        if (fresh.ms !== undefined && a.ms !== undefined) merged.ms = a.ms;
        return merged;
      });
    });
  const addAction = () => setActions((prev) => [...(prev || []), defaultAction("click")]);
  const removeAction = (i: number) =>
    setActions((prev) => (prev ? prev.filter((_, idx) => idx !== i) : prev));
  const moveAction = (i: number, dir: -1 | 1) =>
    setActions((prev) => {
      if (!prev) return prev;
      const next = [...prev];
      const j = i + dir;
      if (j < 0 || j >= next.length) return prev;
      [next[i], next[j]] = [next[j], next[i]];
      return next;
    });

  /* ----- Save ----- */
  const buildPayload = () => {
    const tags = tagsInput
      .split(",")
      .map((t) => t.trim())
      .filter(Boolean);
    const cleanSteps = steps.map((s) => s.trim()).filter((s) => s.length > 0);
    const payload: {
      title: string;
      priority: string;
      steps: string[];
      tags: string[];
      executable_content?: string;
    } = { title: title.trim(), priority, steps: cleanSteps, tags };
    if (actions !== null) {
      payload.executable_content = JSON.stringify(actions);
    }
    return payload;
  };

  const handleSave = async (): Promise<boolean> => {
    setSaving(true);
    setSaveError(null);
    try {
      const updated = await updateTestCase(id, buildPayload());
      setTest(updated);
      setSavedMsg(true);
      return true;
    } catch (e) {
      setSaveError(e instanceof Error ? e.message : "Gagal menyimpan");
      return false;
    } finally {
      setSaving(false);
    }
  };

  const handleSaveAndRun = async () => {
    setSaveRunning(true);
    setSaveError(null);
    try {
      await updateTestCase(id, buildPayload());
      const res = await runTestCase(id);
      router.push(`/runs/${res.run_id}`);
    } catch (e) {
      setSaveError(e instanceof Error ? e.message : "Gagal menyimpan & menjalankan");
      setSaveRunning(false);
    }
  };

  /* ----- Render states ----- */
  if (loading) {
    return (
      <div className="max-w-3xl mx-auto">
        <LoadingSkeleton rows={6} />
      </div>
    );
  }

  if (loadError) {
    return (
      <div className="max-w-3xl mx-auto space-y-4">
        <Link
          href="/tests"
          className="inline-flex items-center gap-1.5 text-sm text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors"
        >
          <ArrowLeft className="w-4 h-4" /> Kembali ke Test Library
        </Link>
        <div className="rounded-[var(--radius)] border border-[var(--danger)]/30 bg-[var(--danger-bg)] px-4 py-3 text-sm text-[var(--danger)]">
          {loadError}
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      {/* Back link */}
      <Link
        href="/tests"
        className="inline-flex items-center gap-1.5 text-sm text-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors"
      >
        <ArrowLeft className="w-4 h-4" /> Kembali ke Test Library
      </Link>

      {/* Header */}
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div className="flex items-center gap-2.5 flex-wrap">
          <h1 className="text-xl font-semibold tracking-tight">Edit Test Case</h1>
          {deterministic && (
            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-[var(--success-bg)] border border-[var(--success)]/20 text-[10px] font-semibold text-[var(--success)]">
              <CheckCircle2 className="w-3 h-3" /> Deterministic
            </span>
          )}
          {test && (
            <span className="inline-flex items-center px-1.5 py-0.5 rounded bg-[var(--bg-subtle)] border border-[var(--border)] text-[10px] font-semibold text-[var(--text-muted)]">
              v{test.version}
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Button variant="secondary" onClick={handleSaveAndRun} isLoading={saveRunning} disabled={saving}>
            <Play className="w-4 h-4" /> Save &amp; Run
          </Button>
          <Button onClick={handleSave} isLoading={saving} disabled={saveRunning}>
            <Save className="w-4 h-4" /> Save
          </Button>
        </div>
      </div>

      {/* Feedback */}
      {savedMsg && (
        <div className="rounded-[var(--radius)] border border-[var(--success)]/30 bg-[var(--success-bg)] px-4 py-2.5 text-sm font-medium text-[var(--success)] flex items-center gap-2">
          <CheckCircle2 className="w-4 h-4" /> Tersimpan ✓
        </div>
      )}
      {saveError && (
        <div className="rounded-[var(--radius)] border border-[var(--danger)]/30 bg-[var(--danger-bg)] px-4 py-3 text-sm text-[var(--danger)]">
          {saveError}
        </div>
      )}

      {/* Section 1 — Info dasar */}
      <Section title="Info Dasar">
        <div className="space-y-4">
          <Input
            label="Nama Test"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Contoh: User bisa login dengan email valid"
          />
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">Priority</label>
              <select
                value={priority}
                onChange={(e) => setPriority(e.target.value)}
                className="w-full h-10 px-3 bg-white border border-[var(--border-default)] rounded-md text-sm text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent)] focus:ring-2 focus:ring-[var(--accent)]/20 transition-colors duration-150"
              >
                <option value="high">High</option>
                <option value="medium">Medium</option>
                <option value="low">Low</option>
              </select>
            </div>
            <Input
              label="Tags"
              value={tagsInput}
              onChange={(e) => setTagsInput(e.target.value)}
              placeholder="login, smoke, auth"
              helperText="Pisahkan dengan koma"
            />
          </div>
        </div>
      </Section>

      {/* Section 2 — Langkah Test */}
      <Section
        title="Langkah Test"
        action={
          <Button variant="ghost" size="sm" onClick={addStep}>
            <Plus className="w-4 h-4" /> Tambah Langkah
          </Button>
        }
      >
        <div className="space-y-2">
          {steps.length === 0 && (
            <p className="text-sm text-[var(--text-muted)] py-2">Belum ada langkah. Klik &quot;Tambah Langkah&quot; untuk menambah.</p>
          )}
          {steps.map((step, i) => (
            <div key={i} className="flex items-center gap-2 group">
              <span className="w-6 shrink-0 text-center text-xs font-semibold text-[var(--text-muted)]">{i + 1}</span>
              <Input
                value={step}
                onChange={(e) => updateStep(i, e.target.value)}
                placeholder={`Langkah ${i + 1}...`}
              />
              <div className="flex items-center gap-0.5 shrink-0">
                <IconBtn onClick={() => moveStep(i, -1)} disabled={i === 0} title="Pindah ke atas">
                  <ChevronUp className="w-4 h-4" />
                </IconBtn>
                <IconBtn onClick={() => moveStep(i, 1)} disabled={i === steps.length - 1} title="Pindah ke bawah">
                  <ChevronDown className="w-4 h-4" />
                </IconBtn>
                <IconBtn onClick={() => removeStep(i)} danger title="Hapus langkah">
                  <Trash2 className="w-4 h-4" />
                </IconBtn>
              </div>
            </div>
          ))}
        </div>
      </Section>

      {/* Section 3 — Aksi Browser (Deterministic) */}
      <Section
        title="Aksi Browser (Deterministic)"
        action={
          actions !== null && (
            <Button variant="ghost" size="sm" onClick={addAction}>
              <Plus className="w-4 h-4" /> Tambah Aksi
            </Button>
          )
        }
      >
        {actions === null ? (
          <EmptyState
            icon={<Monitor className="w-6 h-6" />}
            title="Test case ini belum punya aksi deterministik."
            description="Aksi deterministik adalah rekaman browser action yang bisa di-run ulang persis tanpa AI."
          />
        ) : (
          <div className="space-y-3">
            <p className="text-xs text-[var(--text-muted)] flex items-start gap-1.5 leading-relaxed">
              <CheckCircle2 className="w-3.5 h-3.5 mt-0.5 shrink-0 text-[var(--success)]" />
              Aksi ini dijalankan persis seperti yang tertera saat test di-run ulang (tanpa AI).
            </p>
            {actions.map((action, i) => (
              <ActionCard
                key={i}
                index={i}
                total={actions.length}
                action={action}
                onChange={(patch) => updateAction(i, patch)}
                onChangeType={(type) => setActionType(i, type)}
                onMove={(dir) => moveAction(i, dir)}
                onRemove={() => removeAction(i)}
              />
            ))}
            {actions.length === 0 && (
              <p className="text-sm text-[var(--text-muted)] py-2">Belum ada aksi. Klik &quot;Tambah Aksi&quot; untuk menambah.</p>
            )}
          </div>
        )}
      </Section>

      {/* Bottom actions */}
      <div className="flex items-center justify-end gap-2 pb-6">
        <Button variant="secondary" onClick={handleSaveAndRun} isLoading={saveRunning} disabled={saving}>
          <Play className="w-4 h-4" /> Save &amp; Run
        </Button>
        <Button onClick={handleSave} isLoading={saving} disabled={saveRunning}>
          <Save className="w-4 h-4" /> Save
        </Button>
      </div>
    </div>
  );
}

/* ========== Sub-components ========== */

function IconBtn({
  children,
  onClick,
  disabled,
  danger,
  title,
}: {
  children: React.ReactNode;
  onClick: () => void;
  disabled?: boolean;
  danger?: boolean;
  title?: string;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      title={title}
      className={cn(
        "p-1.5 rounded-[var(--radius-sm)] transition-colors disabled:opacity-30 disabled:cursor-not-allowed",
        danger
          ? "text-[var(--text-muted)] hover:text-[var(--danger)] hover:bg-[var(--danger-bg)]"
          : "text-[var(--text-muted)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]"
      )}
    >
      {children}
    </button>
  );
}

const fieldLabel = "block text-[11px] font-medium text-[var(--text-muted)] mb-1";
const fieldInput =
  "w-full h-9 px-2.5 bg-white border border-[var(--border-default)] rounded-md text-[13px] text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] focus:ring-2 focus:ring-[var(--accent)]/20 transition-colors duration-150";

function ActionCard({
  action,
  index,
  total,
  onChange,
  onChangeType,
  onMove,
  onRemove,
}: {
  action: BrowserAction;
  index: number;
  total: number;
  onChange: (patch: Partial<BrowserAction>) => void;
  onChangeType: (type: string) => void;
  onMove: (dir: -1 | 1) => void;
  onRemove: () => void;
}) {
  const Icon = actionIcons[action.action] || MousePointerClick;
  const a = action.action;

  return (
    <div className="rounded-[var(--radius)] border border-[var(--border)] bg-[var(--bg-subtle)]/50 p-3.5">
      {/* Card header: icon + type select + controls */}
      <div className="flex items-center gap-2 mb-3">
        <span className="w-7 h-7 shrink-0 rounded-[var(--radius-sm)] bg-[var(--accent-bg)] border border-[var(--accent)]/15 flex items-center justify-center text-[var(--accent)]">
          <Icon className="w-3.5 h-3.5" />
        </span>
        <select
          value={a}
          onChange={(e) => onChangeType(e.target.value)}
          className="h-9 px-2.5 bg-white border border-[var(--border-default)] rounded-md text-[13px] font-medium text-[var(--text-primary)] focus:outline-none focus:border-[var(--accent)] focus:ring-2 focus:ring-[var(--accent)]/20 transition-colors duration-150 capitalize"
        >
          {ACTION_TYPES.map((t) => (
            <option key={t} value={t}>
              {t}
            </option>
          ))}
        </select>
        <div className="flex items-center gap-0.5 ml-auto shrink-0">
          <IconBtn onClick={() => onMove(-1)} disabled={index === 0} title="Pindah ke atas">
            <ChevronUp className="w-4 h-4" />
          </IconBtn>
          <IconBtn onClick={() => onMove(1)} disabled={index === total - 1} title="Pindah ke bawah">
            <ChevronDown className="w-4 h-4" />
          </IconBtn>
          <IconBtn onClick={onRemove} danger title="Hapus aksi">
            <Trash2 className="w-4 h-4" />
          </IconBtn>
        </div>
      </div>

      {/* Fields based on action type */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {a === "goto" && (
          <div className="sm:col-span-2">
            <label className={fieldLabel}>URL</label>
            <input
              className={fieldInput}
              value={action.url || ""}
              onChange={(e) => onChange({ url: e.target.value })}
              placeholder="https://example.com/login"
            />
          </div>
        )}

        {(a === "fill" || a === "click" || a === "hover" || a === "select" || a === "assert") && (
          <div className={a === "click" || a === "hover" ? "sm:col-span-2" : ""}>
            <label className={fieldLabel}>Selector</label>
            <input
              className={fieldInput}
              value={action.selector || ""}
              onChange={(e) => onChange({ selector: e.target.value })}
              placeholder="css=..., #id, [data-testid=...]"
            />
          </div>
        )}

        {(a === "fill" || a === "select") && (
          <div>
            <label className={fieldLabel}>Value</label>
            <input
              className={fieldInput}
              value={action.value || ""}
              onChange={(e) => onChange({ value: e.target.value })}
              placeholder={a === "select" ? "Option value" : "Text to fill"}
            />
          </div>
        )}

        {a === "press" && (
          <div>
            <label className={fieldLabel}>Key</label>
            <input
              className={fieldInput}
              value={action.key || ""}
              onChange={(e) => onChange({ key: e.target.value })}
              placeholder="Enter, Tab, Escape..."
            />
          </div>
        )}

        {a === "wait" && (
          <div>
            <label className={fieldLabel}>Duration (ms)</label>
            <input
              type="number"
              min={0}
              className={fieldInput}
              value={action.ms ?? 0}
              onChange={(e) => onChange({ ms: Number(e.target.value) || 0 })}
              placeholder="1000"
            />
          </div>
        )}

        {a === "scroll" && (
          <div>
            <label className={fieldLabel}>Scroll Y (px)</label>
            <input
              type="number"
              className={fieldInput}
              value={action.y ?? 0}
              onChange={(e) => onChange({ y: Number(e.target.value) || 0 })}
              placeholder="500"
            />
          </div>
        )}

        {a === "assert" && (
          <>
            <div>
              <label className={fieldLabel}>Assert Type</label>
              <select
                className={fieldInput}
                value={action.assert || "visible"}
                onChange={(e) => onChange({ assert: e.target.value })}
              >
                {ASSERT_TYPES.map((t) => (
                  <option key={t} value={t}>
                    {t.replace(/_/g, " ")}
                  </option>
                ))}
              </select>
            </div>
            <div className="sm:col-span-2">
              <label className={fieldLabel}>Text</label>
              <input
                className={fieldInput}
                value={action.text || ""}
                onChange={(e) => onChange({ text: e.target.value })}
                placeholder="Expected text / value"
              />
            </div>
          </>
        )}
      </div>
    </div>
  );
}
