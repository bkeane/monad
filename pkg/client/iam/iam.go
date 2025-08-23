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

type IamConfig interface {
	PolicyName() string
	PolicyArn() string
	PolicyDocument() string
	RoleName() string
	RoleArn() string
	RoleDocument() string
	EniRoleName() string
	EniRolePolicyArn() string
	BoundaryPolicyName() string
	BoundaryPolicyArn() string
	Client() *iam.Client
	Tags() []types.Tag
}

type Client struct {
	iam IamConfig
}

func Derive(iam IamConfig) *Client {
	return &Client{
		iam: iam,
	}
}

func (c *Client) Mount(ctx context.Context) error {
	log.Info().
		Str("action", "put").
		Str("role", c.iam.EniRoleName()).
		Msg("iam")

	if err := c.PutEniRole(ctx); err != nil {
		return err
	}

	log.Info().
		Str("action", "put").
		Str("policy", c.iam.PolicyName()).
		Msg("iam")

	if err := c.PutPolicy(ctx); err != nil {
		return err
	}

	log.Info().
		Str("action", "put").
		Str("role", c.iam.PolicyName()).
		Msg("iam")

	if err := c.PutRole(ctx); err != nil {
		return err
	}

	log.Info().
		Str("action", "attach").
		Str("role", c.iam.RoleName()).
		Str("policy", c.iam.PolicyName()).
		Str("boundary", c.iam.BoundaryPolicyName()).
		Msg("iam")

	if err := c.AttachRolePolicy(ctx); err != nil {
		return err
	}

	return nil
}

func (c *Client) Unmount(ctx context.Context) error {
	log.Info().
		Str("action", "detach").
		Str("role", c.iam.RoleName()).
		Str("policy", c.iam.PolicyName()).
		Str("boundary", c.iam.BoundaryPolicyName()).
		Msg("iam")

	if err := c.DetachRolePolicy(ctx); err != nil {
		return err
	}

	log.Info().
		Str("action", "delete").
		Str("role", c.iam.RoleName()).
		Msg("iam")

	if err := c.DeleteRole(ctx); err != nil {
		return err
	}

	log.Info().
		Str("action", "delete").
		Str("policy", c.iam.PolicyName()).
		Msg("iam")

	if err := c.DeletePolicy(ctx); err != nil {
		return err
	}

	return nil
}

// PUT OPERATIONS

func (c *Client) PutPolicy(ctx context.Context) error {
	var err error
	var apiErr smithy.APIError

	create := &iam.CreatePolicyInput{
		PolicyName:     aws.String(c.iam.PolicyName()),
		PolicyDocument: aws.String(c.iam.PolicyDocument()),
	}

	update := &iam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(c.iam.PolicyArn()),
		PolicyDocument: aws.String(c.iam.PolicyDocument()),
		SetAsDefault:   true,
	}

	tag := &iam.TagPolicyInput{
		PolicyArn: aws.String(c.iam.PolicyArn()),
		Tags:      c.iam.Tags(),
	}

	_, err = c.iam.Client().CreatePolicy(ctx, create)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "EntityAlreadyExists":
			if err := c.GCPolicyVersions(ctx, c.iam.PolicyArn()); err != nil {
				return err
			}

			_, err = c.iam.Client().CreatePolicyVersion(ctx, update)
			if err != nil {
				return err
			}

		default:
			return err
		}
	}

	_, err = c.iam.Client().TagPolicy(ctx, tag)
	if err != nil {
		return err
	}

	return nil
}

