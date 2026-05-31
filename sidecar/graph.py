from langgraph.graph import StateGraph, END
from state import AgentState
from agents.planner import planner_node
from agents.writer import writer_node
from agents.critic import critic_node, rewriter_node
from agents.executor import executor_node
from agents.fixer import fixer_node


def route_critic(state: dict) -> str:
    return state.get("next_step", "pass")


def route_executor(state: dict) -> str:
    return state.get("next_step", "passed")


def build_graph():
    graph = StateGraph(AgentState)
    graph.add_node("planner", planner_node)
    graph.add_node("writer", writer_node)
    graph.add_node("critic", critic_node)
    graph.add_node("rewriter", rewriter_node)
    graph.add_node("executor", executor_node)
    graph.add_node("fixer", fixer_node)

    graph.set_entry_point("planner")
    graph.add_edge("planner", "writer")
    graph.add_edge("writer", "critic")
    graph.add_conditional_edges("critic", route_critic, {
        "pass": "executor",
        "fail": "rewriter",
    })
    graph.add_edge("rewriter", "critic")
    graph.add_conditional_edges("executor", route_executor, {
        "passed": END,
        "failed": "fixer",
        "max_fixes": END,
    })
    graph.add_edge("fixer", "executor")

    return graph.compile()
