package iam

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog/log"
)

type IAMResources interface {
	PolicyName() string
	PolicyArn() string
	PolicyDocument() (string, error)
	RoleName() string
	RoleArn() string
	RoleDocument() (string, error)
	EniRoleName() string
	EniRolePolicyArn() string
	// EniRoleDocument string // why does the pattern diverge here?
	BoundaryPolicyName() string
	BoundaryPolicyArn() string
	// BoundaryPolicyDocument() string // why does the pattern diverge here?
	Client() *iam.Client
	Tags() []types.Tag
}

type Client struct {
	iam IAMResources
}

func Init(iam IAMResources) *Client {
	return &Client{
		iam: iam,
	}
}

func (s *Client) Mount(ctx context.Context) error {
	log.Info().
		Str("action", "put").
		Str("role", s.iam.EniRoleName()).
		Msg("iam")

	if err := s.PutEniRole(ctx); err != nil {
		return err
	}

	log.Info().
		Str("action", "put").
		Str("policy", s.iam.PolicyName()).
		Msg("iam")

	if err := s.PutPolicy(ctx); err != nil {
		return err
	}

	log.Info().
		Str("action", "put").
		Str("role", s.iam.PolicyName()).
		Msg("iam")

	if err := s.PutRole(ctx); err != nil {
		return err
	}

	log.Info().
		Str("action", "attach").
		Str("role", s.iam.RoleName()).
		Str("policy", s.iam.PolicyName()).
		Str("boundary", s.iam.BoundaryPolicyName()).
		Msg("iam")

	if err := s.AttachRolePolicy(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Client) Unmount(ctx context.Context) error {
	log.Info().
		Str("action", "detach").
		Str("role", s.iam.RoleName()).
		Str("policy", s.iam.PolicyName()).
		Str("boundary", s.iam.BoundaryPolicyName()).
		Msg("iam")

	if err := s.DetachRolePolicy(ctx); err != nil {
		return err
	}

	log.Info().
		Str("action", "delete").
		Str("role", s.iam.RoleName()).
		Msg("iam")

	if err := s.DeleteRole(ctx); err != nil {
		return err
	}

	log.Info().
		Str("action", "delete").
		Str("policy", s.iam.PolicyName()).
		Msg("iam")

	if err := s.DeletePolicy(ctx); err != nil {
		return err
	}

	return nil
}

// PUT OPERATIONS

func (s *Client) PutPolicy(ctx context.Context) error {
	var apiErr smithy.APIError

	policyDocument, err := s.iam.PolicyDocument()
	if err != nil {
		return err
	}

	create := &iam.CreatePolicyInput{
		PolicyName:     aws.String(s.iam.PolicyName()),
		PolicyDocument: aws.String(policyDocument),
	}

	update := &iam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(s.iam.PolicyArn()),
		PolicyDocument: aws.String(policyDocument),
		SetAsDefault:   true,
	}

	tag := &iam.TagPolicyInput{
		PolicyArn: aws.String(s.iam.PolicyArn()),
		Tags:      s.iam.Tags(),
	}

	_, err = s.iam.Client().CreatePolicy(ctx, create)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "EntityAlreadyExists":
			if err := s.GCPolicyVersions(ctx, s.iam.PolicyArn()); err != nil {
				return err
			}

			_, err = s.iam.Client().CreatePolicyVersion(ctx, update)
			if err != nil {
				return err
			}

		default:
			return err
		}
	}

	_, err = s.iam.Client().TagPolicy(ctx, tag)
	if err != nil {
		return err
	}

	return nil
}

// Always ensure VPC role exists to ensure ENI garbage collection works after lambda deletion
func (s *Client) PutEniRole(ctx context.Context) error {
	var apiErr smithy.APIError

	roleDocument, err := s.iam.RoleDocument()
	if err != nil {
		return err
	}

	create := &iam.CreateRoleInput{
		RoleName:                 aws.String(s.iam.EniRoleName()),
		AssumeRolePolicyDocument: aws.String(roleDocument),
	}

	update := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String(s.iam.EniRoleName()),
		PolicyDocument: aws.String(roleDocument),
	}

	attach := &iam.AttachRolePolicyInput{
		RoleName:  aws.String(s.iam.EniRoleName()),
		PolicyArn: aws.String(s.iam.EniRolePolicyArn()),
	}

	_, err = s.iam.Client().CreateRole(ctx, create)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "EntityAlreadyExists":
			_, err = s.iam.Client().UpdateAssumeRolePolicy(ctx, update)
			if err != nil {
				return err
			}
		default:
			return err
		}
	}

	_, err = s.iam.Client().AttachRolePolicy(ctx, attach)
	if err != nil {
		return err
	}

	return nil
}

