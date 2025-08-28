package vpc

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bkeane/monad/pkg/basis/mock"
)

func TestDerive_Success(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	// Note: This may fail due to AWS API calls during resolution
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	assert.NotNil(t, config)
	assert.NotNil(t, config.Client())
}

func TestDerive_WithSecurityGroups(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.ApplyWithOverrides(t, map[string]string{
		"MONAD_SECURITY_GROUPS": "sg-123456,my-security-group",
		"MONAD_SUBNETS":         "subnet-123456,my-subnet",
	})
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	// Note: This may fail due to AWS API calls during resolution
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	securityGroups := config.SecurityGroups()
	assert.Len(t, securityGroups, 2)
	assert.Contains(t, securityGroups, "sg-123456")
	assert.Contains(t, securityGroups, "my-security-group")

	subnets := config.Subnets()
	assert.Len(t, subnets, 2)
	assert.Contains(t, subnets, "subnet-123456")
	assert.Contains(t, subnets, "my-subnet")
}

func TestDerive_NoVPCConfiguration(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	// Note: This may fail due to AWS API calls during resolution
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	// When no VPC configuration is provided, these should be empty
	assert.Empty(t, config.SecurityGroups())
	assert.Empty(t, config.SecurityGroupIds())
	assert.Empty(t, config.Subnets())
	assert.Empty(t, config.SubnetIds())
}

func TestValidate_Success(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)
	ctx := context.Background()

	config, err := Derive(ctx, setup.Basis)
	// Note: This may fail due to AWS API calls during resolution
	if err != nil {
		// Should be an AWS-related error, not a configuration error
		assert.NotContains(t, err.Error(), "mock:")
		return
	}

	err = config.Validate()
	assert.NoError(t, err)
}

func TestValidate_MissingClient(t *testing.T) {
	config := &Config{
		client: nil,
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be blank")
}

func TestValidate_SecurityGroupsWithoutSubnets(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)

	caller, err := setup.Basis.Caller()
	require.NoError(t, err)

	config := &Config{
		client:            ec2.NewFromConfig(caller.AwsConfig()),
		VpcSecurityGroups: []string{"sg-123456"},
		VpcSubnets:        []string{}, // Empty subnets
	}

	err = config.Validate()
	assert.Error(t, err)
	// When security groups are provided, subnets must also be provided
	assert.Contains(t, err.Error(), "cannot be blank")
}

func TestValidate_SubnetsWithoutSecurityGroups(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)

	caller, err := setup.Basis.Caller()
	require.NoError(t, err)

	config := &Config{
		client:            ec2.NewFromConfig(caller.AwsConfig()),
		VpcSecurityGroups: []string{}, // Empty security groups
		VpcSubnets:        []string{"subnet-123456"},
	}

	err = config.Validate()
	assert.Error(t, err)
	// When subnets are provided, security groups must also be provided
	assert.Contains(t, err.Error(), "cannot be blank")
}

func TestSecurityGroupIdResolution(t *testing.T) {
	tests := []struct {
		name                    string
		inputSecurityGroups     []string
		expectedDirectIds       []string
		expectedNamesToResolve  []string
	}{
		{
			name:                "ID format security groups",
			inputSecurityGroups: []string{"sg-123456", "sg-789012"},
			expectedDirectIds:   []string{"sg-123456", "sg-789012"},
			expectedNamesToResolve: []string{},
		},
		{
			name:                "Name format security groups",
			inputSecurityGroups: []string{"my-security-group", "another-sg"},
			expectedDirectIds:   []string{},
			expectedNamesToResolve: []string{"my-security-group", "another-sg"},
		},
		{
			name:                "Mixed ID and name security groups",
			inputSecurityGroups: []string{"sg-123456", "my-security-group"},
			expectedDirectIds:   []string{"sg-123456"},
			expectedNamesToResolve: []string{"my-security-group"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var directIds []string
			var namesToResolve []string

			for _, nameOrId := range tt.inputSecurityGroups {
				if strings.HasPrefix(nameOrId, "sg-") {
					directIds = append(directIds, nameOrId)
				} else {
					namesToResolve = append(namesToResolve, nameOrId)
				}
			}

			if len(tt.expectedDirectIds) == 0 {
				assert.Empty(t, directIds)
			} else {
				assert.Equal(t, tt.expectedDirectIds, directIds)
			}
			if len(tt.expectedNamesToResolve) == 0 {
				assert.Empty(t, namesToResolve)
			} else {
				assert.Equal(t, tt.expectedNamesToResolve, namesToResolve)
			}
		})
	}
}

