package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDerive_WithEnvironmentVariable(t *testing.T) {
	// Set environment variable
	t.Setenv("MONAD_SERVICE", "my-api-service")

	basis, err := Derive()
	require.NoError(t, err)
	assert.NotNil(t, basis)

	// Should use env var value
	assert.Equal(t, "my-api-service", basis.Name())
}

func TestDerive_WithoutEnvironmentVariable(t *testing.T) {
	// Clear environment variable
	t.Setenv("MONAD_SERVICE", "")

	// Create temp directory and change to it
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	basis, err := Derive()
	require.NoError(t, err)
	assert.NotNil(t, basis)

	// Should use directory basename
	expectedName := filepath.Base(tmpDir)
	assert.Equal(t, expectedName, basis.Name())
}

func TestDerive_WithSpecificDirectoryName(t *testing.T) {
	t.Setenv("MONAD_SERVICE", "")

	// Create temp directory with specific name
	parentDir := t.TempDir()
	serviceDir := filepath.Join(parentDir, "user-authentication-service")
	err := os.Mkdir(serviceDir, 0755)
	require.NoError(t, err)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	
	err = os.Chdir(serviceDir)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	basis, err := Derive()
	require.NoError(t, err)

	assert.Equal(t, "user-authentication-service", basis.Name())
}

func TestDerive_EnvironmentVariablePrecedence(t *testing.T) {
	// Set environment variable to something different than directory
	t.Setenv("MONAD_SERVICE", "env-service")

	// Create temp directory with different name
	parentDir := t.TempDir()
	serviceDir := filepath.Join(parentDir, "directory-service")
	err := os.Mkdir(serviceDir, 0755)
	require.NoError(t, err)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	
	err = os.Chdir(serviceDir)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	basis, err := Derive()
	require.NoError(t, err)

	// Should use env var, not directory name
	assert.Equal(t, "env-service", basis.Name())
}

func TestDerive_ValidationSuccess(t *testing.T) {
	t.Setenv("MONAD_SERVICE", "valid-service")

	basis, err := Derive()
	require.NoError(t, err)
	assert.NotNil(t, basis)

	// Validation should pass
	assert.NoError(t, basis.Validate())
}

func TestDerive_DirectoryNamesVariations(t *testing.T) {
	tests := []struct {
		name        string
		dirName     string
		expectedErr bool
	}{
		{
			name:        "standard directory name",
			dirName:     "api",
			expectedErr: false,
		},
		{
			name:        "directory with hyphens",
			dirName:     "user-service",
			expectedErr: false,
		},
		{
			name:        "directory with underscores",
			dirName:     "payment_processor",
			expectedErr: false,
		},
		{
			name:        "directory with numbers",
			dirName:     "service123",
			expectedErr: false,
		},
		{
			name:        "directory with dots",
			dirName:     "auth.service",
			expectedErr: false,
		},
		{
			name:        "single character",
			dirName:     "a",
			expectedErr: false,
		},
		{
			name:        "long name",
			dirName:     "very-long-service-name-with-many-parts",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MONAD_SERVICE", "")

			// Create temp directory with specific name
			parentDir := t.TempDir()
			serviceDir := filepath.Join(parentDir, tt.dirName)
			err := os.Mkdir(serviceDir, 0755)
			require.NoError(t, err)

			originalDir, err := os.Getwd()
			require.NoError(t, err)
			
			err = os.Chdir(serviceDir)
			require.NoError(t, err)
			defer func() {
				_ = os.Chdir(originalDir)
			}()

			basis, err := Derive()
			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, basis)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.dirName, basis.Name())
			}
		})
	}
}

func TestDerive_RootDirectory(t *testing.T) {
	t.Setenv("MONAD_SERVICE", "")

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	
	// Change to root directory
	err = os.Chdir("/")
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	basis, err := Derive()
	require.NoError(t, err)

	// Root directory basename should be "/"
	assert.Equal(t, "/", basis.Name())
}

func TestDerive_HomeDirectory(t *testing.T) {
	t.Setenv("MONAD_SERVICE", "")

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	
	err = os.Chdir(homeDir)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	basis, err := Derive()
	require.NoError(t, err)

	expectedName := filepath.Base(homeDir)
	assert.Equal(t, expectedName, basis.Name())
	assert.NotEmpty(t, basis.Name())
}

func TestBasis_Accessor(t *testing.T) {
	basis := &Basis{
		ServiceName: "test-service",
	}

	assert.Equal(t, "test-service", basis.Name())
}

