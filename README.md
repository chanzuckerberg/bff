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

## `FATA[0000] unable to fetch ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain`

Your ssh key hasn't been added to your ssh agent. Run ssh-add to add it. Learn more about ssh-add [here](https://www.ssh.com/ssh/add).
```
$ ssh-add
Identity added: directory-to-.ssh/key (user)
```
