export const local = ({
    "": true, "localhost": true, "127.0.0.1": true, '[::1]': true,
})[document.location.hostname];

const debug = {};
export default debug;

local && (window.debug = debug);

export const exports = local ? async (name, factory) => {
    const mod = { exports: {} };
    Object.defineProperty(mod.exports, '__esModule', { value: true });
    factory(mod.exports)
    Object.defineProperty(debug, name, { get: () => mod.exports });
    return mod.exports;
} : async () => { };
