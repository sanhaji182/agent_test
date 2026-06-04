"use client";

import { useEffect, useMemo, useState } from "react";
import {
  createTestList,
  createSchedule,
  getTestCases,
  getTestLists,
  getSchedules,
  runTestList,
  runScheduleNow,
  type Schedule,
  type TestCase,
  type TestList,
} from "@/lib/api";
import { EmptyState, LoadingSkeleton, Section } from "@/components/ui/section";
import { CalendarClock, CheckCircle2, Layers, Pin, PlayCircle, Plus } from "lucide-react";
import { cn } from "@/lib/utils";

export default function SuitesPage() {
  const [lists, setLists] = useState<TestList[]>([]);
  const [cases, setCases] = useState<TestCase[]>([]);
  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [loading, setLoading] = useState(true);
  const [name, setName] = useState("Smoke Regression");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [runningId, setRunningId] = useState<string | null>(null);
  const [scheduleListId, setScheduleListId] = useState("");
  const [scheduleName, setScheduleName] = useState("Nightly Regression");
  const [scheduleFrequency, setScheduleFrequency] = useState("daily");
  const [runningScheduleId, setRunningScheduleId] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([getTestLists(), getTestCases(), getSchedules()])
      .then(([listData, caseData, scheduleData]) => {
        setLists(listData);
        setCases(caseData);
        setSchedules(scheduleData);
        setSelectedIds(caseData.slice(0, 5).map((c) => c.id));
        setScheduleListId(listData[0]?.id || "");
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const caseMap = useMemo(() => new Map(cases.map((c) => [c.id, c])), [cases]);
  const pinned = lists.filter((s) => s.pinned);
  const others = lists.filter((s) => !s.pinned);
  const listSchedules = schedules.filter((s) => s.test_list_id);

  const toggleCase = (id: string) => {
    setSelectedIds((prev) => prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]);
  };

  const createList = async () => {
    if (!name.trim() || selectedIds.length === 0) return;
    const first = caseMap.get(selectedIds[0]);
    const list = await createTestList({
      name,
      project_id: first?.project_id,
      tags: ["manual", "regression"],
      test_case_ids: selectedIds,
      pinned: lists.length === 0,
    });
    setLists((prev) => [list, ...prev]);
  };

  const runList = async (list: TestList) => {
    setRunningId(list.id);
    try {
      const res = await runTestList(list.id);
      if (res.run_ids.length > 0) window.location.href = `/runs/${res.run_ids[0]}`;
    } finally {
      setRunningId(null);
    }
  };

  const createListSchedule = async () => {
    if (!scheduleListId) return;
    const list = lists.find((item) => item.id === scheduleListId);
    const sch = await createSchedule({
      name: scheduleName.trim() || `${list?.name || "Test List"} schedule`,
      project_id: list?.project_id,
      test_list_id: scheduleListId,
      frequency: scheduleFrequency,
      mode: "approved_list",
      enabled: true,
    });
    setSchedules((prev) => [sch, ...prev]);
  };

  const runSchedule = async (sch: Schedule) => {
    setRunningScheduleId(sch.id);
    try {
      const res = await runScheduleNow(sch.id);
      const firstRun = res.run_ids?.[0] || res.run_id;
      if (firstRun) window.location.href = `/runs/${firstRun}`;
    } finally {
      setRunningScheduleId(null);
    }
  };

  if (loading) return <LoadingSkeleton rows={7} />;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-lg font-bold">Test Lists</h1>
        <p className="text-[13px] text-[var(--text-secondary)] mt-0.5">Group approved UI/API tests into repeatable lists and run them as a set.</p>
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-[420px_1fr] gap-5">
        <Section title="Create List" action={<Plus className="w-3.5 h-3.5 text-[var(--accent)]" />}>
          <div className="space-y-4">
            <input value={name} onChange={(e) => setName(e.target.value)} className="input" placeholder="List name" />
            {cases.length === 0 ? (
              <EmptyState icon={<Layers className="w-6 h-6" />} title="No approved test cases" description="Approve a generated plan first, then create a Test List here." />
            ) : (
              <div className="max-h-[360px] overflow-auto space-y-2 pr-1">
                {cases.map((testCase) => {
                  const selected = selectedIds.includes(testCase.id);
                  return (
                    <button
                      key={testCase.id}
                      onClick={() => toggleCase(testCase.id)}
                      className={cn(
                        "w-full text-left rounded-[var(--radius-sm)] border p-3 transition-colors",
                        selected ? "border-[var(--accent)]/40 bg-[var(--accent-bg)]" : "border-[var(--border)] bg-[var(--bg-subtle)] hover:bg-[var(--bg-hover)]"
                      )}
                    >
                      <div className="flex items-start gap-2">
                        <span className={cn("mt-0.5 w-4 h-4 rounded border flex items-center justify-center shrink-0", selected ? "bg-[var(--accent)] border-[var(--accent)] text-white" : "border-[var(--border-strong)]")}>
                          {selected && <CheckCircle2 className="w-3 h-3" />}
                        </span>
                        <div className="min-w-0">
                          <p className="text-[12px] font-semibold text-[var(--text-primary)] truncate">{testCase.title}</p>
                          <p className="text-[11px] text-[var(--text-muted)] truncate">{testCase.feature || testCase.type}</p>
                        </div>
                      </div>
                    </button>
                  );
                })}
              </div>
            )}
            <button
              onClick={createList}
              disabled={!name.trim() || selectedIds.length === 0}
              className="inline-flex items-center justify-center gap-1.5 w-full px-4 py-2 rounded-[var(--radius-sm)] bg-[var(--accent)] text-white text-[13px] font-semibold hover:bg-[var(--accent-hover)] disabled:opacity-50"
            >
              Create Test List
            </button>
          </div>
        </Section>

        <div className="space-y-5">
          {lists.length === 0 ? (
            <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-card)]">
              <EmptyState icon={<Layers className="w-6 h-6" />} title="No test lists yet" description="Select approved test cases and create your first smoke or regression list." />
            </div>
          ) : (
            <>
              {pinned.length > 0 && (
                <Section title="Pinned Lists" action={<Pin className="w-3.5 h-3.5 text-[var(--accent)]" />}>
                  <div className="space-y-2">
                    {pinned.map((list) => <ListRow key={list.id} list={list} caseMap={caseMap} running={runningId === list.id} onRun={() => runList(list)} />)}
                  </div>
                </Section>
              )}
              <Section title="All Lists" action={<span className="text-[11px] text-[var(--text-muted)]">{lists.length} total</span>}>
                <div className="space-y-2">
                  {others.map((list) => <ListRow key={list.id} list={list} caseMap={caseMap} running={runningId === list.id} onRun={() => runList(list)} />)}
                </div>
              </Section>
              <Section title="Recurring Runs" action={<CalendarClock className="w-3.5 h-3.5 text-[var(--accent)]" />}>
                <div className="grid grid-cols-1 lg:grid-cols-[1fr_120px_auto] gap-2">
                  <input value={scheduleName} onChange={(e) => setScheduleName(e.target.value)} className="input" placeholder="Schedule name" />
                  <select value={scheduleFrequency} onChange={(e) => setScheduleFrequency(e.target.value)} className="input">
                    <option value="daily">Daily</option>
                    <option value="weekly">Weekly</option>
                    <option value="monthly">Monthly</option>
                  </select>
                  <button onClick={createListSchedule} disabled={!scheduleListId} className="inline-flex items-center justify-center gap-1.5 px-3 py-2 rounded-[var(--radius-sm)] bg-[var(--accent)] text-white text-[12px] font-semibold hover:bg-[var(--accent-hover)] disabled:opacity-50">
                    <Plus className="w-3.5 h-3.5" /> Schedule
                  </button>
                </div>
                <select value={scheduleListId} onChange={(e) => setScheduleListId(e.target.value)} className="input mt-2">
                  {lists.map((list) => <option key={list.id} value={list.id}>{list.name}</option>)}
                </select>
                <div className="space-y-2 mt-3">
                  {listSchedules.length === 0 ? (
                    <p className="text-[11px] text-[var(--text-muted)]">No Test List schedules yet.</p>
                  ) : listSchedules.map((sch) => (
                    <div key={sch.id} className="flex items-center gap-3 p-3 rounded-lg border border-[var(--border)] bg-[var(--bg-subtle)]">
                      <div className="flex-1 min-w-0">
                        <p className="text-[12px] font-semibold truncate">{sch.name}</p>
                        <p className="text-[11px] text-[var(--text-muted)] truncate">{sch.frequency} - next {new Date(sch.next_run_at).toLocaleString()}</p>
                      </div>
                      <button onClick={() => runSchedule(sch)} disabled={runningScheduleId === sch.id} className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-[var(--radius-sm)] bg-[var(--bg-card)] border border-[var(--border)] text-[11px] font-semibold hover:bg-[var(--bg-hover)] disabled:opacity-50">
                        <PlayCircle className="w-3.5 h-3.5" />
                        {runningScheduleId === sch.id ? "Starting" : "Run"}
                      </button>
                    </div>
                  ))}
                </div>
              </Section>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

function ListRow({ list, caseMap, running, onRun }: { list: TestList; caseMap: Map<string, TestCase>; running: boolean; onRun: () => void }) {
  return (
    <div className="flex items-center gap-3 p-3 rounded-lg border border-[var(--border)] bg-[var(--bg-subtle)]">
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold">{list.name}</span>
          {list.pinned && <Pin className="w-3 h-3 text-[var(--accent)]" />}
        </div>
        <div className="flex gap-1 mt-1 flex-wrap">
          {list.test_case_ids.slice(0, 4).map((id) => (
            <span key={id} className="px-1.5 py-0.5 rounded text-[10px] bg-[var(--bg-card)] border border-[var(--border)] text-[var(--text-muted)]">
              {caseMap.get(id)?.title || id.slice(0, 8)}
            </span>
          ))}
        </div>
      </div>
      <span className="text-[11px] text-[var(--text-muted)]">{list.test_case_ids.length} tests</span>
      <button onClick={onRun} disabled={running} className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-[var(--radius-sm)] bg-[var(--accent)] text-white text-[11px] font-semibold hover:bg-[var(--accent-hover)] disabled:opacity-50">
        <PlayCircle className="w-3.5 h-3.5" />
        {running ? "Starting" : "Run"}
      </button>
    </div>
  );
}
