"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Section, EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { BookOpen, FileText, HelpCircle, ChevronRight, ExternalLink } from "lucide-react";

const documentationSections = [
  {
    category: "Getting Started",
    icon: <BookOpen className="w-5 h-5" />,
    articles: [
      { title: "Quick Start Guide", path: "/docs/quick-start", summary: "Set up GoTest Agent in under 5 minutes" },
      { title: "Installation", path: "/docs/installation", summary: "Install on Linux, macOS, or Windows" },
      { title: "Configuration", path: "/docs/configuration", summary: "Configure your testing environment" },
    ]
  },
  {
    category: "Core Features",
    icon: <FileText className="w-5 h-5" />,
    articles: [
      { title: "AI Test Generation", path: "/docs/ai-generation", summary: "Generate tests using AI-powered analysis" },
      { title: "Record & Playback", path: "/docs/recording", summary: "Capture browser interactions and generate tests" },
      { title: "Multi-Language Support", path: "/docs/multi-language", summary: "Supported frameworks and languages" },
      { title: "Self-Healing Tests", path: "/docs/self-healing", summary: "Automatic test repair on DOM changes" },
    ]
  },
  {
    category: "Advanced Topics",
    icon: <HelpCircle className="w-5 h-5" />,
    articles: [
      { title: "CI/CD Integration", path: "/docs/cicd", summary: "Integrate with GitHub Actions, GitLab CI, Jenkins" },
      { title: "API Reference", path: "/docs/api", summary: "Complete API documentation" },
      { title: "Security Best Practices", path: "/docs/security", summary: "Secure deployment guidelines" },
      { title: "Performance Optimization", path: "/docs/performance", summary: "Optimize test execution speed" },
    ]
  },
  {
    category: "Troubleshooting",
    icon: <ExternalLink className="w-5 h-5" />,
    articles: [
      { title: "Common Issues", path: "/docs/troubleshooting", summary: "Solve common problems quickly" },
      { title: "FAQ", path: "/docs/faq", summary: "Frequently asked questions" },
      { title: "Support", path: "/docs/support", summary: "Get help from our team" },
    ]
  },
];

export default function DocsPage() {
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Simulate loading
    setTimeout(() => setLoading(false), 300);
  }, []);

  if (loading) return <LoadingSkeleton rows={8} />;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-xl font-semibold tracking-tight mb-1">Documentation</h1>
        <p className="text-sm text-[var(--text-muted)] mt-1">Learn how to use GoTest Agent effectively</p>
      </div>

      {/* Search Box (Placeholder for future implementation) */}
      <div className="rounded-lg border border-[var(--border-default)] p-4 bg-white">
        <div className="flex items-center gap-2 text-[var(--text-muted)]">
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <span className="text-sm">Search documentation...</span>
        </div>
      </div>

      {/* Documentation Sections */}
      <div className="space-y-6">
        {documentationSections.map((section, index) => (
          <Section key={index} title={section.category}>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {section.articles.map((article, idx) => (
                <DocCard key={idx} article={article} icon={section.icon} />
              ))}
            </div>
          </Section>
        ))}
      </div>

      {/* Additional Resources */}
      <Section title="Additional Resources">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <ResourceCard 
            title="GitHub Repository"
            description="Source code, issues, and contributions"
            icon={<svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>}
            link="https://github.com/sanhaji182/agent_test"
            external
          />
          <ResourceCard 
            title="Community Support"
            description="Join discussions and get help from the community"
            icon={<svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8h2a2 2 0 012 2v6a2 2 0 01-2 2h-2v4l-4-4H9a1.994 1.994 0 01-1.414-.586m0 0L11 14h4a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2v4l.586-.586z"/></svg>}
            link="#"
          />
        </div>
      </Section>
    </div>
  );
}

function DocCard({ article, icon }: { article: any; icon: React.ReactNode }) {
  return (
    <Link href={article.path} className="group block p-4 rounded-lg border border-[var(--border-default)] bg-white hover:border-[var(--accent)] hover:shadow-sm transition-all">
      <div className="flex items-start gap-3">
        <div className="shrink-0 p-2 bg-gray-100 rounded-lg group-hover:bg-[var(--accent-light)] transition-colors">
          {icon}
        </div>
        <div className="flex-1 min-w-0">
          <h3 className="text-sm font-medium text-[var(--text-primary)] group-hover:text-[var(--accent)]">
            {article.title}
          </h3>
          <p className="text-xs text-[var(--text-muted)] mt-1 line-clamp-2">{article.summary}</p>
        </div>
        <ChevronRight className="w-4 h-4 text-[var(--text-muted)] group-hover:text-[var(--accent)] shrink-0 transition-colors" />
      </div>
    </Link>
  );
}

function ResourceCard({ title, description, icon, link, external }: { 
  title: string; 
  description: string; 
  icon: React.ReactNode;
  link: string;
  external?: boolean;
}) {
  return (
    <a 
      href={link} 
      target={external ? "_blank" : undefined}
      rel={external ? "noopener noreferrer" : undefined}
      className="block p-4 rounded-lg border border-[var(--border-default)] bg-white hover:border-[var(--accent)] hover:shadow-sm transition-all"
    >
      <div className="flex items-start gap-3">
        <div className="shrink-0 p-2 bg-blue-100 rounded-lg text-blue-600">
          {icon}
        </div>
        <div className="flex-1 min-w-0">
          <h3 className="text-sm font-medium text-[var(--text-primary)]">{title}</h3>
          <p className="text-xs text-[var(--text-muted)] mt-1">{description}</p>
          {external && <span className="text-xs text-[var(--text-muted)]">Opens in new window →</span>}
        </div>
      </div>
    </a>
  );
}
