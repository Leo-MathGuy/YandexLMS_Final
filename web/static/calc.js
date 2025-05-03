function handleSubmit(event) {
    event.preventDefault();
    document.getElementById("error").innerHTML = "";

    const form = new FormData(event.target);
    const data = Object.fromEntries(form.entries());

    var xhr = new XMLHttpRequest();
    var url = event.target.action;
    xhr.open("POST", url, true);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.withCredentials = true;
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4) {
            if (xhr.status === 200) {
                if (url.split("/")[5] === "register") {
                    window.location.href = "/login";
                } else {
                    document.cookie =
                        "token=" + xhr.responseText + "; maxAge=1800";
                    window.location.href = "/calc";
                }
            } else {
                document.getElementById("error").innerHTML = xhr.responseText;
                console.log(xhr.status);
                console.log(xhr.readyState);
            }
        }
    };
    xhr.send(JSON.stringify(data));
}

document.getElementById("form").addEventListener("submit", handleSubmit);
document.getElementById("submit").disabled = false;