func (s *Client) PutRole(ctx context.Context) error {
	var apiErr smithy.APIError

	roleDocument, err := s.iam.RoleDocument()
	if err != nil {
		return err
	}

	create := &iam.CreateRoleInput{
		RoleName:                 aws.String(s.iam.RoleName()),
		AssumeRolePolicyDocument: aws.String(roleDocument),
	}

	if s.iam.BoundaryPolicyArn() != "" {
		create.PermissionsBoundary = aws.String(s.iam.BoundaryPolicyArn())
	}

	updatePolicy := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       create.RoleName,
		PolicyDocument: create.AssumeRolePolicyDocument,
	}

	tag := &iam.TagRoleInput{
		RoleName: aws.String(s.iam.RoleName()),
		Tags:     s.iam.Tags(),
	}

	_, err = s.iam.Client().CreateRole(ctx, create)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "EntityAlreadyExists":
			_, err = s.iam.Client().UpdateAssumeRolePolicy(ctx, updatePolicy)
			if err != nil {
				return err
			}

		default:
			return err
		}
	}

	_, err = s.iam.Client().TagRole(ctx, tag)
	if err != nil {
		return err
	}

	return nil
}

func (s *Client) AttachRolePolicy(ctx context.Context) error {
	var apiErr smithy.APIError

	attach := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(s.iam.PolicyArn()),
		RoleName:  aws.String(s.iam.RoleName()),
	}

	_, err := s.iam.Client().AttachRolePolicy(ctx, attach)
	if err != nil {
		return err
	}

	if s.iam.BoundaryPolicyArn() != "" {
		boundary := &iam.PutRolePermissionsBoundaryInput{
			RoleName:            aws.String(s.iam.RoleName()),
			PermissionsBoundary: aws.String(s.iam.BoundaryPolicyArn()),
		}

		_, err := s.iam.Client().PutRolePermissionsBoundary(ctx, boundary)
		if err != nil {
			return err
		}
	} else {
		boundary := &iam.DeleteRolePermissionsBoundaryInput{
			RoleName: aws.String(s.iam.RoleName()),
		}

		_, err := s.iam.Client().DeleteRolePermissionsBoundary(ctx, boundary)
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

func (s *Client) DetachRolePolicy(ctx context.Context) error {
	var apiErr smithy.APIError

	detach := &iam.DetachRolePolicyInput{
		PolicyArn: aws.String(s.iam.PolicyArn()),
		RoleName:  aws.String(s.iam.RoleName()),
	}

	boundary := &iam.DeleteRolePermissionsBoundaryInput{
		RoleName: detach.RoleName,
	}

	_, err := s.iam.Client().DeleteRolePermissionsBoundary(ctx, boundary)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchEntity":
			// in the case that the boundary does not exist, we can continue
			break
		default:
			return err
		}
	}

	_, err = s.iam.Client().DetachRolePolicy(ctx, detach)
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

func (s *Client) DeleteRole(ctx context.Context) error {
	var apiErr smithy.APIError

	delete := &iam.DeleteRoleInput{
		RoleName: aws.String(s.iam.RoleName()),
	}

	_, err := s.iam.Client().DeleteRole(ctx, delete)
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

func (s *Client) DeletePolicy(ctx context.Context) error {
	var apiErr smithy.APIError

	delete := &iam.DeletePolicyInput{
		PolicyArn: aws.String(s.iam.PolicyArn()),
	}

	if err := s.GCPolicyVersions(ctx, s.iam.PolicyArn()); err != nil {
		return err
	}

	_, err := s.iam.Client().DeletePolicy(ctx, delete)
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

func (s *Client) GCPolicyVersions(ctx context.Context, policyArn string) error {
	var apiErr smithy.APIError

	index := &iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(policyArn),
	}

	policyVersions, err := s.iam.Client().ListPolicyVersions(ctx, index)
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

			if _, err = s.iam.Client().DeletePolicyVersion(ctx, delete); err != nil {
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
