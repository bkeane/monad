package saga

import (
	"context"
	"errors"

	"github.com/bkeane/monad/pkg/config/release"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog"
)

type IAM struct {
	release release.Config
	iam     *iam.Client
	log     *zerolog.Logger
}

func (s IAM) Init(ctx context.Context, r release.Config) *IAM {
	return &IAM{
		release: r,
		iam:     iam.NewFromConfig(r.AwsConfig),
		log:     zerolog.Ctx(ctx),
	}
}

func (s *IAM) Do(ctx context.Context) error {
	s.log.Info().Msg("ensuring eni role exists")
	if err := s.PutEniRole(ctx); err != nil {
		return err
	}

	s.log.Info().Msg("ensuring policy exists")
	if err := s.PutPolicy(ctx); err != nil {
		return err
	}

	s.log.Info().Msg("ensuring role exists")
	if err := s.PutRole(ctx); err != nil {
		return err
	}

	s.log.Info().Msg("ensuring role policy attachment")
	if err := s.AttachRolePolicy(ctx); err != nil {
		return err
	}

	return nil
}

func (s *IAM) Undo(ctx context.Context) error {
	s.log.Info().Msg("deleting role policy attachment")
	if err := s.DetachRolePolicy(ctx); err != nil {
		return err
	}

	s.log.Info().Msg("deleting role")
	if err := s.DeleteRole(ctx); err != nil {
		return err
	}

	s.log.Info().Msg("deleting policy")
	if err := s.DeletePolicy(ctx); err != nil {
		return err
	}

	return nil
}

// PUT OPERATIONS

func (s *IAM) PutPolicy(ctx context.Context) error {
	var apiErr smithy.APIError

	policyDocument, err := s.release.PolicyDocument()
	if err != nil {
		return err
	}

	var tagSlice []types.Tag
	for key, value := range s.release.ResourceTags() {
		tagSlice = append(tagSlice, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	create := &iam.CreatePolicyInput{
		PolicyName:     aws.String(s.release.ResourceName()),
		PolicyDocument: aws.String(policyDocument),
	}

	update := &iam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(s.release.PolicyArn()),
		PolicyDocument: aws.String(policyDocument),
		SetAsDefault:   true,
	}

	tag := &iam.TagPolicyInput{
		PolicyArn: aws.String(s.release.PolicyArn()),
		Tags:      tagSlice,
	}

	_, err = s.iam.CreatePolicy(ctx, create)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "EntityAlreadyExists":
			if err := s.GCPolicyVersions(ctx, s.release.PolicyArn()); err != nil {
				return err
			}

			_, err = s.iam.CreatePolicyVersion(ctx, update)
			if err != nil {
				return err
			}

		default:
			return err
		}
	}

	_, err = s.iam.TagPolicy(ctx, tag)
	if err != nil {
		return err
	}

	return nil
}

// Always ensure VPC role exists to ensure ENI garbage collection works after lambda deletion
func (s *IAM) PutEniRole(ctx context.Context) error {
	var apiErr smithy.APIError

	roleDocument, err := s.release.RoleDocument()
	if err != nil {
		return err
	}

	create := &iam.CreateRoleInput{
		RoleName:                 aws.String(s.release.EniRoleName()),
		AssumeRolePolicyDocument: aws.String(roleDocument),
	}

	update := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String(s.release.EniRoleName()),
		PolicyDocument: aws.String(roleDocument),
	}

	attach := &iam.AttachRolePolicyInput{
		RoleName:  aws.String(s.release.EniRoleName()),
		PolicyArn: aws.String(s.release.EniRolePolicyArn()),
	}

	_, err = s.iam.CreateRole(ctx, create)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "EntityAlreadyExists":
			_, err = s.iam.UpdateAssumeRolePolicy(ctx, update)
			if err != nil {
				return err
			}
		default:
			return err
		}
	}

	_, err = s.iam.AttachRolePolicy(ctx, attach)
	if err != nil {
		return err
	}

	return nil
}

func (s *IAM) PutRole(ctx context.Context) error {
	var apiErr smithy.APIError

	roleDocument, err := s.release.RoleDocument()
	if err != nil {
		return err
	}

	var tagSlice []types.Tag
	for key, value := range s.release.ResourceTags() {
		tagSlice = append(tagSlice, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	create := &iam.CreateRoleInput{
		RoleName:                 aws.String(s.release.RoleName()),
		AssumeRolePolicyDocument: aws.String(roleDocument),
	}

	update := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String(s.release.RoleName()),
		PolicyDocument: aws.String(roleDocument),
	}

	tag := &iam.TagRoleInput{
		RoleName: aws.String(s.release.RoleName()),
		Tags:     tagSlice,
	}

	_, err = s.iam.CreateRole(ctx, create)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "EntityAlreadyExists":
			_, err = s.iam.UpdateAssumeRolePolicy(ctx, update)
			if err != nil {
				return err
			}
		default:
			return err
		}
	}

	_, err = s.iam.TagRole(ctx, tag)
	if err != nil {
		return err
	}

	return nil
}

func (s *IAM) AttachRolePolicy(ctx context.Context) error {
	attach := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(s.release.PolicyArn()),
		RoleName:  aws.String(s.release.RoleName()),
	}

	_, err := s.iam.AttachRolePolicy(ctx, attach)
	if err != nil {
		return err
	}

	return nil
}

// DELETE OPERATIONS

func (s *IAM) DetachRolePolicy(ctx context.Context) error {
	var apiErr smithy.APIError

	detach := &iam.DetachRolePolicyInput{
		PolicyArn: aws.String(s.release.PolicyArn()),
		RoleName:  aws.String(s.release.RoleName()),
	}

	_, err := s.iam.DetachRolePolicy(ctx, detach)
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
		RoleName: aws.String(s.release.RoleName()),
	}

	_, err := s.iam.DeleteRole(ctx, delete)
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
		PolicyArn: aws.String(s.release.PolicyArn()),
	}

	if err := s.GCPolicyVersions(ctx, s.release.PolicyArn()); err != nil {
		return err
	}

	_, err := s.iam.DeletePolicy(ctx, delete)
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

	policyVersions, err := s.iam.ListPolicyVersions(ctx, index)
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

			if _, err = s.iam.DeletePolicyVersion(ctx, delete); err != nil {
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
