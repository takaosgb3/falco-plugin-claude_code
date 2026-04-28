"""pytest configuration for the Claude Code Falco scenario Allure report.

Ported from /Users/takaos/lab/falco-plugin-openclaw/e2e/allure/conftest.py.
The CLI options stay identical so existing operator muscle memory carries over;
only the path defaults differ (test/fixtures/hook_events instead of e2e/...).

CLI options:
    --test-results  Path to test-results.json produced by the
                    `go test -tags=allure ./test/integration/...` step.
                    Required.
    --logs-dir      Optional path to a Falco logs directory. Currently unused
                    (the Allure wrapper consumes evidence directly from the
                    test-results.json `evidence` field), kept for parity with
                    the openclaw conftest.
"""

import pytest


def pytest_addoption(parser):
    """Add custom command-line options for E2E test report generation."""
    parser.addoption(
        "--test-results",
        action="store",
        required=True,
        help="Path to test-results.json from `go test -tags=allure ./test/integration/...`",
    )
    parser.addoption(
        "--logs-dir",
        action="store",
        default=None,
        help="Path to Falco logs directory (reserved for future evidence attachments)",
    )


def pytest_configure(config):
    """Stash option values and prewarm fixture loading caches for fair timing."""
    config.test_results = config.getoption("--test-results")
    config.logs_dir = config.getoption("--logs-dir")
    _prewarm_caches()


def _prewarm_caches():
    """Prewarm pattern loading cache so per-test timings exclude file I/O."""
    try:
        from test_e2e_wrapper import load_all_patterns
        load_all_patterns()
    except Exception:
        # Cache prewarming is best-effort; pytest will surface any real
        # import or runtime errors during the actual test phase.
        pass
