#!/bin/bash
set -Eeux

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
mkdir -p /tmp/juscaba/pdfs
mkdir -p /tmp/juscaba/web
mkdir -p ${SCRIPT_DIR}/web/public/data
for ((i=1; i<=$#; i++))
do
    exp=${!i}
    exp_filename=${!i/\//-}

    ./builder -blacklist=/tmp/juscaba/blacklist "-json=/tmp/juscaba/${exp_filename}.json" -pdfs=/tmp/juscaba/pdfs "-expediente=${exp}" -images=true

    pushd ts
    yarn run ts-node create-index.ts /tmp/juscaba/${exp_filename}.json /tmp/juscaba/${exp_filename}-index.json
    popd

    cp /tmp/juscaba/${exp_filename}{-index,}.json public/data
done

yarn build
rm -rf /tmp/juscaba/web/build
mv build /tmp/juscaba/web/
