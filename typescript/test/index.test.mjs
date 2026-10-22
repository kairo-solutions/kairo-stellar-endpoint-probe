import assert from "node:assert/strict";
import { test } from "node:test";
import { parseProbeReport } from "../.build/index.js";

const healthyReport = {
  network: "testnet",
  ok: true,
  checks: [
    { name: "Horizon root", status: "ok", latency_ms: 23 },
    { name: "Soroban RPC latest ledger", status: "ok", ledger_sequence: 1234 },
    { name: "Network consistency", status: "skipped", message: "No expected passphrase" },
  ],
};

test("parses Go probe JSON for TypeScript consumers", () => {
  assert.deepEqual(parseProbeReport(JSON.stringify(healthyReport)), healthyReport);
});

test("accepts failed reports only when a check records the failure", () => {
  const report = {
    network: "custom",
    ok: false,
    checks: [{ name: "Horizon root", status: "fail", message: "timeout" }],
  };
  assert.deepEqual(parseProbeReport(JSON.stringify(report)), report);
  assert.throws(
    () => parseProbeReport(JSON.stringify({ ...report, ok: true })),
    /exactly when a check failed/,
  );
});

test("rejects malformed JSON and invalid check fields", () => {
  assert.throws(() => parseProbeReport("{"), /valid JSON/);
  assert.throws(
    () => parseProbeReport(JSON.stringify({
      ...healthyReport,
      checks: [{ name: "Horizon root", status: "unknown" }],
    })),
    /invalid name or status/,
  );
});

test("rejects inconsistent report status and unsafe numeric check values", () => {
  assert.throws(
    () => parseProbeReport(JSON.stringify({ ...healthyReport, ok: false })),
    /exactly when a check failed/,
  );
  assert.throws(
    () => parseProbeReport(JSON.stringify({
      ...healthyReport,
      checks: [{ name: "Horizon root", status: "ok", latency_ms: -1 }],
    })),
    /non-negative integer/,
  );
});
