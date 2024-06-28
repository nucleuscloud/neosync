---
title: Installing
description: Learn how to install the Neosync CLI onto your operating system of choice.
id: installing
hide_title: false
slug: /cli/installing
---

# Installing the CLI

Instructions on how to get the Neosync CLI installed on your local machine.

Neosync is delivered through a Command Line Interface (CLI) to make it easy for developers to use in their native workflows.
The Neosync CLI lets you view accounts, jobs and sync data locally. To get started with Neosync, follow the instructions below to download the CLI.

## MacOS

Homebrew is the simplest way to install nucleus CLI on the Mac. This can also be used on Linux, as well as on Windows 10 with Windows Subsystem for Linux.

### Homebrew

The easiest way to install the CLI is by using Homebrew. If you don't have Homebrew installed, follow these [instructions](https://docs.brew.sh/Installation). Next, open a new terminal window and use the following command:

```console
brew install neosync
```

You may also install directly from our brew repository:

```console
brew install nucleuscloud/tap/neosync
```

From then on, you can let Homebrew keep Nucleus up to date by running the following command.

```console
brew upgrade
```

## MacOS/Linux Direct Download

Navigate to Neosync [releases](https://github.com/nucleuscloud/neosync/releases) page of the CLI repository in the Nucleus Github. From there you can choose which binary to download based on your machine's architecture.

After you've downloaded and untarred the tarball, move it into your local bin to make it easy to run. If you're using Windows 10/11, see the Windows section below for more details.

**Note: the version listed below may not be the latest. Refer to the Releases page in the link above to retrieve the latest version of the binary.**

```console
tar xzf neosync_0.2.14_darwin_arm64.tar.gz neosync
mv neosync /usr/local/bin/neosync
```

### Verifying your installation

Once you've successfully installed the CLI, verify your installation by following these steps:

1. Open a new terminal window.
2. Type in `neosync help` into your terminal and press enter.
3. If installed successfully, you will see something similar to this help menu

```console
neosync

Terminal UI that interfaces with the Neosync system.

Usage:
  neosync [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  jobs        Parent command for jobs
  login       Login to Neosync
  sync        One off sync job to local resource
  version     Print the client version information
  whoami      Find out who you are

Flags:
      --api-key string   Neosync API Key. Takes precedence over $NEOSYNC_API_KEY
      --config string    config file (default is $HOME/.neosync/neosync.yaml)
  -h, --help             help for neosync
  -v, --version          version for neosync

Use "neosync [command] --help" for more information about a command.
```

Now that you've successfully downloaded the Neosync CLI, you're ready to start building and deploying services. Check out the next section to get familiar with the Neosync CLI commands.

## Windows 10/11 Direct Download

Navigate to Neosync [releases](https://github.com/nucleuscloud/neosync/releases) page of the CLI repository in the Neosync Github. From there you can choose which binary to download based on your machine's architecture for Windows. Some examples are listed below.

After the download has completed, unzip the contents into a new folder. The most important file is neosync.exe. This can be left here, but a more appropriate place to move it would be to a folder such as: `C:\Neosync` or `C:\Apps\Neosync`

Afterwards, this location needs to be added to the system path. This can be done by going into Settings, and searching for "environment variables" in the search bar. Click "Edit Environment Variables for your Account".

The "Path" variable should be edited. User or System is dependent on your preferences.

### Verifying your installation

Once you've successfully installed the CLI, verify your installation by following these steps:

1. Open a new terminal window.
   - Note: the examples below are using Powershell 5.1.x on Windows 11
2. Type in `neosync help` into your terminal and press enter.
3. If installed successfully, you will see something similar to this help menu

```console
neosync
Terminal UI that interfaces with the Neosync system.

Usage:
  neosync [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  jobs        Parent command for jobs
  login       Login to Neosync
  sync        One off sync job to local resource
  version     Print the client version information
  whoami      Find out who you are

Flags:
      --api-key string   Neosync API Key. Takes precedence over $NEOSYNC_API_KEY
      --config string    config file (default is $HOME/.neosync/neosync.yaml)
  -h, --help             help for neosync
  -v, --version          version for neosync

Use "neosync [command] --help" for more information about a command.
```

## Docker

A Docker image is published that matches each official release of Neosync CLI. Each versioned image includes the Neosync CLI release with the same version number.

These images wrap the Neosync executable, allowing you to run Neosync subcommands by passing in their names and arguments as part of `docker run`.

The list of images can be found on [Github](https://github.com/nucleuscloud/neosync/pkgs/container/neosync%2Fcli).

### Configuration

The container will need further configuration so that Neosync can access configuration files, and possibly source code if the plan is to issue source-code deployments with Neosync CLI.

See the example below for how to login to the CLI, and then view a list of environments in a Neosync account.

```console
docker run -it --rm -p 4242:4242 --mount source=neosynccfg,target=/root/.config/.neosync ghcr.io/nucleuscloud/neosync/cli:latest login
```

```console
docker run -it --rm --mount source=neosynccfg,target=/root/.config/.neosync ghcr.io/nucleuscloud/neosync/cli:latest accounts ls
```

The command above will print out a list of environments that are in the account associated with the logged in credentials. Note that the port mapping isn't required here, as that is only necessary during the login flow.

The docker volume is necessary in order to persist the Neosync CLI configuration data between runs. This today is namely used to persist auth data used during the login process so that it can be used with the other CLI commands.
