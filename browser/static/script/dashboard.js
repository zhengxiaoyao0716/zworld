{
    const conn = new WebSocket(`${document.location.protocol.replace('http', 'ws')}//${document.location.host}/ws`);
    conn.addEventListener('open', e => {
        console.info(`connected with '${e.target.url}'`);
    });
    conn.addEventListener('close', e => {
        console.warn(`disconnected with '${e.target.url}'`);
    });
    conn.addEventListener('message', e => {
        const data = JSON.parse(e.data);
        const { ok, code, reason } = data;
        if (!ok) {
            console.error(new Error(reason ? `code: ${code}, reason: ${reason}.` : data));
            return;
        }
        console.info(data);
    });
    function send(action, data) {
        conn.send(JSON.stringify({ action, ...data, '_messageId': Math.random() }));
    }
    window.send = send;
}