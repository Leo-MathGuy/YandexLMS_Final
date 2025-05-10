function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(";").shift();
}

function handleSubmit(event) {
    event.preventDefault();
    document.getElementById("error").innerHTML = "";
    document.getElementById("result").innerHTML = "";

    const form = new FormData(event.target);
    const data = Object.fromEntries(form.entries());
    data["token"] = getCookie("token");

    var xhr = new XMLHttpRequest();
    var url = "/api/v1/calculate";
    xhr.open("POST", url, true);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4) {
            if (xhr.status === 200) {
                document.getElementById("result").innerHTML =
                    "Recieved, ID: " + JSON.parse(xhr.responseText)["id"];
            } else {
                document.getElementById("error").innerHTML = xhr.responseText;
                console.log(xhr.status);
                console.log(xhr.readyState);
            }
        }
        console.log(xhr.status);
        console.log(xhr.readyState);
    };
    xhr.send(JSON.stringify(data));
}

document.getElementById("form1").addEventListener("submit", handleSubmit);
document.getElementById("submit").disabled = false;
