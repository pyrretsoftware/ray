package main

var errorPage = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Error</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/hack-font@3.3.0/build/web/hack.css">
</head>
<style>
    body {
        margin: 0;
        background-color: black;
        font-family: 'Hack', sans-serif;
    }
    .center {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        width: 100vw;
        height: 100vh;
    }
    span, h1 {
        color: white;
    }
    h1 {
        margin: 0;
        font-weight: 400;
    }
    a {
        color: white;
        font-size: 1rem;
        text-decoration: underline black;
        transition: text-decoration 0.2s;
        display: flex;
        align-items: center;
        justify-content: center;
    } 
    a:hover {
        text-decoration: underline;
    }
    .rayver {
        position: absolute;
        right: 0;
        bottom: 0;
        padding: 0.5rem;
    }
</style>

<body>
    <div class="center">
        <h1>${ErrorCode}</h1>
        <a href="/">Back to home
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-chevron-right"><path d="m9 18 6-6-6-6"/></svg>
        </a>
    </div>
    <span class="rayver">ray ${RayVer}</span>
</body>
</html>`

var systemdService string = `[Unit]
Description=ray server (rays)
After=network.target

[Service]
User=${User}
Restart=always
ExecStart=${BinaryPath} --daemon
ExecReload=${BinaryPath} reload
ExecStop=${BinaryPath} stop

[Install]
WantedBy=multi-user.target`

var serviceCommand string = `create rays binpath= ${BinaryPath} start= auto`

var defaultConfig string = `{
    "Projects": [
        {
            "Name": "ray demo",
            "Src": "https://github.com/pyrretsoftwarelabs/ray-demo",
            "Domain": "localhost"
        }
    ]
}`