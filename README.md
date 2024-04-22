# vcrbpkg

# DEPRECATED, PLEASE USE THE OFFICIAL PACKAGER AT: https://docs.veracode.com/r/About_auto_packaging

Unofficial community project automated packaging for Ruby on Rails applications for Veracode Static Analysis.

## What you **must** know about packaging Ruby on Rails applications with Veracode Static Analysis

Veracode Static Analysis **does not analyze Ruby source code**.

Rather it uses the `veracode` Ruby Gem (available at <https://rubygems.org/gems/veracode> ) to **run your application** and in a running application loads all Ruby files and compiles them to YARV (Yet Another Ruby VM) instructions using [RubyVM::InstructionSequence](https://ruby-doc.org/core-2.6/RubyVM/InstructionSequence.html).

This means that when you run `veracode prepare`, **the machine you run on must be setup for your application to run**.

If it does not run without a environment variable with an `API_KEY` and this is not available when running `veracode prepare`, then `veracode prepare` will not be able to run.

## How vcrbpkg helps

`vcrbpkg` (VeraCode RuBy PacKaGe) automates much of the packaging steps mentioned in the ["Ruby on Rails packaging" section on the Veracode Docs](https://docs.veracode.com/r/compilation_ruby):

* Ensuring Ruby is installed globally.
* Ensuring RVM is installed to manage Ruby version for this project.
* Can help check out a repo by URL or work on a local directory.
* Ensuring the directory is a Rails app (Veracode Static Analysis only supports Ruby on Rails applications, not any other kind of Ruby applications).
* Verifies the required Ruby version is supported (but will still package even if it is not as we may occassionally still be able to analyze unsupported versions)
* Verifies the required Rails version is supported (but will still package even if it is not as we may occassionally still be able to analyze unsupported versions)
* Installs the veracode gem.
* Tests if we can use the `production` environment (recommended) but if not, tests if `development` or `test` work.
* Runs `veracode prepare`.

It is designed to work from a local or a CI environment.

Note that because it does a lot of compiling (Ruby as well as native extensions and compiling Ruby code into instructions) a compute heavy machine is recommended.

Also please keep in mind that it **will run (bootstrap) your application**.

## Linux & OS X

### Download

On Linux with go get:

```sh
export GOPATH=`go env GOPATH` &&
export PATH="$GOPATH/bin:$PATH" &&
go install github.com/relaxnow/vcrbpkg/cmd/vcrbpkg@latest
```

### Usage

Package the current working directory:

```sh
~/Projects/my-go-project# vcrbpkg
```

OR package a directory:

```sh
~/# vcrbpkg Projects/my-go-project
```

OR package a repo:

```sh
vcrbpkg https://github.com/OWASP/railsgoat.git
```

vcrbpkg will run `veracode prepare` which will output the zip file at the end.

To have `vcrbpkg` copy the zip package on success add `--out`:

```sh
vcrbpkg railsgoat --out /tmp/railsgoat.zip
```

This zip file can then be uploaded to Veracode Static Analysis.

## Windows

Not currently supported. PRs welcome!

## Development

Build Docker image and get it running:

```sh
docker build . -t vcrbpkg &&
docker run -it --rm -v .:/app --rm vcrbpkg
go run cmd/vcrbpkg/vcrbpkg.go https://github.com/OWASP/railsgoat.git
```
