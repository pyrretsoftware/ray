<h1 align="center">
  <img src="https://raw.githubusercontent.com/pyrretsoftware/ray/refs/heads/main/logo.svg" height="100"></img>
  <br>
  ray
</h1>
<p align="center">
    <i>the self hosted tool for managing your web apps</i>
    <br>
    <a href="https://rdocs.axell.me/">Installation guide</a>
    <span> | </span>
    <a href="https://rdocs.axell.me/">Docs</a>
    <span> | </span>
    <a href="https://github.com/pyrretsoftware/ray/releases">Latest release</a>
    <span> | </span>
    <a href="https://github.com/pyrretsoftware/ray/actions">Latest builds</a>
</p>

# what is ray? 
ray is an open source tool which core purpose is deploying and managing web applications to your own self hosted server.

ray can automatically deploy from a git repo and manage different deployment channels on different branches. ray also includes a reverse proxy (ray router) to, well, route your request to the correct place, but also to manage things like automatically enrolling users to different deployment channels.
# ray's current stage of development
ray is currently very early in development. Right now it's certinaly not recommended to use it in a production environment.
# terms
- rays - ray server (the software running on the server)
- rayc - ray client (client software to remotely communicate with a ray server)
