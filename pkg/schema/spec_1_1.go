package schema

import (
	"path/filepath"

	"github.com/bkeane/monad/internal/git"

	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
)

type Schema_1_1 struct {
	Schema    StringLabel
	Name      StringLabel
	Branch    StringLabel
	Sha       StringLabel
	Origin    StringLabel
	Owner     StringLabel
	Repo      StringLabel
	Dirty     BoolLabel
	Role      FileLabel
	Policy    FileLabel
	Resources FileLabel
	Bus       FolderLabel
}

func (s *Schema_1_1) Encode(git git.Git) error {
	var errs multierror.Error

	// Helper function to reduce repetition
	collect := func(err error) {
		if err != nil {
			errs.Errors = append(errs.Errors, err)
		}
	}

	// Conventions as to where the files are located relative to the path.
	role := filepath.Join(git.Path, "role.json.tmpl")
	policy := filepath.Join(git.Path, "policy.json.tmpl")
	resources := filepath.Join(git.Path, "resources.json.tmpl")
	busRules := filepath.Join(git.Path, "bus")

	log.Debug().Msgf("encoding schema 1.1")
	collect(s.Schema.Encode("1.1"))
	collect(s.Name.Encode(git.BasePath))
	collect(s.Branch.Encode(git.Branch))
	collect(s.Sha.Encode(git.Sha))
	collect(s.Origin.Encode(git.Origin))
	collect(s.Owner.Encode(git.Owner))
	collect(s.Repo.Encode(git.Repo))
	collect(s.Dirty.Encode(git.Dirty))
	collect(s.Role.Encode(role))
	collect(s.Policy.Encode(policy))
	collect(s.Resources.Encode(resources))
	collect(s.Bus.Encode(busRules))

	return errs.ErrorOrNil()
}

func (s *Schema_1_1) Decode(labels map[string]string) error {
	var errs multierror.Error

	collect := func(err error) {
		if err != nil {
			errs.Errors = append(errs.Errors, err)
		}
	}

	collect(s.Schema.Decode(labels))
	collect(s.Name.Decode(labels))
	collect(s.Branch.Decode(labels))
	collect(s.Sha.Decode(labels))
	collect(s.Origin.Decode(labels))
	collect(s.Owner.Decode(labels))
	collect(s.Repo.Decode(labels))
	collect(s.Dirty.Decode(labels))
	collect(s.Role.Decode(labels))
	collect(s.Policy.Decode(labels))
	collect(s.Resources.Decode(labels))
	collect(s.Bus.Decode(labels))

	return errs.ErrorOrNil()
}

func (m *Schema_1_1) EncodedMap() map[string]string {
	mapped := map[string]string{
		m.Schema.Key:    m.Schema.Encoded,
		m.Name.Key:      m.Name.Encoded,
		m.Branch.Key:    m.Branch.Encoded,
		m.Sha.Key:       m.Sha.Encoded,
		m.Origin.Key:    m.Origin.Encoded,
		m.Owner.Key:     m.Owner.Encoded,
		m.Repo.Key:      m.Repo.Encoded,
		m.Dirty.Key:     m.Dirty.Encoded,
		m.Role.Key:      m.Role.Encoded,
		m.Policy.Key:    m.Policy.Encoded,
		m.Resources.Key: m.Resources.Encoded,
	}

	for _, file := range m.Bus.Files {
		mapped[file.Key] = file.Encoded
	}

	return mapped
}

func (m *Schema_1_1) ResourceTags() map[string]string {
	return map[string]string{
		"Origin": m.Origin.Decoded,
		"Branch": m.Branch.Decoded,
		"Sha":    m.Sha.Decoded,
		"Dirty":  strconv.FormatBool(m.Dirty.Decoded),
	}
}

func (m *Schema_1_1) Git() git.Git {
	return git.Git{
		BasePath: m.Name.Decoded,
		Sha:      m.Sha.Decoded,
		Branch:   m.Branch.Decoded,
		Origin:   m.Origin.Decoded,
		Owner:    m.Owner.Decoded,
		Repo:     m.Repo.Decoded,
		Dirty:    m.Dirty.Decoded,
	}
}

func (m *Schema_1_1) ResourceNamePrefix(ownerPrefix bool) string {
	if ownerPrefix {
		return m.Owner.Decoded + "-" + m.Repo.Decoded + "-" + m.Branch.Decoded + "-"
	}

	return m.Repo.Decoded + "-" + m.Branch.Decoded + "-"
}

func (m *Schema_1_1) ResourceName(ownerPrefix bool) string {
	deslashedBranch := strings.ReplaceAll(m.Branch.Decoded, "/", "-")

	if ownerPrefix {
		return m.Owner.Decoded + "-" + m.Repo.Decoded + "-" + deslashedBranch + "-" + m.Name.Decoded
	}

	return m.Repo.Decoded + "-" + deslashedBranch + "-" + m.Name.Decoded
}

func (m *Schema_1_1) ResourcePath(ownerPrefix bool) string {
	if ownerPrefix {
		return m.Owner.Decoded + "/" + m.Repo.Decoded + "/" + m.Branch.Decoded + "/" + m.Name.Decoded
	}

	return m.Repo.Decoded + "/" + m.Branch.Decoded + "/" + m.Name.Decoded
}

func (m *Schema_1_1) ResourcePathPrefix(ownerPrefix bool) string {
	if ownerPrefix {
		return m.Owner.Decoded + "/" + m.Repo.Decoded + "/" + m.Branch.Decoded + "/"
	}

	return m.Repo.Decoded + "/" + m.Branch.Decoded + "/"
}

func (m *Schema_1_1) ImagePath() string {
	return filepath.Join(m.Owner.Decoded, m.Repo.Decoded, m.Name.Decoded)
}

func (m *Schema_1_1) ImageBranchTag(registryUrl string) string {
	return registryUrl + "/" + m.ImagePath() + ":" + m.Branch.Decoded
}

func (m *Schema_1_1) ImageShaTag(registryUrl string) string {
	return registryUrl + "/" + m.ImagePath() + ":" + m.Sha.Decoded
}

func (m *Schema_1_1) ImageTags(registryUrl string) []string {
	return []string{
		m.ImageBranchTag(registryUrl),
		m.ImageShaTag(registryUrl),
	}
}

func (m *Schema_1_1) PolicyDocument(data ...any) (string, error) {
	return m.Policy.Template(data...)
}

func (m *Schema_1_1) RoleDocument(data ...any) (string, error) {
	return m.Role.Template(data...)
}

func (m *Schema_1_1) ResourceDocument(data ...any) (string, error) {
	return m.Resources.Template(data...)
}

func (m *Schema_1_1) EventBridgeDocuments(data ...any) (map[string]string, error) {
	namespacePrefix := m.Bus.Key + "."
	mapped := map[string]string{}

	for _, file := range m.Bus.Files {
		busRule := strings.TrimPrefix(file.Key, namespacePrefix)
		content, err := file.Template(data...)
		if err != nil {
			return nil, err
		}
		mapped[busRule] = content
	}

	return mapped, nil
}
