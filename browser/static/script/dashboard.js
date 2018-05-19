import * as debug from './debug.js';
import * as websocket from './websocket.js';

debug.local && (window.websocket = websocket);

const dashboard = document.querySelector('#dashboard');
const baseArgs = dashboard.querySelector('#baseArgs');

const html = ({ raw }, ...values) => {
    const div = document.createElement('div');
    const safe = v => String(v)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
    div.innerHTML = String.raw({ raw }, ...values.map(safe));
    return div.children;
};

websocket.on('/api/dashboard')(promise => {
    promise.then(({ baseArgs }) => baseArgs.map(({ name, usage, value }) => html`
        <li>
            <span class="name">
                <span>${name}</span>
                <i class="fas fa-info-circle" title="${usage}"></i>
            </span>
            <span class="value">${value || '<none>'}</span>
        </li>
    `[0]
    )).then(eles => {
        baseArgs.innerHTML = '';
        baseArgs.append(...eles);
    });
});
websocket.send('/api/dashboard');
