"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  getAllReviews,
  approveReview,
  rejectReview,
  getUserLabel,
  type Review,
} from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Badge } from "@/components/ui/badge";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";
import { ClipboardCheck, Check, XCircle, Clock, ArrowRight } from "lucide-react";

export default function ReviewsPage() {
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState("pending");
  const [busyId, setBusyId] = useState<string | null>(null);

  const loadReviews = () => {
    return getAllReviews()
      .then((r) => setReviews(r || []))
      .catch((e) => setError(e.message));
  };

  useEffect(() => {
    loadReviews().finally(() => setLoading(false));
  }, []);

  const handleApprove = async (id: string) => {
    setBusyId(id);
    setError(null);
    try {
      await approveReview(id, getUserLabel() || "dashboard", "");
      await loadReviews();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusyId(null);
    }
  };

  const handleReject = async (id: string) => {
    setBusyId(id);
    setError(null);
    try {
      await rejectReview(id, getUserLabel() || "dashboard", "");
      await loadReviews();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusyId(null);
    }
  };

  const filteredReviews = reviews.filter(r => filter === "all" || r.status === filter);

  if (loading) return <LoadingSkeleton rows={6} />;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-xl font-semibold tracking-tight mb-1">Review Test</h1>
        <p className="text-sm text-[var(--text-muted)] mt-1">Review dan setujui test case hasil AI sebelum di-deploy</p>
      </div>

      {/* Error Message */}
      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 p-4">
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}

      {/* Filters - Modern Segment Control */}
      <div className="inline-flex rounded-lg border border-[var(--border-default)] p-0.5 bg-white">
        {[
          { value: "pending", label: "Menunggu" },
          { value: "approved", label: "Disetujui" },
          { value: "rejected", label: "Ditolak" },
        ].map((f) => (
          <button
            key={f.value}
            onClick={() => setFilter(f.value)}
            className={`px-3 py-1.5 text-xs font-medium rounded-md transition-colors ${
              filter === f.value
                ? "bg-blue-600 text-white shadow-sm"
                : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-gray-50"
            }`}
          >
            {f.label} ({getCount(reviews, f.value)})
          </button>
        ))}
      </div>

      {/* Reviews List */}
      {filteredReviews.length === 0 ? (
        <Section title="Tidak ada review">
          <EmptyState
            icon={<ClipboardCheck className="w-8 h-8" />}
            title={!filter || filter === "pending" ? "Tidak ada review yang menunggu" : `No ${filter} reviews`}
            description={!filter || filter === "pending" ? "Semua test case sudah di-review." : "No reviews with this status."}
          />
        </Section>
      ) : (
        <TableContainer>
          <table className="w-full text-left">
            <thead className="bg-gray-50 border-b border-[var(--border-default)]">
              <tr>
                <Th>Review</Th>
                <Th>Tipe</Th>
                <Th>Status</Th>
                <Th>Reviewer</Th>
                <Th>Tanggal</Th>
                <Th align="right">Aksi</Th>
              </tr>
            </thead>
            <tbody>
              {filteredReviews.map((review) => (
                <Tr key={review.id} hover>
                  <Td className="font-medium">
                    <span>Run {review.run_id.slice(0, 8)}</span>
                  </Td>
                  <Td className="text-sm text-[var(--text-muted)]">
                    {formatTipe(review.type)}
                  </Td>
                  <Td>
                    <StatusBadge status={review.status} />
                  </Td>
                  <Td className="text-sm text-[var(--text-muted)] whitespace-nowrap">
                    {review.reviewer || "-"}
                  </Td>
                  <Td className="text-sm text-[var(--text-muted)] whitespace-nowrap">
                    {formatDate(review.created_at)}
                  </Td>
                  <Td align="right" className="space-x-2">
                    {review.status === "pending" && (
                      <>
                        <Button
                          variant="secondary"
                          size="sm"
                          disabled={busyId === review.id}
                          onClick={() => handleApprove(review.id)}
                        >
                          <Check className="w-3.5 h-3.5 mr-1" />
                          Setujui
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          disabled={busyId === review.id}
                          className="text-red-600 hover:text-red-700 hover:bg-red-50"
                          onClick={() => handleReject(review.id)}
                        >
                          <XCircle className="w-3.5 h-3.5" />
                          Tolak
                        </Button>
                      </>
                    )}
                    <Link href={`/reviews/${review.id}`}>
                      <Button variant="ghost" size="sm">
                        Lihat Detail <ArrowRight className="w-3.5 h-3.5 ml-1" />
                      </Button>
                    </Link>
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

function StatusBadge({ status }: { status: string }) {
  if (status === "approved") {
    return (
      <Badge variant="success" size="sm">
        <Check className="w-3 h-3 mr-1" />
        Disetujui
      </Badge>
    );
  } else if (status === "rejected") {
    return (
      <Badge variant="danger" size="sm">
        <XCircle className="w-3 h-3 mr-1" />
        Ditolak
      </Badge>
    );
  } else {
    return (
      <Badge variant="warning" size="sm">
        <Clock className="w-3 h-3 mr-1" />
        Menunggu
      </Badge>
    );
  }
}

function getCount(list: Review[], status: string): number {
  return list.filter(r => r.status === status).length;
}

function formatTipe(type: string): string {
  return type.replace(/_/g, " ").replace(/\b\w/g, l => l.toUpperCase());
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
}
