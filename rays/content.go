package main

var loginPage = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Error</title>
</head>
<style>
    body {
        margin: 0;
        background-color: black;
        font-family: monospace;
    }
    .center {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        width: 100vw;
        height: 100vh;
    }
    span, h1, p {
        color: white;
    }
    h1 {
        margin: 0;
        font-weight: 400;
    }
    a, button {
        color: white;
        text-decoration: underline black;
        transition: text-decoration 0.2s;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
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
    .command {
        margin-top: 0.3rem;
        margin-bottom: 0.55rem;
        border-radius: 0.5rem;
        background-color: rgb(17, 17, 17);
        transition: background-color 0.2s;
    }
    .command:hover {
        background-color: rgb(13, 13, 13);
    }
    .inst {
        display: flex;
        flex-direction: column;
        width: 35rem;
    }
    input {
        outline: none;
        border: none;
        background-color: transparent;
        color: white;
        font-family: inherit;
        padding: 0.55rem;
    }
    .login {
        display: grid;
        grid-template-columns: auto max-content;
        gap: 0.5rem
    }
</style>
<script>
    function login() {
        document.cookie = "ray-auth=" + document.querySelector('#authkey').value
        document.location.reload(true)
    }
</script>
<body>
    <div class="center">
        <div class="inst">
            <span>1. On this server, open up a command line and run the following ray command:</span>
            <p class="command"><span style="padding-left: 0.55rem;">$ </span>
                <input style="padding-left: 0;" type="text" name="" value="rays dev-auth" readonly id="">
            </p>
            <span>2. Take your newly generated authentication key and paste it in the field below:</span>

            <form onsubmit="login()" class="login">
                <p style="padding: 0; margin-bottom: 0;" class="command">
                    <input style="width: 100%;" id="authkey" type="text" name="" required placeholder="authentication key goes here!" id="">
                </p>
                <button style="height: 2.03755rem; width: 2.03755rem; margin-bottom: 0; border: 0; " class="command" href="/">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-chevron-right"><path d="m9 18 6-6-6-6"/></svg>
            </button>
            </form>
        </div>
    </div>
</body>
</html>`

var errorPage = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Error</title>
</head>
<style>
    body {
        margin: 0;
        background-color: black;
        font-family: monospace;
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

//var serviceCommand string = `create rays binpath= ${BinaryPath} start= auto`

var defaultConfig string = `{
    "EnableRayUtil" : true,
    "Projects": [
        {
            "Name": "ray demo",
            "Src": "https://github.com/pyrretsoftwarelabs/ray-demo",
            "Domain": "localhost"
        }
    ]
}`