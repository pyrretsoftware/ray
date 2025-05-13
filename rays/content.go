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
                <input style="padding-left: 0;" type="text" name="" value="sudo rays dev-auth" readonly id="">
            </p>
            <span>2. Take your newly generated authentication key and paste it in the field below:</span>

            <form onsubmit="login()" class="login">
                <p style="padding: 0; margin-bottom: 0;" class="command">
                    <input style="width: 100%;" id="authkey" type="text" name="" required placeholder="authentication key goes here!" id="">
                </p>
                <button style="width: 36px; margin-bottom: 0; border: 0; " class="command" href="/">
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

var messageTemplates = map[string]string{
    "processError" : `<p>The underlying process that is responsible for hosting this service has gone offline. It could be because of an application bug, an issue with routing your request between different servers, or something else. If this service has a status page, consider checking that.</p>
        <p>This is something that needs to be investigated and fixed by this service's owner. Consider contacting them about this issue, and provide the following error message: </p>`,
    "cantResolve" : `<p>An appropriate process to handle the request to this service was not found. It could be this host not pointing to any service, or another issue. This is likely an issue with your client's request, and not an issue with the server or reverse proxy. The following error code was provided:</p>`,
    "requestIssue" : `<p>There is an issue with your issue preventing it from being served. See the below error message: </p>`,
    "unknownError" : `<p>An unknown error was encountered trying to serve your request. This is likely an issue with the ray reverse proxy server, but could also be an issue with something else. The following error code was also provided: </p>`,
}

func getV2ErrorPage(clientState string, rayState string, processState string, messageTemplate string, errorMessage string) string {
    clientColor := "working"
    rayColor := "working"
    processColor := "working"
    if clientState != "working" {
        clientColor = "down"
    }
    if rayState != "working" {
        rayColor = "down"
    }
    if processState != "working" {
        processColor = "down"
    }


    ErrPage := `<!DOCTYPE html>
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
            flex-direction: row;
            align-items: center;
            justify-content: center;
            width: 100vw;
            height: 100vh;
        }
        span, h1, h2, p {
            color: white;
        }
        h1, h2 {
            margin: 0;
            font-weight: 400;
        }
        svg {
            color: white;
        }
        a {
            color: rgb(187, 187, 187);
            text-decoration: none;
            display: flex;
            align-items: center;
            justify-content: center;
        } 
        .rayver {
            position: absolute;
            right: 0;
            bottom: 0;
            padding: 0.5rem;
        }

        .graphic {
            display: flex;
            flex-direction: column;
            margin-right: 2rem;
        }
        .item {
            display: flex;
            flex-direction: row;
            align-items: center;
        }
        .item svg {
            margin-right: 0.3rem;
        }
        .text {
            max-width: 25%;   
        }

        .command {
            margin-top: 0.3rem;
            border-radius: 0.5rem;
            background-color: rgb(17, 17, 17);
            transition: background-color 0.2s;
            padding: 0.55rem;
            margin-bottom: 0;
        }
        .command:hover {
            background-color: rgb(13, 13, 13);
        }
        .working {
            color: aquamarine;
        }
        .down {
            color: orange;
        }
        .line {
            display: flex;
            flex-direction: column;
            margin-top: 0.3rem;
            margin-bottom: 0.3rem;
        }

        @media (orientation: portrait) {
            .graphic {
                display: none;
            }
            .text {
                max-width: 75%;
            }
        }
    </style>

    <body>
        <div class="center">
            <div class="graphic">
                <div class="item">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-globe-icon lucide-globe"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/></svg>
                    <span>Client (you):<span class="` + clientColor + `"> ` + clientState +`</span></span>
                </div class="item">

                <div class="line">
                    <svg style="margin-bottom: -0.1rem;" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-move-up-icon lucide-move-up"><path d="M8 6L12 2L16 6"/><path d="M12 2V22"/></svg>
                    <svg style="margin-top: -0.1rem;" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-move-down-icon lucide-move-down"><path d="M8 18L12 22L16 18"/><path d="M12 2V22"/></svg>
                </div>

                <div class="item">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-pyramid-icon lucide-pyramid"><path d="M2.5 16.88a1 1 0 0 1-.32-1.43l9-13.02a1 1 0 0 1 1.64 0l9 13.01a1 1 0 0 1-.32 1.44l-8.51 4.86a2 2 0 0 1-1.98 0Z"/><path d="M12 2v20"/></svg>
                    <span>Ray:<span class="` + rayColor +`"> ` + rayState + `</span></span>
                </div>

                <div class="line">
                    <svg style="margin-bottom: -0.1rem;" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-move-up-icon lucide-move-up"><path d="M8 6L12 2L16 6"/><path d="M12 2V22"/></svg>
                    <svg style="margin-top: -0.1rem;" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-move-down-icon lucide-move-down"><path d="M8 18L12 22L16 18"/><path d="M12 2V22"/></svg>
                </div>

                <div class="item">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-server-icon lucide-server"><rect width="20" height="8" x="2" y="2" rx="2" ry="2"/><rect width="20" height="8" x="2" y="14" rx="2" ry="2"/><line x1="6" x2="6.01" y1="6" y2="6"/><line x1="6" x2="6.01" y1="18" y2="18"/></svg>
                    <span>Service:<span class="`+ processColor +`"> ` + processState +`</span></span>
                </div>
            </div>
            <div class="text">
                <h1>An issue was encountered</h1>
                ` + messageTemplates[messageTemplate] +`

                <p class="command">
                    <span>` + errorMessage +`</span>
                </p>
            
            </div>
        </div>
        <span class="rayver">ray ` + Version + `</span>
    </body>
    </html>`

    return ErrPage
}

//var serviceCommand string = `create rays binpath= ${BinaryPath} start= auto`