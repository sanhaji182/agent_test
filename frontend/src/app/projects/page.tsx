"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { getProjects, createProject, type Project } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { EmptyState, LoadingSkeleton } from "@/components/ui/section";
import { Plus, Settings, Eye, GitBranch } from "lucide-react";
import { TableContainer, Th, Td, Tr } from "@/components/ui/table";

export default function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [newProject, setNewProject] = useState({ name: "", test_type: "ui" });
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    getProjects().then((p) => setProjects(p || [])).catch((e) => setError(e.message)).finally(() => setLoading(false));
  }, []);

  const filteredProjects = projects.filter(p => 
    p.name.toLowerCase().includes(query.toLowerCase())
  );

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newProject.name.trim() || creating) return;
    setCreating(true);
    setError(null);
    try {
      await createProject({ name: newProject.name.trim(), test_type: newProject.test_type });
      // Refresh list from the server so the new project shows up with its real data.
      const refreshed = await getProjects();
      setProjects(refreshed || []);
      setShowForm(false);
      setNewProject({ name: "", test_type: "ui" });
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setCreating(false);
    }
  };

  if (loading) return <LoadingSkeleton rows={6} />;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-xl font-semibold tracking-tight">Projects</h1>
          <p className="text-sm text-[var(--text-muted)] mt-1">Manage your testing projects and configurations</p>
        </div>
        <Button onClick={() => setShowForm(!showForm)}>
          <Plus className="w-4 h-4" />
          New Project
        </Button>
      </div>

      {/* Add Project Form */}
      {showForm && (
        <div className="bg-white rounded-lg border border-[var(--border-default)] p-6 shadow-xs">
          <h2 className="text-base font-semibold mb-4">Create New Project</h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              label="Project Name"
              placeholder="My Test Project"
              value={newProject.name}
              onChange={(e) => setNewProject(prev => ({ ...prev, name: e.target.value }))}
              required
            />
            
            <div>
              <label htmlFor="test_type" className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">
                Test Type
              </label>
              <select
                id="test_type"
                value={newProject.test_type}
                onChange={(e) => setNewProject(prev => ({ ...prev, test_type: e.target.value }))}
                className="w-full px-3 py-2 bg-white border border-[var(--border-default)] rounded-md text-sm focus:outline-none focus:border-[var(--accent)] focus:ring-2 focus:ring-[var(--accent)]/20"
              >
                <option value="ui">UI Tests</option>
                <option value="api">API Tests</option>
                <option value="mixed">Mixed</option>
              </select>
            </div>
            
            <div className="flex gap-3 pt-4 border-t">
              <Button type="submit" disabled={!newProject.name || creating}>{creating ? "Creating..." : "Create"}</Button>
              <Button type="button" variant="secondary" onClick={() => setShowForm(false)}>Cancel</Button>
            </div>
          </form>
        </div>
      )}

      {/* Search */}
      <div className="relative max-w-md">
        <Input
          placeholder="Search projects..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="pl-9"
        />
      </div>

      {/* Projects List */}
      {filteredProjects.length === 0 ? (
        <div className="rounded-lg border border-[var(--border-default)] bg-white">
          <EmptyState 
            icon={<GitBranch className="w-6 h-6" />}
            title={!query ? "No projects yet" : "No matching projects"}
            description={!query ? "Create your first project to start generating tests." : "Try adjusting your search criteria."}
          />
        </div>
      ) : (
        <TableContainer>
          <table className="w-full text-left">
            <thead className="bg-gray-50 border-b border-[var(--border-default)]">
              <tr>
                <Th>Name</Th>
                <Th>Type</Th>
                <Th>Environment</Th>
                <Th align="right">Actions</Th>
              </tr>
            </thead>
            <tbody>
              {filteredProjects.map((project) => (
                <Tr key={project.id} hover>
                  <Td className="font-medium">{project.name}</Td>
                  <Td>
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize bg-blue-100 text-blue-700">
                      {project.test_type}
                    </span>
                  </Td>
                  <Td>
                    {project.environment ? (
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-700 capitalize">
                        {project.environment}
                      </span>
                    ) : "-"}
                  </Td>
                  <Td align="right" className="space-x-2">
                    <Link href={`/projects/${project.id}/settings`}>
                      <Button variant="ghost" size="sm">
                        <Settings className="w-3.5 h-3.5" />
                      </Button>
                    </Link>
                    <Link href={`/projects/${project.id}`}>
                      <Button variant="ghost" size="sm">
                        <Eye className="w-3.5 h-3.5" />
                      </Button>
                    </Link>
                  </Td>
                </Tr>
              ))}
            </tbody>
          </table>
        </TableContainer>
      )}

      {/* Error Message */}
      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 p-4">
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}
    </div>
  );
}
