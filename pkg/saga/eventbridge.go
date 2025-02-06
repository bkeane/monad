package saga

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"github.com/bkeane/monad/pkg/config/release"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog"
)

type EventBridge struct {
	release     release.Config
	eventbridge *eventbridge.Client
	lambda      *lambda.Client
	log         *zerolog.Logger
}

func (s EventBridge) Init(ctx context.Context, r release.Config) *EventBridge {
	return &EventBridge{
		release:     r,
		eventbridge: eventbridge.NewFromConfig(r.AwsConfig),
		lambda:      lambda.NewFromConfig(r.AwsConfig),
		log:         zerolog.Ctx(ctx),
	}
}

func (s *EventBridge) Do(ctx context.Context) error {
	if err := s.Ensure(ctx); err != nil {
		return err
	}

	if err := s.Prune(ctx); err != nil {
		return err
	}

	return nil
}

func (s *EventBridge) Undo(ctx context.Context) error {
	s.log.Info().Msg("destroying eventbridge rules")
	if err := s.Destroy(ctx); err != nil {
		return err
	}

	return nil
}

// Declarative Operations

func (s *EventBridge) Ensure(ctx context.Context) error {
	if s.release.Substrate.EventBridge.Enable != nil {
		if !*s.release.Substrate.EventBridge.Enable {
			s.log.Info().Msg("destroying eventbridge rules")
			return s.Destroy(ctx)
		}

		if *s.release.Substrate.EventBridge.Enable {
			s.log.Info().Msg("deploying eventbridge rules")
			return s.Deploy(ctx)
		}
	}

	s.log.Info().Msg("leaving eventbridge state unchanged")
	return nil
}

func (s *EventBridge) Prune(ctx context.Context) error {
	undefinedRules, err := s.GetUndefinedRules(ctx)
	if err != nil {
		return err
	}

	for bus, rules := range undefinedRules {
		for name, rule := range rules {
			s.log.Debug().Str("bus", bus).Str("rule", name).Msg("prune rule")
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
			s.log.Debug().Str("bus", bus).Str("rule", name).Msg("put rule")
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
			s.log.Debug().Str("bus", bus).Str("rule", name).Msg("delete rule")
			if err := s.DeleteRule(ctx, rule); err != nil {
				return err
			}
		}
	}

	return nil
}

// PUT Operations

func (s *EventBridge) PutRule(ctx context.Context, rule release.EventBridgeRule) error {
	var apiErr smithy.APIError

	putRuleInput := eventbridge.PutRuleInput{
		EventBusName: aws.String(rule.BusName),
		Name:         aws.String(rule.RuleName),
		Description:  aws.String("managed"),
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
		EventBusName: aws.String(rule.BusName),
		Rule:         aws.String(rule.RuleName),
		Targets: []types.Target{
			{
				Id:  aws.String(s.release.FunctionName()),
				Arn: aws.String(s.release.FunctionArn()),
			},
		},
	}

	putRuleOutput, err := s.eventbridge.PutRule(ctx, &putRuleInput)
	if err != nil {
		return err
	}

	_, err = s.eventbridge.PutTargets(ctx, &putTargetsInput)
	if err != nil {
		return err
	}

	addPermissionsInput := lambda.AddPermissionInput{
		FunctionName: aws.String(s.release.FunctionName()),
		StatementId:  aws.String(s.release.EventBridgeStatementId(rule.BusName, rule.RuleName)),
		Action:       aws.String("lambda:InvokeFunction"),
		Principal:    aws.String("events.amazonaws.com"),
		SourceArn:    aws.String(*putRuleOutput.RuleArn),
	}

	if _, err := s.lambda.AddPermission(ctx, &addPermissionsInput); err != nil {
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
	for key, value := range s.release.ResourceTags() {
		tags = append(tags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	putTagsInput := eventbridge.TagResourceInput{
		ResourceARN: putRuleOutput.RuleArn,
		Tags:        tags,
	}

	if _, err := s.eventbridge.TagResource(ctx, &putTagsInput); err != nil {
		return err
	}

	return nil
}

// DELETE Operations

func (s *EventBridge) DeleteRule(ctx context.Context, rule release.EventBridgeRule) error {
	var apiErr smithy.APIError

	deletePermissionInput := lambda.RemovePermissionInput{
		FunctionName: aws.String(s.release.FunctionName()),
		StatementId:  aws.String(s.release.EventBridgeStatementId(rule.BusName, rule.RuleName)),
	}

	deleteTargetsInput := eventbridge.RemoveTargetsInput{
		EventBusName: aws.String(rule.BusName),
		Rule:         aws.String(rule.RuleName),
		Ids:          []string{s.release.FunctionName()},
	}

	deleteRuleInput := eventbridge.DeleteRuleInput{
		EventBusName: aws.String(rule.BusName),
		Name:         aws.String(rule.RuleName),
	}

	if _, err := s.lambda.RemovePermission(ctx, &deletePermissionInput); err != nil {
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "ResourceNotFoundException":
				break
			default:
				return err
			}
		}
	}

	if _, err := s.eventbridge.RemoveTargets(ctx, &deleteTargetsInput); err != nil {
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "ResourceNotFoundException":
				break
			default:
				return err
			}
		}
	}

	if _, err := s.eventbridge.DeleteRule(ctx, &deleteRuleInput); err != nil {
		return err
	}

	return nil
}

