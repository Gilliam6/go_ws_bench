import { RefinedResponse, ResponseType } from "k6/http";
// import exec from 'k6/execution'

export const vuserLog = (msg) => console.log(`[Vuser: ${__VU}] ${msg}`);
export const vuserDebug = (msg) => console.debug(`[Vuser: ${__VU}] ${msg}`);
export const vuserError = (msg) => console.error(`[Vuser: ${__VU}] ${msg}`);

export const logFailedCheck = (url, response, stepName) =>
    console.log(
        stepName,
        "!!FAILED!!",
        ` | `,
        `RESPONSE CODE: ${JSON.stringify(response.status)}`,
        ` | `,
        `REQUEST URL: ${url}`,
        ` \n `,
        `Request headers: ${JSON.stringify(response.request.headers)}`,
        ` \n `,
        `REQUEST BODY: ${JSON.stringify(response.request.body)}`,
        ` \n `,
        // `RESPONSE HEADERS: ${JSON.stringify(response.headers)}`,
        `RESPONSE BODY: ${JSON.stringify(response.body)}`,
        `\n---------------------------------------------------------------------------------------------------------------------------------------`
    );

export const log = (info, debug = "") => {
    if (__ENV.LOG_LEVEL === "DEBUG") {
        vuserDebug(info + debug);
    } else if (__ENV.LOG_LEVEL === "INFO" || __VU === 1) {
        vuserLog(info);
    }
};

export const debug = (debug) => {
    if (__ENV.LOG_LEVEL === "DEBUG") {
        vuserDebug(debug);
    }
};

export const error = (error) => {
    if (__ENV.LOG_LEVEL === "ERROR"|| __ENV.LOG_LEVEL === "DEBUG" || __ENV.LOG_LEVEL === "INFO") {
        vuserError(error);
    }
};

