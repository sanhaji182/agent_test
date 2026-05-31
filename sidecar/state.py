from typing import Optional
from typing_extensions import TypedDict


class AgentState(TypedDict):
    run_id: str
    project_path: str
    requirements: str
    code_analysis: str
    test_plan: dict
    test_files: list[dict]
    run_result: dict
    fix_attempts: int
    max_fixes: int
    critic_scores: list[float]
    screenshots: list[str]
    next_step: str
    error: Optional[str]
