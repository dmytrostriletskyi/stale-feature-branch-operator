#!/bin/bash
# File .project-version should be changed each merge to master according to semantic versioning.
# As check is going based on Git's diff, the following file is excluded as grep works for it.
#
# Production image version should match the project version in file .project-version.
SUCCESSFUL_EXIST_CODE=0
FAILED_EXIST_CODE=1

if ! [[ $(git diff origin/master...HEAD ':!ops/check-project-version.sh' | grep "b/.project-version") ]]; then
  echo "You forgot to adjust file .project-version according to semantic versioning."
  exit "$((FAILED_EXIST_CODE))"
fi

if ! [[ $(cat configs/production.yml | grep $(cat .project-version)) ]]; then
  echo "Production image version (configs/production.yml) should match the project version in file .project-version."
  exit "$((FAILED_EXIST_CODE))"
fi

exit "$((SUCCESSFUL_EXIST_CODE))"
