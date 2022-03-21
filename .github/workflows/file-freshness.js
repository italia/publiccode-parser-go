import { Octokit } from "https://cdn.skypack.dev/octokit?dts";
import { Md5 } from "https://deno.land/std@0.126.0/hash/md5.ts";
import { $ } from 'https://deno.land/x/zx_deno@1.2.2/mod.mjs'

const AMMINISTRAZIONI_TXT_URL =
  "https://www.indicepa.gov.it/public-services/opendata-read-service.php?dstype=FS&filename=amministrazioni.txt";
const LOCAL_FILE = "../../data/it/ipa_codes.txt";
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

console.log(`Checking ${LOCAL_FILE} freshness...`);

const localFileURL = new URL(LOCAL_FILE, import.meta.url);
const currentHash = await md5(localFileURL);

await $`curl -sL ${AMMINISTRAZIONI_TXT_URL} |
  tee /tmp/amministrazioni.txt |
  tail -n +2 |
  cut -f1 |
  LC_COLLATE=C sort > /tmp/ipa_codes.txt.new`;

const latestHash = await md5("file:///tmp/ipa_codes.txt.new");

// See if we already have an issue opened about it
const results = await octokit.rest.search.issuesAndPullRequests({
  q: `is:issue author:app/github-actions is:open user:${owner} repo:${repo}`,
  sort: "created",
  order: "desc",
});
console.log(`current: ${currentHash}`);
console.log(`latest:  ${latestHash}`);

let diff = ""

try {
  await $`diff -u --label old --label new ${localFileURL.pathname} /tmp/ipa_codes.txt.new`;
} catch (p) {
    if (p == 2) {
        // Exit status 2 from "diff" means something went wrong
        throw p.stderr;
    }

    diff = p.stdout;
}

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
        .replace(/```diff[\s\S]*```/, `\`\`\`diff\n${diff}\n\`\`\``),
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
    title: "Update data/it/ipa_codes.txt to the newest version",
    body: `<!-- ${ISSUE_MARKER} -->
A new version of \`amministrazioni.txt\` is available, we should generate an updated [\`data/it/ipa_codes.txt\`](../blob/master/data/it/ipa_codes.txt) from it.

Run
\`\`\`console
curl -sL '${AMMINISTRAZIONI_TXT_URL}' |
  tail -n +2 |
  cut -f1 |
  LC_COLLATE=C sort > data/it/ipa_codes.txt
\`\`\`
to generate an updated \`ipa_codes.txt\`.

**What changed**

\`\`\`diff
${diff}
\`\`\`
`,
    labels: ["enhancement"],
  });
}
