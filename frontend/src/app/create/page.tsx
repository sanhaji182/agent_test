"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import {
  createRun,
  createAPIKey,
  createRecordingSession,
  createTestCaseFromRecording,
  deleteRecordingSession,
  getAvailableModels,
  getRecordingSession,
  getSettings,
  runTestCase,
  updateRecordingSession,
  type RecordedEvent,
  type RecordingSession,
  type TestCase,
} from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import {
  ArrowLeft,
  ArrowRight,
  ArrowUpDown,
  CheckCircle2,
  ChevronDown,
  Circle,
  Globe,
  MousePointerClick,
  Navigation,
  RefreshCw,
  Sparkles,
  Square,
  Trash2,
  Type,
  Video,
} from "lucide-react";

type Method = "ai" | "record";

const STEPS = ["Target", "Metode", "Konfirmasi"] as const;

const EVENT_ICONS: Record<string, React.ReactNode> = {
  click: <MousePointerClick className="w-3.5 h-3.5" />,
  fill: <Type className="w-3.5 h-3.5" />,
  navigate: <Navigation className="w-3.5 h-3.5" />,
  scroll: <ArrowUpDown className="w-3.5 h-3.5" />,
  assert: <CheckCircle2 className="w-3.5 h-3.5" />,
};

function eventIcon(type: string) {
  return EVENT_ICONS[type] ?? <Circle className="w-3.5 h-3.5" />;
}

