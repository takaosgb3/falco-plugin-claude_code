"""test_e2e_wrapper.py — Allure report wrapper for Claude Code Falco E2E.

Reads test-results.json produced by:

    go test -tags=allure -count=1 ./test/integration/... -run TestL3_Falco_*

and generates one pytest test case per fixture, decorated with Allure
metadata: Epic / Feature / Story hierarchy, Severity (mapped by category),
Markdown description (pattern info table + payload + rule mapping + Falco
alert evidence), and 4 attachment steps.

Ported from /Users/takaos/lab/falco-plugin-openclaw/e2e/allure/
test_e2e_wrapper.py with three adaptations:

  1. PATTERNS_DIR points at test/fixtures/hook_events/ and walks the tree
     (fixtures live in event-name subdirs: PreToolUse/, ConfigChange/, ...).
  2. Pattern info is read from the fixture's `_meta` block plus the top-level
     hook payload fields (`command`, `tool_name`, `file_path`, `permission_
     mode`, ...) so Markdown descriptions reproduce the actual hook event.
  3. SEVERITY_MAP follows requirements v3 §12.1 priority levels (T-001..
     T-018 + T-013-low/high + benign + heartbeat).

Usage:

    cd test/allure && python3 -m pytest test_e2e_wrapper.py \
        --test-results=../../test/integration/results/test-results.json \
        --alluredir=../../allure-results-falco -v
"""

import json
import re
from pathlib import Path
from typing import Any, Optional

import allure
import pytest

# ---------------------------------------------------------------------------
# Paths
# ---------------------------------------------------------------------------

# Project root: test/allure -> test -> repo root.
PROJECT_ROOT = Path(__file__).parent.parent.parent

# Fixture root (recursive). Each subdir is a hook event name plus benign /
# _heartbeat_; each fixture is a JSON file with a `_meta` block.
PATTERNS_DIR = PROJECT_ROOT / "test" / "fixtures" / "hook_events"

# ---------------------------------------------------------------------------
# Severity mapping (requirements v3 §12.1)
# ---------------------------------------------------------------------------

# Categories assigned to Allure severity levels. Keep in lockstep with the
# Falco rule priority declared in `rules/claude-code_rules.yaml`.
SEVERITY_MAP = {
    # CRITICAL (priority 0..1)
    "T-001": allure.severity_level.CRITICAL,
    "T-002": allure.severity_level.CRITICAL,
    "T-003": allure.severity_level.CRITICAL,
    "T-006": allure.severity_level.CRITICAL,
    "T-016": allure.severity_level.CRITICAL,
    # WARNING (priority 4)
    "T-004": allure.severity_level.NORMAL,
    "T-005": allure.severity_level.NORMAL,
    "T-007": allure.severity_level.NORMAL,
    "T-008": allure.severity_level.NORMAL,
    "T-009": allure.severity_level.NORMAL,
    "T-010": allure.severity_level.NORMAL,
    "T-011": allure.severity_level.NORMAL,
    "T-012": allure.severity_level.NORMAL,
    "T-014": allure.severity_level.NORMAL,
    "T-015": allure.severity_level.NORMAL,
    "T-017": allure.severity_level.NORMAL,
    "T-018": allure.severity_level.NORMAL,
    # T-013 split: low NOTICE, high WARNING.
    "T-013": allure.severity_level.NORMAL,
    "T-013-low": allure.severity_level.MINOR,
    "T-013-high": allure.severity_level.NORMAL,
    # Operations
    "benign": allure.severity_level.TRIVIAL,
    "heartbeat": allure.severity_level.MINOR,
}

# ---------------------------------------------------------------------------
# Security keywords (claude_code-specific)
# ---------------------------------------------------------------------------

SECURITY_KEYWORDS = [
    # T-001 dangerous bash
    "rm -rf", "rm -f", "/etc/passwd", "/etc/shadow", "chmod -R 777",
    "mkfs.", "dd if=", "shutdown -h", "/dev/sda",
    "curl pipe sh", "| sh", "| bash", "$(whoami)", "`whoami`",
    # T-002 secret exfiltration
    "AKIA", "ghp_", "github_pat_", "xoxb-", "sk-ant-", "sk-",
    "AWS_SECRET_ACCESS_KEY", "id_rsa", "credentials", ".env",
    "BEGIN PRIVATE KEY", "BEGIN RSA PRIVATE KEY",
    # T-003 permission bypass
    "bypassPermissions", "--dangerously-skip-permissions",
    # T-006 hook disabled
    "disableAllHooks",
    # T-009 sensitive file
    ".aws/credentials", ".kube/config", "/root/",
    # T-010 workspace escape
    "../../", "../../../",
    # T-011 git destructive
    "git push -f", "git reset --hard", "git clean -fdx", "rm -rf .git",
    # T-012 prompt injection
    "ignore previous instructions", "ignore all previous",
    # claude_code generic
    "PreToolUse", "PermissionRequest", "ConfigChange", "claude_code",
]

