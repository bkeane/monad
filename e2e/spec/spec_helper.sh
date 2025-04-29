# shellcheck shell=sh

# Defining variables and functions here will affect all specfiles.
# Change shell options inside a function may cause different behavior,
# so it is better to set them here.
# set -eu

. ./spec/helpers.sh

# This callback function will be invoked only once before loading specfiles.
spec_helper_precheck() {
  # Available functions: info, warn, error, abort, setenv, unsetenv
  # Available variables: VERSION, SHELL_TYPE, SHELL_VERSION
  setenv LOG_LEVEL=warn
  setenv MONAD_REPO=${MONAD_REPO:=monad}
  setenv MONAD_BRANCH=${MONAD_BRANCH:=$(git rev-parse --abbrev-ref HEAD)}
  setenv MONAD_SHA=${MONAD_SHA:=$(git rev-parse HEAD)}
  setenv MONAD_API=${MONAD_API:=kaixo}
  setenv MONAD_HOST=${MONAD_HOST:=$(resolve_api_domain $MONAD_API)}

  : minimum_version "0.28.1"
}

# This callback function will be invoked after a specfile has been loaded.
spec_helper_loaded() {
  :
}

# This callback function will be invoked after core modules has been loaded.
spec_helper_configure() {
  # Available functions: import, before_each, after_each, before_all, after_all
  : import 'support/custom_matcher'
}
