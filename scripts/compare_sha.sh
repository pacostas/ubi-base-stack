#!/usr/bin/env bash
node_versions_need_update=""

sha_node_16=$(docker image inspect registry.access.redhat.com/ubi8/nodejs-16-minimal --format '{{json .RepoDigests}}' | jq -r '.[0]')
saved_sha_node_16=$(cat ./stack-nodejs-16/nodejs-16-sha256 | jq -r '.[0]')

if [ "$sha_node_16" != "$saved_sha_node_16" ]; then
    node_versions_need_update+="16:"
fi

sha_node_18=$(docker image inspect registry.access.redhat.com/ubi8/nodejs-18-minimal --format '{{json .RepoDigests}}' | jq -r '.[0]')
saved_sha_node_18=$(cat ./stack-nodejs-18/nodejs-18-sha256 | jq -r '.[0]')

if [ "$sha_node_18" != "$saved_sha_node_18" ]; then
    node_versions_need_update+="18:"
fi

echo $node_versions_need_update
