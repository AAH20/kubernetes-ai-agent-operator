package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/AAH20/kubernetes-ai-agent-operator/internal/domain"
	"github.com/AAH20/kubernetes-ai-agent-operator/internal/reconcile"
	"github.com/AAH20/kubernetes-ai-agent-operator/internal/render"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "simulate" {
		fmt.Fprintln(os.Stderr, "usage: agentmesh-operator simulate TASK.json --output DIR")
		os.Exit(2)
	}
	flags := flag.NewFlagSet("simulate", flag.ExitOnError)
	output := flags.String("output", "generated", "output directory")
	_ = flags.Parse(os.Args[3:])
	payload, err := os.ReadFile(os.Args[2])
	if err != nil {
		fail(err)
	}
	var task domain.Task
	if err := json.Unmarshal(payload, &task); err != nil {
		fail(err)
	}
	if task.Status.Phase == "" {
		task.Status.Phase = domain.Pending
	}
	reconciler := reconcile.New()
	for i := 0; i < 12; i++ {
		result, err := reconciler.Reconcile(&task, nil)
		if err != nil {
			fail(err)
		}
		if !result.Changed || task.Status.Phase == domain.AwaitingApproval || task.Status.Phase == domain.Succeeded || task.Status.Phase == domain.PolicyRejected || task.Status.Phase == domain.BudgetExceeded {
			break
		}
	}
	if err := render.Bundle(task, *output); err != nil {
		fail(err)
	}
	encoded, _ := json.MarshalIndent(task.Status, "", "  ")
	fmt.Println(string(encoded))
}

func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