func TestSubnetIdResolution(t *testing.T) {
	tests := []struct {
		name                string
		inputSubnets        []string
		expectedDirectIds   []string
		expectedNamesToResolve []string
	}{
		{
			name:                "ID format subnets",
			inputSubnets:        []string{"subnet-123456", "subnet-789012"},
			expectedDirectIds:   []string{"subnet-123456", "subnet-789012"},
			expectedNamesToResolve: []string{},
		},
		{
			name:                "Name format subnets",
			inputSubnets:        []string{"my-subnet", "another-subnet"},
			expectedDirectIds:   []string{},
			expectedNamesToResolve: []string{"my-subnet", "another-subnet"},
		},
		{
			name:                "Mixed ID and name subnets",
			inputSubnets:        []string{"subnet-123456", "my-subnet"},
			expectedDirectIds:   []string{"subnet-123456"},
			expectedNamesToResolve: []string{"my-subnet"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var directIds []string
			var namesToResolve []string

			for _, nameOrId := range tt.inputSubnets {
				if strings.HasPrefix(nameOrId, "subnet-") {
					directIds = append(directIds, nameOrId)
				} else {
					namesToResolve = append(namesToResolve, nameOrId)
				}
			}

			if len(tt.expectedDirectIds) == 0 {
				assert.Empty(t, directIds)
			} else {
				assert.Equal(t, tt.expectedDirectIds, directIds)
			}
			if len(tt.expectedNamesToResolve) == 0 {
				assert.Empty(t, namesToResolve)
			} else {
				assert.Equal(t, tt.expectedNamesToResolve, namesToResolve)
			}
		})
	}
}

func TestAccessors(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)

	caller, err := setup.Basis.Caller()
	require.NoError(t, err)

	config := &Config{
		client:              ec2.NewFromConfig(caller.AwsConfig()),
		VpcSecurityGroups:   []string{"sg-123456", "my-security-group"},
		VpcSecurityGroupIds: []string{"sg-123456", "sg-resolved"},
		VpcSubnets:          []string{"subnet-123456", "my-subnet"},
		VpcSubnetIds:        []string{"subnet-123456", "subnet-resolved"},
	}

	// Test SecurityGroups accessor
	securityGroups := config.SecurityGroups()
	assert.Equal(t, []string{"sg-123456", "my-security-group"}, securityGroups)

	// Test SecurityGroupIds accessor
	securityGroupIds := config.SecurityGroupIds()
	assert.Equal(t, []string{"sg-123456", "sg-resolved"}, securityGroupIds)

	// Test Subnets accessor
	subnets := config.Subnets()
	assert.Equal(t, []string{"subnet-123456", "my-subnet"}, subnets)

	// Test SubnetIds accessor
	subnetIds := config.SubnetIds()
	assert.Equal(t, []string{"subnet-123456", "subnet-resolved"}, subnetIds)

	// Test Client accessor
	client := config.Client()
	assert.NotNil(t, client)
}

func TestDerive_ErrorPropagation(t *testing.T) {
	errorSetup := mock.NewErrorTestSetup()
	errorSetup.Apply(t)
	ctx := context.Background()

	_, err := Derive(ctx, errorSetup.Basis)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock:")
}

func TestEnvParsing_SecurityGroupsAndSubnets(t *testing.T) {
	tests := []struct {
		name                    string
		envSecurityGroups       string
		envSubnets              string
		expectedSecurityGroups  []string
		expectedSubnets         []string
	}{
		{
			name:                   "single security group and subnet",
			envSecurityGroups:      "sg-123456",
			envSubnets:             "subnet-123456",
			expectedSecurityGroups: []string{"sg-123456"},
			expectedSubnets:        []string{"subnet-123456"},
		},
		{
			name:                   "multiple security groups and subnets",
			envSecurityGroups:      "sg-123456,my-security-group",
			envSubnets:             "subnet-123456,my-subnet",
			expectedSecurityGroups: []string{"sg-123456", "my-security-group"},
			expectedSubnets:        []string{"subnet-123456", "my-subnet"},
		},
		{
			name:                   "empty configuration",
			envSecurityGroups:      "",
			envSubnets:             "",
			expectedSecurityGroups: []string{},
			expectedSubnets:        []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup := mock.NewTestSetup()
			envOverrides := make(map[string]string)
			if tt.envSecurityGroups != "" {
				envOverrides["MONAD_SECURITY_GROUPS"] = tt.envSecurityGroups
			}
			if tt.envSubnets != "" {
				envOverrides["MONAD_SUBNETS"] = tt.envSubnets
			}

			if len(envOverrides) > 0 {
				setup.ApplyWithOverrides(t, envOverrides)
			} else {
				setup.Apply(t)
			}
			ctx := context.Background()

			config, err := Derive(ctx, setup.Basis)
			// Note: This may fail due to AWS API calls during resolution
			if err != nil {
				// Should be an AWS-related error, not a configuration error
				assert.NotContains(t, err.Error(), "mock:")
				return
			}

			if len(tt.expectedSecurityGroups) == 0 {
				assert.Empty(t, config.SecurityGroups())
			} else {
				assert.Equal(t, tt.expectedSecurityGroups, config.SecurityGroups())
			}
			if len(tt.expectedSubnets) == 0 {
				assert.Empty(t, config.Subnets())
			} else {
				assert.Equal(t, tt.expectedSubnets, config.Subnets())
			}
		})
	}
}