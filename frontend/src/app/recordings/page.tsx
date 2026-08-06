"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Badge } from "@/components/ui/badge";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";
import { Video, Trash2, Search, Eye, Plus, PlayCircle, AlertTriangle } from "lucide-react";

import { listRecordingSessions, deleteRecordingSession, type RecordingSession } from "@/lib/api";

export default function RecordingsPage() {
  const [sessions, setSessions] = useState<RecordingSession[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [filterStatus, setFilterStatus] = useState("all");
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [extDetected, setExtDetected] = useState(false);

  // Deteksi extension GoTest Recorder via marker DOM.
  useEffect(() => {
    const check = () => setExtDetected(document.documentElement.dataset.gotestRecorder === "1");
    check();
    const t = setInterval(check, 3000);
    return () => clearInterval(t);
  }, []);

  const load = useCallback(() => {
    setLoading(true);
    setError(null);
    listRecordingSessions()
      .then(setSessions)
      .catch((e) => setError(e instanceof Error ? e.message : "Gagal memuat rekaman"))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const filtered = sessions.filter((s) => {
    if (filterStatus !== "all" && s.status !== filterStatus) return false;
    if (query && !`${s.name} ${s.base_url || ""}`.toLowerCase().includes(query.toLowerCase())) return false;
    return true;
  });

  const getCounts = () => ({
    all: sessions.length,
    recording: sessions.filter((s) => s.status === "recording").length,
    completed: sessions.filter((s) => s.status === "completed").length,
    archived: sessions.filter((s) => s.status === "archived").length,
  });

  const handleDelete = async (id: string, name: string) => {
    if (!window.confirm(`Hapus sesi rekaman "${name}" beserta semua event-nya? Tindakan ini tidak bisa dibatalkan.`)) return;
    setDeletingId(id);
    try {
      await deleteRecordingSession(id);
      load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Gagal menghapus sesi");
    } finally {
      setDeletingId(null);
    }
  };

  if (loading) return <LoadingSkeleton rows={6} />;

  const counts = getCounts();

  return (
    <div className="space-y-6">
	      {/* Header */}
	      <div className="flex flex-wrap items-start justify-between gap-3">
	        <div>
	          <h1 className="text-xl font-semibold tracking-tight">Recordings</h1>
	          <p className="text-sm text-[var(--text-muted)] mt-1">Sesi rekaman browser — hasil rekam bisa jadi test case deterministik</p>
	        </div>
	        <Link href="/create?method=record">
	          <Button>
	            <Plus className="w-4 h-4 mr-2" />
	            New Recording
	          </Button>
	        </Link>
	      </div>

	      {/* Peringatan: extension belum terpasang */}
	      {!extDetected && !error && (
	        <div className="rounded-lg border border-[var(--warning)]/30 bg-[var(--warning-bg)] p-4 flex flex-wrap items-center justify-between gap-3" role="status" aria-live="polite">
	          <div className="flex items-start gap-2.5">
	            <AlertTriangle className="w-4 h-4 text-[var(--warning)] shrink-0 mt-0.5" />
	            <div className="text-xs">
	              <p className="font-semibold text-[var(--warning)]">Ekstensi GoTest Recorder belum terpasang di browser ini</p>
	              <p className="text-[var(--text-secondary)] mt-0.5">Tanpa ekstensi, tombol "New Recording" tetap membuat sesi, tapi tidak akan menerima event. Pasang dulu lewat panduan di halaman Rekam (sekali saja).</p>
	            </div>
	          </div>
	          <Link href="/create?method=record">
	            <Button variant="secondary" size="sm">Lihat Panduan Pasang</Button>
	          </Link>
	        </div>
	      )}

      {/* Error state */}
      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 flex items-center justify-between gap-3" role="alert">
          <p className="text-sm text-red-700">{error}</p>
          <Button variant="secondary" size="sm" onClick={load}>Muat Ulang</Button>
        </div>
      )}

      {/* Search & Filter */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
          <Input
            aria-label="Cari rekaman"
            type="search"
            placeholder="Cari rekaman..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="pl-9"
          />
        </div>

        <div className="inline-flex rounded-lg border border-[var(--border-default)] p-0.5 bg-white">
          {[
            { value: "all", label: "All" },
            { value: "recording", label: "Active" },
            { value: "completed", label: "Completed" },
            { value: "archived", label: "Archived" },
          ].map((f) => (
            <button
              key={f.value}
              onClick={() => setFilterStatus(f.value)}
              className={filterStatus === f.value ? "px-3 py-1.5 text-xs font-medium rounded-md bg-blue-600 text-white" : "px-3 py-1.5 text-xs font-medium rounded-md text-[var(--text-secondary)] hover:bg-gray-50"}
            >
              {f.label} ({f.value === "all" ? counts.all : counts[f.value as keyof typeof counts]})
            </button>
          ))}
        </div>
      </div>

      {/* Recordings List */}
      {filtered.length === 0 ? (
        <Section title={query ? "Tidak ada rekaman yang cocok" : "Belum ada rekaman"}>
          <EmptyState
            icon={<Video className="w-8 h-8" />}
            title={!query ? "Belum ada sesi rekaman" : "Tidak ada yang cocok"}
            description={!query ? "Klik 'New Recording' untuk merekam interaksi di browser dan mengubahnya jadi test case." : "Coba ubah kata kunci pencarian."}
            action={!query && (
              <Link href="/create?method=record">
                <Button>
                  <PlayCircle className="w-4 h-4 mr-2" />
                  Mulai Rekam
                </Button>
              </Link>
            )}
          />
        </Section>
      ) : (
        <TableContainer>
          <table className="w-full text-left">
            <thead className="bg-gray-50 border-b border-[var(--border-default)]">
              <tr>
                <Th>Nama</Th>
                <Th>Status</Th>
                <Th>Target</Th>
                <Th>Dibuat</Th>
                <Th align="right">Aksi</Th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((s) => (
                <Tr key={s.id} hover>
                  <Td className="font-medium">
                    <span className="block truncate max-w-[200px]">{s.name || "Tanpa nama"}</span>
                    <span className="text-xs text-[var(--text-muted)] font-mono">{s.id.slice(0, 8)}</span>
                  </Td>
                  <Td>
                    <Badge
                      variant={s.status === "recording" ? "warning" : s.status === "completed" ? "success" : "default"}
                      size="sm"
                    >
                      {s.status}
                    </Badge>
                  </Td>
                  <Td className="text-sm text-[var(--text-muted)] truncate max-w-[200px]">
                    {s.base_url || s.project_path || "—"}
                  </Td>
                  <Td className="text-sm text-[var(--text-muted)] whitespace-nowrap">
                    {new Date(s.created_at).toLocaleDateString(undefined, { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" })}
                  </Td>
                  <Td align="right" className="space-x-2">
                    <Link href={`/recordings/${s.id}`}>
                      <Button variant="ghost" size="sm" aria-label={`Lihat ${s.name || "sesi"}`}>
                        <Eye className="w-3.5 h-3.5" />
                      </Button>
                    </Link>
                    <Button
                      variant="ghost"
                      size="sm"
                      aria-label={`Hapus ${s.name || "sesi"}`}
                      className="text-red-600 hover:text-red-700 hover:bg-red-50"
                      disabled={deletingId === s.id}
                      onClick={() => handleDelete(s.id, s.name)}
                    >
                      <Trash2 className="w-3.5 h-3.5" />
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
