package cli

//
// Basis Config Struct (for annotation-driven options)
//

type BasisConfig struct {
	Owner   string `bind:"--owner,MONAD_OWNER" desc:"git repository owner [default: from git remote]"`
	Repo    string `bind:"--repo,MONAD_REPO" desc:"git repository name [default: from git remote]"`
	Branch  string `bind:"--branch,MONAD_BRANCH" desc:"git repository branch [default: current git branch]"`
	Sha     string `bind:"--sha,MONAD_SHA" desc:"git repository sha [default: current git sha]"`
	Service string `bind:"--service,MONAD_SERVICE" desc:"service name [default: basename $PWD]"`
	Image   string `bind:"--image,MONAD_IMAGE" desc:"service ecr image path [default: owner/repo/service:branch]"`
}