# ---------------------------------------------------------------------------
# Pattern loader
# ---------------------------------------------------------------------------

# Lazy cache; populated by load_all_patterns() and reused across tests.
_patterns_cache: Optional[dict] = None


def load_all_patterns() -> dict:
    """Load all fixture JSON files and return a map of pattern_id ->
    pattern_info.

    pattern_info is a dict with keys:
        category, hook_event_name, tool_name, command, file_path, url,
        permission_mode, mcp_server_name, mcp_tool_name, risk_type,
        risk_score, severity, evidence, expected_detection,
        expected_event_name, expected_risk_type, expected_severity, notes.

    The `_meta` block is the canonical source for fixture_id, category,
    expected_detection, and notes. The top-level fields supply the actual
    hook payload that the Falco plugin parses.
    """
    global _patterns_cache
    if _patterns_cache is not None:
        return _patterns_cache

    pattern_map: dict = {}
    if not PATTERNS_DIR.is_dir():
        _patterns_cache = pattern_map
        return pattern_map

    for json_file in sorted(PATTERNS_DIR.rglob("*.json")):
        try:
            with open(json_file, "r") as f:
                data = json.load(f)
        except (json.JSONDecodeError, OSError):
            continue
        meta = data.get("_meta") or {}
        pid = meta.get("fixture_id") or json_file.stem
        category = meta.get("category", "unknown")

        # Top-level hook fields (skip _meta itself).
        info = {k: v for k, v in data.items() if k != "_meta"}
        info.update({
            "category": category,
            "expected_detection": meta.get("expected_detection", ""),
            "expected_event_name": meta.get("expected_event_name", ""),
            "expected_risk_type": meta.get("expected_risk_type", ""),
            "expected_severity": meta.get("expected_severity", ""),
            "notes": meta.get("notes", ""),
            "_fixture_path": str(json_file.relative_to(PROJECT_ROOT)),
        })
        pattern_map[pid] = info

    _patterns_cache = pattern_map
    return pattern_map


def load_test_results(path: str) -> list:
    """Load test-results.json and return list of result dicts."""
    results_path = Path(path)
    if not results_path.is_file():
        pytest.skip(f"test-results.json not found: {path}")
        return []
    with open(results_path, "r") as f:
        data = json.load(f)
    if not isinstance(data, list):
        pytest.skip(f"test-results.json is not a JSON array: {path}")
        return []
    return data


# ---------------------------------------------------------------------------
# Evidence highlighting
# ---------------------------------------------------------------------------

def highlight_keywords_in_text(text: str, keywords: list, fmt: str = "html") -> str:
    """Highlight security keywords in evidence text.

    Applies keyword highlighting to the plain text content first, then wraps
    in HTML tags. This prevents keyword replacements from corrupting HTML
    markup. Matched text preserves its original casing.
    """
    if fmt == "html":
        escaped_text = _html_escape(text)
        for kw in keywords:
            escaped_kw = _html_escape(kw)
            if escaped_kw.lower() in escaped_text.lower():
                escaped_text = _case_insensitive_replace(
                    escaped_text, escaped_kw,
                    lambda m: f"<mark style='background: #ffeb3b; font-weight: bold;'>{m}</mark>"
                )
        return f"<pre style='font-family: monospace; white-space: pre-wrap;'>{escaped_text}</pre>"
    result = text
    for kw in keywords:
        if kw.lower() in result.lower():
            result = _case_insensitive_replace(result, kw, lambda m: f"**{m}**")
    return result


def _html_escape(text: str) -> str:
    return (text
            .replace("&", "&amp;")
            .replace("<", "&lt;")
            .replace(">", "&gt;")
            .replace('"', "&quot;")
            .replace("'", "&#39;"))


def _case_insensitive_replace(text: str, old: str, new) -> str:
    if callable(new):
        return re.sub(
            re.escape(old),
            lambda m: new(m.group(0)),
            text,
            flags=re.IGNORECASE,
        )
    return re.sub(re.escape(old), new, text, flags=re.IGNORECASE)


# ---------------------------------------------------------------------------
# Result formatting helpers
# ---------------------------------------------------------------------------

def format_rule_match_status(result: dict) -> str:
    """Format rule match status as emoji string for human-readable display."""
    category = result.get("category", "")
    expected = result.get("expected_rule", "")
    rule_match = result.get("rule_match") is True
    detected = result.get("detected") is True

    if category == "benign" and not expected:
        return ("[OK] Expected Not Detected" if not detected
                else "[FAIL] Unexpected Detection")
    if not expected:
        return "[NA] Not Defined"
    if rule_match:
        return "[OK] Match"
    if detected:
        return "[INFO] Preempted (higher-priority rule fired; AT-2 satisfied)"
    return "[FAIL] Not Detected"


