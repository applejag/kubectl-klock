# kubectl krew template repo

There's a lot of scaffolding needed to set up a good kubectl plugin. This repo is a GitHub Template Repo to make it easy to set all of this scaffolding up for a new repo.

The assumptions made are:
1. You'll write your plugin in go
2. You want client-go to interact with the cluster
3. You want all of the kubectl flags available to your plugin
4. Your plugin's home will be a github.com repo
5. Your plugin will work in Linux, MacOS and Windows

## Create your repo

[Start here](https://github.com/replicatedhq/krew-plugin-template/generate) to create a new repo based on this template. This is not a fork, it will make a copy of this repo into your own organization or GitHub account. 

Click that, and create your own version of this repo. Clone it locally. The rest of the steps you will be performing on your local copy.

## Make it yours

Once you have your own repo created locally, change to the directory and run:

```shell
make setup
```

This will prompt you for a few things, such as your GitHub org, repo name and plugin name. The setup application will then update the import paths and code with the data you provided.

(Note, once you've run this step, these instructions will no longer be present in your repo. You can always vew then at [https://github.com/replicatedhq/krew-plugin-template](https://github.com/replicatedhq/krew-plugin-template)).

Commit and check it in to your repo!

```shell
git add .
git commit -m "Updating from template"
git push -u origin master
```

## Write your Plugin

Next, open the pkg/plugin/plugin.go file. This is where you can start writing your plugin.

For an example, take a look at the [outdated](https://github.com/replicatedhq/outdated) plugin that inspired this template.

To make a local build:

```shell
make bin
```

## Creating a release

To create a new release of your plugin, create and push a tag.

```shell
git tag v0.1.0 -m 'initial release'
git push --tags
```

This repo has a built-in GitHub Action that will handle the build process. We use [GoReleaser](https://goreleaser.com) to create tagged releases. This will create all three binaries and push them to the releases page when you push a tag. It will take a few minutes to complete, so be patient.

## Submitting to Krew

Be sure to read the guidelines on the Krew Developer guide before submitting you plugin. This is not automated (yet). We've created a starting point for your plugin manifest, look for it in deploy/krew/plugin.yaml.

## Share!

Finally, we'd love to hear if you've used this template. Let us know on Twitter at @replicatedhq. We've written a few kubectl plugins too, and are always curious to see what other people are working on.
