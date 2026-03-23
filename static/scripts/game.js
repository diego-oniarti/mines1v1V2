/** @type {WebSocket} */
let ws;

let C;
let state;
let Pid;
let timeout;

let timed, time;

const board = [];
const edge = 30;

let bg;

let colors = {
    backgroundA: undefined,
    backgroundB: undefined,
    backgroundC: undefined,
    p1:          undefined,
    p2:          undefined,
    system:      undefined,
};

const states = {
    waitingPlayer: 0,
    countdown:     1,
    yourTurn:      2,
    theirTurn:     3,
    won:           4,
    lost:          5,
    tied:          6,
};

function setup() {
    C = createCanvas(400, 400);
    C.parent("game");

    state = states.waitingPlayer;

    colors.backgroundA = color(220, 220, 220);
    colors.backgroundB = color(200, 200, 200);
    colors.backgroundC = color(130, 130, 130);
    colors.p1          = color(40,  180, 180);
    colors.p2          = color(180, 40,  180);
    colors.system      = color(20,  20,  20);

    let params = new URLSearchParams(document.location.search);
    let lobby = params.get("lobby");

    document.getElementById("gameKey").innerText = lobby;

    const wsUri = `ws://${location.host}/lobby/join?lobby=${lobby}`;
    ws = new WebSocket(wsUri);
    ws.binaryType = "arraybuffer";

    ws.addEventListener("open", () => {
        pingInterval = setInterval(() => {
            ws.send("ping");
        }, 10000);
    });

    ws.addEventListener("message", (event) => {
        if (event.data instanceof ArrayBuffer) {
            console.log("Message received")
            const dataView = new DataView(event.data);
            handleMsg(dataView);
        } else {
            console.log("Received text data:", event.data);
        }
    });
}

let spinnerStart;
function drawTimer(c, T=1000) {
    const dim = min(width, height) * 0.8;

    c = color(
        c.levels[0],
        c.levels[1],
        c.levels[2],
        80
    );

    noFill();
    strokeWeight(15);
    stroke(c);
    circle(width/2,height/2, dim);

    noStroke();
    fill(c);

    const t = (Date.now()-spinnerStart);
    const A = map( t, 0, T, 0, TWO_PI );

    if (parseInt(t/T)%2==0)
        arc(width/2,height/2, dim-15,dim-15, 0-PI/2, A-PI/2);
    else
        arc(width/2,height/2, dim-15,dim-15, A-PI/2, 0-PI/2);
}

function drawBoard() {
    image(bg, 0, 0);
    textSize(edge*0.8);
    noStroke();
    for (let y=0; y<board.length; y++) {
        for (let x=0; x<board[y].length; x++) {
            const cell = board[y][x];
            if (!cell) continue;

            const col = (cell.p==0) ? colors.system : (cell.p==Pid ? colors.p1 : colors.p2);
            const xc = x*edge+edge/2;
            const yc = y*edge+edge/2;

            if (cell.flag) {
                fill(col);
                text("🏳", xc, yc);
                continue;
            }

            fill(colors.backgroundC);
            rect(x*edge,y*edge,edge,edge);
            fill(col);
            if (cell.n > 0) {
                text(cell.n, xc, yc);
            }
            if (cell.n < 0) {
                text("💣", xc, yc);
            }
        }
    }
}

function draw() {
    switch (state) {
        case states.waitingPlayer:
            background(colors.backgroundA);
            noStroke();
            fill(colors.system);
            text("Waiting for player", width/2, height/2);
            textSize(20);
            textAlign(CENTER, CENTER);
            break;
        case states.countdown:
            image(bg, 0, 0)
            drawTimer(color(250, 250, 250, 100));
            break;
        case states.yourTurn:
            drawBoard();
            if (timed) { drawTimer(colors.p1, time) }
            break;
        case states.theirTurn:
            drawBoard();
            if (timed) { drawTimer(colors.p2, time) }
            break;
        case states.won:
            drawBoard();
            noStroke();
            fill(colors.p1);
            text("You Won!", width/2, height/2);
            break;
        case states.lost:
            drawBoard();
            noStroke();
            fill(colors.p2);
            text("You Lost!", width/2, height/2);
            break;
        case states.tied:
            drawBoard();
            noStroke();
            fill(colors.system);
            text("Draw!", width/2, height/2);
            break;
    } 
}

