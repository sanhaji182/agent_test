"use client";

export default function ProjectsPage() {
  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Projects</h1>
      <div className="text-center py-12 text-zinc-500 border rounded-lg">
        <p className="text-lg">No projects registered</p>
        <p className="text-sm mt-1">
          Projects are auto-detected when you run tests via MCP or API.
        </p>
      </div>
    </div>
  );
}
