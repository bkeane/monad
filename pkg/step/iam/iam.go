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

type Resource struct {
	Type string // "role" or "policy"
	Name string
}

type Attachment struct {
	Role     string
	Policy   string
	Boundary string
}

type Summary struct {
	ResourcesCreated  []Resource
	ResourcesDeleted  []Resource
	AttachmentsCreated []Attachment
	AttachmentsDeleted []Attachment
}

type Step struct {
	iam IamConfig
}

func Derive(iam IamConfig) *Step {
	return &Step{
		iam: iam,
	}
}

func (c *Step) Mount(ctx context.Context) error {
	summary, err := c.mount(ctx)
	if err != nil {
		return err
	}

	// Log all resources as action=put for consistency
	for _, resource := range summary.ResourcesCreated {
		log.Info().
			Str("action", "put").
			Str(resource.Type, resource.Name).
			Msg("iam")
	}

	// Log attachments separately
	for _, attachment := range summary.AttachmentsCreated {
		log.Info().
			Str("action", "attach").
			Str("role", attachment.Role).
			Str("policy", attachment.Policy).
			Str("boundary", attachment.Boundary).
			Msg("iam")
	}

	return nil
}

func (c *Step) Unmount(ctx context.Context) error {
	summary, err := c.unmount(ctx)
	if err != nil {
		return err
	}

	// Log detachments first
	for _, attachment := range summary.AttachmentsDeleted {
		log.Info().
			Str("action", "detach").
			Str("role", attachment.Role).
			Str("policy", attachment.Policy).
			Str("boundary", attachment.Boundary).
			Msg("iam")
	}

	// Log deletions
	for _, resource := range summary.ResourcesDeleted {
		log.Info().
			Str("action", "delete").
			Str(resource.Type, resource.Name).
			Msg("iam")
	}

	return nil
}

// Internal methods that return summaries of work done
func (c *Step) mount(ctx context.Context) (Summary, error) {
	var summary Summary

	if err := c.PutEniRole(ctx); err != nil {
		return summary, err
	}
	summary.ResourcesCreated = append(summary.ResourcesCreated, Resource{
		Type: "role",
		Name: c.iam.EniRoleName(),
	})

	if err := c.PutPolicy(ctx); err != nil {
		return summary, err
	}
	summary.ResourcesCreated = append(summary.ResourcesCreated, Resource{
		Type: "policy",
		Name: c.iam.PolicyName(),
	})

	if err := c.PutRole(ctx); err != nil {
		return summary, err
	}
	summary.ResourcesCreated = append(summary.ResourcesCreated, Resource{
		Type: "role",
		Name: c.iam.RoleName(),
	})

	if err := c.AttachRolePolicy(ctx); err != nil {
		return summary, err
	}
	summary.AttachmentsCreated = append(summary.AttachmentsCreated, Attachment{
		Role:     c.iam.RoleName(),
		Policy:   c.iam.PolicyName(),
		Boundary: c.iam.BoundaryPolicyName(),
	})

	return summary, nil
}

func (c *Step) unmount(ctx context.Context) (Summary, error) {
	var summary Summary

	if err := c.DetachRolePolicy(ctx); err != nil {
		return summary, err
	}
	summary.AttachmentsDeleted = append(summary.AttachmentsDeleted, Attachment{
		Role:     c.iam.RoleName(),
		Policy:   c.iam.PolicyName(),
		Boundary: c.iam.BoundaryPolicyName(),
	})

	if err := c.DeleteRole(ctx); err != nil {
		return summary, err
	}
	summary.ResourcesDeleted = append(summary.ResourcesDeleted, Resource{
		Type: "role",
		Name: c.iam.RoleName(),
	})

	if err := c.DeletePolicy(ctx); err != nil {
		return summary, err
	}
	summary.ResourcesDeleted = append(summary.ResourcesDeleted, Resource{
		Type: "policy",
		Name: c.iam.PolicyName(),
	})

	return summary, nil
}

// PUT OPERATIONS

func (c *Step) PutPolicy(ctx context.Context) error {
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
func (c *Step) PutEniRole(ctx context.Context) error {
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

func (c *Step) PutRole(ctx context.Context) error {
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

func (c *Step) AttachRolePolicy(ctx context.Context) error {
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

func (c *Step) DetachRolePolicy(ctx context.Context) error {
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

func (c *Step) DeleteRole(ctx context.Context) error {
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

func (c *Step) DeletePolicy(ctx context.Context) error {
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

func (c *Step) GCPolicyVersions(ctx context.Context, policyArn string) error {
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
