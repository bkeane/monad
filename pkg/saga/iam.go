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
		Str("name", s.config.EniRoleName()).
		Str("arn", s.config.EniRoleArn()).
		Msg("ensuring eni role")

	if err := s.PutEniRole(ctx); err != nil {
		return err
	}

	log.Info().
		Str("name", s.config.ResourceName()).
		Str("arn", s.config.PolicyArn()).
		Msg("ensuring policy")

	if err := s.PutPolicy(ctx); err != nil {
		return err
	}

	log.Info().
		Str("name", s.config.ResourceName()).
		Str("arn", s.config.RoleArn()).
		Msg("ensuring role")

	if err := s.PutRole(ctx); err != nil {
		return err
	}

	log.Info().
		Str("role", s.config.RoleArn()).
		Str("policy", s.config.PolicyArn()).
		Str("boundary", s.config.BoundaryPolicyArn()).
		Msg("ensuring role policy attachment")

	if err := s.AttachRolePolicy(ctx); err != nil {
		return err
	}

	return nil
}

func (s *IAM) Undo(ctx context.Context) error {
	log.Info().
		Str("role", s.config.RoleArn()).
		Str("policy", s.config.PolicyArn()).
		Str("boundary", s.config.BoundaryPolicyArn()).
		Msg("deleting role policy attachment")

	if err := s.DetachRolePolicy(ctx); err != nil {
		return err
	}

	log.Info().
		Str("name", s.config.ResourceName()).
		Str("arn", s.config.RoleArn()).
		Msg("deleting role")

	if err := s.DeleteRole(ctx); err != nil {
		return err
	}

	log.Info().
		Str("name", s.config.ResourceName()).
		Str("arn", s.config.PolicyArn()).
		Msg("deleting policy")

	if err := s.DeletePolicy(ctx); err != nil {
		return err
	}

	return nil
}

// PUT OPERATIONS

func (s *IAM) PutPolicy(ctx context.Context) error {
	var apiErr smithy.APIError

	policyDocument, err := s.config.PolicyDocument()
	if err != nil {
		return err
	}

	var tagSlice []types.Tag
	for key, value := range s.config.Tags() {
		tagSlice = append(tagSlice, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	create := &iam.CreatePolicyInput{
		PolicyName:     aws.String(s.config.ResourceName()),
		PolicyDocument: aws.String(policyDocument),
	}

	update := &iam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(s.config.PolicyArn()),
		PolicyDocument: aws.String(policyDocument),
		SetAsDefault:   true,
	}

	tag := &iam.TagPolicyInput{
		PolicyArn: aws.String(s.config.PolicyArn()),
		Tags:      tagSlice,
	}

	_, err = s.config.Iam.Client.CreatePolicy(ctx, create)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "EntityAlreadyExists":
			if err := s.GCPolicyVersions(ctx, s.config.PolicyArn()); err != nil {
				return err
			}

			_, err = s.config.Iam.Client.CreatePolicyVersion(ctx, update)
			if err != nil {
				return err
			}

		default:
			return err
		}
	}

	_, err = s.config.Iam.Client.TagPolicy(ctx, tag)
	if err != nil {
		return err
	}

	return nil
}

// Always ensure VPC role exists to ensure ENI garbage collection works after lambda deletion
func (s *IAM) PutEniRole(ctx context.Context) error {
	var apiErr smithy.APIError

	roleDocument, err := s.config.RoleDocument()
	if err != nil {
		return err
	}

	create := &iam.CreateRoleInput{
		RoleName:                 aws.String(s.config.EniRoleName()),
		AssumeRolePolicyDocument: aws.String(roleDocument),
	}

	update := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String(s.config.EniRoleName()),
		PolicyDocument: aws.String(roleDocument),
	}

	attach := &iam.AttachRolePolicyInput{
		RoleName:  aws.String(s.config.EniRoleName()),
		PolicyArn: aws.String(s.config.EniRolePolicyArn()),
	}

	_, err = s.config.Iam.Client.CreateRole(ctx, create)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "EntityAlreadyExists":
			_, err = s.config.Iam.Client.UpdateAssumeRolePolicy(ctx, update)
			if err != nil {
				return err
			}
		default:
			return err
		}
	}

	_, err = s.config.Iam.Client.AttachRolePolicy(ctx, attach)
	if err != nil {
		return err
	}

	return nil
}

