<h1 align="center">
  <img src="https://raw.githubusercontent.com/pyrretsoftware/ray/refs/heads/main/logo.svg" height="100"></img>
  <br>
  ray
</h1>
<p align="center">
    <i>the self hosted tool for managing your web apps</i>
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
ray is an open source tool which core purpose is deploying and managing web applications to your own self hosted server.

ray can automatically deploy from a git repo and manage different deployment channels on different branches. ray also includes a reverse proxy (ray router) to, well, route your request to the correct place, but also to manage things like automatically enrolling users to different deployment channels.
# ray's current stage of development
ray is currently early in development. The latest release (v1.0.0) is considered to be stable but some extra features are not guaranteed to work and it could contain bugs, so it might be smart to wait before using ray in production.
# terms
- rays - ray server (the software running on the server)
- rayinstall - ray installer (utility to help with installing ray installation packages, .rpack files)