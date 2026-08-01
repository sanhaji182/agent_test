"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Badge } from "@/components/ui/badge";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";
import { ClipboardCheck, Search, Check, XCircle, Clock, ArrowRight } from "lucide-react";

export default function ReviewsPage() {
  const [reviews, setReviews] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState("pending");

  useEffect(() => {
    // TODO: Fetch real reviews data
    setTimeout(() => {
      setReviews([
        { 
          id: "rev_1", 
          title: "Checkout Flow Improvements", 
          testCount: 12,
          status: "pending",
          createdAt: "2026-07-31T10:00:00Z"
        },
        { 
          id: "rev_2", 
          title: "API Endpoint Validation", 
          testCount: 8,
          status: "approved",
          createdAt: "2026-07-30T14:30:00Z"
        },
        { 
          id: "rev_3", 
          title: "Authentication Tests", 
          testCount: 5,
          status: "rejected",
          createdAt: "2026-07-29T09:15:00Z"
        },
      ]);
      setLoading(false);
    }, 500);
  }, []);

  const filteredReviews = reviews.filter(r => filter === "all" || r.status === filter);

  if (loading) return <LoadingSkeleton rows={6} />;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-xl font-semibold tracking-tight mb-1">Test Reviews</h1>
        <p className="text-sm text-[var(--text-muted)] mt-1">Review and approve AI-generated test cases before deployment</p>
      </div>

      {/* Filters - Modern Segment Control */}
      <div className="inline-flex rounded-lg border border-[var(--border-default)] p-0.5 bg-white">
        {[
          { value: "pending", label: "Pending" },
          { value: "approved", label: "Approved" },
          { value: "rejected", label: "Rejected" },
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
            {f.label} ({getCount(filteredReviews, f.value)})
          </button>
        ))}
      </div>

      {/* Reviews List */}
      {filteredReviews.length === 0 ? (
        <Section title="No reviews found">
          <EmptyState 
            icon={<ClipboardCheck className="w-8 h-8" />}
            title={!filter || filter === "pending" ? "No pending reviews" : `No ${filter} reviews`}
            description={!filter || filter === "pending" ? "All test cases have been reviewed." : "No reviews with this status."}
          />
        </Section>
      ) : (
        <TableContainer>
          <table className="w-full text-left">
            <thead className="bg-gray-50 border-b border-[var(--border-default)]">
              <tr>
                <Th>Title</Th>
                <Th>Tests</Th>
                <Th>Status</Th>
                <Th>Date</Th>
                <Th align="right">Actions</Th>
              </tr>
            </thead>
            <tbody>
              {filteredReviews.map((review) => (
                <Tr key={review.id} hover>
                  <Td className="font-medium">
                    <span>{review.title}</span>
                  </Td>
                  <Td className="text-sm text-[var(--text-muted)]">{review.testCount} tests</Td>
                  <Td>
                    <StatusBadge status={review.status} />
                  </Td>
                  <Td className="text-sm text-[var(--text-muted)] whitespace-nowrap">
                    {formatDate(review.createdAt)}
                  </Td>
                  <Td align="right" className="space-x-2">
                    {review.status === "pending" && (
                      <>
                        <Button variant="secondary" size="sm">
                          <Check className="w-3.5 h-3.5 mr-1" />
                          Approve
                        </Button>
                        <Button variant="ghost" size="sm" className="text-red-600 hover:text-red-700 hover:bg-red-50">
                          <XCircle className="w-3.5 h-3.5" />
                          Reject
                        </Button>
                      </>
                    )}
                    <Link href={`/reviews/${review.id}`}>
                      <Button variant="ghost" size="sm">
                        View Details <ArrowRight className="w-3.5 h-3.5 ml-1" />
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
        Approved
      </Badge>
    );
  } else if (status === "rejected") {
    return (
      <Badge variant="danger" size="sm">
        <XCircle className="w-3 h-3 mr-1" />
        Rejected
      </Badge>
    );
  } else {
    return (
      <Badge variant="warning" size="sm">
        <Clock className="w-3 h-3 mr-1" />
        Pending
      </Badge>
    );
  }
}

function getCount(list: any[], status: string): number {
  return list.filter(r => r.status === status).length;
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
}
