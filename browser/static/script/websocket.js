const listeners = {};

export const ready = new Promise(resolve => {
    const conn = new WebSocket(`${document.location.protocol.replace('http', 'ws')}//${document.location.host}/ws`);
    conn.addEventListener('open', e => {
        console.info(`connected with '${e.target.url}'`);
        resolve(conn);
    });
    conn.addEventListener('close', e => console.warn(`disconnected with '${e.target.url}'`));
    conn.addEventListener('message', e => {
        const data = JSON.parse(e.data);
        const { action, _messageId: id, ok, code, reason, ...body } = data;
        if (!action) {
            console.error(new Error(`missing 'action' argument in ${data}`));
            return;
        }
        const handlers = listeners[action] || [];
        const msgHandler = handlers[id];
        delete (handlers[id]);
        if (!msgHandler && !handlers.length) { // 无监听者，采用默认处理方案
            const tag = `[websocket] action: ${action}, _messageId: ${id}.`;
            ok ? console.info(tag, body) :
                console.error(tag, new Error(reason ? `code: ${code}, reason: ${reason}.` : body));
            return;
        }
        const promise = ok ? Promise.resolve(body) : Promise.reject({ code, reason, ...body });
        [msgHandler, ...handlers].forEach(async handler => handler && handler(promise));
    })
});

export const on = (action, messageId) => {
    listeners[action] || (listeners[action] = []);
    return handler => {
        const index = messageId ? (listeners[action][messageId] = handler, messageId) :
            listeners[action].push(handler) - 1;
        const off = () => delete listeners[action][index];
        return off;
    }
}

export const send = (action, data) => new Promise(resolve => {
    const id = Math.random();
    on(action, id)(resolve);
    ready.then(conn => conn.send(JSON.stringify({ action, _messageId: id, ...data })));
});