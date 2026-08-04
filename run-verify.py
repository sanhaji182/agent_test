import subprocess, json, urllib.request, http.cookiejar, time

# Bypass macOS system proxy for localhost API calls (urllib picks up the
# system proxy from System Configuration, which breaks calls to localhost).
proxy_handler = urllib.request.ProxyHandler({})

api_key = subprocess.run(
    ["docker", "compose", "exec", "-T", "backend", "printenv", "API_KEY"],
    capture_output=True, text=True
).stdout.strip()

BASE = "http://localhost:8081"
cj = http.cookiejar.CookieJar()
opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(cj), proxy_handler)

def req(method, path, payload=None):
    data = json.dumps(payload).encode() if payload is not None else None
    r = urllib.request.Request(BASE + path, data=data, method=method,
                               headers={"Content-Type": "application/json"})
    return json.loads(opener.open(r, timeout=120).read().decode())

req("POST", "/api/v1/auth/login", {"api_key": api_key})

t0 = time.time()
created = req("POST", "/api/v1/runs", {
    "project_path": "/app",
    "requirements": "Test https://asisten.digital/ : verify the homepage loads successfully and the main heading is visible.",
    "mode": "simple",
    "test_type": "ui",
})
run_id = created.get("run_id")
print(f"created: {created}", flush=True)
if not run_id:
    print("no run_id")
    exit(0)

print("run_id:", run_id, flush=True)

state = None
while time.time() - t0 < 900:  # 15 min max
    run = req("GET", "/api/v1/runs/" + run_id)
    new_state = run.get("state")
    if new_state != state:
        print(f"[{int(time.time()-t0)}s] state: {new_state}", flush=True)
        state = new_state
    if state in ("done", "failed", "cancelled", "simulated"):
        break
    time.sleep(5)

print("\n=== FINAL ===")
print("state:", run.get("state"))
print("elapsed:", round(time.time()-t0), "s")
print("error:", (run.get("error") or "")[:300])
rr = run.get("run_result") or {}
if rr:
    print(f"result: passed={rr.get('passed')} failed={rr.get('failed')} total={rr.get('total')}")
tp = run.get("test_plan") or {}
if tp:
    print("test plan summary:", tp.get("summary", "")[:200])
    for sc in (tp.get("scenarios") or [])[:4]:
        print(f"  - [{sc.get('priority')}] {sc.get('name')} ({len(sc.get('steps') or [])} steps)")
print("test files:", len(run.get("test_files") or []))
