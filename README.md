# bff
Breaking, Feature, Fix - a tool for managing semantic versioning.

A tool for implementing semantic versioning, but with a better name, inspired by [this tweet](https://twitter.com/kadikraman/status/1051935326028091392).

# Install

## Homebrew

```
brew tap chanzuckerberg/tap
brew install bff
```

## Platform Agnostic

You can use the godownloader script like so–

```
curl -L https://raw.githubusercontent.com/chanzuckerberg/bff/master/download.sh | sh
```

See the script for parameters.

# Common Errors

- Branch errors
You might see this if your git head's branch is different from the default branch. 
```
➜  aws-oidc git:(main) bff bump
Please only release versions from master.
SHAs on branches could go away if a branch is rebased or squashed.
latestVersionTag "0.22.3+96d6e06"
fileversion "0.22.3"
FATA[0000] tag does not match VERSION file              
make: *** [release] Error 1
```
To change the default branch in `head` from master to main, run this:
```
git remote set-head origin main
```
Then try running the bff command again. Make sure the default branch is clean!


- `FATA[0000] unable to fetch ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain`

Your ssh key hasn't been added to your ssh agent. Run ssh-add to add it. Learn more about ssh-add [here](https://www.ssh.com/ssh/add).
```
$ ssh-add
Identity added: directory-to-.ssh/key (user)
```
