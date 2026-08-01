"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Section, EmptyState } from "@/components/ui/section";
import { Badge } from "@/components/ui/badge";
import { 
  Settings, KeyRound, Shield, Database, Terminal, Globe, FileText,
  Save, CheckCircle2, AlertTriangle
} from "lucide-react";

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState<"general" | "ai" | "integrations">("general");
  const [settings, setSettings] = useState({
    // General settings
    appName: "GoTest Agent",
    timezone: "UTC",
    language: "en",
    
    // AI settings
    defaultProvider: "anthropic",
    maxTokens: 4096,
    temperature: 0.7,
    
    // Integration settings
    enableWebhooks: true,
    slackWebhookUrl: "",
    githubRepo: "",
  });

  const handleSave = () => {
    // TODO: Implement save functionality
    alert("Settings saved!");
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-xl font-semibold tracking-tight mb-1">Settings</h1>
        <p className="text-sm text-[var(--text-muted)]">Configure your testing environment and preferences</p>
      </div>

      {/* Tabs - Modern Segmented Control */}
      <div className="inline-flex rounded-lg border border-[var(--border-default)] p-0.5 bg-white">
        {[
          { id: "general", label: "General", icon: Settings },
          { id: "ai", label: "AI Configuration", icon: Terminal },
          { id: "integrations", label: "Integrations", icon: Globe },
        ].map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id as typeof activeTab)}
            className={`flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-md transition-colors ${
              activeTab === tab.id
                ? "bg-blue-600 text-white shadow-sm"
                : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-gray-50"
            }`}
          >
            <tab.icon className="w-4 h-4" />
            {tab.label}
          </button>
        ))}
      </div>

      {/* General Settings Tab */}
      {activeTab === "general" && (
        <Section title="General Settings">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">Application Name</label>
              <Input
                value={settings.appName}
                onChange={(e) => setSettings(prev => ({ ...prev, appName: e.target.value }))}
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">Timezone</label>
                <select
                  className="w-full px-3 py-2 bg-white border border-[var(--border-default)] rounded-md text-sm focus:outline-none focus:border-[var(--accent)] focus:ring-2 focus:ring-[var(--accent)]/20"
                  value={settings.timezone}
                  onChange={(e) => setSettings(prev => ({ ...prev, timezone: e.target.value }))}
                >
                  <option value="UTC">UTC</option>
                  <option value="America/New_York">Eastern Time</option>
                  <option value="America/Los_Angeles">Pacific Time</option>
                  <option value="Europe/London">London</option>
                  <option value="Asia/Tokyo">Tokyo</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">Language</label>
                <select
                  className="w-full px-3 py-2 bg-white border border-[var(--border-default)] rounded-md text-sm focus:outline-none focus:border-[var(--accent)] focus:ring-2 focus:ring-[var(--accent)]/20"
                  value={settings.language}
                  onChange={(e) => setSettings(prev => ({ ...prev, language: e.target.value }))}
                >
                  <option value="en">English</option>
                  <option value="id">Bahasa Indonesia</option>
                </select>
              </div>
            </div>
          </div>
        </Section>
      )}

      {/* AI Configuration Tab */}
      {activeTab === "ai" && (
        <Section title="AI Provider Configuration">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">Default LLM Provider</label>
              <select
                className="w-full px-3 py-2 bg-white border border-[var(--border-default)] rounded-md text-sm focus:outline-none focus:border-[var(--accent)] focus:ring-2 focus:ring-[var(--accent)]/20"
                value={settings.defaultProvider}
                onChange={(e) => setSettings(prev => ({ ...prev, defaultProvider: e.target.value }))}
              >
                <option value="anthropic">Anthropic Claude</option>
                <option value="openai">OpenAI GPT</option>
                <option value="google">Google Gemini</option>
                <option value="deepseek">DeepSeek</option>
              </select>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">Max Tokens</label>
                <Input
                  type="number"
                  value={settings.maxTokens}
                  onChange={(e) => setSettings(prev => ({ ...prev, maxTokens: parseInt(e.target.value) || 4096 }))}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">Temperature (0-1)</label>
                <Input
                  type="number"
                  min="0"
                  max="1"
                  step="0.1"
                  value={settings.temperature}
                  onChange={(e) => setSettings(prev => ({ ...prev, temperature: parseFloat(e.target.value) || 0.7 }))}
                />
              </div>
            </div>

            <div className="rounded-lg bg-blue-50 border border-blue-200 p-4 mt-4">
              <div className="flex items-start gap-3">
                <AlertTriangle className="w-5 h-5 text-blue-600 shrink-0 mt-0.5" />
                <div>
                  <p className="text-sm font-medium text-blue-800">API Key Configuration</p>
                  <p className="text-xs text-blue-700 mt-1">
                    Configure provider-specific API keys in the secure vault. Keys are encrypted at rest.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </Section>
      )}

      {/* Integrations Tab */}
      {activeTab === "integrations" && (
        <Section title="External Integrations">
          <div className="space-y-4">
            {/* Webhooks */}
            <div className="flex items-center justify-between p-4 bg-white border border-[var(--border-default)] rounded-lg">
              <div className="flex items-center gap-3">
                <FileText className="w-5 h-5 text-[var(--text-muted)]" />
                <div>
                  <p className="text-sm font-medium text-[var(--text-primary)]">GitHub Webhooks</p>
                  <p className="text-xs text-[var(--text-muted)] mt-0.5">Auto-detect test drift on code changes</p>
                </div>
              </div>
              <Badge variant={settings.enableWebhooks ? "success" : "default"} size="sm">
                {settings.enableWebhooks ? "Enabled" : "Disabled"}
              </Badge>
            </div>

            {/* Slack */}
            <div>
              <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">Slack Webhook URL</label>
              <Input
                placeholder="https://hooks.slack.com/services/..."
                value={settings.slackWebhookUrl}
                onChange={(e) => setSettings(prev => ({ ...prev, slackWebhookUrl: e.target.value }))}
                helperText="Receive test failure notifications in Slack"
              />
            </div>

            {/* GitHub */}
            <div>
              <label className="block text-sm font-medium text-[var(--text-primary)] mb-1.5">GitHub Repository</label>
              <Input
                placeholder="owner/repo"
                value={settings.githubRepo}
                onChange={(e) => setSettings(prev => ({ ...prev, githubRepo: e.target.value }))}
                helperText="Monitor this repository for test drift"
              />
            </div>
          </div>
        </Section>
      )}

      {/* Save Button */}
      <div className="flex justify-end gap-3 pt-4 border-t">
        <Button variant="secondary" onClick={() => window.location.reload()}>Cancel</Button>
        <Button onClick={handleSave}>
          <CheckCircle2 className="w-4 h-4 mr-2" />
          Save Changes
        </Button>
      </div>
    </div>
  );
}
