// Package helptext holds the long help and examples for every command as
// string constants, kept in one place so help text is reviewed like
// documentation rather than scattered through command definitions.
package helptext

const RootShort = "Manage App Store Connect from the command line"

const RootLong = `asctl is a CLI for managing App Store Connect: TestFlight, builds, releases,
code signing, team users, and sales reports.

It wraps the App Store Connect API around real release workflows: invite and
sync beta testers, assign builds to groups, submit versions for review, and
inspect what is live — from a terminal, script, or CI job.

Basic workflow:

  asctl auth init
  asctl apps use com.example.myapp
  asctl beta groups create alpha
  asctl beta testers sync --group alpha --file testers.csv
  asctl builds assign 42 --group alpha

Guide: https://github.com/vedanta/asctl`

const VersionShort = "Show CLI version"

const VersionLong = `Show the CLI version and build information.`
