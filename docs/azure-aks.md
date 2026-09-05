# Azure AKS production path

Target components:

- private AKS cluster with Azure CNI and network policy;
- dedicated Kata-capable node pool where supported and validated;
- Azure Workload Identity for the controller and each executor;
- Key Vault CSI only for short-lived, task-scoped material;
- ACR images pinned by digest with SBOM and signature admission;
- Azure Monitor managed Prometheus and OpenTelemetry Collector;
- Private Link for registries, Key Vault, storage and control services;
- GitHub Actions OIDC for build and deployment;
- Azure Policy assignments and Defender for Containers as independently priced options.

Do not deploy the placeholder image digest in the examples. Live validation requires an authorized subscription, regional capability checks, a cost budget and teardown evidence.
