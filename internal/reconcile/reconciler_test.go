package reconcile

import (
	"testing"
	"time"

	"github.com/AAH20/kubernetes-ai-agent-operator/internal/domain"
)

func task() domain.Task {
	return domain.Task{Generation: 1, Spec: domain.Spec{TaskID: "azure-change-001", Objective: "prepare Azure change", AcceptanceCriteria: []string{"plan passes"}, Capabilities: map[string]string{"terraform.plan": "allow", "terraform.apply": "approval", "github.merge": "deny"}, DataClassification: "restricted", TenantTrust: "untrusted_multi_tenant", MutationImpact: "production", MaximumCostUSD: 5, EstimatedAgentUSD: 1.2, ContextRefs: []string{"git:repo@main"}, IdempotencyKey: "change-001-v1", PolicyVersion: "policy-v1"}, Status: domain.Status{Phase: domain.Pending}}
}

func reconcileUntil(t *testing.T, r *Reconciler, item *domain.Task, target domain.Phase, approvals []domain.Approval) {
	t.Helper()
	for i := 0; i < 12 && item.Status.Phase != target; i++ {
		if _, err := r.Reconcile(item, approvals); err != nil {
			t.Fatal(err)
		}
	}
}

func TestProductionTaskStopsForApproval(t *testing.T) {
	r, item := New(), task()
	reconcileUntil(t, r, &item, domain.AwaitingApproval, nil)
	if item.Status.Phase != domain.AwaitingApproval {
		t.Fatalf("phase=%s", item.Status.Phase)
	}
	if item.Status.RuntimeProfile != "dedicated-vm-sandbox" {
		t.Fatalf("runtime=%s", item.Status.RuntimeProfile)
	}
}

func TestExactApprovalResumesLifecycle(t *testing.T) {
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	r, item := New(), task()
	r.Now = func() time.Time { return now }
	reconcileUntil(t, r, &item, domain.AwaitingApproval, nil)
	approval := domain.Approval{TaskID: item.Spec.TaskID, Capability: "terraform.apply", ArtifactDigest: item.Status.ArtifactDigest, PolicyVersion: item.Spec.PolicyVersion, Approver: "platform-owner", ExpiresAt: now.Add(time.Hour)}
	reconcileUntil(t, r, &item, domain.Succeeded, []domain.Approval{approval})
	if item.Status.Phase != domain.Succeeded || len(item.Status.ReceiptSHA256) != 64 {
		t.Fatalf("status=%+v", item.Status)
	}
}

func TestChangedArtifactInvalidatesApproval(t *testing.T) {
	now := time.Now()
	r, item := New(), task()
	r.Now = func() time.Time { return now }
	reconcileUntil(t, r, &item, domain.AwaitingApproval, nil)
	approval := domain.Approval{TaskID: item.Spec.TaskID, Capability: "terraform.apply", ArtifactDigest: "sha256:old", PolicyVersion: item.Spec.PolicyVersion, Approver: "owner", ExpiresAt: now.Add(time.Hour)}
	result, _ := r.Reconcile(&item, []domain.Approval{approval})
	if result.Changed || item.Status.Phase != domain.AwaitingApproval {
		t.Fatal("stale approval must not resume")
	}
}

func TestBudgetFailsBeforeSandbox(t *testing.T) {
	r, item := New(), task()
	item.Spec.MaximumCostUSD = 0.1
	_, _ = r.Reconcile(&item, nil)
	if item.Status.Phase != domain.BudgetExceeded || item.Status.SandboxClaim != "" {
		t.Fatalf("status=%+v", item.Status)
	}
}

func TestUnknownAuthorityFailsClosed(t *testing.T) {
	r, item := New(), task()
	item.Spec.Capabilities["cloud.delete"] = "maybe"
	_, _ = r.Reconcile(&item, nil)
	if item.Status.Phase != domain.PolicyRejected {
		t.Fatalf("phase=%s", item.Status.Phase)
	}
}

func TestTerminalReconciliationIsIdempotent(t *testing.T) {
	r, item := New(), task()
	item.Status.Phase = domain.PolicyRejected
	result, _ := r.Reconcile(&item, nil)
	if result.Changed || result.Requeue {
		t.Fatal("terminal state must be stable")
	}
}
