package saga

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"github.com/bkeane/monad/pkg/param"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog/log"
)

type EventBridgeRule struct {
	BusName  string
	RuleName string
	Document string
}

type EventBridge struct {
	config param.Aws
}

func (s EventBridge) Init(ctx context.Context, c param.Aws) *EventBridge {
	return &EventBridge{
		config: c,
	}
}

func (s *EventBridge) Do(ctx context.Context) error {
	var action string
	var busName string
	if s.config.EventBridge.RuleTemplate != "" {
		action = "put"
		busName = s.config.EventBridge.BusName
	} else {
		action = "delete"
		busName = "*"
	}

	log.Info().
		Str("bus", busName).
		Str("rule", s.config.ResourceName()).
		Str("action", action).
		Msg("ensuring eventbridge rules")

	if err := s.Ensure(ctx); err != nil {
		return err
	}

	if err := s.Prune(ctx); err != nil {
		return err
	}

	return nil
}

func (s *EventBridge) Undo(ctx context.Context) error {
	log.Info().
		Str("bus", "*").
		Str("rule", s.config.ResourceName()).
		Str("action", "delete").
		Msg("destroying eventbridge rules")

	if err := s.Destroy(ctx); err != nil {
		return err
	}

	return nil
}

// Declarative Operations
func (s *EventBridge) Ensure(ctx context.Context) error {
	if s.config.EventBridge.RuleTemplate == "" {
		return s.Destroy(ctx)
	}

	return s.Deploy(ctx)
}

func (s *EventBridge) Prune(ctx context.Context) error {
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

func (s *EventBridge) Deploy(ctx context.Context) error {
	definedRules, err := s.GetDefinedRules(ctx)
	if err != nil {
		return err
	}

	for bus, rules := range definedRules {
		for name, rule := range rules {
			log.Debug().Str("bus", bus).Str("rule", name).Msg("put rule")
			if err := s.PutRule(ctx, rule); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *EventBridge) Destroy(ctx context.Context) error {
	rules, err := s.GetAssociatedRules(ctx)
	if err != nil {
		return err
	}

	for bus, rules := range rules {
		for name, rule := range rules {
			log.Debug().Str("bus", bus).Str("rule", name).Msg("delete rule")
			if err := s.DeleteRule(ctx, rule); err != nil {
				return err
			}
		}
	}

	return nil
}

// PUT Operations
func (s *EventBridge) PutRule(ctx context.Context, rule EventBridgeRule) error {
	var apiErr smithy.APIError

	putRuleInput := eventbridge.PutRuleInput{
		EventBusName: aws.String(rule.BusName),
		Name:         aws.String(rule.RuleName),
		Description:  aws.String("managed by monad"),
		State:        types.RuleStateEnabled,
	}

	// This is an inconsistency in the API interface of the eventbridge API.
	// All but one pattern in the API is considered an EventPattern (always JSON).
	// The ScheduleExpression is the odd one out (always a stringy cron-type expression).

	if strings.HasPrefix(rule.Document, "cron(") || strings.HasPrefix(rule.Document, "rate(") {
		scheduleExpression := chomp(rule.Document)
		putRuleInput.ScheduleExpression = aws.String(scheduleExpression)
	} else {
		putRuleInput.EventPattern = aws.String(rule.Document)
	}

	putTargetsInput := eventbridge.PutTargetsInput{
		EventBusName: aws.String(s.config.EventBridge.BusName),
		Rule:         aws.String(s.config.ResourceName()),
		Targets: []types.Target{
			{
				Id:  aws.String(s.config.ResourceName()),
				Arn: aws.String(s.config.FunctionArn()),
			},
		},
	}

	putRuleOutput, err := s.config.EventBridge.Client.PutRule(ctx, &putRuleInput)
	if err != nil {
		return err
	}

	_, err = s.config.EventBridge.Client.PutTargets(ctx, &putTargetsInput)
	if err != nil {
		return err
	}

	addPermissionsInput := lambda.AddPermissionInput{
		FunctionName: aws.String(s.config.ResourceName()),
		StatementId:  aws.String(s.config.EventBridgeStatementId()),
		Action:       aws.String("lambda:InvokeFunction"),
		Principal:    aws.String("events.amazonaws.com"),
		SourceArn:    aws.String(*putRuleOutput.RuleArn),
	}

	if _, err := s.config.Lambda.Client.AddPermission(ctx, &addPermissionsInput); err != nil {
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "ResourceConflictException":
				break
			default:
				return err
			}
		}
	}

	var tags []types.Tag
	for key, value := range s.config.Tags() {
		tags = append(tags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	putTagsInput := eventbridge.TagResourceInput{
		ResourceARN: putRuleOutput.RuleArn,
		Tags:        tags,
	}

	if _, err := s.config.EventBridge.Client.TagResource(ctx, &putTagsInput); err != nil {
		return err
	}

	return nil
}

// DELETE Operations
func (s *EventBridge) DeleteRule(ctx context.Context, rule EventBridgeRule) error {
	var apiErr smithy.APIError

	deletePermissionInput := lambda.RemovePermissionInput{
		FunctionName: aws.String(s.config.ResourceName()),
		StatementId:  aws.String(s.config.EventBridgeStatementId()),
	}

	deleteTargetsInput := eventbridge.RemoveTargetsInput{
		EventBusName: aws.String(rule.BusName),
		Rule:         aws.String(rule.RuleName),
		Ids:          []string{s.config.ResourceName()},
	}

	deleteRuleInput := eventbridge.DeleteRuleInput{
		EventBusName: aws.String(rule.BusName),
		Name:         aws.String(rule.RuleName),
	}

	if _, err := s.config.Lambda.Client.RemovePermission(ctx, &deletePermissionInput); err != nil {
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "ResourceNotFoundException":
				break
			default:
				return err
			}
		}
	}

	if _, err := s.config.EventBridge.Client.RemoveTargets(ctx, &deleteTargetsInput); err != nil {
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "ResourceNotFoundException":
				break
			default:
				return err
			}
		}
	}

	if _, err := s.config.EventBridge.Client.DeleteRule(ctx, &deleteRuleInput); err != nil {
		return err
	}

	return nil
}

