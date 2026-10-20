# Coding standards

The standards a diff is reviewed against live in these sources. Apply every rule in each one the diff touches.

- `docs/adr/`: the decisions behind the architecture. A diff that contradicts an ADR without superseding it is a violation.
- `docs/toolchain.md`: the conventions and pitfalls of each tool, config and test kind. List its sections with `grep -n '^## ' docs/toolchain.md` and read the section for everything the diff touches. Its `## Agent hooks` table names the source file of each hook rule.
- `GLOSSARY.md`: identifiers, test names and docs use its terms.
- Doc drift: a diff that renames, removes or changes something a doc describes updates that doc. Grep `docs/`, `GLOSSARY.md` and `.env.example` for every identifier, target and setting the diff renames or removes.

Tooling already enforces the rest, so a review skips it: formatting, golangci-lint (`.golangci.yml`), go-arch-lint (`.go-arch-lint.yml`), the architecture tests in `tests/architecture/`, generated-code drift (`make drift-check`) and the agent hooks' comment guard.
