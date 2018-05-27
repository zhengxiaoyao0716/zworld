import * as debug from './debug.js';
import * as websocket from './websocket.js';

debug.local && (window.websocket = websocket);

const html = ({ raw }, ...values) => {
    const div = document.createElement('div');
    const mapper = v => {
        if (v instanceof Array || v instanceof HTMLCollection) {
            return Array.from(v).map(mapper).join('');
        }
        if (v instanceof HTMLElement) {
            return v.outerHTML;
        }
        return v == null ? "" : String(v);
    }
    div.innerHTML = String.raw({ raw }, ...values.map(mapper));
    return div.children.length > 1 ? div.children : div.children[0];
};

const safe = value => String(value)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');

Array.from(document.querySelectorAll('#dashboard>div.card')).forEach(card => {
    card.appendChild(html`<h3>
        <a href="#${card.id}" id="hash">${card.getAttribute('data-card-title')}</a>
        <a href="javascript:;" id="fold">⊿</a>
    </h3>`);
    card.querySelector('#fold').addEventListener('click', () =>
        card.classList.contains('fold')
            ? card.classList.remove('fold')
            : card.classList.add('fold')
    );
    const ul = html`<ul></ul>`;
    card.appendChild(ul);
    websocket.on('/api/dashboard')(promise => {
        promise.then(data =>
            data[card.id].map(([name, usage, value]) => html`
                <li>
                    <span class="name">${name}</span>
                    ${usage && html`<i class="desc">${usage}</i>`}
                    <span class="value">${value}</span>
                </li>
            `)
        ).then(eles => {
            ul.innerHTML = '';
            ul.append(...eles);
        });
    });
});

websocket.send('/api/dashboard');
