package eventbridge

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	eventbridgetypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog/log"
)

type EventBridgeConfig interface {
	BusName() string
	Rules() map[string]string
	PermissionStatementId() string
	Client() *eventbridge.Client
	Tags() []eventbridgetypes.Tag
}

type LambdaConfig interface {
	FunctionName() string
	FunctionArn() string
	Client() *lambda.Client
	Tags() map[string]string
}

type EventBridgeRule struct {
	BusName  string
	RuleName string
	Document string
}

type Rule struct {
	BusName  string
	RuleName string
}

type Summary struct {
	RulesCreated []Rule
	RulesDeleted []Rule
}

type Step struct {
	eventbridge EventBridgeConfig
	lambda      LambdaConfig
}

func Derive(eventbridge EventBridgeConfig, lambda LambdaConfig) *Step {
	return &Step{
		eventbridge: eventbridge,
		lambda:      lambda,
	}
}

func (s *Step) Mount(ctx context.Context) error {
	// Call internal unmount silently (don't log deletes) to clean up existing resources
	if _, err := s.unmount(ctx); err != nil {
		return err
	}

	// Call internal mount and log only the creates as action=put
	summary, err := s.mount(ctx)
	if err != nil {
		return err
	}

	// Log only action=put for consistency with other services
	for _, rule := range summary.RulesCreated {
		log.Info().
			Str("action", "put").
			Str("bus", rule.BusName).
			Str("rule", rule.RuleName).
			Msg("eventbridge")
	}

	return nil
}

func (s *Step) Unmount(ctx context.Context) error {
	summary, err := s.unmount(ctx)
	if err != nil {
		return err
	}

	// Log action=delete
	for _, rule := range summary.RulesDeleted {
		log.Info().
			Str("action", "delete").
			Str("bus", rule.BusName).
			Str("rule", rule.RuleName).
			Msg("eventbridge")
	}

	return nil
}

// Internal methods that return summaries of work done
func (s *Step) mount(ctx context.Context) (Summary, error) {
	var summary Summary

	definedRules, err := s.GetDefinedRules(ctx)
	if err != nil {
		return summary, err
	}

	for bus, rules := range definedRules {
		for name, rule := range rules {
			log.Debug().Str("bus", bus).Str("rule", name).Msg("put rule")
			if err := s.PutRule(ctx, rule); err != nil {
				return summary, err
			}
			
			// Record what was created
			summary.RulesCreated = append(summary.RulesCreated, Rule{
				BusName:  bus,
				RuleName: name,
			})
		}
	}

	return summary, nil
}

func (s *Step) unmount(ctx context.Context) (Summary, error) {
	var summary Summary

	rules, err := s.GetAssociatedRules(ctx)
	if err != nil {
		return summary, err
	}

	for bus, rules := range rules {
		for name, rule := range rules {
			log.Debug().Str("bus", bus).Str("rule", name).Msg("delete rule")
			if err := s.DeleteRule(ctx, rule); err != nil {
				return summary, err
			}
			
			// Record what was deleted
			summary.RulesDeleted = append(summary.RulesDeleted, Rule{
				BusName:  bus,
				RuleName: name,
			})
		}
	}

	if err := s.prune(ctx); err != nil {
		return summary, err
	}

	return summary, nil
}

func (s *Step) prune(ctx context.Context) error {
	undefinedRules, err := s.GetUndefinedRules(ctx)
	if err != nil {
		return err
	}

	for bus, rules := range undefinedRules {
		for name, rule := range rules {
			log.Debug().Str("bus", bus).Str("rule", name).Msg("prune rule")
			if err := s.DeleteRule(ctx, rule); err != nil {
				return err
			}
		}
	}

	return nil
}

// PUT Operations
func (s *Step) PutRule(ctx context.Context, rule EventBridgeRule) error {
	var apiErr smithy.APIError

	putRuleInput := eventbridge.PutRuleInput{
		EventBusName: aws.String(rule.BusName),
		Name:         aws.String(rule.RuleName),
		Description:  aws.String("managed by monad"),
		State:        eventbridgetypes.RuleStateEnabled,
	}

	// This is an inconsistency in the interface of the eventbridge API.
	// All but one pattern in the API is considered an EventPattern (always JSON).
	// The ScheduleExpression is the odd one out (always a stringy cron-type expression).

	if strings.HasPrefix(rule.Document, "cron(") || strings.HasPrefix(rule.Document, "rate(") {
		putRuleInput.ScheduleExpression = aws.String(rule.Document)
	} else {
		putRuleInput.EventPattern = aws.String(rule.Document)
	}

	putTargetsInput := eventbridge.PutTargetsInput{
		EventBusName: aws.String(rule.BusName),
		Rule:         aws.String(rule.RuleName),
		Targets: []eventbridgetypes.Target{
			{
				Id:  aws.String(s.lambda.FunctionName()),
				Arn: aws.String(s.lambda.FunctionArn()),
			},
		},
	}

	putRuleOutput, err := s.eventbridge.Client().PutRule(ctx, &putRuleInput)
	if err != nil {
		return err
	}

	_, err = s.eventbridge.Client().PutTargets(ctx, &putTargetsInput)
	if err != nil {
		return err
	}

	addPermissionsInput := lambda.AddPermissionInput{
		FunctionName: aws.String(s.lambda.FunctionName()),
		StatementId:  aws.String(s.eventbridge.PermissionStatementId()),
		Action:       aws.String("lambda:InvokeFunction"),
		Principal:    aws.String("events.amazonaws.com"),
		SourceArn:    aws.String(*putRuleOutput.RuleArn),
	}

	if _, err := s.lambda.Client().AddPermission(ctx, &addPermissionsInput); err != nil {
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "ResourceConflictException":
				break
			default:
				return err
			}
		}
	}

	putTagsInput := eventbridge.TagResourceInput{
		ResourceARN: putRuleOutput.RuleArn,
		Tags:        s.eventbridge.Tags(),
	}

	if _, err := s.eventbridge.Client().TagResource(ctx, &putTagsInput); err != nil {
		return err
	}

	return nil
}

