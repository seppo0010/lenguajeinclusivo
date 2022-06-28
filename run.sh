#!/bin/bash
set -Eeux

mkdir -p /tmp/juscaba/pdfs
mkdir -p public/data
for ((i=1; i<=$#; i++))
do
    exp=${!i}
    exp_filename=${!i/\//-}

    ./builder "-json=public/data/${exp_filename}.json" -pdfs=/tmp/juscaba/pdfs "-expediente=${exp}" -images=${READ_IMAGES:-true} "-blacklist=${BLACKLIST_REGEX:-}" "-mirror-base-url=${MIRROR_BASE_URL:-}"

    pushd ts
    yarn run ts-node create-index.ts ../public/data/${exp_filename}.json ../public/data/${exp_filename}-index.json
    popd
done

yarn build

rm -rf /tmp/juscaba/build
chmod -R 777 build
mv build /tmp/juscaba/
