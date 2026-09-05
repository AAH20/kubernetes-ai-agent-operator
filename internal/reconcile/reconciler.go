package reconcile

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AAH20/kubernetes-ai-agent-operator/internal/domain"
)

type Reconciler struct {
	Now func() time.Time
}

func New() *Reconciler { return &Reconciler{Now: time.Now} }

func (r *Reconciler) Reconcile(task *domain.Task, approvals []domain.Approval) (domain.Result, error) {
	if err := validate(task.Spec); err != nil {
		return r.transition(task, domain.PolicyRejected, err.Error()), nil
	}
	switch task.Status.Phase {
	case "", domain.Pending:
		return r.admit(task), nil
	case domain.Admitted:
		return r.transition(task, domain.SandboxProvisioning, "runtime profile selected"), nil
	case domain.SandboxProvisioning:
		task.Status.SandboxClaim = "claim-" + task.Spec.TaskID
		return r.transition(task, domain.Running, "sandbox claim intent emitted"), nil
	case domain.Running:
		task.Status.ArtifactDigest = artifactDigest(task.Spec)
		return r.transition(task, domain.Evaluating, "execution artifact recorded"), nil
	case domain.Evaluating:
		if len(task.Status.RequiredApprovals) > 0 {
			return r.transition(task, domain.AwaitingApproval, "exact approval required"), nil
		}
		return r.transition(task, domain.Delivering, "evaluation gates passed"), nil
	case domain.AwaitingApproval:
		if !r.allApproved(task, approvals) {
			return domain.Result{Changed: false, Requeue: true, RequeueAfter: time.Minute}, nil
		}
		return r.transition(task, domain.Delivering, "artifact-bound approval verified"), nil
	case domain.Delivering:
		result := r.transition(task, domain.Succeeded, "artifact delivered idempotently")
		task.Status.ReceiptSHA256 = receipt(task)
		return result, nil
	default:
		return domain.Result{Changed: false, Requeue: false}, nil
	}
}

func (r *Reconciler) admit(task *domain.Task) domain.Result {
	spec := task.Spec
	profile, sandboxUSD := "rootless-developer", 0.10
	if spec.DataClassification == "confidential" || spec.TenantTrust == "untrusted_multi_tenant" {
		profile, sandboxUSD = "kata-microvm", 0.40
	}
	if spec.DataClassification == "restricted" || spec.MutationImpact == "production" {
		profile, sandboxUSD = "dedicated-vm-sandbox", 0.75
	}
	task.Status.EstimatedCostUSD = round(spec.EstimatedAgentUSD+sandboxUSD, 4)
	if task.Status.EstimatedCostUSD > spec.MaximumCostUSD {
		return r.transition(task, domain.BudgetExceeded, "estimated cost exceeds task budget")
	}
	for capability, authority := range spec.Capabilities {
		switch authority {
		case "approval":
			task.Status.RequiredApprovals = append(task.Status.RequiredApprovals, capability)
		case "deny":
			task.Status.DeniedCapabilities = append(task.Status.DeniedCapabilities, capability)
		}
	}
	sort.Strings(task.Status.RequiredApprovals)
	sort.Strings(task.Status.DeniedCapabilities)
	task.Status.RuntimeProfile = profile
	return r.transition(task, domain.Admitted, "policy and budget admission passed")
}

func (r *Reconciler) allApproved(task *domain.Task, approvals []domain.Approval) bool {
	for _, capability := range task.Status.RequiredApprovals {
		matched := false
		for _, approval := range approvals {
			if approval.TaskID == task.Spec.TaskID && approval.Capability == capability &&
				approval.ArtifactDigest == task.Status.ArtifactDigest && approval.PolicyVersion == task.Spec.PolicyVersion &&
				approval.Approver != "" && approval.ExpiresAt.After(r.Now()) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func (r *Reconciler) transition(task *domain.Task, phase domain.Phase, condition string) domain.Result {
	if task.Status.Phase == phase {
		return domain.Result{}
	}
	task.Status.Phase = phase
	task.Status.ObservedGeneration = task.Generation
	task.Status.Conditions = append(task.Status.Conditions, condition)
	return domain.Result{Changed: true, Requeue: phase != domain.Succeeded && phase != domain.PolicyRejected && phase != domain.BudgetExceeded}
}

func validate(spec domain.Spec) error {
	if spec.TaskID == "" || spec.Objective == "" || len(spec.AcceptanceCriteria) == 0 {
		return fmt.Errorf("task identity, objective and acceptance criteria are required")
	}
	if len(spec.IdempotencyKey) < 8 {
		return fmt.Errorf("idempotency key is too short")
	}
	if spec.MaximumCostUSD <= 0 || spec.EstimatedAgentUSD < 0 {
		return fmt.Errorf("invalid task economics")
	}
	if spec.PolicyVersion == "" {
		return fmt.Errorf("policy version is required")
	}
	for capability, authority := range spec.Capabilities {
		if capability == "" || (authority != "allow" && authority != "approval" && authority != "deny") {
			return fmt.Errorf("invalid capability authority")
		}
	}
	for _, ref := range spec.ContextRefs {
		if !(strings.HasPrefix(ref, "git:") || strings.HasPrefix(ref, "artifact:") || strings.HasPrefix(ref, "graph:")) {
			return fmt.Errorf("unsupported context reference")
		}
	}
	return nil
}

func artifactDigest(spec domain.Spec) string {
	payload, _ := json.Marshal(struct{ TaskID, IdempotencyKey, PolicyVersion string }{spec.TaskID, spec.IdempotencyKey, spec.PolicyVersion})
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func receipt(task *domain.Task) string {
	copyTask := *task
	copyTask.Status.ReceiptSHA256 = ""
	payload, _ := json.Marshal(copyTask)
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func round(value float64, places int) float64 {
	format := fmt.Sprintf("%%.%df", places)
	var rounded float64
	fmt.Sscanf(fmt.Sprintf(format, value), "%f", &rounded)
	return rounded
}