def map_severity(result: dict) -> Any:
    """Map test result category to Allure severity level."""
    category = result.get("category", "")
    return SEVERITY_MAP.get(category, allure.severity_level.NORMAL)


def _format_latency(result: dict) -> str:
    latency = result.get("latency_ms", -1)
    if latency is None or latency < 0:
        return "N/A"
    return f"{latency}ms"


def _payload_summary(pattern_info: dict) -> str:
    """Build a human-readable payload string from the fixture's hook fields."""
    parts = []
    for k in ("hook_event_name", "tool_name", "command", "file_path",
              "url", "permission_mode", "mcp_server_name", "mcp_tool_name"):
        v = pattern_info.get(k)
        if v not in (None, "", 0):
            parts.append(f"{k}={v}")
    return " ".join(parts) if parts else "N/A"


def _build_description(result: dict, pattern_info: dict) -> str:
    """Build rich Markdown description for Allure report detail view."""
    pattern_id = result["pattern_id"]
    category = result.get("category", "unknown")
    status = result.get("status", "unknown")
    expected_rule = result.get("expected_rule", "")
    matched_rule = result.get("matched_rule", "")
    matched_rules = result.get("matched_rules") or []
    evidence = result.get("evidence") or "No evidence recorded"

    notes = pattern_info.get("notes") or "N/A"
    expected_detection = pattern_info.get("expected_detection") or "N/A"
    expected_event = pattern_info.get("expected_event_name") or "N/A"
    fixture_path = pattern_info.get("_fixture_path") or "N/A"
    payload_display = _payload_summary(pattern_info)

    if matched_rules:
        matched_display = ", ".join(f"`{r}`" for r in matched_rules)
    else:
        matched_display = f"`{matched_rule or 'N/A'}`"

    detected = result.get("detected", False)
    detection_count = "1 / 1" if detected else "0 / 1"
    match_status = format_rule_match_status(result)

    return (
        "## Attack Pattern Information\n"
        "\n"
        "| Item | Value |\n"
        "|------|-------|\n"
        f"| **Pattern ID** | `{pattern_id}` |\n"
        f"| **Description** | {notes} |\n"
        f"| **Category** | `{category}` |\n"
        f"| **Severity (Allure)** | `{map_severity(result).value.upper()}` |\n"
        f"| **Fixture** | `{fixture_path}` |\n"
        f"| **Expected Hook Event** | `{expected_event}` |\n"
        "\n"
        "## Attack Details\n"
        "\n"
        f"- **Payload**: `{payload_display}`\n"
        f"- **Expected Detection**: `{expected_detection}`\n"
        "\n"
        "## Test Execution Results\n"
        "\n"
        f"- **Status**: `{status.upper()}`\n"
        f"- **Detection Count**: {detection_count}\n"
        f"- **Latency**: {_format_latency(result)}\n"
        "\n"
        "## Rule Mapping\n"
        "\n"
        "| Item | Value |\n"
        "|------|-------|\n"
        f"| **Expected Rule** | `{expected_rule or 'N/A'}` |\n"
        f"| **Matched Rule** | {matched_display} |\n"
        f"| **Rule Match** | {match_status} |\n"
        "\n"
        "## Detection Evidence\n"
        "\n"
        "```\n"
        f"{evidence}\n"
        "```\n"
    )


# ---------------------------------------------------------------------------
# pytest hooks and test function
# ---------------------------------------------------------------------------

def pytest_generate_tests(metafunc):
    """Dynamically parametrize tests from test-results.json."""
    if "result" in metafunc.fixturenames:
        test_results_path = metafunc.config.getoption("--test-results")
        results = load_test_results(test_results_path)
        if results:
            metafunc.parametrize(
                "result", results,
                ids=[f"{i:02d}_{r['pattern_id']}" for i, r in enumerate(results, 1)],
            )


def _epic_for(category: str) -> str:
    """Return the Allure Epic label for a category. Single-Epic strategy."""
    return "Claude Code E2E Security Tests"


