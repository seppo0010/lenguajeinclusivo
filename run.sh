#!/bin/bash
set -Eeux

exp=$1
exp_filename=${1/\//-}
SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
mkdir -p /tmp/juscaba/pdfs
mkdir -p /tmp/juscaba/web
mkdir -p ${SCRIPT_DIR}/web/public/data

./builder -blacklist=/tmp/juscaba/blacklist "-json=/tmp/juscaba/${exp_filename}.json" -pdfs=/tmp/juscaba/pdfs "-expediente=${exp}" -images=true

pushd ts
yarn run ts-node create-index.ts /tmp/juscaba/${exp_filename}.json /tmp/juscaba/${exp_filename}-index.json
popd

cp /tmp/juscaba/${exp_filename}{-index,}.json ${SCRIPT_DIR}/web/public/data

yarn build
mv build/* /tmp/juscaba/web
