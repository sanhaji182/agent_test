import asyncio
import uuid
from fastapi import FastAPI, BackgroundTasks
from pydantic import BaseModel
from typing import Optional
from graph import build_graph

app = FastAPI(title="GoTest Agent LangGraph Sidecar")
graph = build_graph()

# In-memory job store
jobs: dict[str, dict] = {}


class RunRequest(BaseModel):
    run_id: str
    project_path: str
    requirements: str = ""
    code_analysis: str = ""
    max_fixes: int = 3


class RunResponse(BaseModel):
    job_id: str
    status: str


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/agent/run", response_model=RunResponse)
async def start_run(req: RunRequest, background_tasks: BackgroundTasks):
    job_id = str(uuid.uuid4())
    jobs[job_id] = {"status": "running", "result": None}
    background_tasks.add_task(execute_graph, job_id, req)
    return RunResponse(job_id=job_id, status="running")


@app.get("/agent/{job_id}")
def get_status(job_id: str):
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
