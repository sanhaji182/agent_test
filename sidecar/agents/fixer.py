from langchain_anthropic import ChatAnthropic
from langchain_core.messages import HumanMessage
import json


llm = ChatAnthropic(model="claude-sonnet-4-5")


def fixer_node(state: dict) -> dict:
    failures = json.dumps(state.get("run_result", {}).get("failures", []))
    files = json.dumps(state["test_files"])
    prompt = f"""These tests failed:
{failures}

Original files: {files}

Fix the failing tests. Return JSON array: [{{"name": "...", "content": "..."}}]"""

    resp = llm.invoke([HumanMessage(content=prompt)])
    content = resp.content
    if "```json" in content:
        content = content.split("```json")[1].split("```")[0]
    elif "```" in content:
        content = content.split("```")[1].split("```")[0]

    fixed = json.loads(content.strip())
    return {"test_files": fixed, "fix_attempts": state.get("fix_attempts", 0) + 1}
