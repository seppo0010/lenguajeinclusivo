#!/bin/bash
set -Eeux

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
mkdir -p ${SCRIPT_DIR}/web/public/data
for ((i=1; i<=$#; i++))
do
    exp=${!i}
    exp_filename=${!i/\//-}

    ./builder -blacklist=/tmp/juscaba/blacklist "-json=/tmp/juscaba/${exp_filename}.json" -pdfs=./pdfs "-expediente=${exp}" -images=true

    pushd ts
    yarn run ts-node create-index.ts /tmp/juscaba/${exp_filename}.json /tmp/juscaba/${exp_filename}-index.json
    popd

    cp /tmp/juscaba/${exp_filename}{-index,}.json ${SCRIPT_DIR}/web/public/data
done

yarn
yarn build

mkdir -p /tmp/juscaba/web
rm -rf /tmp/juscaba/web/build
chmod -R 777 build
mv build /tmp/juscaba/web/
