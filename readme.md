<h1 align="center">
  <img src="https://raw.githubusercontent.com/pyrretsoftware/ray/refs/heads/main/logo.svg" height="100"></img>
  <br>
  ray
</h1>
<p align="center">
    <i>easily self-host your stuff</i>
    <br>
    <a href="https://rdocs.axell.me/guides/install">Installation guide</a>
    <span> | </span>
    <a href="https://rdocs.axell.me/">Docs</a>
    <span> | </span>
    <a href="https://github.com/pyrretsoftware/ray/releases">Latest release</a>
    <span> | </span>
    <a href="https://github.com/pyrretsoftware/ray/actions">Latest unstable builds</a>
</p>

# what is ray? 
At its core, ray is a comprehensive system for deploying, managing, and routing web applications on self-hosted servers. Given a project configuration file, ray can build, deploy and make your project accessible through ray's reverse proxy (ray router). Ray can automatically deploy and update from a remote git repository, manage different deployment channels on different branches, handle authenticating users on private deployment channels, load balance your application to other ray servers, monitor your projects, notify you if anything goes wrong, and a lot more.

# ray's current stage of development
ray is currently early in development. The latest release (v1.0.0) is considered to be somewhat stable but some features, mainly TLS, are not fully tested and could contain bugs, so it might be smart to wait before using ray in production.
# components
- rays - ray server (the software running on the server)
- rray - remote ray (tool to help you easily connect to and manage remote ray servers)
- rayinstall - ray installer (utility to help with installing)
