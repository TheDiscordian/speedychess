function changeCSS(cssFile) {
	document.getElementById("pagestyle").setAttribute("href", cssFile);
}
function lightmode() {
	document.body.style.background = "#D3D3D3";
	document.body.style.color = "black";
	document.getElementById("connectionbox").style.backgroundColor = "white";
	document.getElementById("maindiv").style.background = "white";
	document.getElementById("pawnpromo").style.backgroundColor = "white";
}
function darkmode() {
	document.body.style.background = "#404040";
	document.body.style.color = "white";
	document.getElementById("connectionbox").style.backgroundColor = "#101010";
	document.getElementById("maindiv").style.background = "#101010";
	document.getElementById("pawnpromo").style.backgroundColor = "#101010";
}