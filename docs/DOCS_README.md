# Osmosis Documentation Guide

The documentation for Osmosis is be hosted at:

- <https://docs.osmosis.zone/>

This site is generated from files in this [`docs` directory for `master`](https://github.com/osmosis-labs/osmosis/tree/master/docs)


### Vuepress
This site was generated using vuepress. [Documentation can be found here](https://vuepress.vuejs.org/).


### Config.js

The [config.js](./.vuepress/config.js) contains most of the configuration used for the site generation.


### Local Development

```
git clone https://github.com/osmosis-labs/osmosis.git
```

Make sure you are inside the docs folder
``` 
cd docs
yarn install
```

Run & watch files for changes
``` 
yarn dev
```


## Production build & Automated Deployment to Github Pages

This site can be deployed to Github pages. All that needs to be done is use the git action included in the root directory under `.github/workflows/docbuild.yml`. This action will deploy the site inside a branch called gh-pages. This branch can then be configured to be served as a website in the repository settings. The git action looks like this: 

```yaml
name: Build and Deploy
on:
  push:
    branches:
      - main
jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout 🛎️
        uses: actions/checkout@v2.3.1

      - name: Install and Build 🔧 # This will create version inside the 'build' folder.
        run: |
          cd docs
          npm install
          npm run build

      - name: Deploy 🚀
        uses: JamesIves/github-pages-deploy-action@4.1.5
        with:
          branch: gh-pages # The branch the action should deploy to.
          folder: docs/src/.vuepress/dist
```
