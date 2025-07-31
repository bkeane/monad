package saga

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go"
	"github.com/bkeane/monad/pkg/param"
	"github.com/rs/zerolog/log"
)

type IAM struct {
	config param.Aws
}

func (s IAM) Init(ctx context.Context, c param.Aws) *IAM {
	return &IAM{
		config: c,
	}
}

func (s *IAM) Do(ctx context.Context) error {
	log.Info().
		Str("action", "put").
		Str("role", s.config.IAM().EniRoleName()).
		Msg("iam")

	if err := s.PutEniRole(ctx); err != nil {
		return err
	}

	log.Info().
		Str("action", "put").
		Str("policy", s.config.Schema().Name()).
		Msg("iam")

	if err := s.PutPolicy(ctx); err != nil {
		return err
	}

	log.Info().
		Str("action", "put").
		Str("role", s.config.Schema().Name()).
		Msg("iam")

	if err := s.PutRole(ctx); err != nil {
		return err
	}

	log.Info().
		Str("action", "attach").
		Str("role", s.config.Schema().Name()).
		Str("policy", s.config.Schema().Name()).
		Str("boundary", s.config.IAM().BoundaryPolicy()).
		Msg("iam")

	if err := s.AttachRolePolicy(ctx); err != nil {
		return err
	}

	return nil
}

func (s *IAM) Undo(ctx context.Context) error {
	log.Info().
		Str("action", "detach").
		Str("role", s.config.Schema().Name()).
		Str("policy", s.config.Schema().Name()).
		Str("boundary", s.config.IAM().BoundaryPolicy()).
		Msg("iam")

	if err := s.DetachRolePolicy(ctx); err != nil {
		return err
	}

	log.Info().
		Str("action", "delete").
		Str("role", s.config.Schema().Name()).
		Msg("iam")

	if err := s.DeleteRole(ctx); err != nil {
		return err
	}

	log.Info().
		Str("action", "delete").
		Str("policy", s.config.Schema().Name()).
		Msg("iam")

	if err := s.DeletePolicy(ctx); err != nil {
		return err
	}

	return nil
}

// PUT OPERATIONS

func (s *IAM) PutPolicy(ctx context.Context) error {
	var apiErr smithy.APIError

	policyDocument, err := s.config.IAM().PolicyDocument()
	if err != nil {
		return err
	}

	var tagSlice []types.Tag
	for key, value := range s.config.Schema().Tags() {
		tagSlice = append(tagSlice, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	create := &iam.CreatePolicyInput{
		PolicyName:     aws.String(s.config.Schema().Name()),
		PolicyDocument: aws.String(policyDocument),
	}

	update := &iam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(s.config.IAM().PolicyArn()),
		PolicyDocument: aws.String(policyDocument),
		SetAsDefault:   true,
	}

	tag := &iam.TagPolicyInput{
		PolicyArn: aws.String(s.config.IAM().PolicyArn()),
		Tags:      tagSlice,
	}

	_, err = s.config.IAM().Client().CreatePolicy(ctx, create)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "EntityAlreadyExists":
			if err := s.GCPolicyVersions(ctx, s.config.IAM().PolicyArn()); err != nil {
				return err
			}

			_, err = s.config.IAM().Client().CreatePolicyVersion(ctx, update)
			if err != nil {
				return err
			}

		default:
			return err
		}
	}

	_, err = s.config.IAM().Client().TagPolicy(ctx, tag)
	if err != nil {
		return err
	}

	return nil
}

// Always ensure VPC role exists to ensure ENI garbage collection works after lambda deletion
func (s *IAM) PutEniRole(ctx context.Context) error {
	var apiErr smithy.APIError

	roleDocument, err := s.config.IAM().RoleDocument()
	if err != nil {
		return err
	}

	create := &iam.CreateRoleInput{
		RoleName:                 aws.String(s.config.IAM().EniRoleName()),
		AssumeRolePolicyDocument: aws.String(roleDocument),
	}

	update := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String(s.config.IAM().EniRoleName()),
		PolicyDocument: aws.String(roleDocument),
	}

	attach := &iam.AttachRolePolicyInput{
		RoleName:  aws.String(s.config.IAM().EniRoleName()),
		PolicyArn: aws.String(s.config.IAM().EniRolePolicyArn()),
	}

	_, err = s.config.IAM().Client().CreateRole(ctx, create)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "EntityAlreadyExists":
			_, err = s.config.IAM().Client().UpdateAssumeRolePolicy(ctx, update)
			if err != nil {
				return err
			}
		default:
			return err
		}
	}

	_, err = s.config.IAM().Client().AttachRolePolicy(ctx, attach)
	if err != nil {
		return err
	}

	return nil
}

