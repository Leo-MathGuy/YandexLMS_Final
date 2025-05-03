var loggedin = document.getElementById("login");
var logout = document.getElementById("register");

function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(";").shift();
}

function decodeJwt(token) {
    try {
        const [header, payload, signature] = token.split(".");
        const decodedPayload = atob(payload);
        const claims = JSON.parse(decodedPayload);
        return claims["user"];
    } catch (error) {
        return null;
    }
}

function deleteAllCookies() {
    document.cookie.split(";").forEach((cookie) => {
        const eqPos = cookie.indexOf("=");
        const name = eqPos > -1 ? cookie.substring(0, eqPos) : cookie;
        document.cookie = name + "=;expires=Thu, 01 Jan 1970 00:00:00 GMT";
    });
}

var cookie = getCookie("token");
var decoded = decodeJwt(cookie);
if (decoded !== null) {
    loggedin.href = "#";
    loggedin.innerHTML = "Logged in as: " + decoded;

    logout.href = "";
    logout.innerHTML = "Log out";
    logout.onclick = function (event) {
        deleteAllCookies();
        window.location.href = "/calc";
    };
}
