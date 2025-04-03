const rayutil = {
    "switchChannel" : () => {
        document.cookie = "ray-channel=" + prompt("What channel would you like to switch to?")
        document.cookie = "ray-enrolled-at=" + Date.now()
        document.location.reload(true)
    },
}

addEventListener("load", (event) => {
    if ("$${Message}" != "") {
        if ("$${MsgType}" == "login") {
            if (sessionStorage.getItem('loginShown')) {
                return
            }
            sessionStorage.setItem('loginShown', true)
        }

        const html = '<div id="__ray-notification" style="z-index: 50000; position: fixed; height: max-content; width: max-content; padding: 0.5rem; top: 0.5rem; left: 0.5rem; background-color: black; display: flex; flex-direction: row; color: white; border-radius: 0.25rem; align-items: center; transition: opacity 0.5s; opacity: 1;"><svg style="margin-right: 0.25rem;" xmlns="http://www.w3.org/2000/svg" width="20" height="20" $${Icon}></svg><span style="font-family: monospace; font-size: 1rem;">$${Message}</span></div>'
        document.body.insertAdjacentHTML("beforeend", html)
        
        setTimeout(() => { 
            document.getElementById("__ray-notification").style.opacity = 0
            setTimeout(() => {
                document.getElementById("__ray-notification").remove()
            }, 500)
        }, 1500)
    }
});