func (s *IAM) PutRole(ctx context.Context) error {
	var apiErr smithy.APIError

	roleDocument, err := s.config.RoleDocument()
	if err != nil {
		return err
	}

	var tagSlice []types.Tag
	for key, value := range s.config.Tags() {
		tagSlice = append(tagSlice, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	create := &iam.CreateRoleInput{
		RoleName:                 aws.String(s.config.ResourceName()),
		AssumeRolePolicyDocument: aws.String(roleDocument),
	}

	if s.config.BoundaryPolicyArn() != "" {
		create.PermissionsBoundary = aws.String(s.config.BoundaryPolicyArn())
	}

	updatePolicy := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       create.RoleName,
		PolicyDocument: create.AssumeRolePolicyDocument,
	}

	tag := &iam.TagRoleInput{
		RoleName: aws.String(s.config.ResourceName()),
		Tags:     tagSlice,
	}

	_, err = s.config.Iam.Client.CreateRole(ctx, create)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "EntityAlreadyExists":
			_, err = s.config.Iam.Client.UpdateAssumeRolePolicy(ctx, updatePolicy)
			if err != nil {
				return err
			}

		default:
			return err
		}
	}

	_, err = s.config.Iam.Client.TagRole(ctx, tag)
	if err != nil {
		return err
	}

	return nil
}

func (s *IAM) AttachRolePolicy(ctx context.Context) error {
	var apiErr smithy.APIError

	attach := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(s.config.PolicyArn()),
		RoleName:  aws.String(s.config.ResourceName()),
	}

	_, err := s.config.Iam.Client.AttachRolePolicy(ctx, attach)
	if err != nil {
		return err
	}

	if s.config.BoundaryPolicyArn() != "" {
		boundary := &iam.PutRolePermissionsBoundaryInput{
			RoleName:            aws.String(s.config.ResourceName()),
			PermissionsBoundary: aws.String(s.config.BoundaryPolicyArn()),
		}

		_, err := s.config.Iam.Client.PutRolePermissionsBoundary(ctx, boundary)
		if err != nil {
			return err
		}
	} else {
		boundary := &iam.DeleteRolePermissionsBoundaryInput{
			RoleName: aws.String(s.config.ResourceName()),
		}

		_, err := s.config.Iam.Client.DeleteRolePermissionsBoundary(ctx, boundary)
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
		PolicyArn: aws.String(s.config.PolicyArn()),
		RoleName:  aws.String(s.config.ResourceName()),
	}

	boundary := &iam.DeleteRolePermissionsBoundaryInput{
		RoleName: detach.RoleName,
	}

	_, err := s.config.Iam.Client.DeleteRolePermissionsBoundary(ctx, boundary)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchEntity":
			// in the case that the boundary does not exist, we can continue
			break
		default:
			return err
		}
	}

	_, err = s.config.Iam.Client.DetachRolePolicy(ctx, detach)
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
		RoleName: aws.String(s.config.ResourceName()),
	}

	_, err := s.config.Iam.Client.DeleteRole(ctx, delete)
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
		PolicyArn: aws.String(s.config.PolicyArn()),
	}

	if err := s.GCPolicyVersions(ctx, s.config.PolicyArn()); err != nil {
		return err
	}

	_, err := s.config.Iam.Client.DeletePolicy(ctx, delete)
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

	policyVersions, err := s.config.Iam.Client.ListPolicyVersions(ctx, index)
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

			if _, err = s.config.Iam.Client.DeletePolicyVersion(ctx, delete); err != nil {
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
