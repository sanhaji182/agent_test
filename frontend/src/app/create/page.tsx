"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { createRun } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { LoadingSkeleton } from "@/components/ui/section";

export default function CreatePage() {
  const router = useRouter();
  const [formData, setFormData] = useState({
    project_path: "",
    requirements: "",
    base_url: "",
    name: "",
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.project_path.trim() || !formData.requirements.trim()) return;
    
    setLoading(true);
    setError(null);
    
    try {
      const result = await createRun({
        ...formData,
        mode: "simple", // Use simple mode for now
      });
      
      if (result && result.run_id) {
        router.push(`/runs/${result.run_id}`);
      } else {
        setError("Failed to create test run. Please try again.");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (field: string, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-xl font-semibold tracking-tight mb-2">Create Test Run</h1>
        <p className="text-sm text-[var(--text-muted)]">Define your requirements and let AI generate comprehensive tests</p>
      </div>

      {/* Form Card */}
      <div className="bg-white rounded-lg border border-[var(--border-default)] p-6 shadow-xs">
        <form onSubmit={handleSubmit} className="space-y-5">
          {/* Project Path */}
          <Input
            label="Project Path"
            placeholder="/path/to/your/project"
            value={formData.project_path}
            onChange={(e) => handleChange("project_path", e.target.value)}
            required
            helperText="Local path to the project you want to test"
          />

          {/* Requirements */}
          <div>
            <label htmlFor="requirements" className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">
              Test Requirements <span className="text-red-500">*</span>
            </label>
            <textarea
              id="requirements"
              value={formData.requirements}
              onChange={(e) => handleChange("requirements", e.target.value)}
              placeholder="Describe what behaviors you want to test... Example: 'Test user login flow including valid credentials, invalid password, and account lockout'"
              className="w-full h-32 px-3 py-2 bg-white border border-[var(--border-default)] rounded-md text-sm resize-none focus:outline-none focus:border-[var(--accent)] focus:ring-2 focus:ring-[var(--accent)]/20"
              required
            />
            <p className="mt-1 text-xs text-[var(--text-muted)]">Be specific about scenarios, edge cases, and expected behaviors.</p>
          </div>

          {/* Base URL (Optional) */}
          <Input
            label="Base URL (Optional)"
            placeholder="https://app.example.com"
            value={formData.base_url}
            onChange={(e) => handleChange("base_url", e.target.value)}
            helperText="The base URL where the application is running"
          />

          {/* Test Name (Optional) */}
          <Input
            label="Test Name (Optional)"
            placeholder="My Login Test"
            value={formData.name}
            onChange={(e) => handleChange("name", e.target.value)}
            helperText="Give your test a meaningful name"
          />

          {/* Error Message */}
          {error && (
            <div className="rounded-lg bg-red-50 border border-red-200 p-3">
              <p className="text-sm text-red-700">{error}</p>
            </div>
          )}

          {/* Actions */}
          <div className="flex gap-3 pt-4 border-t border-[var(--border-default)]">
            <Button 
              type="submit" 
              disabled={loading || !formData.project_path.trim() || !formData.requirements.trim()}
              className="w-40"
            >
              {loading ? (
                <span className="flex items-center gap-2">
                  <svg className="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  Creating...
                </span>
              ) : (
                "Create Test"
              )}
            </Button>
            <Button 
              type="button" 
              variant="secondary" 
              onClick={() => router.back()}
              className="w-32"
            >
              Cancel
            </Button>
          </div>
        </form>
      </div>

      {/* Tips Section */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <h3 className="text-sm font-semibold text-blue-800 mb-2 flex items-center gap-2">
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          Tips for Better Results
        </h3>
        <ul className="text-sm text-blue-700 space-y-1 ml-6 list-disc">
          <li>Be specific about user flows and acceptance criteria</li>
          <li>Include both happy paths and edge cases</li>
          <li>Specify expected behaviors, not implementation details</li>
          <li>Consider different authentication states (logged in/out)</li>
          <li>Think about data validation and error handling</li>
        </ul>
      </div>
    </div>
  );
}
