# Installation

Following these installation instructions will result in a fully functional "dev" build which means you will have total freedom to modify any part of this program as much as you like. It's all completely open source, there are no binaries and all 3rd party dependencies are included as source code in this repo for you to compile yourself, which means that this project is entirely self-contained within this repo.

## Step 1. Install Go

Follow the instructions [here](https://go.dev/doc/install). 

Verify your install by checking that `go version` prints the version of go.

## Step 2. Clone this repo and compile hoedown. 

Follow these steps:

```
$ git clone git@github.com:1f604/supersimplewiki.git
$ cd supersimplewiki/dependencies/hoedownv3.08/
$ make
```

Verify that the compilation has succeeded by checking that there is a `hoedown` binary in the `supersimplewiki/dependencies/hoedownv3.08/` directory. Check the version by typing this command:

```
$ ./hoedown --version
```

It should print this:

```
Built with Hoedown 3.0.7.
```

## Step 3. ???

## Step 4. Profit!







