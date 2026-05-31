from langchain_anthropic import ChatAnthropic
from langchain_core.messages import HumanMessage
import json


llm = ChatAnthropic(model="claude-sonnet-4-5")


def planner_node(state: dict) -> dict:
    prompt = f"""Analyze this codebase and create a test plan.

Code analysis: {state['code_analysis']}
Requirements: {state['requirements']}

Return JSON: {{"summary": "...", "scenarios": [{{"name": "...", "priority": "high|medium|low", "steps": ["..."]}}]}}"""

    resp = llm.invoke([HumanMessage(content=prompt)])
    content = resp.content
    if "```json" in content:
        content = content.split("```json")[1].split("```")[0]
    elif "```" in content:
        content = content.split("```")[1].split("```")[0]

    plan = json.loads(content.strip())
    return {"test_plan": plan}