func TestBasis_Validate(t *testing.T) {
	tests := []struct {
		name       string
		basis      *Basis
		wantErr    bool
	}{
		{
			name: "valid basis",
			basis: &Basis{
				ServiceName: "valid-service",
			},
			wantErr: false,
		},
		{
			name: "empty service name",
			basis: &Basis{
				ServiceName: "",
			},
			wantErr: true,
		},
		{
			name: "whitespace only service name",
			basis: &Basis{
				ServiceName: "   ",
			},
			wantErr: false, // ozzo validation considers whitespace as "present"
		},
		{
			name: "service name with special characters",
			basis: &Basis{
				ServiceName: "my-service_v2.0",
			},
			wantErr: false,
		},
		{
			name: "very long service name",
			basis: &Basis{
				ServiceName: "very-long-service-name-that-might-exceed-normal-limits-but-should-still-be-valid",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.basis.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDerive_EmptyEnvironmentVariableUsesDirectory(t *testing.T) {
	// Explicitly set empty string (vs unset)
	t.Setenv("MONAD_SERVICE", "")

	// Create temp directory with specific name
	parentDir := t.TempDir()
	serviceDir := filepath.Join(parentDir, "fallback-service")
	err := os.Mkdir(serviceDir, 0755)
	require.NoError(t, err)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	
	err = os.Chdir(serviceDir)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	basis, err := Derive()
	require.NoError(t, err)

	// Should fall back to directory name when env var is empty
	assert.Equal(t, "fallback-service", basis.Name())
}

func TestDerive_ValidatesAfterDerivation(t *testing.T) {
	// This test ensures that validation is called during Derive()
	// We can't easily create a scenario where computed values fail validation
	// because directory names are always non-empty strings
	// But we can verify the validation flow works
	
	t.Setenv("MONAD_SERVICE", "test-service")

	basis, err := Derive()
	require.NoError(t, err)
	assert.NotNil(t, basis)

	// Manual validation should also pass
	assert.NoError(t, basis.Validate())
}

func TestDerive_PathVariations(t *testing.T) {
	tests := []struct {
		name         string
		pathSegments []string
		expectedName string
	}{
		{
			name:         "nested service directory",
			pathSegments: []string{"services", "microservices", "auth"},
			expectedName: "auth",
		},
		{
			name:         "project structure",
			pathSegments: []string{"project", "backend", "api", "users"},
			expectedName: "users",
		},
		{
			name:         "single level",
			pathSegments: []string{"simple"},
			expectedName: "simple",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MONAD_SERVICE", "")

			// Build nested directory structure
			baseDir := t.TempDir()
			currentPath := baseDir
			
			for _, segment := range tt.pathSegments {
				currentPath = filepath.Join(currentPath, segment)
				err := os.Mkdir(currentPath, 0755)
				require.NoError(t, err)
			}

			originalDir, err := os.Getwd()
			require.NoError(t, err)
			
			err = os.Chdir(currentPath)
			require.NoError(t, err)
			defer func() {
				_ = os.Chdir(originalDir)
			}()

			basis, err := Derive()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedName, basis.Name())
		})
	}
}

func TestDerive_ConsistentOutput(t *testing.T) {
	// Test that multiple calls with same conditions produce identical results
	t.Setenv("MONAD_SERVICE", "consistent-service")

	basis1, err1 := Derive()
	require.NoError(t, err1)

	basis2, err2 := Derive()
	require.NoError(t, err2)

	// Should be identical
	assert.Equal(t, basis1.Name(), basis2.Name())
}

func TestDerive_UnicodeDirectoryNames(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping unicode test in short mode")
	}

	tests := []struct {
		name    string
		dirName string
	}{
		{
			name:    "chinese characters",
			dirName: "服务",
		},
		{
			name:    "russian characters", 
			dirName: "сервис",
		},
		{
			name:    "emoji",
			dirName: "api🚀service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MONAD_SERVICE", "")

			// Create temp directory with unicode name
			parentDir := t.TempDir()
			serviceDir := filepath.Join(parentDir, tt.dirName)
			err := os.Mkdir(serviceDir, 0755)
			require.NoError(t, err)

			originalDir, err := os.Getwd()
			require.NoError(t, err)
			
			err = os.Chdir(serviceDir)
			require.NoError(t, err)
			defer func() {
				_ = os.Chdir(originalDir)
			}()

			basis, err := Derive()
			require.NoError(t, err)
			assert.Equal(t, tt.dirName, basis.Name())
		})
	}
}