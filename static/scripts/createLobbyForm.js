/** @param {HTMLFormElement} form */
function createLobby(form) {
    const params = new URLSearchParams();

    params.append('mode', form.mode.value);
    params.append('timed', form.timed.checked); 
    params.append('time', form.time.value);
    params.append('width', form.width.value);
    params.append('height', form.height.value);
    params.append('bombs', form.bombs.value);

    fetch('/lobby/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: params.toString() 
    }).then(resp=>{
        if (!resp.ok) {
            alert("Post failed! Try again.");
            return;
        }
        resp.text().then(data=>{
            switch (form.mode.value) {
                case "1v1":
                    window.location.href = `game1v1.html?lobby=${data}`;
                    break;
                case "singleplayer":
                    window.location.href = `gameSingle.html?lobby=${data}`;
                    break;
            }
        }).catch(e=>console.log(e));
    });

    return false;
}