// DELETE Operations
func (s *Step) DeleteRule(ctx context.Context, rule EventBridgeRule) error {
	var apiErr smithy.APIError

	deletePermissionInput := lambda.RemovePermissionInput{
		FunctionName: aws.String(s.lambda.FunctionName()),
		StatementId:  aws.String(s.eventbridge.PermissionStatementId()),
	}

	deleteTargetsInput := eventbridge.RemoveTargetsInput{
		EventBusName: aws.String(rule.BusName),
		Rule:         aws.String(rule.RuleName),
		Ids:          []string{s.lambda.FunctionName()},
	}

	deleteRuleInput := eventbridge.DeleteRuleInput{
		EventBusName: aws.String(rule.BusName),
		Name:         aws.String(rule.RuleName),
	}

	if _, err := s.lambda.Client().RemovePermission(ctx, &deletePermissionInput); err != nil {
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "ResourceNotFoundException":
				break
			default:
				return err
			}
		}
	}

	if _, err := s.eventbridge.Client().RemoveTargets(ctx, &deleteTargetsInput); err != nil {
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "ResourceNotFoundException":
				break
			default:
				return err
			}
		}
	}

	if _, err := s.eventbridge.Client().DeleteRule(ctx, &deleteRuleInput); err != nil {
		return err
	}

	return nil
}

// GET Operations
func (s *Step) GetDefinedRules(ctx context.Context) (map[string]map[string]EventBridgeRule, error) {
	// This code can now handle multiple defined rules
	busName := s.eventbridge.BusName()
	rules := s.eventbridge.Rules()

	// Only create rules if rules are defined
	if len(rules) == 0 {
		return map[string]map[string]EventBridgeRule{}, nil
	}

	// If no bus is explicitly configured, use default bus
	if busName == "" {
		busName = "default"
	}

	ruleMap := map[string]map[string]EventBridgeRule{}
	// Initialize the inner map if it doesn't exist
	if _, exists := ruleMap[busName]; !exists {
		ruleMap[busName] = make(map[string]EventBridgeRule)
	}

	// Create EventBridgeRule for each defined rule
	for ruleName, document := range rules {
		ruleMap[busName][ruleName] = EventBridgeRule{
			BusName:  busName,
			RuleName: ruleName,
			Document: document,
		}
	}

	return ruleMap, nil
}

func (s *Step) GetUndefinedRules(ctx context.Context) (map[string]map[string]EventBridgeRule, error) {
	definedRules, err := s.GetDefinedRules(ctx)
	if err != nil {
		return nil, err
	}

	associatedRules, err := s.GetAssociatedRules(ctx)
	if err != nil {
		return nil, err
	}

	undefinedRules := make(map[string]map[string]EventBridgeRule)
	for bus, rules := range associatedRules {
		for rule := range rules {
			if _, exists := definedRules[bus][rule]; !exists {
				if _, exists := undefinedRules[bus]; !exists {
					undefinedRules[bus] = make(map[string]EventBridgeRule)
				}

				undefinedRules[bus][rule] = associatedRules[bus][rule]
			}
		}
	}

	return undefinedRules, nil
}

func (s *Step) GetAssociatedRules(ctx context.Context) (map[string]map[string]EventBridgeRule, error) {
	listBuses := &eventbridge.ListEventBusesInput{}

	buses, err := s.eventbridge.Client().ListEventBuses(ctx, listBuses)
	if err != nil {
		return nil, err
	}

	associatedRules := make(map[string]map[string]EventBridgeRule)
	for _, bus := range buses.EventBuses {
		associatedRules[*bus.Name] = make(map[string]EventBridgeRule)

		listRuleNames := &eventbridge.ListRuleNamesByTargetInput{
			TargetArn:    aws.String(s.lambda.FunctionArn()),
			EventBusName: bus.Name,
		}

		target, err := s.eventbridge.Client().ListRuleNamesByTarget(ctx, listRuleNames)
		if err != nil {
			return nil, err
		}

		for _, associated := range target.RuleNames {
			listRules := &eventbridge.ListRulesInput{
				EventBusName: bus.Name,
				NamePrefix:   &associated,
			}

			output, err := s.eventbridge.Client().ListRules(ctx, listRules)
			if err != nil {
				return nil, err
			}

			for _, rule := range output.Rules {
				if associated == *rule.Name {
					if rule.ScheduleExpression != nil {
						associatedRules[*bus.Name][*rule.Name] = EventBridgeRule{
							BusName:  *bus.Name,
							RuleName: *rule.Name,
							Document: *rule.ScheduleExpression,
						}
					} else {
						associatedRules[*bus.Name][*rule.Name] = EventBridgeRule{
							BusName:  *bus.Name,
							RuleName: *rule.Name,
							Document: *rule.EventPattern,
						}
					}
				}
			}
		}
	}

	return associatedRules, nil
}

