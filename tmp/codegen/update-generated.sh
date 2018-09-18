#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

vendor/k8s.io/code-generator/generate-groups.sh \
deepcopy \
github.com/jakub-bacic/database-k8s-operator/pkg/generated \
github.com/jakub-bacic/database-k8s-operator/pkg/apis \
jakub-bacic:v1alpha1 \
--go-header-file "./tmp/codegen/boilerplate.go.txt"
