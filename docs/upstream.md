# Upstream contribution strategy

The most credible distribution route is a small, generic contribution to Kubernetes Agent Sandbox after the local operator is exercised.

Candidate contributions:

1. An approval-aware infrastructure-agent example using `SandboxClaim`.
2. OpenTelemetry correlation across claim, sandbox and agent task.
3. Azure Workload Identity and Kata-on-AKS documentation with verified constraints.
4. AgentInfraBench lifecycle and isolation scenarios.

Upstream work must use upstream terminology, tests and contribution rules. It should not advertise A2Z SOC or force AgentMesh-specific CRDs into a general project.
