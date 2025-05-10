function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(";").shift();
}

function handleSubmit1(event) {
    event.preventDefault();
    document.getElementById("error1").innerHTML = "";
    document.getElementById("result1").innerHTML = "";

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
                document.getElementById("result1").innerHTML =
                    "Recieved, ID: " + JSON.parse(xhr.responseText)["id"];
            } else {
                document.getElementById("error1").innerHTML = xhr.responseText;
                console.log(xhr.status);
                console.log(xhr.readyState);
            }
        }
        console.log(xhr.status);
        console.log(xhr.readyState);
    };
    xhr.send(JSON.stringify(data));
}

function handleSubmit2(event) {
    event.preventDefault();
    document.getElementById("error2").innerHTML = "";
    document.getElementById("result2").innerHTML = "";

    const form = new FormData(event.target);
    formdata = Object.fromEntries(form.entries());
    data = { token: getCookie("token") };

    var xhr = new XMLHttpRequest();
    var url = "/api/v1/expressions/" + formdata["id"];
    xhr.open("POST", url, true);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4) {
            if (xhr.status === 200) {
                document.getElementById("result2").innerHTML = JSON.stringify(
                    xhr.responseText
                );
            } else {
                document.getElementById("error2").innerHTML = xhr.responseText;
                console.log(xhr.status);
                console.log(xhr.readyState);
            }
        }
        console.log(xhr.status);
        console.log(xhr.readyState);
    };
    xhr.send(JSON.stringify(data));
}

document.getElementById("form1").addEventListener("submit", handleSubmit1);
document.getElementById("submit1").disabled = false;

document.getElementById("form2").addEventListener("submit", handleSubmit2);
document.getElementById("submit2").disabled = false;
