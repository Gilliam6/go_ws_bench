import  ws  from "k6/ws";
import { check, sleep } from "k6";
import {getGigaHost} from "./config.js";
import {debug, log} from "./log.js";
const closeTimeout = Number(__ENV.CLOSE_TIMEOUT);

export default function () {
    const ws_url = `${getGigaHost()}/stream`;
    const params = {
        headers: {
            'Origin': 'http://localhost'
        }
    };

    log(ws_url);
    const res = ws.connect(ws_url, params, socketHandler);
    check(res, { "Play Socket: status is 101": r => r && r.status === 101 });
}


function socketHandler(socket) {
    socket.on("open", () => onOpen());
    socket.on("message", (message) => onMessage(message));
    socket.on("close", code => onClose(code));
    socket.on("error", error => onError(error));
    socket.setTimeout(() => {
        socket.close();
    }, closeTimeout);
}

function onError(error) {
    log("web socket error: ", error);
}

function onOpen() {
    log("client connected");
}

function onClose(code) {
    log("client disconnected: ", code);
}

function onMessage(message) {
}