// ----------- //

function handleMsg(view) {
    switch (state) {
        case states.waitingPlayer:
            handleStart(view);
            break;
        case states.countdown:
            handleFirstMove(view);
            break;
        default:
            handleGame(view);
    }
}

/** @param {DataView} view */
function handleStart(view) {
    const W = view.getUint16(0);
    const H = view.getUint16(2);
    time = view.getUint32(4);
    const P = view.getInt8(8);
    timed = view.getUint8(9)!=0;

    Pid = P;
    C.resize(W*edge, H*edge);
    bg = createGraphics(W*edge, H*edge);

    bg.noStroke();
    for (let y=0; y<H; y++) {
        board.push([]);
        for (let x=0; x<W; x++) {
            board[y].push(undefined);
            if ((x+y)%2==0) {
                bg.fill(colors.backgroundA);
            }else{
                bg.fill(colors.backgroundB);
            }
            bg.rect(x*edge, y*edge, edge+1, edge+1);
        }
    }

    spinnerStart = Date.now();
    state = states.countdown;
}

/** @param {DataView} view */
function handleFirstMove(view) {
    console.log(view)
    let i=0
    while (i<view.byteLength) {
        let x = view.getUint16(i)
        let y = view.getUint16(i+2)
        let n = view.getUint8(i+4)
        board[y][x] = {
            p:    0,
            flag: false,
            n:    n,
        };
        i+=5
    }

    if (Pid==1) {
        state = states.yourTurn
    }else{
        state = states.theirTurn
    }
    spinnerStart = Date.now();
}

/** @param {DataView} view */
function handleGame(view) {
    console.log(view);
    const header = view.getUint8(0)

    const player = header & 0b00111111;
    const aaa = header >> 6;

    switch (aaa) {
        case 0b10: // game over
            const x = view.getUint16(1);
            const y = view.getUint16(3);
            board[y][x] = {
                flag: false,
                n:    -1,
                p:    player,
            }
        case 0b11: // game over and timeout
            if (player==Pid) {
                state = states.lost
            }else{
                state = states.won
            }
            break;
        case 0b01: // tie
            state = states.tied;
    }

    let i = 1;
    let flag = false;
    while (i<view.byteLength) {
        const x = view.getUint16(i);
        const y = view.getUint16(i+2);
        const m = view.getUint8(i+4);

        const fBit = (m&16)>0
        flag |= fBit;

        if (fBit && board[y][x]) {
            board[y][x] = null;
        }else{
            board[y][x] = {
                flag: fBit,
                n:    m&0b1111,
                p:    player,
            };
        }

        i+=5;
    }

    if (!flag) {
        switch (state) {
            case states.yourTurn:
                state = states.theirTurn
                break;
            case states.theirTurn:
                state = states.yourTurn
                break;
        }
        spinnerStart = Date.now()
    }

}

// ----------- //

function keyPressed() {

}

function mousePressed() {
    const X = mouseX;
    const Y = mouseY;
    if (X<0 || X>width || Y<0 || Y>height) return true;
    if (mouseButton != LEFT && mouseButton != RIGHT) return true;

    const bf = new ArrayBuffer(5);
    const dw = new DataView(bf);

    dw.setUint16(0, parseInt(X/edge));
    dw.setUint16(2, parseInt(Y/edge));
    dw.setUint8(4, mouseButton==LEFT ? 0 : 1);

    ws.send(bf);

    return false;
}

document.oncontextmenu = function() {
    const X = mouseX;
    const Y = mouseY;
    if (X<0 || X>width || Y<0 || Y>height) return true;
    return false;
}
// ----------- //
