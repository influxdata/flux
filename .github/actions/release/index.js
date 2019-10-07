const core = require('@actions/core');
const github = require('@actions/github');
const exec = require('@actions/exec');
const fs = require('fs');

const version = core.getInput('ref')
  .replace("refs/tags/", "");

async function generateChangelog(output) {
  var env = Object.assign({}, process.env);
  env["GO111MODULE"] = "on";

  await exec.exec("go", [
    "run",
    "github.com/influxdata/changelog",
    "generate",
    "--version",
    version,
    "--commit-url",
    "https://github.com/influxdata/flux/commit",
    "-o",
    output,
  ], {
    env: env,
  });
}

async function release(releaseNotes) {
  var env = Object.assign({}, process.env);
  env["GO111MODULE"] = "on";
  env["GITHUB_TOKEN"] = core.getInput("repo-token");

  await exec.exec("go", [
    "run",
    "github.com/goreleaser/goreleaser",
    "release",
    "--rm-dist",
    "--release-notes",
    releaseNotes,
  ], {
    env: env,
  });
}

async function run() {
  try {
    await generateChangelog("release-notes.txt");
    await release("release-notes.txt");
  } catch (error) {
    core.setFailed(error.message);
  }
  core.setOutput("version", version);
}

run();
