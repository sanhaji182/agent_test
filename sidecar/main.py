# GoTest Agent — LangGraph multi-agent sidecar
#
# EXPERIMENTAL: Not wired to cmd/server. Agent.executeAdvanced branches on
# Mode="advanced" + SidecarClient, but NewServer never constructs SidecarClient.
# To wire: add sidecar URL config and construct agent.SidecarClient in NewServer.
#
import asyncio
import os
import uuid
from fastapi import FastAPI, BackgroundTasks, Depends, Header, HTTPException
from pydantic import BaseModel
from typing import Optional
from graph import build_graph

app = FastAPI(title="GoTest Agent LangGraph Sidecar")
graph = build_graph()

# In-memory job store (Phase 1 — will be moved to Redis per ADR-003)
jobs: dict[str, dict] = {}

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
