# Changelog

All notable changes to the claude-code Falco plugin will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Phase 1 scaffold: plugin skeleton (`cmd/plugin-sdk/plugin.go`), parser
  skeleton (`pkg/parser/`), `claude-code-security-logger` skeleton
  (`cmd/claude-code-security-logger/`), Falco configs (local/Linux/docker),
  T-001 reference rule, README, Makefile (P002 `-buildmode=c-shared` set),
  and CI/E2E/Release workflows. T-002..T-018 rules, full parser logic, and
  E2E tests are added in Phase 2/3/4.

## [0.1.0] - YYYY-MM-DD (planned)

Initial release per requirements v3 §24 (v0.1 Minimum Viable Scope).
