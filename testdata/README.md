# Add new testcase files to this directory

**tl;dr**:

* `v0.1/valid/name.yml`: A valid v0.2 publiccode.yml testing the `name` key.
* `v0.1/invalid/name_missing.yml`: An invalid v0.1 publiccode.yml where the mandatory
  `name` key is missing.
* `v0.2/invalid/latin1_encoded.yml`: An invalid v0.2 publiccode.yml with the wrong
  encoding.

Rules are not enforced but help keeping things tidy.

## Path

Put files in `VERSION/valid/` for valid YAML files for that version and
`VERSION/invalid/` for invalid YAML files.

Where `VERSION` is the version of the core standard prefixed by `v`.

**Examples**

* `v0.2/valid/`
* `v0.2/invalid/`
* `v0.1/invalid/`

## File contents

Start from `valid.minimal.yml` and edit it to contain the subject of
your test.

Add comments only to the section / key being tested to draw attention
to it and, if the testcase needs more context, add it as a comment at
the top of the file.

## Filename

### Testcases validating keys

Filenames for testcases testing a particular key follow this format:

`KEYNAME_REASON.yml`

Where `KEYNAME` is the name of the key being checked in the test (eg. `name`).
Use an underscore as a separator for nested keys (eg. `maintenance_contracts`).

`REASON` is the scenario the testcase is checking. It should be descriptive and
short. Good `REASON`s are:

* `nil`: The key is nil
* `empty`: The key is an empty string
* `missing`: The key is missing
* `invalid`: Generic, the key format is wrong.

Examples (not necessarly real):

* `valid/key_nil.yml`: We expect the key to be nil
* `invalid/maintenance_contracts_empty.yml`: We expect maintenance/contracts key
  to contain something.
* `invalid/longDescription_invalid.yml`: We are testing an unspecified invalid
  value in longDescription.
* `invalid/longDescription_blocked_word.yml`: We expect longDescription doesn't contain
  any blocked word.

### Other testcases

Use a short and descriptive name for testcases not focusing on
a particular key, eg `file_encoding.yml`
