<?php
?>
<!doctype html>
<html lang="en" class="h-100">
    <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <title>Adam</title>
        <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-QWTKZyjpPEjISv5WaRU9OFeRpok6YctnYmDr5pNlyT2bRjXh0JMhjY6hW+ALEwIH" crossorigin="anonymous">
        <link href="css/extra.css" rel="stylesheet">
        <meta name="theme-color" content="#7952b3">
    </head>
    <body class="d-flex h-100 text-center text-white bg-dark">
        <div class="cover-container d-flex w-100 h-100 p-3 mx-auto flex-column">
            <header class="mb-auto">
                <div>
                    <h3 class="float-md-start mb-0">Adam's Website</h3>
                    <nav class="nav nav-masthead justify-content-center float-md-end">
                        <a class="nav-link" href="https://github.com/lemonman44/adam-site">Repository</a>
                        <a class="nav-link" href="resume/Resume.pdf">Download Resume</a>
                        <a class="nav-link" href="mailto:adamlehman2018@gmail.com">Email Me</a>
                    </nav>
                </div>
            </header>
            <main class="px-3" style="margin-top: 100px;">
                <h1>Welcome to my website.</h1>
                <p class="lead">
                    This website is a simple PHP frontend and a Golang backend, each Dockerized and deployed through GitHub Actions CI/CD on to an Amazon Elastic Kubernetes Service.
                </p>
                <p class="lead">
                    Below is a small, (hopefully) fun chat box where anything you type will be displayed to every other active connection to this site.
                </p>
                <p class="lead" style="display: grid; height: 300px;">
                    <span style="font-size: 1rem;align-self: end;text-align: right;">Active Connections: <span id="numconns">1</span></span>
                    <textarea name="chat" id="chat" disabled></textarea>
                    <label id="chatlabel" for="chat" style="font-size: 0.25rem;">Connecting to GloboChat™©®ª°...</label>
                </p>
            </main>
            <footer class="mt-auto text-white-50">
                <!-- <p style="font-size: 0.5rem;">Website by Adam Lehman, with a simple template thanks to <a href="https://getbootstrap.com/docs/5.0/examples/">Bootstrap</a>, also you can follow me <a href="https://adamclehman.bsky.social">here</a>.</p> -->
                <p style="font-size: 0.5rem;">Website by Adam Lehman, with a simple template thanks to <a href="https://getbootstrap.com/docs/5.0/examples/">Bootstrap</a>.</p>
            </footer>
        </div>
        <script src="https://code.jquery.com/jquery-3.2.1.slim.min.js" integrity="sha384-KJ3o2DKtIkvYIK3UENzmM7KCkRr/rE9/Qpg6aAZGJwFDMVNA/GpGFF93hXpG5KkN" crossorigin="anonymous"></script>
        <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js" integrity="sha384-YvpcrYf0tY3lHB60NNkmXc5s9fDVZLESaAA55NDzOxhy9GkcIdslK1eN7N6jIeHz" crossorigin="anonymous"></script>
        <script type="text/javascript">
            $(document).ready(function() {
                initChat();
            });

            function initChat() {
                let label = $("#chatlabel");
                let conns = $("#numconns");
                let chat = $("#chat");
                label.text("Connecting to GloboChat™©®ª°...");
                chat.prop("disabled", true);
                console.log("Attempting Websocket Connection");
                const chatsocket = new WebSocket("<?php echo "$_ENV[GO_SITE_HTTP_HOST]/api/go/sockets/chat"; ?>");
                chatsocket.onopen = (event) => {
                    console.log("Opened Websocket Connection");
                    label.text("Connected to GloboChat™©®ª°");
                    chat.prop("disabled", false);
                    chat.get(0).addEventListener("input", (event) => {
                        let key = event.data;
                        if(event.inputType == "insertLineBreak") key = "\r\n";
                        // if(event.inputType == "deleteContentBackward") key = "\b \b"; // doesnt work but im sure theres a correct backspace character to use in this situation
                        if(key) chatsocket.send(key);
                    });
                    chatsocket.onmessage = (event) => {
                        console.log("data received", event);
                        json = JSON.parse(event.data);
                        conns.text(json.numconns);
                        chat.val(chat.val() + json.message);
                        chat.scrollTop(chat[0].scrollHeight)
                    };
                    chatsocket.onclose = (event) => {
                        console.log("Connection Closed, will attempt to reconnect");
                        initChat();
                    };
                };
                chatsocket.onerror = (event) => {
                    console.log("error", event);
                }
            }
        </script>
    </body>
</html>