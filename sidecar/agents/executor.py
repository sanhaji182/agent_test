import os
import httpx


GOTEST_API_URL = os.getenv("GOTEST_API_URL", "http://app:8080")
GOTEST_API_KEY = os.getenv("GOTEST_API_KEY", "")


def executor_node(state: dict) -> dict:
    """Trigger test execution via Go API and return results."""
    headers = {"X-Api-Key": GOTEST_API_KEY} if GOTEST_API_KEY else {}

    try:
        resp = httpx.post(
            f"{GOTEST_API_URL}/api/v1/runs",
            json={
                "project_path": state["project_path"],
                "requirements": state["requirements"],
            },
            headers=headers,
            timeout=300,
        )
        result = resp.json()
    except Exception as e:
        return {"run_result": {"error": str(e)}, "next_step": "failed"}

    failed = result.get("failed", 0)
    fix_attempts = state.get("fix_attempts", 0)
    max_fixes = state.get("max_fixes", 3)

    if failed == 0:
        next_step = "passed"
    elif fix_attempts >= max_fixes:
        next_step = "max_fixes"
    else:
        next_step = "failed"

    return {"run_result": result, "next_step": next_step, "fix_attempts": fix_attempts}