// GET Operations
func (s *EventBridge) GetDefinedRules(ctx context.Context) (map[string]map[string]EventBridgeRule, error) {
	// This code _can_ handle many defined rules, but monad currently will only support one until more are necessary.
	document, err := s.config.RuleDocument()
	if err != nil {
		return nil, err
	}

	if document == "" {
		return map[string]map[string]EventBridgeRule{}, nil
	}

	ruleMap := map[string]map[string]EventBridgeRule{}
	// Initialize the inner map if it doesn't exist
	if _, exists := ruleMap[s.config.EventBridge.BusName]; !exists {
		ruleMap[s.config.EventBridge.BusName] = make(map[string]EventBridgeRule)
	}
	ruleMap[s.config.EventBridge.BusName][s.config.ResourceName()] = EventBridgeRule{
		BusName:  s.config.EventBridge.BusName,
		RuleName: s.config.ResourceName(),
		Document: document,
	}

	return ruleMap, nil
}

func (s *EventBridge) GetUndefinedRules(ctx context.Context) (map[string]map[string]EventBridgeRule, error) {
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

func (s *EventBridge) GetAssociatedRules(ctx context.Context) (map[string]map[string]EventBridgeRule, error) {
	listBuses := &eventbridge.ListEventBusesInput{}

	buses, err := s.config.EventBridge.Client.ListEventBuses(ctx, listBuses)
	if err != nil {
		return nil, err
	}

	associatedRules := make(map[string]map[string]EventBridgeRule)
	for _, bus := range buses.EventBuses {
		associatedRules[*bus.Name] = make(map[string]EventBridgeRule)

		listRuleNames := &eventbridge.ListRuleNamesByTargetInput{
			TargetArn:    aws.String(s.config.FunctionArn()),
			EventBusName: bus.Name,
		}

		target, err := s.config.EventBridge.Client.ListRuleNamesByTarget(ctx, listRuleNames)
		if err != nil {
			return nil, err
		}

		for _, associated := range target.RuleNames {
			listRules := &eventbridge.ListRulesInput{
				EventBusName: bus.Name,
				NamePrefix:   &associated,
			}

			output, err := s.config.EventBridge.Client.ListRules(ctx, listRules)
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

// Utility
func chomp(s string) string {
	s = strings.TrimLeftFunc(s, unicode.IsSpace)
	s = strings.TrimRightFunc(s, unicode.IsSpace)
	return s
}
