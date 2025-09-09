package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bkeane/monad/pkg/basis/mock"
)

func TestDerive_Success(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)

	scaffold, err := Derive(setup.Basis)
	require.NoError(t, err)
	assert.NotNil(t, scaffold)
	assert.NotNil(t, scaffold.defaults)
}

func TestDerive_ErrorPropagation(t *testing.T) {
	errorSetup := mock.NewErrorTestSetup()
	errorSetup.Apply(t)

	_, err := Derive(errorSetup.Basis)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock:")
}

func TestList_Success(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)

	scaffold, err := Derive(setup.Basis)
	require.NoError(t, err)

	languages, err := scaffold.List()
	require.NoError(t, err)
	assert.NotEmpty(t, languages)

	// Check that we have expected languages
	expectedLanguages := []string{"go", "node", "python", "ruby", "shell"}
	for _, expected := range expectedLanguages {
		assert.Contains(t, languages, expected, "Should contain %s language", expected)
	}

	// Verify all returned languages are directories
	for _, lang := range languages {
		assert.True(t, strings.TrimSpace(lang) != "", "Language name should not be empty")
		assert.False(t, strings.Contains(lang, "/"), "Language name should not contain path separators")
	}
}

func TestList_EmptyResult(t *testing.T) {
	// This test verifies the List method handles the embedded filesystem correctly
	setup := mock.NewTestSetup()
	setup.Apply(t)

	scaffold, err := Derive(setup.Basis)
	require.NoError(t, err)

	languages, err := scaffold.List()
	require.NoError(t, err)
	
	// We expect to have languages since templates are embedded
	assert.NotEmpty(t, languages)
	assert.Greater(t, len(languages), 0)
}

func TestCreate_ValidLanguage(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)

	scaffold, err := Derive(setup.Basis)
	require.NoError(t, err)

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test-scaffold")

	// Test creating a Go scaffold
	err = scaffold.Create("go", testDir)
	require.NoError(t, err)

	// Verify files were created
	entries, err := os.ReadDir(testDir)
	require.NoError(t, err)
	assert.NotEmpty(t, entries)

	// Check for expected files (should have at least Dockerfile and main.go)
	var hasDockerfile, hasMainGo bool
	for _, entry := range entries {
		if entry.Name() == "Dockerfile" {
			hasDockerfile = true
		}
		if entry.Name() == "main.go" {
			hasMainGo = true
		}
	}

	assert.True(t, hasDockerfile, "Should create Dockerfile")
	assert.True(t, hasMainGo, "Should create main.go")
}

func TestCreate_InvalidLanguage(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)

	scaffold, err := Derive(setup.Basis)
	require.NoError(t, err)

	tempDir := t.TempDir()

	// Test with invalid language
	err = scaffold.Create("invalid-language", tempDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid language type 'invalid-language'")
}

func TestCreate_EmptyLanguage(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)

	scaffold, err := Derive(setup.Basis)
	require.NoError(t, err)

	tempDir := t.TempDir()

	// Test with empty language
	err = scaffold.Create("", tempDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid language type ''")
}

func TestCreate_DefaultTargetDirectory(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)

	scaffold, err := Derive(setup.Basis)
	require.NoError(t, err)

	// Save current directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Change to temp directory
	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Test with empty target directory (should default to ".")
	err = scaffold.Create("shell", "")
	require.NoError(t, err)

	// Verify files were created in current directory
	entries, err := os.ReadDir(".")
	require.NoError(t, err)
	assert.NotEmpty(t, entries)

	// Should have main.sh for shell template
	var hasMainSh bool
	for _, entry := range entries {
		if entry.Name() == "main.sh" {
			hasMainSh = true
		}
	}
	assert.True(t, hasMainSh, "Should create main.sh in current directory")
}

func TestCreate_AllLanguages(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)

	scaffold, err := Derive(setup.Basis)
	require.NoError(t, err)

	languages, err := scaffold.List()
	require.NoError(t, err)

	// Test creating scaffold for each available language
	for _, language := range languages {
		t.Run(language, func(t *testing.T) {
			tempDir := t.TempDir()
			testDir := filepath.Join(tempDir, "test-"+language)

			err := scaffold.Create(language, testDir)
			require.NoError(t, err, "Should be able to create %s scaffold", language)

			// Verify directory was created and has files
			entries, err := os.ReadDir(testDir)
			require.NoError(t, err)
			assert.NotEmpty(t, entries, "Should create files for %s", language)

			// All languages should have a Dockerfile
			var hasDockerfile bool
			for _, entry := range entries {
				if entry.Name() == "Dockerfile" {
					hasDockerfile = true
				}
			}
			assert.True(t, hasDockerfile, "Should create Dockerfile for %s", language)
		})
	}
}

