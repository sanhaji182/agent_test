"use client";

import { useEffect, useState } from "react";
import { getAllReviews, approveReview, rejectReview, type Review } from "@/lib/api";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { ClipboardCheck, Check, X } from "lucide-react";
import Link from "next/link";

export default function ReviewsPage() {
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getAllReviews().then(setReviews).catch(() => {}).finally(() => setLoading(false));
  }, []);

  const handleApprove = async (id: string) => {
    await approveReview(id, "admin", "Approved");
    getAllReviews().then(setReviews);
  };

  const handleReject = async (id: string) => {
    await rejectReview(id, "admin", "Rejected");
    getAllReviews().then(setReviews);
  };

  if (loading) return <LoadingSkeleton rows={5} />;

  const pending = reviews.filter((r) => r.status === "pending");
  const resolved = reviews.filter((r) => r.status !== "pending");

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold">Review Queue</h1>
        <p className="text-sm text-[var(--text-secondary)] mt-0.5">Approve or reject generated test plans, scripts, and fixes</p>
      </div>

      <Section title="Pending Reviews" action={<span className="text-[11px] text-[var(--text-muted)]">{pending.length} pending</span>}>
        {pending.length === 0 ? (
          <EmptyState icon={<ClipboardCheck className="w-6 h-6" />} title="No pending reviews" description="Reviews are created when the agent generates test plans, scripts, or fix suggestions." />
        ) : (
          <div className="space-y-2">
            {pending.map((rev) => (
              <div key={rev.id} className="flex items-center gap-3 p-3 rounded-lg border border-[var(--border)] bg-[var(--bg-subtle)]">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-[var(--warning-bg)] text-[var(--warning)]">{rev.type.replace("_", " ")}</span>
                    <Link href={`/runs/${rev.run_id}`} className="text-xs font-mono text-[var(--accent)] hover:underline">{rev.run_id.slice(0, 8)}</Link>
                  </div>
                  <p className="text-[11px] text-[var(--text-muted)] mt-0.5">{new Date(rev.created_at).toLocaleString()}</p>
                </div>
                <button onClick={() => handleApprove(rev.id)} className="p-2 rounded-lg hover:bg-[var(--success-bg)] text-[var(--success)]" title="Approve">
                  <Check className="w-4 h-4" />
                </button>
                <button onClick={() => handleReject(rev.id)} className="p-2 rounded-lg hover:bg-[var(--danger-bg)] text-[var(--danger)]" title="Reject">
                  <X className="w-4 h-4" />
                </button>
              </div>
            ))}
          </div>
        )}
      </Section>

      {resolved.length > 0 && (
        <Section title="Resolved">
          <div className="space-y-1.5">
            {resolved.map((rev) => (
              <div key={rev.id} className="flex items-center gap-3 px-3 py-2 rounded-lg">
                <span className={`px-2 py-0.5 rounded text-[10px] font-semibold ${rev.status === "approved" ? "bg-[var(--success-bg)] text-[var(--success)]" : "bg-[var(--danger-bg)] text-[var(--danger)]"}`}>{rev.status}</span>
                <span className="text-xs text-[var(--text-secondary)]">{rev.type.replace("_", " ")}</span>
                <span className="text-[11px] text-[var(--text-muted)]">by {rev.reviewer || "—"}</span>
              </div>
            ))}
          </div>
        </Section>
      )}
    </div>
  );
}
