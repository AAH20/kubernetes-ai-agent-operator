# Architecture and invariants

The task controller never executes generated code. It admits intent, creates execution resources, validates results and records status. Kubernetes Agent Sandbox owns sandbox reconciliation.

Production invariants:

- no provider or cloud credential in the operator image;
- workload identity scoped to tenant, task and expiry;
- default-deny sandbox networking;
- untrusted tenants use a kernel-isolated runtime or dedicated node/VM;
- external writes use idempotency keys;
- approval binds task, capability, artifact digest and policy version;
- finalizers retain evidence before sandbox deletion;
- status updates use observed generation and optimistic concurrency;
- controller restart does not duplicate delivery.

The v0.1 reconciliation kernel implements state progression, budget rejection and approval binding. Kubernetes API watches, finalizers and optimistic-concurrency retries are roadmap work, not current claims.