export default function CreatePage() {
  const router = useRouter();
  const [step, setStep] = useState(0);
  const [method, setMethod] = useState<Method>("ai");

  const [formData, setFormData] = useState({
    project_path: "",
    name: "",
    requirements: "",
    model: "",
  });

  // Model override (advanced)
  const [defaultModel, setDefaultModel] = useState("");
  const [models, setModels] = useState<string[]>([]);
  const [fetchingModels, setFetchingModels] = useState(false);
  const [modelsError, setModelsError] = useState<string | null>(null);

  // Recording state
  const [session, setSession] = useState<RecordingSession | null>(null);
  const [events, setEvents] = useState<RecordedEvent[]>([]);
  const [startingSession, setStartingSession] = useState(false);
  const [stoppingSession, setStoppingSession] = useState(false);
  const [testCase, setTestCase] = useState<TestCase | null>(null);

	  const [submitting, setSubmitting] = useState(false);
	  const [error, setError] = useState<string | null>(null);

	  // Recorder setup: deteksi extension, salin URL backend, buat API key.
	  const [extDetected, setExtDetected] = useState(false);
	  const [copiedUrl, setCopiedUrl] = useState(false);
	  const [generatedKey, setGeneratedKey] = useState<string | null>(null);
	  const [keyCopied, setKeyCopied] = useState(false);
	  const [genKeyErr, setGenKeyErr] = useState<string | null>(null);
	  const backendUrl = (typeof window !== "undefined" ? (process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080") : "http://localhost:8080");

	  // Deteksi extension via marker DOM (data-gotest-recorder) — poll tiap 2s
	  // selama langkah Rekam aktif.
	  useEffect(() => {
	    if (method !== "record") return;
	    const check = () => {
	      if (typeof document !== "undefined") {
	        setExtDetected(document.documentElement.dataset.gotestRecorder === "1");
	      }
	    };
	    check();
	    const t = setInterval(check, 2000);
	    return () => clearInterval(t);
	  }, [method]);

	  const copyText = async (text: string, kind: "url" | "key") => {
	    try {
	      await navigator.clipboard.writeText(text);
	      if (kind === "url") { setCopiedUrl(true); setTimeout(() => setCopiedUrl(false), 2000); }
	      else { setKeyCopied(true); setTimeout(() => setKeyCopied(false), 2000); }
	    } catch {
	      // fallback
	      const ta = document.createElement("textarea");
	      ta.value = text; document.body.appendChild(ta); ta.select();
	      document.execCommand("copy"); document.body.removeChild(ta);
	      if (kind === "url") { setCopiedUrl(true); setTimeout(() => setCopiedUrl(false), 2000); }
	      else { setKeyCopied(true); setTimeout(() => setKeyCopied(false), 2000); }
	    }
	  };

	  const handleGenKey = async () => {
	    setGenKeyErr(null);
	    setGeneratedKey(null);
	    try {
	      const res = await createAPIKey("recorder-extension", "viewer");
	      if (res.key) {
	        setGeneratedKey(res.key);
	        copyText(res.key, "key");
	      } else {
	        setGenKeyErr("Key tidak kembali dari server — cek kembali atau pakai API_KEY utama.");
	      }
	    } catch (e) {
	      setGenKeyErr(e instanceof Error ? e.message : "Gagal membuat API key (perlu akses admin).");
	    }
	  };

	  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const handleChange = (field: string, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

	  // Load the current default model from Settings so we can hint it in the field.
	  useEffect(() => {
	    getSettings()
	      .then((s) => { if (s.llm_model) setDefaultModel(s.llm_model); })
	      .catch(() => {});
	  }, []);

	  // Support deep-link: /create?method=record (tombol "Record Test" di dashboard)
	  // langsung memilih metode Rekam. Juga prefill URL bila ?url= disertakan.
	  useEffect(() => {
	    if (typeof window === "undefined") return;
	    const qs = new URLSearchParams(window.location.search);
	    const url = qs.get("url") || "";
	    if (url) {
	      setFormData((prev) => ({ ...prev, project_path: url }));
	    }
	    // Deep-link ?method=record: pilih metode Rekam, tapi JANGAN lompati
	    // langkah Target — URL aplikasi tetap wajib diisi dulu di Step 1.
	    // (Tombol "Record Test" di dashboard mengarah ke sini.)
	    if (qs.get("method") === "record") {
	      setMethod("record");
	      if (url) {
	        setStep(1);
	      }
	    }
	  }, []);

  const stopPolling = useCallback(() => {
    if (pollRef.current) {
      clearInterval(pollRef.current);
      pollRef.current = null;
    }
  }, []);

  // Poll for recorded events only while a session is active AND we're on the
  // Metode step with "Rekam" selected. Cleans up on unmount and when stopping.
  useEffect(() => {
    stopPolling();
    if (session && step === 1 && method === "record" && session.status !== "completed") {
      pollRef.current = setInterval(() => {
        getRecordingSession(session.id)
          .then((res) => {
            setSession(res.session);
            setEvents(res.events || []);
          })
          .catch(() => {});
      }, 2000);
    }
    return stopPolling;
  }, [session, step, method, stopPolling]);

  // Fetch events once right after creating a session.
  const refreshEvents = useCallback((sessionId: string) => {
    getRecordingSession(sessionId)
      .then((res) => {
        setSession(res.session);
        setEvents(res.events || []);
      })
      .catch(() => {});
  }, []);

  const handleFetchModels = async () => {
    setFetchingModels(true);
    setModelsError(null);
    try {
      const res = await getAvailableModels();
      if (res.error) {
        setModelsError(res.error);
        setModels([]);
      } else {
        setModels(res.models || []);
      }
    } catch (e) {
      setModelsError(e instanceof Error ? e.message : "failed to fetch models");
      setModels([]);
    } finally {
      setFetchingModels(false);
    }
  };

	  const handleStartSession = async () => {
	    // Validasi: URL aplikasi wajib (sesi rekam butuh base_url untuk replay).
	    if (!formData.project_path.trim()) {
	      setError("Isi dulu URL aplikasi di langkah Target sebelum mulai merekam.");
	      setStep(0);
	      return;
	    }
	    setStartingSession(true);
	    setError(null);
	    try {
	      const s = await createRecordingSession({
	        name: formData.name.trim() || `Recording ${new Date().toLocaleTimeString()}`,
	        project_path: formData.project_path,
	        base_url: formData.project_path,
	      });
	      setSession(s);
	      setEvents([]);
	      refreshEvents(s.id);
	    } catch (err) {
	      setError(err instanceof Error ? err.message : "Gagal membuat sesi rekam");
	    } finally {
	      setStartingSession(false);
	    }
	  };

  const handleStopAndConvert = async () => {
    if (!session) return;
    setStoppingSession(true);
    setError(null);
    try {
      await updateRecordingSession(session.id, { status: "completed" });
      stopPolling();
      const tc = await createTestCaseFromRecording(session.id);
      setTestCase(tc);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal membuat test case dari rekaman");
    } finally {
      setStoppingSession(false);
    }
  };

  const handleDeleteSession = async () => {
    if (!session) return;
    stopPolling();
    try {
      await deleteRecordingSession(session.id);
    } catch {
      // Session mungkin sudah terhapus di server; tetap reset state lokal.
    }
    setSession(null);
    setEvents([]);
    setTestCase(null);
  };

  const canNextStep1 = formData.project_path.trim().length > 0;
  const canNextStep2 =
    method === "ai"
      ? formData.requirements.trim().length > 0
      : testCase !== null;

  const handleSubmit = async () => {
    setSubmitting(true);
    setError(null);
    try {
      let runId: string | undefined;
      if (method === "record") {
        if (!testCase) return;
        const result = await runTestCase(testCase.id);
        runId = result?.run_id;
      } else {
        const payload: Parameters<typeof createRun>[0] = {
          project_path: formData.project_path,
          requirements: formData.requirements,
          mode: "simple",
        };
        if (formData.model.trim()) payload.model = formData.model.trim();
        const result = await createRun(payload);
        runId = result?.run_id;
      }

      if (runId) {
        router.push(`/runs/${runId}`);
      } else {
        setError("Gagal membuat test run. Silakan coba lagi.");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Terjadi kesalahan");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-xl font-semibold tracking-tight mb-2">Create Test Run</h1>
        <p className="text-sm text-[var(--text-muted)]">Buat test baru dalam 3 langkah mudah</p>
      </div>

      {/* Step Indicator */}
      <div className="flex items-center gap-2">
        {STEPS.map((label, i) => {
          const active = i === step;
          const done = i < step;
          return (
            <div key={label} className="flex items-center gap-2 flex-1 last:flex-none">
              <div className="flex items-center gap-2">
                <div
                  className={cn(
                    "w-6 h-6 rounded-full flex items-center justify-center text-[11px] font-semibold border transition-colors",
                    done
                      ? "bg-[var(--accent)] border-[var(--accent)] text-white"
                      : active
                        ? "border-[var(--accent)] text-[var(--accent)] bg-[var(--accent-bg)]"
                        : "border-[var(--border-default)] text-[var(--text-muted)] bg-white"
                  )}
                >
                  {done ? <CheckCircle2 className="w-3.5 h-3.5" /> : i + 1}
                </div>
                <span
                  className={cn(
                    "text-xs font-medium whitespace-nowrap",
                    active ? "text-[var(--text-primary)]" : "text-[var(--text-muted)]"
                  )}
                >
                  {label}
                </span>
              </div>
              {i < STEPS.length - 1 && (
                <div className="flex-1 h-px bg-[var(--border-default)] mx-1" />
              )}
            </div>
          );
        })}
      </div>

      {/* Form Card */}
      <div className="bg-white rounded-lg border border-[var(--border-default)] p-6 shadow-xs">

        {/* ============ STEP 1 — TARGET ============ */}
        {step === 0 && (
          <div className="space-y-5">
            <Input
              label="URL Aplikasi"
              placeholder="https://asisten.digital/"
              value={formData.project_path}
              onChange={(e) => handleChange("project_path", e.target.value)}
              leftIcon={<Globe className="w-4 h-4" />}
              required
              helperText="Alamat aplikasi atau website yang mau dites."
            />
            <Input
              label="Nama Test (opsional)"
              placeholder="My Login Test"
              value={formData.name}
              onChange={(e) => handleChange("name", e.target.value)}
            />
            <div className="flex justify-end pt-4 border-t border-[var(--border-default)]">
              <Button
                type="button"
                disabled={!canNextStep1}
                onClick={() => setStep(1)}
              >
                Lanjut
                <ArrowRight className="w-4 h-4" />
              </Button>
            </div>
          </div>
        )}

        {/* ============ STEP 2 — METODE ============ */}
        {step === 1 && (
          <div className="space-y-5">
            <div className="grid gap-3">
              {/* Card: Deskripsi AI */}
              <button
                type="button"
                onClick={() => setMethod("ai")}
                className={cn(
                  "text-left rounded-lg border p-4 transition-colors",
                  method === "ai"
                    ? "border-[var(--accent)] bg-[var(--accent-bg)] ring-1 ring-[var(--accent)]/30"
                    : "border-[var(--border-default)] bg-white hover:border-[var(--border-strong)]"
                )}
              >
                <div className="flex items-start gap-3">
                  <div
                    className={cn(
                      "w-8 h-8 rounded-[var(--radius-sm)] flex items-center justify-center shrink-0",
                      method === "ai"
                        ? "bg-[var(--accent)] text-white"
                        : "bg-[var(--bg-subtle)] text-[var(--text-muted)]"
                    )}
                  >
                    <Sparkles className="w-4 h-4" />
                  </div>
                  <div>
                    <p className="text-sm font-semibold text-[var(--text-primary)]">Deskripsi AI</p>
                    <p className="text-xs text-[var(--text-muted)] mt-0.5 leading-relaxed">
                      Jelaskan apa yang mau dites dalam bahasa bebas. AI akan menganalisis halaman, membuat, lalu menjalankan test secara otomatis.
                    </p>
                  </div>
                </div>
              </button>

              {/* Card: Rekam (Extension) */}
              <button
                type="button"
                onClick={() => setMethod("record")}
                className={cn(
                  "text-left rounded-lg border p-4 transition-colors",
                  method === "record"
                    ? "border-[var(--accent)] bg-[var(--accent-bg)] ring-1 ring-[var(--accent)]/30"
                    : "border-[var(--border-default)] bg-white hover:border-[var(--border-strong)]"
                )}
              >
                <div className="flex items-start gap-3">
                  <div
                    className={cn(
                      "w-8 h-8 rounded-[var(--radius-sm)] flex items-center justify-center shrink-0",
                      method === "record"
                        ? "bg-[var(--accent)] text-white"
                        : "bg-[var(--bg-subtle)] text-[var(--text-muted)]"
                    )}
                  >
                    <Video className="w-4 h-4" />
                  </div>
                  <div>
                    <p className="text-sm font-semibold text-[var(--text-primary)]">Rekam (Extension)</p>
                    <p className="text-xs text-[var(--text-muted)] mt-0.5 leading-relaxed">
                      Rekam interaksi kamu di browser seperti Katalon. Hasil rekaman jadi test case yang bisa di-run ulang persis (deterministik).
                    </p>
                  </div>
                </div>
              </button>
            </div>

            {/* AI path: requirements textarea */}
            {method === "ai" && (
              <div>
                <label htmlFor="requirements" className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">
                  Apa yang mau dites? <span className="text-red-500">*</span>
                </label>
                <textarea
                  id="requirements"
                  value={formData.requirements}
                  onChange={(e) => handleChange("requirements", e.target.value)}
                  placeholder="Pastikan homepage https://asisten.digital/ berhasil dimuat dan heading utama terlihat."
                  className="w-full h-32 px-3 py-2 bg-white border border-[var(--border-default)] rounded-md text-sm resize-none focus:outline-none focus:border-[var(--accent)] focus:ring-2 focus:ring-[var(--accent)]/20"
                  required
                />
                <p className="mt-1 text-xs text-[var(--text-muted)]">Tulis dalam bahasa bebas — makin spesifik makin bagus hasilnya.</p>
              </div>
            )}

	            {/* Record path: recording panel */}
	            {method === "record" && (
	              <div className="space-y-4">
	                {/* Deteksi extension */}
	                <div className={`rounded-lg border p-3 flex items-center gap-3 ${extDetected ? "border-[var(--success)]/30 bg-[var(--success-bg)]" : "border-[var(--warning)]/30 bg-[var(--warning-bg)]"}`} role="status" aria-live="polite">
	                  <CheckCircle2 className={`w-5 h-5 shrink-0 ${extDetected ? "text-[var(--success)]" : "text-[var(--warning)]"}`} />
	                  <div className="text-xs">
	                    {extDetected ? (
	                      <p className="font-semibold text-[var(--success)]">Ekstensi GoTest Recorder terdeteksi ✓</p>
	                    ) : (
	                      <>
	                        <p className="font-semibold text-[var(--warning)]">Ekstensi belum terpasang</p>
	                        <p className="text-[var(--text-secondary)] mt-0.5">Pasang dulu (sekali saja), lalu muat ulang halaman ini.</p>
	                      </>
	                    )}
	                  </div>
	                </div>

	                {/* Setup: panduan + salin URL + buat key */}
	                <div className="rounded-lg border border-[var(--border-default)] bg-[var(--bg-subtle)] p-3 space-y-3">
	                  <p className="text-xs font-semibold text-[var(--text-primary)]">Setup ekstensi (sekali saja):</p>
	                  <ol className="text-xs text-[var(--text-secondary)] space-y-1.5 list-decimal ml-4">
	                    <li>Buka <code className="font-mono text-[11px] bg-white px-1 py-0.5 rounded border border-[var(--border-default)]">chrome://extensions</code></li>
	                    <li>Aktifkan <strong>Developer mode</strong> (pojok kanan atas)</li>
	                    <li>Klik <strong>Load unpacked</strong> → pilih folder <code className="font-mono text-[11px] bg-white px-1 py-0.5 rounded border border-[var(--border-default)]">chrome-extension/</code></li>
	                    <li>Buka popup ekstensi → isi Backend URL &amp; API key di bawah → <strong>Save Settings</strong></li>
	                    <li>Muat ulang halaman ini — indikator di atas berubah jadi ✓</li>
	                  </ol>
	                  <div className="flex flex-wrap gap-2 pt-1">
	                    <Button type="button" variant="secondary" size="sm" onClick={() => copyText(backendUrl, "url")}>
	                      {copiedUrl ? "✓ Tersalin" : "Salin Backend URL"}
	                    </Button>
	                    <Button type="button" variant="secondary" size="sm" onClick={handleGenKey}>
	                      {generatedKey ? "Key dibuat — tersalin ✓" : "Buat & Salin API Key"}
	                    </Button>
	                  </div>
	                  {genKeyErr && <p className="text-xs text-[var(--danger)]">{genKeyErr}</p>}
	                  {generatedKey && (
	                    <p className="text-[11px] text-[var(--text-muted)] leading-relaxed">
	                      Key <code className="font-mono bg-white px-1 py-0.5 rounded border border-[var(--border-default)]">{generatedKey.slice(0, 12)}…</code> sudah disalin —
	                      tempel ke field <strong>API Key</strong> di popup ekstensi. Key hanya ditampilkan sekali.
	                    </p>
	                  )}
	                </div>

                {!session ? (
                  <Button
                    type="button"
                    onClick={handleStartSession}
                    isLoading={startingSession}
                    disabled={startingSession}
                  >
                    <Video className="w-4 h-4" />
                    {startingSession ? "Membuat sesi…" : "Mulai Sesi Rekam"}
                  </Button>
                ) : (
                  <div className="space-y-3">
                    {/* Session status */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <span className="relative flex h-2.5 w-2.5">
                          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-[var(--danger)] opacity-60" />
                          <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-[var(--danger)]" />
                        </span>
                        <span className="text-xs font-medium text-[var(--text-primary)]">
                          {session.name}
                        </span>
                        <Badge variant={session.status === "completed" ? "success" : "danger"} size="sm">
                          {session.status}
                        </Badge>
                      </div>
                      <span className="text-xs text-[var(--text-muted)]">
                        {events.length} event
                      </span>
                    </div>

                    {/* Live events */}
                    {events.length > 0 ? (
                      <div className="rounded-lg border border-[var(--border-default)] bg-white max-h-48 overflow-y-auto divide-y divide-[var(--border)]">
                        {events.slice(-8).map((ev) => (
                          <div key={ev.id} className="flex items-center gap-2.5 px-3 py-2">
                            <span className="text-[var(--text-muted)] shrink-0">{eventIcon(ev.event_type)}</span>
                            <span className="text-xs font-medium text-[var(--text-primary)] capitalize shrink-0 w-16">
                              {ev.event_type}
                            </span>
                            <span className="text-xs text-[var(--text-secondary)] font-mono truncate flex-1">
                              {ev.selector || ev.url || "—"}
                            </span>
                            {ev.value && (
                              <span className="text-[11px] text-[var(--text-muted)] truncate max-w-[120px]">
                                {ev.value}
                              </span>
                            )}
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="rounded-lg border border-dashed border-[var(--border-default)] py-6 text-center">
                        <p className="text-xs text-[var(--text-muted)]">Belum ada event. Mulai klik-klik di browser…</p>
                      </div>
                    )}

	                    <p className="text-xs text-[var(--text-muted)] leading-relaxed">
	                      Buka extension (icon GoTest Agent di toolbar), klik <strong>Attach</strong> pada sesi ini di daftar "Recent Sessions", lalu klik-klik di aplikasi target seperti biasa. Event akan muncul di sini secara live.
	                    </p>

                    {/* Recording actions */}
                    {!testCase ? (
                      <div className="flex flex-wrap gap-2">
                        <Button
                          type="button"
                          variant="secondary"
                          onClick={handleStopAndConvert}
                          isLoading={stoppingSession}
                          disabled={events.length === 0 || stoppingSession}
                        >
                          <Square className="w-3.5 h-3.5" />
                          {stoppingSession ? "Memproses…" : "Berhenti & Jadikan Test Case"}
                        </Button>
                        <Button
                          type="button"
                          variant="ghost"
                          onClick={handleDeleteSession}
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                          Hapus Sesi
                        </Button>
                      </div>
                    ) : (
                      <div className="rounded-lg border border-[var(--success)]/30 bg-[var(--success-bg)] p-3 flex items-center gap-3">
                        <CheckCircle2 className="w-4 h-4 text-[var(--success)] shrink-0" />
                        <div className="flex-1 min-w-0">
                          <p className="text-xs font-semibold text-[var(--text-primary)] truncate">
                            {testCase.title}
                          </p>
                          <p className="text-[11px] text-[var(--text-muted)]">
                            {testCase.steps?.length || 0} langkah · siap dijalankan
                          </p>
                        </div>
                        <Badge variant="success" size="sm">Deterministic</Badge>
                      </div>
                    )}
                  </div>
                )}
              </div>
            )}

            {/* Error */}
            {error && (
              <div className="rounded-lg bg-red-50 border border-red-200 p-3">
                <p className="text-sm text-red-700">{error}</p>
              </div>
            )}

            {/* Actions */}
            <div className="flex justify-between pt-4 border-t border-[var(--border-default)]">
              <Button type="button" variant="secondary" onClick={() => setStep(0)}>
                <ArrowLeft className="w-4 h-4" />
                Kembali
              </Button>
              <Button
                type="button"
                disabled={!canNextStep2}
                onClick={() => setStep(2)}
              >
                Lanjut
                <ArrowRight className="w-4 h-4" />
              </Button>
            </div>
          </div>
        )}

        {/* ============ STEP 3 — KONFIRMASI ============ */}
        {step === 2 && (
          <div className="space-y-5">
            {/* Summary */}
            <div className="rounded-lg border border-[var(--border-default)] divide-y divide-[var(--border)]">
              <SummaryRow label="Target URL" value={formData.project_path} />
              <SummaryRow label="Nama" value={formData.name.trim() || "—"} />
              <SummaryRow label="Metode" value={method === "ai" ? "Deskripsi AI" : "Rekam"} />
              {method === "ai" ? (
                <div className="px-4 py-3">
                  <p className="text-xs font-medium text-[var(--text-muted)] mb-1">Requirements</p>
                  <p className="text-sm text-[var(--text-primary)] whitespace-pre-wrap leading-relaxed">
                    {formData.requirements}
                  </p>
                </div>
              ) : testCase ? (
                <div className="px-4 py-3">
                  <p className="text-xs font-medium text-[var(--text-muted)] mb-1">Test Case</p>
                  <div className="flex items-center gap-2 flex-wrap">
                    <p className="text-sm text-[var(--text-primary)] font-medium">{testCase.title}</p>
                    <span className="text-xs text-[var(--text-muted)]">· {testCase.steps?.length || 0} langkah</span>
                    <Badge variant="success" size="sm">Deterministic</Badge>
                  </div>
                </div>
              ) : null}
            </div>

            {/* Advanced */}
            {method === "ai" && (
              <details className="group rounded-lg border border-[var(--border-default)]">
                <summary className="flex items-center justify-between px-4 py-3 cursor-pointer select-none text-sm font-medium text-[var(--text-secondary)] hover:text-[var(--text-primary)]">
                  Advanced (opsional)
                  <ChevronDown className="w-4 h-4 transition-transform group-open:rotate-180" />
                </summary>
                <div className="px-4 pb-4 pt-1 border-t border-[var(--border)]">
                  <div className="flex items-center justify-between mb-1.5">
                    <label htmlFor="model" className="block text-sm font-medium text-[var(--text-primary)]">
                      AI Model <span className="text-[var(--text-muted)] font-normal">(opsional)</span>
                    </label>
                    <Button
                      type="button"
                      variant="secondary"
                      size="sm"
                      onClick={handleFetchModels}
                      disabled={fetchingModels}
                    >
                      <RefreshCw className={`w-3.5 h-3.5 ${fetchingModels ? "animate-spin" : ""}`} />
                      {fetchingModels ? "Mengambil…" : "Ambil Model"}
                    </Button>
                  </div>
                  <Input
                    id="model"
                    list="available-models"
                    placeholder={defaultModel ? `Default: ${defaultModel}` : "Kosongkan untuk pakai model dari Settings"}
                    value={formData.model}
                    onChange={(e) => handleChange("model", e.target.value)}
                    leftIcon={<Sparkles className="w-4 h-4" />}
                    helperText={modelsError ? undefined : "Kosongkan untuk pakai model dari Settings."}
                    error={modelsError || undefined}
                  />
                  <datalist id="available-models">
                    {models.map((m) => (
                      <option key={m} value={m} />
                    ))}
                  </datalist>
                </div>
              </details>
            )}

            {/* Error */}
            {error && (
              <div className="rounded-lg bg-red-50 border border-red-200 p-3">
                <p className="text-sm text-red-700">{error}</p>
              </div>
            )}

            {/* Actions */}
            <div className="flex justify-between pt-4 border-t border-[var(--border-default)]">
              <Button type="button" variant="secondary" onClick={() => setStep(1)}>
                <ArrowLeft className="w-4 h-4" />
                Kembali
              </Button>
              <Button
                type="button"
                onClick={handleSubmit}
                isLoading={submitting}
                disabled={submitting || (method === "record" && !testCase)}
              >
                {submitting
                  ? "Memproses…"
                  : method === "ai"
                    ? "Buat & Jalankan Test"
                    : "Jalankan Test Case"}
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function SummaryRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between px-4 py-3 gap-4">
      <span className="text-xs font-medium text-[var(--text-muted)] shrink-0">{label}</span>
      <span className="text-sm text-[var(--text-primary)] text-right truncate">{value}</span>
    </div>
  );
}
