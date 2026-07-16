#!/bin/bash
# Generate OpenAPI clients
# Usage: ./generate.sh [language]

LANG=${1:-python}

case $LANG in
  python)
    openapi-generator generate -i sdk/openapi/openapi.yaml -g python -o sdk/clients/python
    ;;
  go)
    openapi-generator generate -i sdk/openapi/openapi.yaml -g go -o sdk/clients/go
    ;;
  typescript)
    openapi-generator generate -i sdk/openapi/openapi.yaml -g typescript-axios -o sdk/clients/typescript
    ;;
  java)
    openapi-generator generate -i sdk/openapi/openapi.yaml -g java -o sdk/clients/java
    ;;
  *)
    echo "Generating all clients..."
    openapi-generator generate -i sdk/openapi/openapi.yaml -g python -o sdk/clients/python
    openapi-generator generate -i sdk/openapi/openapi.yaml -g go -o sdk/clients/go
    openapi-generator generate -i sdk/openapi/openapi.yaml -g typescript-axios -o sdk/clients/typescript
    openapi-generator generate -i sdk/openapi/openapi.yaml -g java -o sdk/clients/java
    ;;
esac
echo "Client generated: sdk/clients/$LANG"
