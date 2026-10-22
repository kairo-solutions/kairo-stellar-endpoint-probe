import { parseProbeReport } from "@kairo-solutions/stellar-endpoint-probe-types";

const report = parseProbeReport("{}");
const overallStatus: boolean = report.ok;
const checkStatus: "ok" | "fail" | "skipped" | undefined = report.checks[0]?.status;
const ledgerSequence: number | undefined = report.checks[0]?.ledger_sequence;

void [overallStatus, checkStatus, ledgerSequence];
