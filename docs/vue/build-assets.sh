#!/bin/bash
set -e

# We are using the main branch of mermaid-cli for icon support.
# In the future, when this is in a release, we should lock to a specific version.

# remove all png diagrams
mkdir -p assets/diagrams
rm -f assets/diagrams/*.png

# generate deployment diagrams
for file in mermaid/deployments/*.md; do
    npx github:mermaid-js/mermaid-cli \
    --iconPacks @iconify-json/bitcoin-icons @iconify-json/logos \
    --theme dark \
    --backgroundColor transparent \
    --input "$file" \
    --cssFile "$(dirname "$file")/style.css" \
    --outputFormat png \
    --output "assets/diagrams/$(basename "$file" .md).png"
done

# generate git diagrams
for file in mermaid/git/*.md; do
    npx github:mermaid-js/mermaid-cli \
    --iconPacks @iconify-json/bitcoin-icons @iconify-json/logos \
    --theme neutral \
    --backgroundColor transparent \
    --input "$file" \
    --outputFormat png \
    --output "assets/diagrams/$(basename "$file" .md).png"
done

# trim all empty space from the diagrams
mogrify -trim +repage assets/diagrams/*.png