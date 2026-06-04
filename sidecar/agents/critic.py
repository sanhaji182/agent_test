from langchain_anthropic import ChatAnthropic
from langchain_core.messages import HumanMessage
import json


llm = ChatAnthropic(model="claude-sonnet-4-5")


def critic_node(state: dict) -> dict:
    files = json.dumps(state["test_files"])
    prompt = f"""You are a strict QA Test Critic (Enterprise Level). Review these Playwright test files for quality:
{files}

Critique the tests ruthlessly on:
1. Assertions (Are they robust or flaky?)
2. Selector quality (Are they using stable data-testids or fragile DOM paths?)
3. Edge cases (Did the planner miss negative scenarios?)

Return JSON: {{"score": 0-100, "suggestions": ["..."]}}"""

    resp = llm.invoke([HumanMessage(content=prompt)])
    content = resp.content
    if "```json" in content:
        content = content.split("```json")[1].split("```")[0]
    elif "```" in content:
        content = content.split("```")[1].split("```")[0]

    review = json.loads(content.strip())
    scores = state.get("critic_scores", [])
    scores.append(review.get("score", 50))
    next_step = "pass" if review.get("score", 0) >= 70 else "fail"
    return {"critic_scores": scores, "next_step": next_step}


def rewriter_node(state: dict) -> dict:
    files = json.dumps(state["test_files"])
    prompt = f"""Improve these test files based on critic feedback.
Current files: {files}
Rewrite with better assertions, selectors, and edge case coverage.
Return JSON array: [{{"name": "...", "content": "..."}}]"""

    resp = llm.invoke([HumanMessage(content=prompt)])
    content = resp.content
    if "```json" in content:
        content = content.split("```json")[1].split("```")[0]
    elif "```" in content:
        content = content.split("```")[1].split("```")[0]

    files = json.loads(content.strip())
    return {"test_files": files}
