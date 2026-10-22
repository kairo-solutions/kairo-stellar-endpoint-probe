export type CheckStatus = "ok" | "fail" | "skipped";

export interface ProbeCheck {
  name: string;
  status: CheckStatus;
  endpoint?: string;
  latency_ms?: number;
  ledger_sequence?: number;
  message?: string;
}

export interface ProbeReport {
  network: string;
  ok: boolean;
  checks: ProbeCheck[];
}

export function parseProbeReport(json: string): ProbeReport {
  let value: unknown;
  try {
    value = JSON.parse(json);
  } catch (error) {
    throw new TypeError("report must be valid JSON", { cause: error });
  }
  if (!isRecord(value)) {
    throw new TypeError("report must be a JSON object");
  }
  if (typeof value.network !== "string" || value.network.length === 0) {
    throw new TypeError("report.network must be a non-empty string");
  }
  if (typeof value.ok !== "boolean") {
    throw new TypeError("report.ok must be a boolean");
  }
  if (!Array.isArray(value.checks)) {
    throw new TypeError("report.checks must be an array");
  }

  const checks = value.checks.map(parseCheck);
  const hasFailure = checks.some((check) => check.status === "fail");
  if (value.ok === hasFailure) {
    throw new TypeError("report.ok must be false exactly when a check failed");
  }
  return { network: value.network, ok: value.ok, checks };
}

function parseCheck(value: unknown, index: number): ProbeCheck {
  if (!isRecord(value)
    || typeof value.name !== "string"
    || value.name.length === 0
    || !isCheckStatus(value.status)) {
    throw new TypeError(`report.checks[${index}] has an invalid name or status`);
  }
  if (value.endpoint !== undefined && typeof value.endpoint !== "string") {
    throw new TypeError(`report.checks[${index}].endpoint must be a string`);
  }
  if (value.message !== undefined && typeof value.message !== "string") {
    throw new TypeError(`report.checks[${index}].message must be a string`);
  }
  for (const field of ["latency_ms", "ledger_sequence"] as const) {
    if (value[field] !== undefined
      && (!Number.isSafeInteger(value[field]) || Number(value[field]) < 0)) {
      throw new TypeError(`report.checks[${index}].${field} must be a non-negative integer`);
    }
  }
  return {
    name: value.name,
    status: value.status,
    ...(value.endpoint === undefined ? {} : { endpoint: value.endpoint as string }),
    ...(value.message === undefined ? {} : { message: value.message as string }),
    ...(value.latency_ms === undefined ? {} : { latency_ms: value.latency_ms as number }),
    ...(value.ledger_sequence === undefined ? {} : { ledger_sequence: value.ledger_sequence as number }),
  };
}

function isCheckStatus(value: unknown): value is CheckStatus {
  return value === "ok" || value === "fail" || value === "skipped";
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
