set unstable

# List help
[private]
default:
    @just --list --unsorted

# build and install monad to ~/.local/bin
install:
    go build -o ~/.local/bin/monad cmd/monad/main.go

# build docker images
[script]
build target='':
    if [ -z "{{ target }}" ]; then
        docker buildx bake --list targets
        exit 1
    fi

    BRANCH=$(git rev-parse --abbrev-ref HEAD) \
    docker buildx bake {{ target }} --progress=plain

# develop docs
docs:
    cd .docs/vue && npm run dev

# package docs
[script]
build-docs:
    # We are using the main branch of mermaid-cli for icon support.
    # In the future, when this is in a release, we should lock to a specific version.

    # remove all svg diagrams
    rm .docs/vue/assets/diagrams/*.png

    # generate deployment diagrams
    for file in .docs/mermaid/deployments/*.md; do
        npx github:mermaid-js/mermaid-cli \
        --iconPacks @iconify-json/bitcoin-icons @iconify-json/logos \
        --theme dark \
        --backgroundColor transparent \
        --input $file \
        --cssFile $(dirname $file)/style.css \
        --outputFormat png \
        --output .docs/vue/assets/diagrams/$(basename $file .md).png
    done

    # generate git diagrams
    for file in .docs/mermaid/git/*.md; do
         npx github:mermaid-js/mermaid-cli \
        --iconPacks @iconify-json/bitcoin-icons @iconify-json/logos \
        --theme neutral \
        --backgroundColor transparent \
        --input $file \
        --outputFormat png \
        --output .docs/vue/assets/diagrams/$(basename $file .md).png
    done

    # trim all empty space from the diagrams
    mogrify -trim +repage .docs/vue/assets/diagrams/*.png

    cd .docs/vue && npm run build