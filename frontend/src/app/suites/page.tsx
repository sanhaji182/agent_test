"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Badge } from "@/components/ui/badge";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";
import { Layers, Search, Plus, PlayCircle, CalendarClock, CheckCircle2 } from "lucide-react";

export default function SuitesPage() {
  const [suites, setSuites] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [showModal, setShowModal] = useState(false);

  useEffect(() => {
    // TODO: Fetch real suites data
    setTimeout(() => {
      setSuites([
        { 
          id: "suite_1", 
          name: "E2E Checkout Flow", 
          description: "Complete checkout process tests",
          testCount: 12,
          status: "active",
          lastRun: "2026-07-31T14:00:00Z"
        },
        { 
          id: "suite_2", 
          name: "User Authentication", 
          description: "Login, logout, password reset flows",
          testCount: 8,
          status: "active",
          lastRun: "2026-07-31T12:00:00Z"
        },
        { 
          id: "suite_3", 
          name: "API Integration Tests", 
          description: "Backend API endpoint validation",
          testCount: 24,
          status: "archived",
          lastRun: "2026-07-25T10:00:00Z"
        },
      ]);
      setLoading(false);
    }, 500);
  }, []);

  const filteredSuites = suites.filter(s => 
    s.name.toLowerCase().includes(query.toLowerCase()) ||
    s.description?.toLowerCase().includes(query.toLowerCase())
  );

  if (loading) return <LoadingSkeleton rows={6} />;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-semibold tracking-tight mb-1">Test Suites</h1>
          <p className="text-sm text-[var(--text-muted)] mt-1">Organized collections of related test cases</p>
        </div>
        <Button onClick={() => setShowModal(true)}>
          <Plus className="w-4 h-4 mr-2" />
          New Suite
        </Button>
      </div>

      {/* Search & Filter */}
      <div className="relative max-w-md">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-muted)]" />
        <Input
          placeholder="Search suites..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="pl-9"
        />
      </div>

      {/* Suites Grid */}
      {filteredSuites.length === 0 ? (
        <Section title="No test suites found">
          <EmptyState 
            icon={<Layers className="w-8 h-8" />}
            title={!query ? "No test suites yet" : "No matching suites"}
            description={!query ? "Create your first test suite to organize your tests." : "Try adjusting your search criteria."}
            action={!query && (
              <Button onClick={() => setShowModal(true)}>
                <Plus className="w-4 h-4 mr-2" />
                Create Suite
              </Button>
            )}
          />
        </Section>
      ) : (
        <TableContainer>
          <table className="w-full text-left">
            <thead className="bg-gray-50 border-b border-[var(--border-default)]">
              <tr>
                <Th>Suite Name</Th>
                <Th>Status</Th>
                <Th>Tests</Th>
                <Th>Last Run</Th>
                <Th align="right">Actions</Th>
              </tr>
            </thead>
            <tbody>
              {filteredSuites.map((suite) => (
                <Tr key={suite.id} hover>
                  <Td className="font-medium">
                    <span>{suite.name}</span>
                  </Td>
                  <Td>
                    <StatusBadge status={suite.status} />
                  </Td>
                  <Td className="text-sm text-[var(--text-muted)]">{suite.testCount} tests</Td>
                  <Td className="text-sm text-[var(--text-muted)] whitespace-nowrap">
                    {formatDate(suite.lastRun)}
                  </Td>
                  <Td align="right" className="space-x-2">
                    <Link href={`/runs?suite=${suite.id}`}>
                      <Button variant="secondary" size="sm">
                        <PlayCircle className="w-3.5 h-3.5 mr-1" />
                        Run
                      </Button>
                    </Link>
                    <Link href={`/suites/${suite.id}`}>
                      <Button variant="ghost" size="sm">Details</Button>
                    </Link>
                  </Td>
                </Tr>
              ))}
            </tbody>
          </table>
        </TableContainer>
      )}

      {/* Create Suite Modal */}
      {showModal && (
        <SuiteModal onClose={() => setShowModal(false)} />
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

function SuiteModal({ onClose }: { onClose: () => void }) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    
    // TODO: Implement suite creation
    alert("Suite created!");
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6">
        <h2 className="text-lg font-semibold mb-4">Create New Test Suite</h2>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="suite-name" className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">
              Suite Name <span className="text-red-500">*</span>
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
            <Button type="button" variant="secondary" onClick={onClose}>Cancel</Button>
            <Button type="submit" disabled={!name.trim()}>Create Suite</Button>
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
