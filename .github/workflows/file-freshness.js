import { Octokit } from "https://cdn.skypack.dev/octokit?dts";
import { Md5 } from "https://deno.land/std@0.126.0/hash/md5.ts";

const AMMINISTRAZIONI_TXT_URL =
  "https://www.indicepa.gov.it/public-services/opendata-read-service.php?dstype=FS&filename=amministrazioni.txt";
const LOCAL_FILE = "../../data/amministrazioni.txt";
const ISSUE_MARKER = "#auto-issue-files-freshness-checker#";

async function md5(url) {
  const hash = new Md5();

  const res = await fetch(url);
  if (res.body) {
    for await (const chunk of res.body) {
      hash.update(chunk);
    }
  }

  return hash.toString();
}

const octokit = new Octokit({ auth: Deno.env.get("GITHUB_TOKEN") });

const [ owner, repo ] = Deno.env.get("GITHUB_REPOSITORY").split("/");

console.log("Checking ${LOCAL_FILE} freshness...")

const currentHash = await md5(new URL(LOCAL_FILE, import.meta.url));
const latestHash = await md5(AMMINISTRAZIONI_TXT_URL);

// See if we already have an issue opened about it
const results = await octokit.rest.search.issuesAndPullRequests({
  q: `is:issue author:app/github-actions is:open user:${owner} repo:${repo}`,
  sort: "created",
  order: "desc",
});

console.log(`current: ${currentHash}`);
console.log(`latest:  ${latestHash}`);

// If we do, update the current issue
if (new RegExp(ISSUE_MARKER).test(results?.data?.items[0]?.body)) {
  const issue = await octokit.rest.issues.get({
    owner,
    repo,
    issue_number: results.data.items[0].number,
  });

  let params = {};

  // Close the issue if the file has been updated in the repo
  if (currentHash === latestHash) {
    await octokit.rest.issues.createComment({
      owner,
      repo,
      issue_number: issue.data.number,
      body: "The file is up to date now, closing.",
    });
    params = {
      state: "closed",
    };
  } else {
    params = {
      body: issue.data.body
        .replace(/\(current-md5: .*\)/, `(current-md5: \`${currentHash}\`)`)
        .replace(/\(latest-md5: .*\)/, `(latest-md5: \`${latestHash}\`)`),
    };
  }

  console.log(`Found issue #${issue.data.number}, updating...`);
  await octokit.rest.issues.update({
    owner,
    repo,
    issue_number: issue.data.number,
    ...params,
  });
} else if (currentHash !== latestHash) {
  console.log("Creating new issue...");
  await octokit.rest.issues.create({
    owner,
    repo,
    title: "Update amministrazioni.txt to the newest version",
    body: `<!-- ${ISSUE_MARKER} -->
A new version of \`amministrazioni.txt\` is available, we should update it.

* [Current file in the repo](../blob/master/data/amministrazioni.txt) (current-md5: \`${currentHash}\`)
* [**Latest file available**](${AMMINISTRAZIONI_TXT_URL}) (latest-md5: \`${latestHash}\`)
`,
    labels: ["enhancement"],
  });
}
