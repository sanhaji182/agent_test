"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Badge } from "@/components/ui/badge";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";
import { Download, FileJson, CheckCircle2, Code2, PlayCircle } from "lucide-react";

export default function ExportsPage() {
  const [runs, setRuns] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // TODO: Fetch real runs data
    setTimeout(() => {
      setRuns([
        { id: "run_1", name: "Checkout Flow Test", date: "2026-07-31" },
        { id: "run_2", name: "Login Authentication", date: "2026-07-30" },
        { id: "run_3", name: "User Registration", date: "2026-07-29" },
      ]);
      setLoading(false);
    }, 500);
  }, []);

  if (loading) return <LoadingSkeleton rows={6} />;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-xl font-semibold tracking-tight mb-1">Test Exports</h1>
        <p className="text-sm text-[var(--text-muted)]">Export test scripts in various formats for your CI/CD pipeline</p>
      </div>

      {/* Export Formats Info */}
      <Section title="Supported Formats">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <FormatCard 
            icon={<PlayCircle className="w-6 h-6" />}
            name="Playwright"
            description="TypeScript/JavaScript"
            language="ts"
          />
          <FormatCard 
            icon={<Code2 className="w-6 h-6" />}
            name="Cypress"
            description="JavaScript"
            language="js"
          />
          <FormatCard 
            icon={<FileJson className="w-6 h-6" />}
            name="Selenium"
            description="Python"
            language="py"
          />
          <FormatCard 
            icon={<Download className="w-6 h-6" />}
            name="JUnit XML"
            description="Test reports"
            language="xml"
          />
        </div>
      </Section>

      {/* Recent Runs Available for Export */}
      <Section title="Recent Runs" action={<Link href="/runs"><span className="text-xs text-blue-600 hover:text-blue-700">View All →</span></Link>}>
        {runs.length === 0 ? (
          <EmptyState 
            icon={<CheckCircle2 className="w-8 h-8" />}
            title="No completed runs"
            description="Run tests first before exporting them to your preferred framework."
          />
        ) : (
          <TableContainer>
            <table className="w-full text-left">
              <thead className="bg-gray-50 border-b border-[var(--border-default)]">
                <tr>
                  <Th>Test Name</Th>
                  <Th>Passed</Th>
                  <Th>Date</Th>
                  <Th align="right">Actions</Th>
                </tr>
              </thead>
              <tbody>
                {runs.map((run) => (
                  <Tr key={run.id} hover>
                    <Td className="font-medium">{run.name}</Td>
                    <Td><Badge variant="success" size="sm">✓</Badge></Td>
                    <Td className="text-sm text-[var(--text-muted)] whitespace-nowrap">{formatDate(run.date)}</Td>
                    <Td align="right" className="space-x-2">
                      <Button variant="secondary" size="sm">
                        <Download className="w-3.5 h-3.5 mr-1" />
                        Export
                      </Button>
                    </Td>
                  </Tr>
                ))}
              </tbody>
            </table>
          </TableContainer>
        )}
      </Section>
    </div>
  );
}

function FormatCard({ 
  icon, 
  name, 
  description,
  language 
}: { 
  icon: React.ReactNode; 
  name: string; 
  description: string;
  language: string;
}) {
  return (
    <div className="flex items-start gap-3 p-4 rounded-lg bg-white border border-[var(--border-default)]">
      <div className="shrink-0 p-2 bg-gray-100 rounded-lg">
        {icon}
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-semibold text-[var(--text-primary)]">{name}</p>
        <p className="text-xs text-[var(--text-muted)] mt-0.5">{description}</p>
        <p className="text-[10px] text-[var(--text-muted)] uppercase tracking-wide mt-1">Language: {language}</p>
      </div>
    </div>
  );
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}
