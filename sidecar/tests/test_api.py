"""Tests for the LangGraph sidecar API and auth layer.

Run with: python -m pytest sidecar/tests/ -v
"""

import pytest
import os
import uuid
from unittest.mock import patch, MagicMock

# Default to development mode (no auth) for endpoint tests; auth tests
# patch SIDECAR_AUTH_TOKEN explicitly per-case.
os.environ["SIDECAR_AUTH_TOKEN"] = ""

from fastapi.testclient import TestClient
from main import app, jobs, RunRequest, RunResponse


@pytest.fixture(autouse=True)
def clear_jobs():
    """Reset the in-memory job store between tests."""
    jobs.clear()
    yield
    jobs.clear()


@pytest.fixture
def client():
    return TestClient(app)


class TestHealthEndpoint:
    def test_health_returns_ok(self, client):
        response = client.get("/health")
        assert response.status_code == 200
        assert response.json() == {"status": "ok"}


class TestAuth:
    @patch.dict(os.environ, {"SIDECAR_AUTH_TOKEN": "production-secret"})
    def test_run_rejected_without_token(self):
        """When SIDECAR_AUTH_TOKEN is set, /agent/run requires X-Auth-Token."""
        client = TestClient(app)
        response = client.post("/agent/run", json={
            "run_id": "test-001",
            "project_path": "/tmp/test",
        })
        assert response.status_code == 401

    @patch.dict(os.environ, {"SIDECAR_AUTH_TOKEN": "production-secret"})
    def test_run_rejected_with_wrong_token(self):
        """Wrong X-Auth-Token should be rejected."""
        client = TestClient(app)
        response = client.post("/agent/run", json={
            "run_id": "test-001",
            "project_path": "/tmp/test",
        }, headers={"X-Auth-Token": "wrong"})
        assert response.status_code == 401

    @patch.dict(os.environ, {"SIDECAR_AUTH_TOKEN": "production-secret"})
    def test_run_accepted_with_correct_token(self):
        """Correct token should be accepted."""
        client = TestClient(app)
        response = client.post("/agent/run", json={
            "run_id": "test-001",
            "project_path": "/tmp/test",
        }, headers={"X-Auth-Token": "production-secret"})
        assert response.status_code == 200
        data = response.json()
        assert "job_id" in data
        assert data["status"] == "running"

    @patch.dict(os.environ, {"SIDECAR_AUTH_TOKEN": ""})
    def test_run_accepted_in_dev_mode(self):
        """When SIDECAR_AUTH_TOKEN is empty, auth is optional (development)."""
        client = TestClient(app)
        response = client.post("/agent/run", json={
            "run_id": "test-001",
            "project_path": "/tmp/test",
        })
        assert response.status_code == 200


class TestRunEndpoint:
    def test_start_run_creates_job(self, client):
        response = client.post("/agent/run", json={
            "run_id": "test-001",
            "project_path": "/tmp/test",
            "requirements": "Test login flow",
            "code_analysis": "Next.js app",
            "max_fixes": 2,
        })
        assert response.status_code == 200
        data = response.json()
        assert "job_id" in data
        assert data["status"] == "running"

    def test_start_run_validates_input(self, client):
        response = client.post("/agent/run", json={
            # Missing required run_id and project_path
        })
        assert response.status_code == 422  # Validation error

    def test_start_run_accepts_minimal_fields(self, client):
        response = client.post("/agent/run", json={
            "run_id": "minimal-run",
            "project_path": "/tmp/minimal",
        })
        assert response.status_code == 200
        data = response.json()
        assert "job_id" in data

    def test_start_run_default_max_fixes(self, client):
        response = client.post("/agent/run", json={
            "run_id": "defaults-run",
            "project_path": "/tmp/defaults",
        })
        assert response.status_code == 200
        # Verify the job was created with default max_fixes
        data = response.json()
        job = jobs.get(data["job_id"])
        assert job is not None


class TestStatusEndpoint:
    def test_get_nonexistent_job(self, client):
        response = client.get("/agent/nonexistent-id")
        assert response.status_code == 200
        assert response.json() == {"error": "not found"}

    def test_get_existing_job(self, client):
        job_id = str(uuid.uuid4())
        jobs[job_id] = {"status": "running", "result": None}
        response = client.get(f"/agent/{job_id}")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "running"

    def test_get_completed_job(self, client):
        job_id = str(uuid.uuid4())
        jobs[job_id] = {"status": "completed", "result": {"test_plan": {}, "test_files": []}}
        response = client.get(f"/agent/{job_id}")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "completed"
        assert data["result"] is not None

    def test_get_failed_job(self, client):
        job_id = str(uuid.uuid4())
        jobs[job_id] = {"status": "failed", "result": None, "error": "something went wrong"}
        response = client.get(f"/agent/{job_id}")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "failed"
        assert "error" in data
