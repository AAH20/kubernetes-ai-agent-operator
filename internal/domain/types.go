package domain

import "time"

type Phase string

const (
	Pending             Phase = "Pending"
	Admitted            Phase = "Admitted"
	SandboxProvisioning Phase = "SandboxProvisioning"
	Running             Phase = "Running"
	Evaluating          Phase = "Evaluating"
	AwaitingApproval    Phase = "AwaitingApproval"
	Delivering          Phase = "Delivering"
	Succeeded           Phase = "Succeeded"
	PolicyRejected      Phase = "PolicyRejected"
	BudgetExceeded      Phase = "BudgetExceeded"
)

type Spec struct {
	TaskID             string            `json:"taskId"`
	Objective          string            `json:"objective"`
	AcceptanceCriteria []string          `json:"acceptanceCriteria"`
	Capabilities       map[string]string `json:"capabilities"`
	DataClassification string            `json:"dataClassification"`
	TenantTrust        string            `json:"tenantTrust"`
	MutationImpact     string            `json:"mutationImpact"`
	MaximumCostUSD     float64           `json:"maximumCostUSD"`
	EstimatedAgentUSD  float64           `json:"estimatedAgentUSD"`
	ContextRefs        []string          `json:"contextRefs"`
	IdempotencyKey     string            `json:"idempotencyKey"`
	PolicyVersion      string            `json:"policyVersion"`
}

type Approval struct {
	TaskID         string    `json:"taskId"`
	Capability     string    `json:"capability"`
	ArtifactDigest string    `json:"artifactDigest"`
	PolicyVersion  string    `json:"policyVersion"`
	Approver       string    `json:"approver"`
	ExpiresAt      time.Time `json:"expiresAt"`
}

type Status struct {
	Phase              Phase    `json:"phase"`
	ObservedGeneration int64    `json:"observedGeneration"`
	RuntimeProfile     string   `json:"runtimeProfile,omitempty"`
	SandboxClaim       string   `json:"sandboxClaim,omitempty"`
	ArtifactDigest     string   `json:"artifactDigest,omitempty"`
	EstimatedCostUSD   float64  `json:"estimatedCostUSD,omitempty"`
	RequiredApprovals  []string `json:"requiredApprovals,omitempty"`
	DeniedCapabilities []string `json:"deniedCapabilities,omitempty"`
	Conditions         []string `json:"conditions,omitempty"`
	ReceiptSHA256      string   `json:"receiptSHA256,omitempty"`
}

type Task struct {
	Generation int64  `json:"generation"`
	Spec       Spec   `json:"spec"`
	Status     Status `json:"status"`
}

type Result struct {
	Changed      bool
	Requeue      bool
	RequeueAfter time.Duration
}
