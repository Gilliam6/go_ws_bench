export function getGigaStreamId() {
    const streamId = __ENV.GIGASPACE_STREAMID;
    if (streamId === undefined) {
        return "test8001";
    }
    return streamId
}

export function getCloseTimeout() {
    const timeout = __ENV.CLOSE_TIMEOUT;
    if (timeout === undefined) {
        return 60000;
    }
    return Number(timeout);
}
export function getGigaScenario() {
    const scenario = __ENV.GIGASPACE_SCENARIO;
    if (scenario === undefined) {
        return "test";
    }
    return scenario
}

export function getGigaHost() {
    const host = __ENV.GIGASPACE_HOST;
    if (host === undefined) {
        return "127.0.0.1";
    }
    return host;
}

export function getGigaPort() {
    const port = __ENV.GIGASPACE_PORT;
    if (port === undefined) {
        return "80";
    }
    return port;
}