func (s *IAM) PutRole(ctx context.Context) error {
	var apiErr smithy.APIError

	roleDocument, err := s.config.IAM().RoleDocument()
	if err != nil {
		return err
	}

	var tagSlice []types.Tag
	for key, value := range s.config.Schema().Tags() {
		tagSlice = append(tagSlice, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	create := &iam.CreateRoleInput{
		RoleName:                 aws.String(s.config.Schema().Name()),
		AssumeRolePolicyDocument: aws.String(roleDocument),
	}

	if s.config.IAM().BoundaryPolicyArn() != "" {
		create.PermissionsBoundary = aws.String(s.config.IAM().BoundaryPolicyArn())
	}

	updatePolicy := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       create.RoleName,
		PolicyDocument: create.AssumeRolePolicyDocument,
	}

	tag := &iam.TagRoleInput{
		RoleName: aws.String(s.config.Schema().Name()),
		Tags:     tagSlice,
	}

	_, err = s.config.IAM().Client().CreateRole(ctx, create)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "EntityAlreadyExists":
			_, err = s.config.IAM().Client().UpdateAssumeRolePolicy(ctx, updatePolicy)
			if err != nil {
				return err
			}

		default:
			return err
		}
	}

	_, err = s.config.IAM().Client().TagRole(ctx, tag)
	if err != nil {
		return err
	}

	return nil
}

func (s *IAM) AttachRolePolicy(ctx context.Context) error {
	var apiErr smithy.APIError

	attach := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(s.config.IAM().PolicyArn()),
		RoleName:  aws.String(s.config.Schema().Name()),
	}

	_, err := s.config.IAM().Client().AttachRolePolicy(ctx, attach)
	if err != nil {
		return err
	}

	if s.config.IAM().BoundaryPolicyArn() != "" {
		boundary := &iam.PutRolePermissionsBoundaryInput{
			RoleName:            aws.String(s.config.Schema().Name()),
			PermissionsBoundary: aws.String(s.config.IAM().BoundaryPolicyArn()),
		}

		_, err := s.config.IAM().Client().PutRolePermissionsBoundary(ctx, boundary)
		if err != nil {
			return err
		}
	} else {
		boundary := &iam.DeleteRolePermissionsBoundaryInput{
			RoleName: aws.String(s.config.Schema().Name()),
		}

		_, err := s.config.IAM().Client().DeleteRolePermissionsBoundary(ctx, boundary)
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "NoSuchEntity":
				return nil
			default:
				return err
			}
		}
	}

	return nil
}

// DELETE OPERATIONS

func (s *IAM) DetachRolePolicy(ctx context.Context) error {
	var apiErr smithy.APIError

	detach := &iam.DetachRolePolicyInput{
		PolicyArn: aws.String(s.config.IAM().PolicyArn()),
		RoleName:  aws.String(s.config.Schema().Name()),
	}

	boundary := &iam.DeleteRolePermissionsBoundaryInput{
		RoleName: detach.RoleName,
	}

	_, err := s.config.IAM().Client().DeleteRolePermissionsBoundary(ctx, boundary)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchEntity":
			// in the case that the boundary does not exist, we can continue
			break
		default:
			return err
		}
	}

	_, err = s.config.IAM().Client().DetachRolePolicy(ctx, detach)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchEntity":
			// in the case that the policy or role does not exist, we can return nil
			return nil
		default:
			return err
		}
	}

	return nil
}

func (s *IAM) DeleteRole(ctx context.Context) error {
	var apiErr smithy.APIError

	delete := &iam.DeleteRoleInput{
		RoleName: aws.String(s.config.Schema().Name()),
	}

	_, err := s.config.IAM().Client().DeleteRole(ctx, delete)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchEntity":
			return nil
		default:
			return err
		}
	}

	return nil
}

func (s *IAM) DeletePolicy(ctx context.Context) error {
	var apiErr smithy.APIError

	delete := &iam.DeletePolicyInput{
		PolicyArn: aws.String(s.config.IAM().PolicyArn()),
	}

	if err := s.GCPolicyVersions(ctx, s.config.IAM().PolicyArn()); err != nil {
		return err
	}

	_, err := s.config.IAM().Client().DeletePolicy(ctx, delete)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchEntity":
			return nil
		default:
			return err
		}
	}

	return nil
}

// Util

func (s *IAM) GCPolicyVersions(ctx context.Context, policyArn string) error {
	var apiErr smithy.APIError

	index := &iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(policyArn),
	}

	policyVersions, err := s.config.IAM().Client().ListPolicyVersions(ctx, index)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchEntity":
			return nil
		default:
			return err
		}
	}

	for _, version := range policyVersions.Versions {
		if !version.IsDefaultVersion {
			delete := &iam.DeletePolicyVersionInput{
				PolicyArn: aws.String(policyArn),
				VersionId: version.VersionId,
			}

			if _, err = s.config.IAM().Client().DeletePolicyVersion(ctx, delete); err != nil {
				if errors.As(err, &apiErr) {
					switch apiErr.ErrorCode() {
					case "NoSuchEntity":
						return nil
					default:
						return err
					}
				}
			}
		}
	}

	return nil
}
