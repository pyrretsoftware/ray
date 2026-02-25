<h1 align="center">
  <img src="https://raw.githubusercontent.com/pyrretsoftware/ray/refs/heads/main/logo.svg" height="100"></img>
  <br>
  ray
</h1>
<p align="center">
    <i>easily self-host your stuff</i>
    <br>
    <a href="https://ray.pyrret.com/guides/installation">Installation guide</a>
    <span> | </span>
    <a href="https://ray.pyrret.com">Docs</a>
    <span> | </span>
    <a href="https://github.com/pyrretsoftware/ray/releases">Latest release</a>
    <span> | </span>
    <a href="https://github.com/pyrretsoftware/ray/actions/workflows/package.yml">Latest unstable builds</a>
    <span> | </span>
    <a href="https://discord.gg/aesgfB6EnX">Chat</a>
</p>

# what is ray? 
ray is a comprehensive system for deploying, managing, and routing web applications on self-hosted servers. Given a project configuration file, ray can build, deploy and make your project accessible through ray's reverse proxy (ray router). Ray can automatically deploy and update from a remote git repository, manage different deployment channels on different branches, handle authenticating users on private deployment channels, load balance your application to other ray servers, monitor your projects, notify you if anything goes wrong, and a lot more.

# getting started
first, go through [the installation guide](https://ray.pyrret.com/guides/installation/). Then, we recommend the guide ["deploying a project"](https://ray.pyrret.com/guides/deploying-a-project/) to learn some of the basics.

after that, you might want to explore some of ray's features using either the [docs features section](https://ray.pyrret.com/features/) or looking at the configuartion references ([1](https://ray.pyrret.com/reference/rayconfig/), [2](https://ray.pyrret.com/reference/projectconfig/)).

you can also [join the discord](https://discord.gg/aesgfB6EnX) if you have any questions.

# components and sister repositories
in this repo:
- **rays** stands for ray server and is the main component of ray.
- **rayc** stands for ray client and interacts with comlines exposed by rays.
- **rayinstall** is the installation utility for rays.
- **raydoc** is ray's automatic documentation tool.

sister repos:
- [ray website and docs](https://github.com/pyrretsoftware/raydocs) - hosted at [ray.pyyret.com](https://ray.pyrret.com)
- [comline package](https://github.com/pyrretsoftware/comline) - go package to interact with comlines
- [modernstatus](https://github.com/pyrretsoftware/modernstatus) - basic looking raystatus implementation

# ray's current stage of development
ray is currently slowly receiving new updates and is not actively developed. The latest release (v3.0.0) is considered to be pretty stable and i currently use it for everything i personally host. It has been tested extensively but remember that ray is a hobby project that comes with no warranty. 

# known issues (in the latest release)
There are currently no known issues in the latest release.