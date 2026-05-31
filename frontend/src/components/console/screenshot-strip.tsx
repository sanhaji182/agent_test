"use client";

import { useState } from "react";
import { Image as ImageIcon, X } from "lucide-react";
import { EmptyState } from "@/components/ui/section";

const API = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

function fullUrl(u: string) {
  return u.startsWith("http") ? u : `${API}${u}`;
}

// Strip thumbnail screenshot dengan lightbox preview
export function ScreenshotStrip({ screenshots }: { screenshots?: string[] }) {
  const [preview, setPreview] = useState<string | null>(null);

  if (!screenshots || screenshots.length === 0) {
    return (
      <EmptyState
        icon={<ImageIcon className="w-6 h-6" />}
        title="No screenshots captured"
        description="Screenshots are captured automatically when a test step fails."
      />
    );
  }

  return (
    <>
      <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
        {screenshots.map((url, i) => (
          <button
            key={i}
            onClick={() => setPreview(fullUrl(url))}
            className="group relative aspect-video rounded-lg border border-[var(--border)] bg-[var(--bg-subtle)] overflow-hidden hover:border-[var(--accent)] transition-colors"
          >
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={fullUrl(url)}
              alt={`Screenshot ${i + 1}`}
              className="w-full h-full object-cover"
            />
            <div className="absolute bottom-0 inset-x-0 px-2 py-1 bg-black/60 text-white text-[10px] truncate opacity-0 group-hover:opacity-100 transition-opacity">
              {url.split("/").pop()}
            </div>
          </button>
        ))}
      </div>

      {/* Lightbox */}
      {preview && (
        <div
          className="fixed inset-0 z-50 bg-black/80 flex items-center justify-center p-8"
          onClick={() => setPreview(null)}
        >
          <button className="absolute top-4 right-4 p-2 text-white/80 hover:text-white">
            <X className="w-5 h-5" />
          </button>
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src={preview} alt="Preview" className="max-w-full max-h-full rounded-lg" />
        </div>
      )}
    </>
  );
}