// Always ensure VPC role exists to ensure ENI garbage collection works after lambda deletion
func (c *Client) PutEniRole(ctx context.Context) error {
	var err error
	var apiErr smithy.APIError

	create := &iam.CreateRoleInput{
		RoleName:                 aws.String(c.iam.EniRoleName()),
		AssumeRolePolicyDocument: aws.String(c.iam.RoleDocument()),
	}

	update := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String(c.iam.EniRoleName()),
		PolicyDocument: aws.String(c.iam.RoleDocument()),
	}

	attach := &iam.AttachRolePolicyInput{
		RoleName:  aws.String(c.iam.EniRoleName()),
		PolicyArn: aws.String(c.iam.EniRolePolicyArn()),
	}

	_, err = c.iam.Client().CreateRole(ctx, create)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "EntityAlreadyExists":
			_, err = c.iam.Client().UpdateAssumeRolePolicy(ctx, update)
			if err != nil {
				return err
			}
		default:
			return err
		}
	}

	_, err = c.iam.Client().AttachRolePolicy(ctx, attach)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) PutRole(ctx context.Context) error {
	var err error
	var apiErr smithy.APIError

	create := &iam.CreateRoleInput{
		RoleName:                 aws.String(c.iam.RoleName()),
		AssumeRolePolicyDocument: aws.String(c.iam.RoleDocument()),
	}

	if c.iam.BoundaryPolicyArn() != "" {
		create.PermissionsBoundary = aws.String(c.iam.BoundaryPolicyArn())
	}

	updatePolicy := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       create.RoleName,
		PolicyDocument: create.AssumeRolePolicyDocument,
	}

	tag := &iam.TagRoleInput{
		RoleName: aws.String(c.iam.RoleName()),
		Tags:     c.iam.Tags(),
	}

	_, err = c.iam.Client().CreateRole(ctx, create)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "EntityAlreadyExists":
			_, err = c.iam.Client().UpdateAssumeRolePolicy(ctx, updatePolicy)
			if err != nil {
				return err
			}

		default:
			return err
		}
	}

	_, err = c.iam.Client().TagRole(ctx, tag)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) AttachRolePolicy(ctx context.Context) error {
	var apiErr smithy.APIError

	attach := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(c.iam.PolicyArn()),
		RoleName:  aws.String(c.iam.RoleName()),
	}

	_, err := c.iam.Client().AttachRolePolicy(ctx, attach)
	if err != nil {
		return err
	}

	if c.iam.BoundaryPolicyArn() != "" {
		boundary := &iam.PutRolePermissionsBoundaryInput{
			RoleName:            aws.String(c.iam.RoleName()),
			PermissionsBoundary: aws.String(c.iam.BoundaryPolicyArn()),
		}

		_, err := c.iam.Client().PutRolePermissionsBoundary(ctx, boundary)
		if err != nil {
			return err
		}
	} else {
		boundary := &iam.DeleteRolePermissionsBoundaryInput{
			RoleName: aws.String(c.iam.RoleName()),
		}

		_, err := c.iam.Client().DeleteRolePermissionsBoundary(ctx, boundary)
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

func (c *Client) DetachRolePolicy(ctx context.Context) error {
	var apiErr smithy.APIError

	detach := &iam.DetachRolePolicyInput{
		PolicyArn: aws.String(c.iam.PolicyArn()),
		RoleName:  aws.String(c.iam.RoleName()),
	}

	boundary := &iam.DeleteRolePermissionsBoundaryInput{
		RoleName: detach.RoleName,
	}

	_, err := c.iam.Client().DeleteRolePermissionsBoundary(ctx, boundary)
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchEntity":
			// in the case that the boundary does not exist, we can continue
			break
		default:
			return err
		}
	}

	_, err = c.iam.Client().DetachRolePolicy(ctx, detach)
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

func (c *Client) DeleteRole(ctx context.Context) error {
	var apiErr smithy.APIError

	delete := &iam.DeleteRoleInput{
		RoleName: aws.String(c.iam.RoleName()),
	}

	_, err := c.iam.Client().DeleteRole(ctx, delete)
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

func (c *Client) DeletePolicy(ctx context.Context) error {
	var apiErr smithy.APIError

	delete := &iam.DeletePolicyInput{
		PolicyArn: aws.String(c.iam.PolicyArn()),
	}

	if err := c.GCPolicyVersions(ctx, c.iam.PolicyArn()); err != nil {
		return err
	}

	_, err := c.iam.Client().DeletePolicy(ctx, delete)
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

func (c *Client) GCPolicyVersions(ctx context.Context, policyArn string) error {
	var apiErr smithy.APIError

	index := &iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(policyArn),
	}

	policyVersions, err := c.iam.Client().ListPolicyVersions(ctx, index)
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

			if _, err = c.iam.Client().DeletePolicyVersion(ctx, delete); err != nil {
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
