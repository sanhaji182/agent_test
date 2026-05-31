from langchain_anthropic import ChatAnthropic
from langchain_core.messages import HumanMessage
import json


llm = ChatAnthropic(model="claude-sonnet-4-5")


def writer_node(state: dict) -> dict:
    plan = json.dumps(state["test_plan"])
    prompt = f"""Write Playwright TypeScript test files for this test plan:
{plan}

Code context: {state['code_analysis']}

Return JSON array: [{{"name": "test-name.spec.ts", "content": "..."}}]"""

    resp = llm.invoke([HumanMessage(content=prompt)])
    content = resp.content
    if "```json" in content:
        content = content.split("```json")[1].split("```")[0]
    elif "```" in content:
        content = content.split("```")[1].split("```")[0]

    files = json.loads(content.strip())
    return {"test_files": files}
