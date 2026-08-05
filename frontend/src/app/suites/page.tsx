"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Badge } from "@/components/ui/badge";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";
import { Layers, Search, Plus, PlayCircle, CalendarClock, CheckCircle2 } from "lucide-react";

import { getSuites, createSuite, type Suite } from "@/lib/api";

export default function SuitesPage() {
  const [suites, setSuites] = useState<Suite[]>([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [showModal, setShowModal] = useState(false);

  const loadSuites = () => {
    setLoading(true);
    getSuites()
      .then(setSuites)
      .catch(console.error)
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadSuites();
  }, []);

  const filteredSuites = suites.filter(s =>
    s.name.toLowerCase().includes(query.toLowerCase()) ||
    (s.tags || []).some(t => t.toLowerCase().includes(query.toLowerCase()))
  );

  if (loading) return <LoadingSkeleton rows={6} />;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-semibold tracking-tight mb-1">Test Suite</h1>
          <p className="text-sm text-[var(--text-muted)] mt-1">Organized collections of related test cases</p>
        </div>
        <Button onClick={() => setShowModal(true)}>
          <Plus className="w-4 h-4 mr-2" />
          Suite Baru
        </Button>
      </div>

      {/* Search & Filter */}
      <div className="relative max-w-md">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
        <Input
          placeholder="Cari suite..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="pl-9"
        />
      </div>

      {/* Suites Grid */}
      {filteredSuites.length === 0 ? (
        <Section title="Tidak ada test suite">
          <EmptyState 
            icon={<Layers className="w-8 h-8" />}
            title={!query ? "Belum ada test suite" : "Tidak ada suite yang cocok"}
            description={!query ? "Buat test suite pertamamu untuk mengatur test." : "Coba ubah kriteria pencarian."}
            action={!query && (
              <Button onClick={() => setShowModal(true)}>
                <Plus className="w-4 h-4 mr-2" />
                Buat Suite
              </Button>
            )}
          />
        </Section>
      ) : (
        <TableContainer>
          <table className="w-full text-left">
            <thead className="bg-gray-50 border-b border-[var(--border-default)]">
              <tr>
                <Th>Nama Suite</Th>
                <Th>Status</Th>
                <Th>Test</Th>
                <Th>Run Terakhir</Th>
                <Th align="right">Aksi</Th>
              </tr>
            </thead>
            <tbody>
              {filteredSuites.map((suite) => (
                <Tr key={suite.id} hover>
                  <Td className="font-medium">
                    <span>{suite.name}</span>
                  </Td>
                  <Td>
                    <StatusBadge status={suite.pinned ? "active" : "inactive"} />
                  </Td>
                  <Td className="text-sm text-[var(--text-muted)]">{suite.run_ids?.length || 0} tests</Td>
                  <Td className="text-sm text-[var(--text-muted)] whitespace-nowrap">
                    {formatDate(suite.created_at)}
                  </Td>
                  <Td align="right" className="space-x-2">
                    <Link href={`/runs?suite=${suite.id}`}>
                      <Button variant="secondary" size="sm">
                        <PlayCircle className="w-3.5 h-3.5 mr-1" />
                        Run
                      </Button>
                    </Link>
                    <Link href={`/suites/${suite.id}`}>
                      <Button variant="ghost" size="sm">Detail</Button>
                    </Link>
                  </Td>
                </Tr>
              ))}
            </tbody>
          </table>
        </TableContainer>
      )}

      {/* Buat Suite Modal */}
      {showModal && (
        <SuiteModal
          onClose={() => setShowModal(false)}
          onCreated={() => {
            setShowModal(false);
            loadSuites();
          }}
        />
      )}
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  if (status === "active") {
    return (
      <Badge variant="success" size="sm">
        Active
      </Badge>
    );
  } else if (status === "archived") {
    return (
      <Badge variant="default" size="sm">
        Archived
      </Badge>
    );
  } else {
    return (
      <Badge variant="warning" size="sm">
        Inactive
      </Badge>
    );
  }
}

function SuiteModal({ onClose, onCreated }: { onClose: () => void; onCreated: () => void }) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    setSubmitting(true);
    setError(null);
    try {
      await createSuite({
        name: name.trim(),
        tags: description.trim() ? [description.trim()] : [],
      });
      onCreated();
    } catch (err) {
      console.error(err);
      setError(err instanceof Error ? err.message : "Failed to create suite");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6">
        <h2 className="text-lg font-semibold mb-4">Create New Test Suite</h2>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="suite-name" className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">
              Nama Suite <span className="text-red-500">*</span>
            </label>
            <Input
              id="suite-name"
              placeholder="e.g., E2E Checkout Flow"
              value={name}
              onChange={(e) => setName(e.target.value)}
              autoFocus
            />
          </div>

          <div>
            <label htmlFor="suite-desc" className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">
              Description
            </label>
            <textarea
              id="suite-desc"
              placeholder="Describe what this suite tests..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full px-3 py-2 bg-white border border-[var(--border-default)] rounded-md text-sm resize-none focus:outline-none focus:border-[var(--accent)] focus:ring-2 focus:ring-[var(--accent)]/20"
              rows={3}
            />
          </div>

          <div className="flex justify-end gap-3 pt-4 border-t">
            {error && (
              <p className="mr-auto text-xs text-red-600 self-center">{error}</p>
            )}
            <Button type="button" variant="secondary" onClick={onClose}>Cancel</Button>
            <Button type="submit" disabled={!name.trim() || submitting}>
              {submitting ? "Creating…" : "Buat Suite"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffHour = Math.floor(diffMs / 3600000);
  const diffDay = Math.floor(diffMs / 86400000);
  
  if (diffHour < 24) return `${diffHour}h ago`;
  if (diffDay < 7) return `${diffDay}d ago`;
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}
