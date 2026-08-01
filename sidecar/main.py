# GoTest Agent — LangGraph multi-agent sidecar
#
# EXPERIMENTAL: Not wired to cmd/server. Agent.executeAdvanced branches on
# Mode="advanced" + SidecarClient, but NewServer never constructs SidecarClient.
# To wire: add sidecar URL config and construct agent.SidecarClient in NewServer.
#
import asyncio
import json
import os
import sqlite3
import threading
import uuid
from collections.abc import Iterator, MutableMapping
from pathlib import Path
from fastapi import FastAPI, BackgroundTasks, Depends, Header, HTTPException
from pydantic import BaseModel
from typing import Optional
from graph import build_graph

app = FastAPI(title="GoTest Agent LangGraph Sidecar")
graph = build_graph()

class JobStore(MutableMapping[str, dict]):
    """Dict-like sidecar job store.

    The default is in-memory storage for local development and existing tests.
    Set SIDECAR_JOBS_DB to a SQLite file path to persist job state across
    sidecar restarts without introducing another runtime dependency.
    """

    def __init__(self):
        self._memory: dict[str, dict] = {}
        self._lock = threading.RLock()
        self._initialized_paths: set[str] = set()

    def _db_path(self) -> str:
        return os.getenv("SIDECAR_JOBS_DB", "").strip()

    def _connect(self, path: str) -> sqlite3.Connection:
        db_path = Path(path)
        db_path.parent.mkdir(parents=True, exist_ok=True)
        conn = sqlite3.connect(str(db_path), timeout=5)
        conn.row_factory = sqlite3.Row
        with self._lock:
            if path not in self._initialized_paths:
                conn.execute(
                    """
                    CREATE TABLE IF NOT EXISTS jobs (
                        id TEXT PRIMARY KEY,
                        payload TEXT NOT NULL,
                        created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
                    )
                    """
                )
                conn.commit()
                self._initialized_paths.add(path)
        return conn

    @staticmethod
    def _decode(payload: str) -> dict:
        data = json.loads(payload)
        return data if isinstance(data, dict) else {"status": "failed", "result": None, "error": "invalid persisted payload"}

    def __getitem__(self, key: str) -> dict:
        path = self._db_path()
        if path:
            with self._connect(path) as conn:
                row = conn.execute("SELECT payload FROM jobs WHERE id = ?", (key,)).fetchone()
                if row is None:
                    raise KeyError(key)
                return self._decode(row["payload"])
        with self._lock:
            return self._memory[key]

    def __setitem__(self, key: str, value: dict) -> None:
        path = self._db_path()
        if path:
            payload = json.dumps(value, default=str, sort_keys=True)
            with self._connect(path) as conn:
                conn.execute(
                    """
                    INSERT INTO jobs (id, payload, created_at, updated_at)
                    VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                    ON CONFLICT(id) DO UPDATE SET
                        payload = excluded.payload,
                        updated_at = CURRENT_TIMESTAMP
                    """,
                    (key, payload),
                )
                conn.commit()
            return
        with self._lock:
            self._memory[key] = value

    def __delitem__(self, key: str) -> None:
        path = self._db_path()
        if path:
            with self._connect(path) as conn:
                result = conn.execute("DELETE FROM jobs WHERE id = ?", (key,))
                conn.commit()
                if result.rowcount == 0:
                    raise KeyError(key)
            return
        with self._lock:
            del self._memory[key]

    def __iter__(self) -> Iterator[str]:
        path = self._db_path()
        if path:
            with self._connect(path) as conn:
                rows = conn.execute("SELECT id FROM jobs ORDER BY created_at ASC").fetchall()
                return iter([row["id"] for row in rows])
        with self._lock:
            return iter(list(self._memory.keys()))

    def __len__(self) -> int:
        path = self._db_path()
        if path:
            with self._connect(path) as conn:
                row = conn.execute("SELECT COUNT(*) AS count FROM jobs").fetchone()
                return int(row["count"])
        with self._lock:
            return len(self._memory)

    def clear(self) -> None:
        path = self._db_path()
        if path:
            with self._connect(path) as conn:
                conn.execute("DELETE FROM jobs")
                conn.commit()
            return
        with self._lock:
            self._memory.clear()


jobs = JobStore()

# Internal service auth — shared token between backend and sidecar.
# This replaces the previous GOTEST_API_KEY pattern which passed the
# user-facing API key across an unauthenticated boundary.
# Read at request time (not import time) so token rotation and tests
# that patch the environment behave correctly.


class RunRequest(BaseModel):
    run_id: str
    project_path: str
    requirements: str = ""
    code_analysis: str = ""
    max_fixes: int = 3


class RunResponse(BaseModel):
    job_id: str
    status: str


def verify_auth(x_auth_token: Optional[str] = Header(None)) -> None:
    """Validate internal service token. If SIDECAR_AUTH_TOKEN is unset,
    the sidecar is running in development mode and auth is optional."""
    expected = os.getenv("SIDECAR_AUTH_TOKEN", "")
    if not expected:
        return  # Development mode — no auth required
    if not x_auth_token or x_auth_token != expected:
        raise HTTPException(status_code=401, detail="unauthorized")


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/agent/run", response_model=RunResponse)
async def start_run(req: RunRequest, background_tasks: BackgroundTasks, _auth: None = Depends(verify_auth)):
    """Start a LangGraph execution job. Requires X-Auth-Token header
    matching SIDECAR_AUTH_TOKEN when in production."""
    job_id = str(uuid.uuid4())
    jobs[job_id] = {"status": "running", "result": None}
    background_tasks.add_task(execute_graph, job_id, req)
    return RunResponse(job_id=job_id, status="running")


@app.get("/agent/{job_id}")
def get_status(job_id: str, _auth: None = Depends(verify_auth)):
    """Get the status of a LangGraph job."""
    job = jobs.get(job_id)
    if not job:
        return {"error": "not found"}
    return job


async def execute_graph(job_id: str, req: RunRequest):
    try:
        initial_state = {
            "run_id": req.run_id,
            "project_path": req.project_path,
            "requirements": req.requirements,
            "code_analysis": req.code_analysis,
            "test_plan": {},
            "test_files": [],
            "run_result": {},
            "fix_attempts": 0,
            "max_fixes": req.max_fixes,
            "critic_scores": [],
            "screenshots": [],
            "next_step": "",
            "error": None,
        }
        result = await asyncio.to_thread(graph.invoke, initial_state)
        jobs[job_id] = {"status": "completed", "result": result}
    except Exception as e:
        jobs[job_id] = {"status": "failed", "result": None, "error": str(e)}
