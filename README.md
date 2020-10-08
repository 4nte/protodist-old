<p align="center">
<img width="600" src="assets/logo2.png" alt="Protodist" title="Protodist" />
</p>

![docker](https://img.shields.io/github/go-mod/go-version/4nte/protodist)
![docker](https://img.shields.io/docker/pulls/antegulin/protodist)
![version](https://img.shields.io/github/v/release/4nte/protodist?sort=semver)
![license](https://img.shields.io/github/license/4nte/protodist)

**This project is in early development!**

Docs and a first alpha version is coming soon.


all-in-one tool for compiling, bundling, and distributing protobuf packages via Git repositories (Batteries included).

### Why?
Idea behind protodist is to simplify and automate the entire process of compiling and distributing protobuf packages using Git Repositories.
Protodist requires a `protodist.yaml` configuration file to understand how you want your protobuf packages to be compiled and distributed.


### Install

```shell script
curl -sfL https://raw.githubusercontent.com/4nte/protodist/master/install.sh | sh
```