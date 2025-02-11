package schema

import (
	"path/filepath"

	"github.com/bkeane/monad/internal/git"
)

type Spec interface {
	Encode(git git.Git) error
	Decode(labels map[string]string) error
	EncodedMap() map[string]string
	ResourceName(ownerPrefix bool) string
	ResourceNamePrefix(ownerPrefix bool) string
	ResourcePath(ownerPrefix bool) string
	ResourcePathPrefix(ownerPrefix bool) string
	ResourceTags() map[string]string
	ImagePath() string
	ImageBranchTag(registryUrl string) string
	ImageShaTag(registryUrl string) string
	ImageTags(registryUrl string) []string
	PolicyDocument(templateData ...any) (string, error)
	RoleDocument(templateData ...any) (string, error)
	EventBridgeDocuments(templateData ...any) (map[string]string, error)
	ResourceDocument(templateData ...any) (string, error)
	Git() git.Git
}

// This methodology is not complete, but it has the ingredients to implement...
// map[major][minor]Spec_{major}, where major decides interface and minor decides implementation.
// Generics will probably be needed to implement the semi-dynamic return schema types.

func init() {
	Version["latest"] = Version["1.1"]
}

var Version = map[string]Spec{
	"1.1": &Schema_1_1{
		Schema: StringLabel{
			Description: "Spec schema version string",
			Key:         "org.kaixo.monad.schema",
			Required:    true,
		},
		Name: StringLabel{
			Description: "Function name string",
			Key:         "org.kaixo.monad.name",
			Required:    true,
		},
		Branch: StringLabel{
			Description: "Git branch string",
			Key:         "org.kaixo.monad.git.branch",
			Required:    true,
		},
		Sha: StringLabel{
			Description: "Git sha string",
			Key:         "org.kaixo.monad.git.sha",
			Required:    true,
		},
		Origin: StringLabel{
			Description: "Git origin string",
			Key:         "org.kaixo.monad.git.origin",
			Required:    true,
		},
		Owner: StringLabel{
			Description: "Git owner string",
			Key:         "org.kaixo.monad.git.owner",
			Required:    true,
		},
		Repo: StringLabel{
			Description: "Git repo string",
			Key:         "org.kaixo.monad.git.repo",
			Required:    true,
		},
		Dirty: BoolLabel{
			Description: "Git dirty state",
			Key:         "org.kaixo.monad.git.dirty",
			Required:    true,
		},
		Role: FileLabel{
			Description: "Role template file",
			Key:         "org.kaixo.monad.role",
			Default:     filepath.Join("defaults", "role.json.tmpl"),
			Required:    false,
		},
		Policy: FileLabel{
			Description: "Policy template file",
			Key:         "org.kaixo.monad.policy",
			Default:     filepath.Join("defaults", "policy.json.tmpl"),
			Required:    false,
		},
		Resources: FileLabel{
			Description: "Resources template file",
			Key:         "org.kaixo.monad.resources",
			Default:     filepath.Join("defaults", "resources.json.tmpl"),
			Required:    false,
		},
		Bus: FolderLabel{
			Description: "Bus rule templates path",
			Key:         "org.kaixo.monad.bus",
			Required:    false,
		},
	},
}