def _feature_for(category: str) -> str:
    """Return the Allure Feature label. Per spec: T-001..T-018 + benign +
    heartbeat. T-013-low / T-013-high are folded under T-013 so the Behaviors
    tree shows one feature per requirement-level threat ID."""
    if category.startswith("T-013"):
        return "T-013 Agent / Subagent Risk"
    feature_titles = {
        "T-001": "T-001 Dangerous Bash Command",
        "T-002": "T-002 Secret Exfiltration Attempt",
        "T-003": "T-003 Permission Bypass Mode",
        "T-004": "T-004 Suspicious Permission Update",
        "T-005": "T-005 Claude Settings Modified",
        "T-006": "T-006 Hook Disabled Or Modified",
        "T-007": "T-007 MCP Config Changed",
        "T-008": "T-008 Suspicious MCP Tool Use",
        "T-009": "T-009 Sensitive File Read",
        "T-010": "T-010 Workspace Escape",
        "T-011": "T-011 Destructive Git Operation",
        "T-012": "T-012 Prompt Injection Pattern",
        "T-014": "T-014 Agent Runaway Tool Storm",
        "T-015": "T-015 External Fetch With Sensitive Context",
        "T-016": "T-016 Config Policy Downgrade",
        "T-017": "T-017 Skill Or Command Shell Risk",
        "T-018": "T-018 Channel Or MCP Push Risk",
        "benign": "Benign (no false positives)",
        "heartbeat": "Plugin Heartbeat (health)",
    }
    return feature_titles.get(category, category)


def test_e2e_detection(result):
    """E2E pattern detection test — one test case per fixture.

    Reads a single result entry from test-results.json (parametrized via
    pytest_generate_tests above), decorates it with Allure metadata, and
    asserts the recorded `status` is "passed". The 4 attachment steps mirror
    the openclaw report: result JSON, evidence with HTML highlighting,
    rule-mapping summary, and a final verification block.
    """
    pattern_id = result["pattern_id"]
    category = result.get("category", "unknown")

    allure.dynamic.epic(_epic_for(category))
    allure.dynamic.feature(_feature_for(category))
    allure.dynamic.story(pattern_id)
    allure.dynamic.severity(map_severity(result))
    allure.dynamic.label("category", category)
    allure.dynamic.label("layer", "Level 3 — Falco-in-the-loop")

    patterns = load_all_patterns()
    pattern_info = patterns.get(pattern_id, {})
    description = _build_description(result, pattern_info)
    allure.dynamic.description(description)

    # Step 1: Test execution result (JSON attachment, openclaw-compatible).
    with allure.step("Test Execution Result"):
        allure.attach(
            json.dumps(result, indent=2),
            name=f"{pattern_id}-result.json",
            attachment_type=allure.attachment_type.JSON,
        )

    # Step 2: Detection evidence with keyword highlighting.
    evidence = result.get("evidence") or ""
    if evidence:
        highlighted = highlight_keywords_in_text(
            evidence, SECURITY_KEYWORDS, "html",
        )
        html_doc = (
            "<!DOCTYPE html><html><head><meta charset='UTF-8'>"
            "<style>"
            "body{font-family:monospace;padding:15px;background:#1a1a1a;"
            "color:#e0e0e0;line-height:1.5}"
            "pre{white-space:pre-wrap;word-wrap:break-word}"
            "mark{background:#FFFF00;color:#000;padding:1px 3px;border-radius:2px}"
            "h3{color:#4CAF50;border-bottom:1px solid #333;padding-bottom:6px}"
            "</style></head><body>"
            f"<h3>Detection Evidence — {pattern_id}</h3>"
            f"{highlighted}"
            "</body></html>"
        )
        with allure.step("Detection Evidence (Highlighted)"):
            allure.attach(
                html_doc,
                name="Detection Evidence (HTML)",
                attachment_type=allure.attachment_type.HTML,
            )

    # Step 3: Rule mapping verification.
    with allure.step("Rule Mapping"):
        expected_rule = result.get("expected_rule", "")
        matched_rule = result.get("matched_rule", "")
        match_status = format_rule_match_status(result)
        mapping_text = (
            f"Expected Rule: {expected_rule or 'N/A'}\n"
            f"Matched Rule:  {matched_rule or 'N/A'}\n"
            f"Rule Match:    {match_status}"
        )
        allure.attach(
            mapping_text,
            name="Rule Mapping",
            attachment_type=allure.attachment_type.TEXT,
        )

    # Step 4: Verification result.
    with allure.step("Verification"):
        if result["status"] == "passed":
            allure.attach(
                f"Test passed: {pattern_id}\n"
                f"Category: {category}\n"
                f"Detected: {result.get('detected')}\n"
                f"Rule Match: {result.get('rule_match')}\n",
                name="Verification",
                attachment_type=allure.attachment_type.TEXT,
            )
        else:
            allure.attach(
                f"Test failed: {pattern_id}\n"
                f"Detected: {result.get('detected')}\n"
                f"Expected: {result.get('expected_rule')}\n"
                f"Matched:  {result.get('matched_rule')}\n",
                name="Verification",
                attachment_type=allure.attachment_type.TEXT,
            )

    assert result["status"] == "passed", (
        f"Pattern {pattern_id} ({category}): "
        f"status={result['status']}, "
        f"detected={result.get('detected')}, "
        f"expected_rule={result.get('expected_rule', '')}, "
        f"matched_rule={result.get('matched_rule', '')}"
    )
