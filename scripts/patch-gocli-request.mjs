#!/usr/bin/env node
/**
 * Patches the goctl-generated gocliRequest.ts:
 * 1. Auth headers (Authorization, X-User-Id, X-Sender-Id, X-Tenant-Id)
 * 2. 401 → triggerAuthFailure()
 * 3. GET/DELETE: use req directly as query params, no body
 * 4. POST/PUT/PATCH: goctl convention is webapi.method(url, queryParams, bodyReq)
 *    — when config is present it is the real body; req goes to query string
 */
import { readFileSync, writeFileSync } from 'fs';
import { resolve } from 'path';

const filePath = resolve('./ui/web/src/client/gocliRequest.ts');
let content = readFileSync(filePath, 'utf8');

if (content.includes('getAuthHeaders')) {
  console.log('gocliRequest.ts already patched, skipping');
  process.exit(0);
}

// 1. Auth import
content = `import { getAuthHeaders, triggerAuthFailure } from './auth';\n` + content;

// 2. Auth headers in fetch
content = content.replace(
  /('Content-Type':\s*'application\/json')/,
  `$1,\n            ...getAuthHeaders()`
);

// 3. 401 handling
content = content.replace(
  /return response\.json\(\);/,
  `if (response.status === 401) { triggerAuthFailure(); }\n    return response.json();`
);

// 4. Replace entire api() body with corrected logic
//    Match from the if-block through the end of the switch
content = content.replace(
  /if \(url\.match\(\/:\/\) \|\| method\.match\(\/get\|delete\/i\)\) \{[\s\S]*?url = genUrl\(url, req(\??\.params \|\| req\??\.forms)?\);\s*\}\s*method = method\.toLocaleLowerCase\(\) as Method;\s*switch \(method\) \{[\s\S]*?default:\s*return request\(\{method: 'post', url, data: req, config\}\);\s*\}/,
  `method = method.toLocaleLowerCase() as Method;

    if (method === 'get' || method === 'delete') {
        url = genUrl(url, req);
        return request({method, url, data: undefined});
    }

    // goctl generates webapi.post(url, queryParams, bodyReq)
    // req = queryParams (often {}), config = bodyReq (actual body)
    const body = config !== undefined ? config : req;
    const queryParams = config !== undefined ? req : undefined;
    if (queryParams && typeof queryParams === 'object' && Object.keys(queryParams as object).length > 0) {
        url = genUrl(url, queryParams);
    }

    switch (method) {
        case 'put':
            return request({method: 'put', url, data: body});
        case 'post':
            return request({method: 'post', url, data: body});
        case 'patch':
            return request({method: 'patch', url, data: body});
        default:
            return request({method: 'post', url, data: body});
    }`
);

writeFileSync(filePath, content);
console.log('gocliRequest.ts patched');