// GET Operations

func (s *EventBridge) GetDefinedRules(ctx context.Context) (map[string]map[string]release.EventBridgeRule, error) {
	return s.release.EventBridgeRules(s.release.Substrate.EventBridge.BusName)
}

func (s *EventBridge) GetUndefinedRules(ctx context.Context) (map[string]map[string]release.EventBridgeRule, error) {
	definedRules, err := s.GetDefinedRules(ctx)
	if err != nil {
		return nil, err
	}

	associatedRules, err := s.GetAssociatedRules(ctx)
	if err != nil {
		return nil, err
	}

	undefinedRules := make(map[string]map[string]release.EventBridgeRule)
	for bus, rules := range associatedRules {
		for rule := range rules {
			if _, exists := definedRules[bus][rule]; !exists {
				if _, exists := undefinedRules[bus]; !exists {
					undefinedRules[bus] = make(map[string]release.EventBridgeRule)
				}

				undefinedRules[bus][rule] = associatedRules[bus][rule]
			}
		}
	}

	return undefinedRules, nil
}

func (s *EventBridge) GetAssociatedRules(ctx context.Context) (map[string]map[string]release.EventBridgeRule, error) {
	listBuses := &eventbridge.ListEventBusesInput{}

	buses, err := s.eventbridge.ListEventBuses(ctx, listBuses)
	if err != nil {
		return nil, err
	}

	associatedRules := make(map[string]map[string]release.EventBridgeRule)
	for _, bus := range buses.EventBuses {
		associatedRules[*bus.Name] = make(map[string]release.EventBridgeRule)

		listRuleNames := &eventbridge.ListRuleNamesByTargetInput{
			TargetArn:    aws.String(s.release.FunctionArn()),
			EventBusName: bus.Name,
		}

		target, err := s.eventbridge.ListRuleNamesByTarget(ctx, listRuleNames)
		if err != nil {
			return nil, err
		}

		for _, associated := range target.RuleNames {
			listRules := &eventbridge.ListRulesInput{
				EventBusName: bus.Name,
				NamePrefix:   &associated,
			}

			output, err := s.eventbridge.ListRules(ctx, listRules)
			if err != nil {
				return nil, err
			}

			for _, rule := range output.Rules {
				if associated == *rule.Name {
					if rule.ScheduleExpression != nil {
						associatedRules[*bus.Name][*rule.Name] = release.EventBridgeRule{
							BusName:  *bus.Name,
							RuleName: *rule.Name,
							Document: *rule.ScheduleExpression,
						}
					} else {
						associatedRules[*bus.Name][*rule.Name] = release.EventBridgeRule{
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
