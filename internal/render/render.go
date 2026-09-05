package render

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AAH20/kubernetes-ai-agent-operator/internal/domain"
)

func Bundle(task domain.Task, output string) error {
	if err := os.MkdirAll(output, 0o755); err != nil {
		return err
	}
	status, _ := json.MarshalIndent(task, "", "  ")
	if err := os.WriteFile(filepath.Join(output, "task-status.json"), append(status, '\n'), 0o644); err != nil {
		return err
	}
	claim := fmt.Sprintf("apiVersion: extensions.agents.x-k8s.io/v1beta1\nkind: SandboxClaim\nmetadata:\n  name: %s\n  labels:\n    agentmesh.io/task-id: %s\nspec:\n  sandboxTemplateRef:\n    name: %s\n", task.Status.SandboxClaim, task.Spec.TaskID, task.Status.RuntimeProfile)
	if err := os.WriteFile(filepath.Join(output, "sandboxclaim.yaml"), []byte(claim), 0o644); err != nil {
		return err
	}
	evidence := map[string]any{"taskId": task.Spec.TaskID, "phase": task.Status.Phase, "artifactDigest": task.Status.ArtifactDigest, "estimatedCostUSD": task.Status.EstimatedCostUSD, "policyVersion": task.Spec.PolicyVersion, "receiptSHA256": task.Status.ReceiptSHA256, "provenance": "synthetic-local-reconciliation"}
	evidenceJSON, _ := json.MarshalIndent(evidence, "", "  ")
	if err := os.WriteFile(filepath.Join(output, "evidence-receipt.json"), append(evidenceJSON, '\n'), 0o644); err != nil {
		return err
	}
	report := fmt.Sprintf("# AgentMesh Operator lifecycle\n\n| Field | Value |\n|---|---|\n| Task | `%s` |\n| Phase | **%s** |\n| Runtime | %s |\n| Estimated cost | $%.4f |\n| Artifact | `%s` |\n\n## Required approvals\n\n%s\n\n## Evidence boundary\n\nThis output is a local deterministic reconciliation. It does not prove cluster provisioning or external delivery.\n", task.Spec.TaskID, task.Status.Phase, task.Status.RuntimeProfile, task.Status.EstimatedCostUSD, task.Status.ArtifactDigest, strings.Join(task.Status.RequiredApprovals, ", "))
	return os.WriteFile(filepath.Join(output, "executive-report.md"), []byte(report), 0o644)
}
