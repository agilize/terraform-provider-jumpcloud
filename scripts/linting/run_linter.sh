#!/bin/bash

echo "Terraform Provider Linting Helper"
echo "=================================="
echo 

# Function to display usage
show_usage() {
  echo "Usage: $0 [options]"
  echo
  echo "Options:"
  echo "  --all        Run all lint checks"
  echo "  --critical   Run only critical checks"
  echo "  --r001       Check R001 (ResourceData.Set with string literal)"
  echo "  --r014       Check R014 (interface{} parameter name)"
  echo "  --r017       Check R017 (Schema attributes stability)"
  echo "  --r019       Check R019 (HasChanges → HasChangesExcept)"
  echo "  --v013       Check V013 (SchemaValidateFunc → validation.StringInSlice)"
  echo "  --at         Check all acceptance test issues"
  echo
  echo "If no options are provided, this help message is displayed."
}

# Check if no arguments were provided
if [ $# -eq 0 ]; then
  show_usage
  exit 0
fi

# Parse arguments
while [ "$1" != "" ]; do
  case $1 in
    --all )
      echo "Running all linting checks"
      tfproviderlint ./...
      exit 0
      ;;
    --critical )
      echo "Running only critical checks"
      ./scripts/linting/check_critical_lint.sh
      exit 0
      ;;
    --r001 )
      echo "Checking R001: ResourceData.Set() key argument should be string literal"
      tfproviderlint -AT=false -R=false -S=false -V=false -R001=true ./...
      ;;
    --r014 )
      echo "Checking R014: interface{} parameter should be named meta"
      tfproviderlint -AT=false -R=false -S=false -V=false -R014=true ./...
      ;;
    --r017 )
      echo "Checking R017: schema attributes should be stable across Terraform runs"
      tfproviderlint -AT=false -R=false -S=false -V=false -R017=true ./...
      ;;
    --r019 )
      echo "Checking R019: d.HasChanges() has many arguments, consider d.HasChangesExcept()"
      tfproviderlint -AT=false -R=false -S=false -V=false -R019=true ./...
      ;;
    --v013 )
      echo "Checking V013: custom SchemaValidateFunc should be replaced with validation.StringInSlice()"
      tfproviderlint -AT=false -R=false -S=false -V=false -V013=true ./...
      ;;
    --at )
      echo "Checking all acceptance test issues"
      tfproviderlint -AT=true -R=false -S=false -V=false ./...
      ;;
    * )
      show_usage
      exit 1
      ;;
  esac
  shift
done 