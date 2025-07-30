<h1 align="center">
  <img src="https://raw.githubusercontent.com/pyrretsoftware/ray/refs/heads/main/logo.svg" height="100"></img>
  <br>
  ray
</h1>
<p align="center">
    <i>easily self-host your stuff</i>
    <br>
    <a href="https://docs.ray.pyrret.com/guides/install">Installation guide</a>
    <span> | </span>
    <a href="https://docs.ray.pyrret.com/">Docs</a>
    <span> | </span>
    <a href="https://github.com/pyrretsoftware/ray/releases">Latest release</a>
    <span> | </span>
    <a href="https://github.com/pyrretsoftware/ray/actions">Latest unstable builds</a>
</p>

# what is ray? 
ray is a comprehensive system for deploying, managing, and routing web applications on self-hosted servers. Given a project configuration file, ray can build, deploy and make your project accessible through ray's reverse proxy (ray router). Ray can automatically deploy and update from a remote git repository, manage different deployment channels on different branches, handle authenticating users on private deployment channels, load balance your application to other ray servers, monitor your projects, notify you if anything goes wrong, and a lot more.

# ray's current stage of development
ray is currently early in development. The latest release (v2.1.0) is considered to be pretty stable and i currently use it for everything i personally host. Beware though, TLS (https) has not been tested yet, and there could certanly be bugs in this release.

# known issues (in the latest release)
These are bugs that are known to be in the current release, most of them fixed already, but not pushed out in a release.
* *(no issues are known in the latest release at this time)*

# components
- rays - ray server (the software running on the server)
- rray - remote ray (tool to help you easily connect to and manage remote ray servers)
- rayinstall - ray installer (utility to help with installing)