func TestCreate_ExistingFiles(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)

	scaffold, err := Derive(setup.Basis)
	require.NoError(t, err)

	tempDir := t.TempDir()

	// Create an existing file
	existingFile := filepath.Join(tempDir, "main.go")
	err = os.WriteFile(existingFile, []byte("// existing content"), 0644)
	require.NoError(t, err)

	// Create scaffold - should skip existing files
	err = scaffold.Create("go", tempDir)
	require.NoError(t, err)

	// Verify existing file content wasn't overwritten
	content, err := os.ReadFile(existingFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "// existing content")
}

func TestCreate_WithTemplateOptions(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)

	tempDir := t.TempDir()

	// Test with template writing options enabled
	scaffold := &Scaffold{
		WritePolicy: true,
		WriteRole:   true,
		WriteRule:   true,
		WriteEnv:    true,
	}

	defaults, err := setup.Basis.Defaults()
	require.NoError(t, err)
	scaffold.defaults = defaults

	err = scaffold.Create("go", tempDir)
	require.NoError(t, err)

	// Verify template files were created
	expectedFiles := []string{"policy.json.tmpl", "role.json.tmpl", "rule.json.tmpl", ".env.tmpl"}
	for _, filename := range expectedFiles {
		filePath := filepath.Join(tempDir, filename)
		_, err := os.Stat(filePath)
		assert.NoError(t, err, "Should create %s", filename)

		// Verify file has content
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.NotEmpty(t, content, "%s should not be empty", filename)
	}
}

func TestLanguageValidation(t *testing.T) {
	setup := mock.NewTestSetup()
	setup.Apply(t)

	scaffold, err := Derive(setup.Basis)
	require.NoError(t, err)

	languages, err := scaffold.List()
	require.NoError(t, err)

	// Test that all listed languages can be used to create scaffolds
	for _, language := range languages {
		t.Run("validate_"+language, func(t *testing.T) {
			tempDir := t.TempDir()
			
			// This should not return an error since the language is in the list
			err := scaffold.Create(language, tempDir)
			assert.NoError(t, err, "Language %s from List() should be valid for Create()", language)
		})
	}
}

func TestTemplate_DirectoryStructure(t *testing.T) {
	// Test that the embedded templates have the expected structure
	entries, err := Templates.ReadDir("templates")
	require.NoError(t, err)
	assert.NotEmpty(t, entries)

	// Each entry should be a directory
	for _, entry := range entries {
		assert.True(t, entry.IsDir(), "Template entry %s should be a directory", entry.Name())
		
		// Each language directory should have files
		langPath := filepath.Join("templates", entry.Name())
		langFiles, err := Templates.ReadDir(langPath)
		require.NoError(t, err, "Should be able to read %s directory", entry.Name())
		assert.NotEmpty(t, langFiles, "Language %s should have template files", entry.Name())
	}
}

func TestScaffold_EnvironmentVariables(t *testing.T) {
	// Test that environment variables are properly parsed
	tests := []struct {
		name        string
		envVars     map[string]string
		expectPolicy bool
		expectRole   bool
		expectRule   bool
		expectEnv    bool
	}{
		{
			name:         "no env vars",
			envVars:      map[string]string{},
			expectPolicy: false,
			expectRole:   false,
			expectRule:   false,
			expectEnv:    false,
		},
		{
			name: "policy enabled",
			envVars: map[string]string{
				"MONAD_SCAFFOLD_POLICY": "true",
			},
			expectPolicy: true,
			expectRole:   false,
			expectRule:   false,
			expectEnv:    false,
		},
		{
			name: "all enabled",
			envVars: map[string]string{
				"MONAD_SCAFFOLD_POLICY": "true",
				"MONAD_SCAFFOLD_ROLE":   "true",
				"MONAD_SCAFFOLD_RULE":   "true",
				"MONAD_SCAFFOLD_ENV":    "true",
			},
			expectPolicy: true,
			expectRole:   true,
			expectRule:   true,
			expectEnv:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			setup := mock.NewTestSetup()
			setup.Apply(t)

			// Create scaffold and check if env vars are parsed correctly
			scaffold, err := Derive(setup.Basis)
			require.NoError(t, err)

			assert.Equal(t, tt.expectPolicy, scaffold.WritePolicy)
			assert.Equal(t, tt.expectRole, scaffold.WriteRole)
			assert.Equal(t, tt.expectRule, scaffold.WriteRule)
			assert.Equal(t, tt.expectEnv, scaffold.WriteEnv)
		})
	}
}