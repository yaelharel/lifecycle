## Changelog

A simple script that generates the changelog for lifecycle based on a lifecycle version (aka milestone).

### Usage

#### Github Action

```yaml
  - name: Generate changelog
    uses: actions/github-script@v1
    id: changelog
    with:
      github-token: ${{secrets.GITHUB_TOKEN}}
      result-encoding: string
      script: |
        const path = require('path');
        const scriptPath = path.resolve('.github/workflows/build/changelog/index.js');
        require(scriptPath)({core, github, repository: "${{ env.GITHUB_REPOSITORY }}", version: "${{ env.LIFECYCLE_VERSION }}" });
```

#### Local

To run/test locally:

```shell
# install deps
npm install

# set required info
export GITHUB_TOKEN="<GITHUB_PAT_TOKEN>"
export LIFECYCLE_VERSION="<LIFECYCLE_VERSION>"

# run locally
npm run local
```

Notice that a file `changelog.md` is created as well for further inspection.