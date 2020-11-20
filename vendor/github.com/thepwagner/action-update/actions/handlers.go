package actions

import (
	"context"
	"fmt"

	parser "github.com/caarlos0/env/v6"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

type Handlers struct {
	IssueComment       func(context.Context, *github.IssueCommentEvent) error
	PullRequest        func(context.Context, *github.PullRequestEvent) error
	Release            func(context.Context, *github.ReleaseEvent) error
	RepositoryDispatch func(context.Context, *github.RepositoryDispatchEvent) error
	Schedule           func(context.Context) error
	WorkflowDispatch   func(context.Context) error
}

// Handle invokes the appropriate handler for an given actions Environment.
func (h *Handlers) Handle(ctx context.Context, env *Environment) error {
	// Is there a handler for this event?
	log := logrus.WithField("event_name", env.GitHubEventName)
	handler := h.handler(env.GitHubEventName)
	if handler == nil {
		log.Warn("unhandled event")
		return nil
	}

	// Parse event and invoke handler:
	evt, err := env.ParseEvent()
	if err != nil {
		return err
	}
	if err := handler(ctx, evt); err != nil {
		return err
	}
	log.Debug("handler complete")
	return nil
}

// ParseAndHandle hydrates an `env:""` annotated struct, then Handle()s
func (h *Handlers) ParseAndHandle(ctx context.Context, env interface{}) error {
	aEnv, ok := env.(ActionEnvironment)
	if !ok {
		return fmt.Errorf("env should embed actions.Environment")
	}

	if err := parser.Parse(env); err != nil {
		return fmt.Errorf("parsing environment: %w", err)
	}

	actionEnv := aEnv.env()
	logrus.SetLevel(actionEnv.LogLevel())
	return h.Handle(ctx, actionEnv)
}

func (h *Handlers) handler(event string) func(context.Context, interface{}) error {
	switch event {
	case "issue_comment":
		if h.IssueComment == nil {
			return nil
		}
		return func(ctx context.Context, evt interface{}) error {
			return h.IssueComment(ctx, evt.(*github.IssueCommentEvent))
		}

	case "pull_request":
		if h.PullRequest == nil {
			return nil
		}
		return func(ctx context.Context, evt interface{}) error {
			return h.PullRequest(ctx, evt.(*github.PullRequestEvent))
		}

	case "release":
		if h.Release == nil {
			return nil
		}
		return func(ctx context.Context, evt interface{}) error {
			return h.Release(ctx, evt.(*github.ReleaseEvent))
		}

	case "repository_dispatch":
		if h.RepositoryDispatch == nil {
			return nil
		}
		return func(ctx context.Context, evt interface{}) error {
			return h.RepositoryDispatch(ctx, evt.(*github.RepositoryDispatchEvent))
		}

	case "schedule":
		if h.Schedule == nil {
			return nil
		}
		return func(ctx context.Context, _ interface{}) error {
			return h.Schedule(ctx)
		}

	case "workflow_dispatch":
		if h.WorkflowDispatch == nil {
			return nil
		}
		return func(ctx context.Context, _ interface{}) error {
			return h.WorkflowDispatch(ctx)
		}

	default:
		return nil
	}
}
